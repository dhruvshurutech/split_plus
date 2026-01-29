package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

type searchExpensePayload struct {
	Title        string           `json:"title"`
	Notes        string           `json:"notes,omitempty"`
	Amount       string           `json:"amount"`
	CurrencyCode string           `json:"currency_code,omitempty"`
	Date         string           `json:"date"`
	CategoryID   string           `json:"category_id,omitempty"`
	Payments     []paymentPayload `json:"payments"`
	Splits       []splitPayload   `json:"splits"`
}

func createExpenseWithCategory(t *testing.T, h *testHarness, userID, groupID string, payload searchExpensePayload) string {
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
	return out.Expense.ID
}

func createCategory(t *testing.T, h *testHarness, userID, groupID, name, icon, color string) string {
	t.Helper()
	status, sr := h.doJSON(http.MethodPost, fmt.Sprintf("/groups/%s/categories", groupID), map[string]string{
		"name":  name,
		"icon":  icon,
		"color": color,
	}, h.authHeader(userID))
	if status != http.StatusCreated {
		t.Fatalf("create category expected 201, got %d (err=%v)", status, sr.Error)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sr.Data, &out); err != nil {
		t.Fatalf("unmarshal category: %v", err)
	}
	return out.ID
}

func TestSearchExpenses(t *testing.T) {
	h := newHarness(t)

	// 1) Users
	userA := createUser(t, h, "a@example.com")
	userB := createUser(t, h, "b@example.com")
	userC := createUser(t, h, "c@example.com")

	// 2) Group
	groupID := createGroup(t, h, userA.ID, "Search Trip", "USD")

	// Invite & Join
	tokenB := inviteUserToGroup(t, h, userA.ID, groupID, userB.Email)
	tokenC := inviteUserToGroup(t, h, userA.ID, groupID, userC.Email)
	joinGroup(t, h, userB.ID, tokenB)
	joinGroup(t, h, userC.ID, tokenC)

	// 3) Create Categories
	foodCatID := createCategory(t, h, userA.ID, groupID, "Food", "🍔", "#FF0000")
	entCatID := createCategory(t, h, userA.ID, groupID, "Entertainment", "🎬", "#00FF00")

	// 4) Create Expenses
	// Exp1: Lunch (Food, Paid by A, 50, 2024-01-01)
	exp1 := createExpenseWithCategory(t, h, userA.ID, groupID, searchExpensePayload{
		Title:      "Lunch",
		Notes:      "Team lunch",
		Amount:     "50.00",
		Date:       "2024-01-01",
		CategoryID: foodCatID,
		Payments: []paymentPayload{{UserID: userA.ID, Amount: "50.00"}},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "16.66", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "16.67", SplitType: "equal"},
			{UserID: userC.ID, AmountOwned: "16.67", SplitType: "equal"},
		},
	})

	// Exp2: Dinner (Food, Paid by B, 100, 2024-01-02, Notes="Pizza")
	exp2 := createExpenseWithCategory(t, h, userA.ID, groupID, searchExpensePayload{
		Title:      "Dinner",
		Notes:      "Pizza night",
		Amount:     "100.00",
		Date:       "2024-01-02",
		CategoryID: foodCatID,
		Payments: []paymentPayload{{UserID: userB.ID, Amount: "100.00"}},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "33.33", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "33.33", SplitType: "equal"},
			{UserID: userC.ID, AmountOwned: "33.34", SplitType: "equal"},
		},
	})

	// Exp3: Movie (Entertainment, Paid by A, 30, 2024-01-03)
	exp3 := createExpenseWithCategory(t, h, userA.ID, groupID, searchExpensePayload{
		Title:      "Movie",
		Amount:     "30.00",
		Date:       "2024-01-03",
		CategoryID: entCatID,
		Payments: []paymentPayload{{UserID: userA.ID, Amount: "30.00"}},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "10.00", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "10.00", SplitType: "equal"},
			{UserID: userC.ID, AmountOwned: "10.00", SplitType: "equal"},
		},
	})

	// Exp4: Taxi (No Category, Paid by C, 20, 2024-01-04)
	exp4 := createExpenseWithCategory(t, h, userA.ID, groupID, searchExpensePayload{
		Title:  "Taxi",
		Amount: "20.00",
		Date:   "2024-01-04",
		Payments: []paymentPayload{{UserID: userC.ID, Amount: "20.00"}},
		Splits: []splitPayload{
			{UserID: userA.ID, AmountOwned: "6.66", SplitType: "equal"},
			{UserID: userB.ID, AmountOwned: "6.67", SplitType: "equal"},
			{UserID: userC.ID, AmountOwned: "6.67", SplitType: "equal"},
		},
	})

	// Helper to run search and assertions
	runSearch := func(name string, params map[string]string, expectedCount int, expectedIDs []string) {
		t.Run(name, func(t *testing.T) {
			path := fmt.Sprintf("/groups/%s/expenses/search?", groupID)
			for k, v := range params {
				path += fmt.Sprintf("%s=%s&", k, v)
			}
			
			status, sr := h.doJSON(http.MethodGet, path, nil, h.authHeader(userA.ID))
			if status != http.StatusOK {
				t.Fatalf("search expected 200, got %d (err=%v)", status, sr.Error)
			}
			
			var data []struct{
				Expense struct {
					ID string `json:"id"`
				} `json:"expense"`
			}
			if err := json.Unmarshal(sr.Data, &data); err != nil {
				t.Fatalf("unmarshal search result: %v", err)
			}
			
			if len(data) != expectedCount {
				t.Errorf("expected %d results, got %d", expectedCount, len(data))
			}
			
			// Verify IDs if provided
			if expectedIDs != nil {
				foundMap := make(map[string]bool)
				for _, item := range data {
					foundMap[item.Expense.ID] = true
				}
				for _, id := range expectedIDs {
					if !foundMap[id] {
						t.Errorf("expected ID %s not found in results", id)
					}
				}
			}
		})
	}

	// 5) Run Searches
	
	// Query: "Lunch"
	runSearch("Query Title", map[string]string{"q": "Lunch"}, 1, []string{exp1})
	
	// Query: "Pizza" (in notes)
	runSearch("Query Notes", map[string]string{"q": "Pizza"}, 1, []string{exp2})
	
	// Category: Food
	runSearch("Category Food", map[string]string{"category_id": foodCatID}, 2, []string{exp1, exp2})
	
	// CreatedBy/Payer: A (Paid for Lunch and Movie)
	// Note: CreatedBy in my test setup is always userA (passed to createExpenseWithCategory), 
	// BUT 'payer_id' filter checks who paid. PayerID logic:
	// Exp1: A, Exp2: B, Exp3: A, Exp4: C
	runSearch("Payer A", map[string]string{"payer_id": userA.ID}, 2, []string{exp1, exp3})
	
	// Min Amount: 40 (Lunch=50, Dinner=100)
	runSearch("Min Amount 40", map[string]string{"min_amount": "40"}, 2, []string{exp1, exp2})
	
	// Max Amount: 40 (Movie=30, Taxi=20)
	runSearch("Max Amount 40", map[string]string{"max_amount": "40"}, 2, []string{exp3, exp4})
	
	// Date Range: 2024-01-02 to 2024-01-03 (Dinner, Movie)
	runSearch("Date Range", map[string]string{"start_date": "2024-01-02", "end_date": "2024-01-03"}, 2, []string{exp2, exp3})
	
	// Ower: C (owes in all)
	runSearch("Ower C", map[string]string{"ower_id": userC.ID}, 4, []string{exp1, exp2, exp3, exp4})
	
	// Combined: Category Food + Paid by A -> Lunch
	runSearch("Combined Food + Payer A", map[string]string{"category_id": foodCatID, "payer_id": userA.ID}, 1, []string{exp1})
}
