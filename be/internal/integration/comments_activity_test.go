package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCommentsAndActivityFlow(t *testing.T) {
	h := newHarness(t)

	// 1) Users
	userA := createUser(t, h, "ca_a@example.com")
	userB := createUser(t, h, "ca_b@example.com")

	// 2) Group
	groupID := createGroup(t, h, userA.ID, "Activity Group", "USD")
	tokenB := inviteUserToGroup(t, h, userA.ID, groupID, userB.Email)
	joinGroup(t, h, userB.ID, tokenB)

	// 3) Create Expense (Should log 'expense_created')
	expenseID := createExpense(t, h, userA.ID, groupID, createExpensePayload{
		Title:  "Lunch",
		Amount: "20.00",
		Date:   "2024-01-10",
		Payments: []paymentPayload{
			{UserID: userA.ID, Amount: "20.00"},
		},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "10.00", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "10.00", SplitType: "equal"},
		},
	})

	// 4) Add Comment (Should log 'comment_added')
	_ = createComment(t, h, userB.ID, groupID, expenseID, "Nice lunch!")

	// 5) List Comments
	listComments(t, h, userA.ID, groupID, expenseID, 1)

	// 6) Update Expense (Should log 'expense_updated')
	// We'll update amount to 22.00
	newAmount := "22.00"
	half := "11.00"
	updatePayload := createExpensePayload{
		Title:  "Lunch Updated",
		Amount: newAmount,
		Date:   "2024-01-10",
		Payments: []paymentPayload{
			{UserID: userA.ID, Amount: newAmount},
		},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: half, SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: half, SplitType: "equal"},
		},
	}

	status, sr := h.doJSON(http.MethodPut, fmt.Sprintf("/groups/%s/expenses/%s", groupID, expenseID), updatePayload, h.authHeader(userA.ID))
	if status != http.StatusOK {
		t.Fatalf("update expense expected 200, got %d (err=%v)", status, sr.Error)
	}

	// 7) Create Settlement (Should log 'settlement_created')
	settlementID := createSettlement(t, h, userA.ID, groupID, userB.ID, userA.ID, "10.00", "USD")

	// 8) Complete Settlement (Should log 'settlement_completed')
	updateSettlementStatus(t, h, userA.ID, groupID, settlementID, "completed")

	// 9) Verify Activity Feed
	// We expect: settlement_completed, settlement_created, expense_updated, comment_added, expense_created
	// (Check order desc)
	activities := listGroupActivities(t, h, userA.ID, groupID)
	if len(activities) < 5 {
		t.Fatalf("expected at least 5 activities, got %d", len(activities))
	}

	// Simple verification of order/types
	// Note: activity order depends on timestamp granularity but usually sequential
	expectedActions := []string{"settlement_completed", "settlement_created", "expense_updated", "comment_added", "expense_created"}
	for i, action := range expectedActions {
		if i >= len(activities) {
			break
		}
		if activities[i].Action != action {
			t.Errorf("expected activity %d to be %s, got %s", i, action, activities[i].Action)
		}
	}
}

func createComment(t *testing.T, h *testHarness, userID, groupID, expenseID, comment string) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/expenses/%s/comments", groupID, expenseID), map[string]any{
		"comment": comment,
	}, h.authHeader(userID))
	if status != http.StatusCreated {
		t.Fatalf("create comment expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	return out.ID
}

func listComments(t *testing.T, h *testHarness, userID, groupID, expenseID string, expectedCount int) {
	t.Helper()
	status, sr := h.doJSON(http.MethodGet, fmt.Sprintf("/groups/%s/expenses/%s/comments", groupID, expenseID), nil, h.authHeader(userID))
	if status != http.StatusOK {
		t.Fatalf("list comments expected 200, got %d (err=%v)", status, sr.Error)
	}
	var comments []map[string]any
	if err := json.Unmarshal(sr.Data, &comments); err != nil {
		t.Fatalf("unmarshal comments: %v", err)
	}
	if len(comments) != expectedCount {
		t.Fatalf("expected %d comments, got %d", expectedCount, len(comments))
	}
}

type activityResponse struct {
	Action string `json:"action"`
}

func listGroupActivities(t *testing.T, h *testHarness, userID, groupID string) []activityResponse {
	t.Helper()
	status, sr := h.doJSON(http.MethodGet, fmt.Sprintf("/groups/%s/activity", groupID), nil, h.authHeader(userID))
	if status != http.StatusOK {
		t.Fatalf("list activity expected 200, got %d (err=%v)", status, sr.Error)
	}
	var activities []activityResponse
	if err := json.Unmarshal(sr.Data, &activities); err != nil {
		t.Fatalf("unmarshal activities: %v", err)
	}
	return activities
}
