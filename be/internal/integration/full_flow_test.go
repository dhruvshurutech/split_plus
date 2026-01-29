package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestFullFlow_HappyPath(t *testing.T) {
	h := newHarness(t)

	// 1) Users
	userA := createUser(t, h, "a@example.com")
	userB := createUser(t, h, "b@example.com")
	userC := createUser(t, h, "c@example.com")

	// 2) Group
	groupID := createGroup(t, h, userA.ID, "Trip", "USD")

	// invite + join B, C
	tokenB := inviteUserToGroup(t, h, userA.ID, groupID, userB.Email)
	tokenC := inviteUserToGroup(t, h, userA.ID, groupID, userC.Email)
	joinGroup(t, h, userB.ID, tokenB)
	joinGroup(t, h, userC.ID, tokenC)

	// 3) Expense (Spotify: A paid, split equally among A,B,C)
	expenseID := createExpense(t, h, userA.ID, groupID, createExpensePayload{
		Title:  "Spotify",
		Notes:  "Monthly subscription",
		Amount: "9.99",
		Date:   "2024-01-03",
		Payments: []paymentPayload{
			{UserID: userA.ID, Amount: "9.99", PaymentMethod: "card"},
		},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "3.33", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "3.33", SplitType: "equal"},
			{UserID: userC.ID, AmountOwned: "3.33", SplitType: "equal"},
		},
	})
	if expenseID == "" {
		t.Fatalf("expected expenseID")
	}

	// 4) Balances endpoints should respond OK (log error body on failure)
	{
		status, sr := h.doJSON(http.MethodGet, fmt.Sprintf("/groups/%s/balances", groupID), nil, h.authHeader(userA.ID))
		if status != http.StatusOK {
			t.Fatalf("expected balances 200, got %d, err=%+v", status, sr.Error)
		}
		status, sr = h.doJSON(http.MethodGet, fmt.Sprintf("/groups/%s/debts", groupID), nil, h.authHeader(userA.ID))
		if status != http.StatusOK {
			t.Fatalf("expected debts 200, got %d, err=%+v", status, sr.Error)
		}
	}

	// 5) Settlement (B pays A)
	settlementID := createSettlement(t, h, userA.ID, groupID, userB.ID, userA.ID, "3.33", "USD")
	updateSettlementStatus(t, h, userA.ID, groupID, settlementID, "completed")

	// 6) Recurring expense (daily coffee)
	recurringID := createRecurringExpense(t, h, userA.ID, groupID, createRecurringPayload{
		Title:          "Coffee",
		Notes:          "Daily coffee",
		Amount:         "5.00",
		CurrencyCode:   "USD",
		RepeatInterval: "daily",
		StartDate:      "2024-01-04",
		Payments: []paymentPayload{
			{UserID: userA.ID, Amount: "5.00", PaymentMethod: "cash"},
		},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "2.50", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "2.50", SplitType: "equal"},
		},
	})

	// manual generate should create an expense
	status, _ := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/recurring-expenses/%s/generate", groupID, recurringID), nil, h.authHeader(userA.ID))
	if status != http.StatusCreated {
		t.Fatalf("expected recurring generate 201, got %d", status)
	}
}

// TestFriendFlow_HappyPath exercises the direct friend expenses & settlements flow (no groups).
func TestFriendFlow_HappyPath(t *testing.T) {
	h := newHarness(t)

	// 1) Users
	userA := createUser(t, h, "friend-a@example.com")
	userB := createUser(t, h, "friend-b@example.com")

	// 2) Friend request + accept
	requestID := sendFriendRequest(t, h, userA.ID, userB.ID)
	acceptFriendRequest(t, h, userB.ID, requestID)

	// 3) Friend expense: A pays 10, split equally A/B
	friendExpensePayload := createExpensePayload{
		Title:  "Dinner",
		Notes:  "Friend dinner",
		Amount: "10.00",
		Date:   "2024-01-05",
		Payments: []paymentPayload{
			{UserID: userA.ID, Amount: "10.00", PaymentMethod: "cash"},
		},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "5.00", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "5.00", SplitType: "equal"},
		},
	}

	status, sr := h.doJSON(
		http.MethodPost,
		fmt.Sprintf("/friends/%s/expenses", userB.ID),
		friendExpensePayload,
		h.authHeader(userA.ID),
	)
	if status != http.StatusCreated {
		t.Fatalf("create friend expense expected 201, got %d (err=%v)", status, sr.Error)
	}

	// 4) List friend expenses for both sides should be 200
	if status, sr = h.doJSON(
		http.MethodGet,
		fmt.Sprintf("/friends/%s/expenses", userB.ID),
		nil,
		h.authHeader(userA.ID),
	); status != http.StatusOK {
		t.Fatalf("list friend expenses (A->B) expected 200, got %d (err=%v)", status, sr.Error)
	}
	if status, sr = h.doJSON(
		http.MethodGet,
		fmt.Sprintf("/friends/%s/expenses", userA.ID),
		nil,
		h.authHeader(userB.ID),
	); status != http.StatusOK {
		t.Fatalf("list friend expenses (B->A) expected 200, got %d (err=%v)", status, sr.Error)
	}

	// 5) Friend settlement: B pays A back 5
	friendSettlementPayload := map[string]any{
		"payer_id":      userB.ID,
		"payee_id":      userA.ID,
		"amount":        "5.00",
		"currency_code": "USD",
		"status":        "pending",
	}

	if status, sr = h.doJSON(
		http.MethodPost,
		fmt.Sprintf("/friends/%s/settlements", userB.ID),
		friendSettlementPayload,
		h.authHeader(userA.ID),
	); status != http.StatusCreated {
		t.Fatalf("create friend settlement expected 201, got %d (err=%v)", status, sr.Error)
	}

	// 6) List friend settlements should be 200 for both sides
	if status, sr = h.doJSON(
		http.MethodGet,
		fmt.Sprintf("/friends/%s/settlements", userB.ID),
		nil,
		h.authHeader(userA.ID),
	); status != http.StatusOK {
		t.Fatalf("list friend settlements (A->B) expected 200, got %d (err=%v)", status, sr.Error)
	}
	if status, sr = h.doJSON(
		http.MethodGet,
		fmt.Sprintf("/friends/%s/settlements", userA.ID),
		nil,
		h.authHeader(userB.ID),
	); status != http.StatusOK {
		t.Fatalf("list friend settlements (B->A) expected 200, got %d (err=%v)", status, sr.Error)
	}
}

type createdUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func createUser(t *testing.T, h *testHarness, email string) createdUser {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, "/users", map[string]any{
		"email":    email,
		"password": "password123",
	}, nil)
	if status != http.StatusCreated {
		t.Fatalf("create user expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out createdUser
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal user: %v", err)
	}
	if out.ID == "" {
		t.Fatalf("expected user id")
	}
	out.Email = email // Manually set since API might not return it or we just know it
	return out
}

func sendFriendRequest(t *testing.T, h *testHarness, requesterID, friendID string) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, "/friends/requests", map[string]any{
		"friend_id": friendID,
	}, h.authHeader(requesterID))
	if status != http.StatusCreated {
		t.Fatalf("send friend request expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal friend request: %v", err)
	}
	if out.ID == "" {
		t.Fatalf("expected friend request id")
	}
	return out.ID
}

func acceptFriendRequest(t *testing.T, h *testHarness, requesterID, requestID string) {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/friends/requests/%s/accept", requestID), nil, h.authHeader(requesterID))
	if status != http.StatusOK {
		t.Fatalf("accept friend request expected 200, got %d (err=%v)", status, sr.Error)
	}
}

func createGroup(t *testing.T, h *testHarness, userID, name, currency string) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, "/groups", map[string]any{
		"name":          name,
		"description":   "test group",
		"currency_code": currency,
	}, h.authHeader(userID))
	if status != http.StatusCreated {
		t.Fatalf("create group expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal group: %v", err)
	}
	if out.ID == "" {
		t.Fatalf("expected group id")
	}
	return out.ID
}

func inviteUserToGroup(t *testing.T, h *testHarness, inviterID, groupID, email string) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/invitations", groupID), map[string]any{
		"email": email,
		"role":  "member",
	}, h.authHeader(inviterID))
	if status != http.StatusCreated {
		t.Fatalf("invite expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal invitation: %v", err)
	}
	return out.Token
}

func joinGroup(t *testing.T, h *testHarness, userID, token string) {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/invitations/%s/accept", token), nil, h.authHeader(userID))
	if status != http.StatusOK {
		t.Fatalf("accept invitation expected 200, got %d (err=%v)", status, sr.Error)
	}
}

type paymentPayload struct {
	UserID        string `json:"user_id"`
	Amount        string `json:"amount"`
	PaymentMethod string `json:"payment_method,omitempty"`
}

type splitPayload struct {
	UserID      string `json:"user_id"`
	AmountOwned string `json:"amount_owned"`
	SplitType   string `json:"split_type,omitempty"`
}

type createExpensePayload struct {
	Title        string           `json:"title"`
	Notes        string           `json:"notes,omitempty"`
	Amount       string           `json:"amount"`
	CurrencyCode string           `json:"currency_code,omitempty"`
	Date         string           `json:"date"`
	Payments     []paymentPayload `json:"payments"`
	Splits       []splitPayload   `json:"splits"`
}

func createExpense(t *testing.T, h *testHarness, userID, groupID string, payload createExpensePayload) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/expenses", groupID), payload, h.authHeader(userID))
	if status != http.StatusCreated {
		t.Fatalf("create expense expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		Expense struct {
			ID string `json:"id"`
		} `json:"expense"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal create expense: %v", err)
	}
	if out.Expense.ID == "" {
		t.Fatalf("expected expense.id")
	}
	return out.Expense.ID
}

func createSettlement(t *testing.T, h *testHarness, requesterID, groupID, payerID, payeeID, amount, currency string) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/settlements", groupID), map[string]any{
		"payer_id":      payerID,
		"payee_id":      payeeID,
		"amount":        amount,
		"currency_code": currency,
		"status":        "pending",
	}, h.authHeader(requesterID))
	if status != http.StatusCreated {
		t.Fatalf("create settlement expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal settlement: %v", err)
	}
	if out.ID == "" {
		t.Fatalf("expected settlement id")
	}
	return out.ID
}

func updateSettlementStatus(t *testing.T, h *testHarness, requesterID, groupID, settlementID, statusVal string) {
	t.Helper()
	status, sr := h.doJSON(http.MethodPatch, fmt.Sprintf("/groups/%s/settlements/%s/status", groupID, settlementID), map[string]any{
		"status": statusVal,
	}, h.authHeader(requesterID))
	if status != http.StatusOK {
		t.Fatalf("update settlement status expected 200, got %d (err=%v)", status, sr.Error)
	}
}

type createRecurringPayload struct {
	Title          string           `json:"title"`
	Notes          string           `json:"notes,omitempty"`
	Amount         string           `json:"amount"`
	CurrencyCode   string           `json:"currency_code,omitempty"`
	RepeatInterval string           `json:"repeat_interval"`
	DayOfMonth     *int             `json:"day_of_month,omitempty"`
	DayOfWeek      *int             `json:"day_of_week,omitempty"`
	StartDate      string           `json:"start_date"`
	EndDate        *string          `json:"end_date,omitempty"`
	Payments       []paymentPayload `json:"payments"`
	Splits         []splitPayload   `json:"splits"`
}

func createRecurringExpense(t *testing.T, h *testHarness, userID, groupID string, payload createRecurringPayload) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/recurring-expenses", groupID), payload, h.authHeader(userID))
	if status != http.StatusCreated {
		t.Fatalf("create recurring expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal recurring: %v", err)
	}
	if out.ID == "" {
		t.Fatalf("expected recurring id")
	}
	return out.ID
}
