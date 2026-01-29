package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestGroupInvitationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	h := newHarness(t)

	// 1. Create two users
	owner := createUser(t, h, "owner-"+randomString()+"@example.com")
	inviteeEmail := "invitee-" + randomString() + "@example.com"
	invitee := createUser(t, h, inviteeEmail) // User C exists in DB already

	// 2. Create Group
	groupID := createGroup(t, h, owner.ID, "Invitation Test Group", "USD")

	// 3. Invite User (by email of invitee)
	// POST /groups/{group_id}/invitations
	token := createInvitation(t, h, owner.ID, groupID, inviteeEmail)

	// 4. Create Expense assigned to Pending User
	// We need to look up pending user ID.
	// We can do this by using a helper that inspects DB, or just listing invitations?
	// The pending user is created when invitation is created.
	// But to create expense via API we need pending_user_id.
	// Let's rely on a helper that queries the DB directly via h.pool.
	pendingUser := getPendingUserByEmail(t, h, inviteeEmail)

	// Create expense: Owner pays 100, Owner owes 50, PendingUser owes 50
	createExpenseWithPendingUser(t, h, owner.ID, groupID, pendingUser.ID, "Pending User Expense", "100.00")

	// 5. Accept Invitation
	// POST /invitations/{token}/accept
	acceptInvitation(t, h, invitee.ID, token)

	// 6. Verify Membership
	// List group members and check status
	members := listGroupMembers(t, h, owner.ID, groupID)
	found := false
	for _, m := range members {
		if m.UserID == invitee.ID {
			if m.Status != "active" {
				t.Errorf("expected active status, got %s", m.Status)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("invitee not found in group members")
	}

	// 7. Verify Expense Claiming
	// List expenses for group
	expenses := listExpenses(t, h, owner.ID, groupID)
	if len(expenses) == 0 {
		t.Fatalf("expected expenses, got 0")
	}
	lastExpenseID := expenses[len(expenses)-1].Expense.ID

	// Get splits for expense
	splits := getExpenseSplits(t, h, owner.ID, groupID, lastExpenseID)
	splitFound := false
	for _, s := range splits {
		if s.UserID == invitee.ID {
			splitFound = true
			break
		}
	}
	if !splitFound {
		t.Errorf("claimed expense split not found for invitee")
	}
}

func TestSmartJoinFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	h := newHarness(t)

	// Setup: Owner and Group
	owner := createUser(t, h, "owner-"+randomString()+"@example.com")
	groupID := createGroup(t, h, owner.ID, "Smart Join Group", "USD")

	t.Run("Scenario: New User Registration via Join", func(t *testing.T) {
		email := "new-user-" + randomString() + "@example.com"
		token := createInvitation(t, h, owner.ID, groupID, email)

		// Call Smart Join without auth header
		payload := map[string]string{
			"password": "newpassword123",
			"name":     "New User Name",
		}
		status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/invitations/%s/join", token), payload, nil)
		if status != http.StatusOK {
			t.Fatalf("smart join new user expected 200, got %d (err=%v)", status, sr.Error)
		}

		var out struct {
			User struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			} `json:"user"`
		}
		if err := json.Unmarshal(sr.Data, &out); err != nil {
			t.Fatalf("unmarshal smart join response: %v", err)
		}

		if out.User.Email != email {
			t.Errorf("expected email %s, got %s", email, out.User.Email)
		}

		// Verify membership
		members := listGroupMembers(t, h, owner.ID, groupID)
		found := false
		for _, m := range members {
			if m.UserID == out.User.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("newly joined user not found in group members")
		}
	})

	t.Run("Scenario: Existing User Login via Join", func(t *testing.T) {
		email := "existing-user-" + randomString() + "@example.com"
		password := "password123"
		user := createUserWithPassword(t, h, email, password)
		token := createInvitation(t, h, owner.ID, groupID, email)

		// Call Smart Join without auth header but with password
		payload := map[string]string{
			"password": password,
		}
		status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/invitations/%s/join", token), payload, nil)
		if status != http.StatusOK {
			t.Fatalf("smart join existing user expected 200, got %d (err=%v)", status, sr.Error)
		}

		var out struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		json.Unmarshal(sr.Data, &out)

		if out.User.ID != user.ID {
			t.Errorf("expected user id %s, got %s", user.ID, out.User.ID)
		}
	})

	t.Run("Scenario: Logged-in User Seamless Join", func(t *testing.T) {
		email := "logged-in-" + randomString() + "@example.com"
		user := createUser(t, h, email)
		token := createInvitation(t, h, owner.ID, groupID, email)

		// Call Smart Join WITH auth header, no password needed
		status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/invitations/%s/join", token), nil, h.authHeader(user.ID))
		if status != http.StatusOK {
			t.Fatalf("smart join logged in expected 200, got %d (err=%v)", status, sr.Error)
		}
	})

	t.Run("Scenario: Multi-Group Merge", func(t *testing.T) {
		email := "merge-user-" + randomString() + "@example.com"

		// 1. Invite to Group A
		groupA := groupID
		tokenA := createInvitation(t, h, owner.ID, groupA, email)
		pendingUser := getPendingUserByEmail(t, h, email)

		// 2. Create another group B
		groupB := createGroup(t, h, owner.ID, "Group B", "USD")
		_ = createInvitation(t, h, owner.ID, groupB, email) // Second invite

		// 3. Add expenses in both groups for pending user
		createExpenseWithPendingUser(t, h, owner.ID, groupA, pendingUser.ID, "Expense A", "100.00")
		createExpenseWithPendingUser(t, h, owner.ID, groupB, pendingUser.ID, "Expense B", "200.00")

		// 4. Bob joins Group A (creates account)
		payload := map[string]string{"password": "password123"}
		status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/invitations/%s/join", tokenA), payload, nil)
		if status != http.StatusOK {
			t.Fatalf("merge join expected 200, got %d (err=%v)", status, sr.Error)
		}

		var out struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		}
		json.Unmarshal(sr.Data, &out)
		bobID := out.User.ID

		// 5. Verify Expense B in Group B is now linked to Bob's real ID
		expensesB := listExpenses(t, h, owner.ID, groupB)
		if len(expensesB) == 0 {
			t.Fatalf("expected expenses in group B")
		}
		lastExpenseBID := expensesB[len(expensesB)-1].Expense.ID

		splits := getExpenseSplits(t, h, owner.ID, groupB, lastExpenseBID)
		found := false
		for _, s := range splits {
			if s.UserID == bobID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expense in group B was not merged to real user ID")
		}
	})
}

func createUserWithPassword(t *testing.T, h *testHarness, email, password string) createdUser {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, "/users", map[string]any{
		"email":    email,
		"password": password,
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create user expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out createdUser
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal user: %v", err)
	}
	out.Email = email
	return out
}

func createInvitation(t *testing.T, h *testHarness, userID, groupID, email string) string {
	t.Helper()
	payload := map[string]string{
		"email": email,
		"role":  "member",
		"name":  "Invitee Name",
	}
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/invitations", groupID), payload, h.authHeader(userID))
	if status != http.StatusCreated {
		t.Fatalf("create invitation expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal invitation: %v", err)
	}
	return out.Token
}

func acceptInvitation(t *testing.T, h *testHarness, userID, token string) {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/invitations/%s/accept", token), nil, h.authHeader(userID))
	if status != http.StatusOK {
		t.Fatalf("accept invitation expected 200, got %d (err=%v)", status, sr.Error)
	}
}

func getPendingUserByEmail(t *testing.T, h *testHarness, email string) struct{ ID string } {
	t.Helper()
	// Direct DB query
	var id string
	err := h.pool.QueryRow(context.Background(), "SELECT id FROM pending_users WHERE email=$1", email).Scan(&id)
	if err != nil {
		t.Fatalf("failed to get pending user: %v", err)
	}
	return struct{ ID string }{ID: id}
}

func createExpenseWithPendingUser(t *testing.T, h *testHarness, payerID, groupID, pendingUserID, title, amount string) {
	t.Helper()

	// Parse amount to float for simple splitting
	var amt float64
	fmt.Sscanf(amount, "%f", &amt)
	half := fmt.Sprintf("%.2f", amt/2.0)

	payload := map[string]interface{}{
		"title":         title,
		"amount":        amount,
		"currency_code": "USD",
		"date":          "2024-01-01",
		"payments": []map[string]interface{}{
			{"user_id": payerID, "amount": amount},
		},
		"splits": []map[string]interface{}{
			{"user_id": payerID, "amount_owned": half, "split_type": "equal"},
			{"pending_user_id": pendingUserID, "amount_owned": half, "split_type": "equal"},
		},
	}
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/expenses", groupID), payload, h.authHeader(payerID))
	if status != http.StatusCreated {
		t.Fatalf("create expense expected 201, got %d (err=%v)", status, sr.Error)
	}
}

func listGroupMembers(t *testing.T, h *testHarness, userID, groupID string) []struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
} {
	t.Helper()
	status, sr := h.doJSON(http.MethodGet, fmt.Sprintf("/groups/%s/members", groupID), nil, h.authHeader(userID))
	if status != http.StatusOK {
		t.Fatalf("list members expected 200, got %d", status)
	}
	var members []struct {
		UserID string `json:"user_id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(sr.Data, &members); err != nil {
		t.Fatalf("unmarshal members: %v", err)
	}
	return members
}

func listExpenses(t *testing.T, h *testHarness, userID, groupID string) []struct {
	Expense struct {
		ID string `json:"id"`
	} `json:"expense"`
} {
	t.Helper()
	status, sr := h.doJSON(http.MethodGet, fmt.Sprintf("/groups/%s/expenses", groupID), nil, h.authHeader(userID))
	if status != http.StatusOK {
		t.Fatalf("list expenses expected 200, got %d", status)
	}
	var expenses []struct {
		Expense struct {
			ID string `json:"id"`
		} `json:"expense"`
	}
	if err := json.Unmarshal(sr.Data, &expenses); err != nil {
		t.Fatalf("unmarshal expenses: %v", err)
	}
	return expenses
}

func getExpenseSplits(t *testing.T, h *testHarness, userID, groupID, expenseID string) []struct {
	UserID string `json:"user_id"`
	Amount string `json:"amount_owned"`
} {
	t.Helper()
	status, sr := h.doJSON(http.MethodGet, fmt.Sprintf("/groups/%s/expenses/%s", groupID, expenseID), nil, h.authHeader(userID))
	if status != http.StatusOK {
		t.Fatalf("get expense expected 200, got %d (err=%v)", status, sr.Error)
	}
	var resp struct {
		Splits []struct {
			UserID string `json:"user_id"`
			Amount string `json:"amount_owned"`
		} `json:"splits"`
	}
	if err := json.Unmarshal(sr.Data, &resp); err != nil {
		t.Fatalf("unmarshal expense: %v", err)
	}
	return resp.Splits
}

func randomString() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
