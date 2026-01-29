# Expense Service Test Cases

This document provides a comprehensive overview of all test cases for the expense creation functionality.

---

## 📋 Table of Contents

1. [Service Layer Tests](#service-layer-tests)
   - [Validation Error Cases](#validation-error-cases)
   - [Successful Cases](#successful-cases)
2. [Handler Layer Tests](#handler-layer-tests)

---

## 🔧 Service Layer Tests

**Test Function:** `TestExpenseService_CreateExpense`  
**File:** `internal/service/expense_service_test.go`

### Validation Error Cases

These tests verify that the service properly rejects invalid inputs.

#### 1. **Empty Title**
- **Test Name:** `empty title`
- **Input:**
  - Title: `""` (empty)
  - Amount: `"100.00"`
  - Payments: 1 payment of `"100.00"`
  - Splits: 1 split of `"100.00"`
- **Expected Error:** `"title is required"`
- **What it tests:** Title validation - ensures expense must have a title

---

#### 2. **Invalid Amount - Zero**
- **Test Name:** `invalid amount - zero`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"0"`
  - Payments: 1 payment of `"0"`
  - Splits: 1 split of `"0"`
- **Expected Error:** `ErrInvalidAmount`
- **What it tests:** Expense amount must be greater than zero

---

#### 3. **Invalid Amount - Negative**
- **Test Name:** `invalid amount - negative`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"-10.00"`
  - Payments: 1 payment of `"-10.00"`
  - Splits: 1 split of `"-10.00"`
- **Expected Error:** `ErrInvalidAmount`
- **What it tests:** Expense amount cannot be negative

---

#### 4. **Invalid Amount - Invalid String**
- **Test Name:** `invalid amount - invalid string`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"not-a-number"`
  - Payments: 1 payment of `"not-a-number"`
  - Splits: 1 split of `"not-a-number"`
- **Expected Error:** `ErrInvalidAmount`
- **What it tests:** Amount must be a valid decimal number

---

#### 5. **No Payments**
- **Test Name:** `no payments`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: `[]` (empty array)
  - Splits: 1 split of `"100.00"`
- **Expected Error:** `"at least one payment is required"`
- **What it tests:** Expense must have at least one payment

---

#### 6. **No Splits**
- **Test Name:** `no splits`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"100.00"`
  - Splits: `[]` (empty array)
- **Expected Error:** `"at least one split is required"`
- **What it tests:** Expense must have at least one split

---

#### 7. **Payment Total Mismatch - Less Than Expense**
- **Test Name:** `payment total mismatch - less than expense`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"50.00"` (total = 50, but expense = 100)
  - Splits: 1 split of `"100.00"`
- **Expected Error:** `ErrPaymentTotalMismatch`
- **What it tests:** Sum of all payments must equal expense amount

---

#### 8. **Payment Total Mismatch - Exceeds Expense**
- **Test Name:** `payment total mismatch - exceeds expense`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"150.00"` (total = 150, but expense = 100)
  - Splits: 1 split of `"100.00"`
- **Expected Error:** `ErrPaymentTotalMismatch`
- **What it tests:** Sum of all payments cannot exceed expense amount

---

#### 9. **Invalid Payment Amount - Zero**
- **Test Name:** `invalid payment amount - zero`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"0"` (invalid - must be > 0)
  - Splits: 1 split of `"100.00"`
- **Expected Error:** `ErrInvalidAmount`
- **What it tests:** Individual payment amounts must be greater than zero

---

#### 10. **Invalid Payment Amount - Negative**
- **Test Name:** `invalid payment amount - negative`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"-10.00"` (invalid - cannot be negative)
  - Splits: 1 split of `"100.00"`
- **Expected Error:** `ErrInvalidAmount`
- **What it tests:** Individual payment amounts cannot be negative

---

#### 11. **Split Total Mismatch - Less Than Expense**
- **Test Name:** `split total mismatch - less than expense`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"100.00"`
  - Splits: 1 split of `"50.00"` (total = 50, but expense = 100)
- **Expected Error:** `ErrSplitTotalMismatch`
- **What it tests:** Sum of all splits must equal expense amount

---

#### 12. **Split Total Mismatch - Exceeds Expense**
- **Test Name:** `split total mismatch - exceeds expense`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"100.00"`
  - Splits: 1 split of `"150.00"` (total = 150, but expense = 100)
- **Expected Error:** `ErrSplitTotalMismatch`
- **What it tests:** Sum of all splits cannot exceed expense amount

---

#### 13. **Invalid Split Amount - Negative**
- **Test Name:** `invalid split amount - negative`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"100.00"`
  - Splits: 1 split of `"-10.00"` (invalid - cannot be negative)
- **Expected Error:** `ErrInvalidAmount`
- **What it tests:** Individual split amounts cannot be negative (can be zero for some users)

---

#### 14. **Not Group Member**
- **Test Name:** `not group member`
- **Input:**
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment of `"100.00"`
  - Splits: 1 split of `"100.00"`
  - CreatedBy: User who is NOT a member of the group
- **Expected Error:** `ErrNotGroupMember`
- **What it tests:** Only group members can create expenses

---

### Successful Cases

These tests verify that valid expense creation scenarios work correctly.

#### 15. **Valid Equal Split - Single Payer** ✅
- **Test Name:** `valid equal split single payer`
- **Scenario:** Split by equally
- **Input:**
  - Title: `"Dinner"`
  - Amount: `"100.00"`
  - Payments: 
    - User 1: `"100.00"` (single payer covers full amount)
  - Splits:
    - User 1: `"50.00"` (SplitType: `"equal"`)
    - User 2: `"50.00"` (SplitType: `"equal"`)
- **Expected Result:** ✅ Success (no error)
- **What it tests:** 
  - One person pays the full expense
  - Expense is split equally between two people (50/50)
  - Payment total (100) = Expense amount (100) ✓
  - Split total (50 + 50 = 100) = Expense amount (100) ✓

---

#### 16. **Valid Split by Percentage - Multiple Payers** ✅
- **Test Name:** `valid split by percentage multiple payers`
- **Scenario:** Split by percentage with multiple payers
- **Input:**
  - Title: `"Trip"`
  - Amount: `"100.00"`
  - Payments:
    - User 1: `"60.00"` (60% of expense)
    - User 2: `"40.00"` (40% of expense)
  - Splits:
    - User 1: `"60.00"` (SplitType: `"percentage"`)
    - User 2: `"40.00"` (SplitType: `"percentage"`)
- **Expected Result:** ✅ Success (no error)
- **What it tests:**
  - Multiple people pay for the expense (60 + 40 = 100)
  - Expense is split by percentage (60% / 40%)
  - Payment total (60 + 40 = 100) = Expense amount (100) ✓
  - Split total (60 + 40 = 100) = Expense amount (100) ✓

---

#### 17. **Valid Fixed Unequal Split - Single Payer** ✅
- **Test Name:** `valid fixed unequal split single payer`
- **Scenario:** Split by fixed (unequal) amounts
- **Input:**
  - Title: `"Groceries"`
  - Amount: `"150.00"`
  - Payments:
    - User 1: `"150.00"` (single payer covers full amount)
  - Splits:
    - User 1: `"100.00"` (SplitType: `"fixed"`)
    - User 2: `"50.00"` (SplitType: `"fixed"`)
- **Expected Result:** ✅ Success (no error)
- **What it tests:**
  - One person pays the full expense
  - Expense is split unequally (100 + 50 = 150)
  - Payment total (150) = Expense amount (150) ✓
  - Split total (100 + 50 = 150) = Expense amount (150) ✓

---

#### 18. **Valid Multiple Payers - Equal Split** ✅
- **Test Name:** `valid multiple payers with equal split`
- **Scenario:** Expense with multiple payers, split equally
- **Input:**
  - Title: `"Restaurant"`
  - Amount: `"200.00"`
  - Payments:
    - User 1: `"100.00"`
    - User 2: `"50.00"`
    - User 3: `"50.00"`
    - Total: 100 + 50 + 50 = 200 ✓
  - Splits:
    - User 1: `"66.67"` (SplitType: `"equal"`)
    - User 2: `"66.67"` (SplitType: `"equal"`)
    - User 3: `"66.66"` (SplitType: `"equal"`)
    - Total: 66.67 + 66.67 + 66.66 = 200 ✓
- **Expected Result:** ✅ Success (no error)
- **What it tests:**
  - Three people pay for the expense (100 + 50 + 50 = 200)
  - Expense is split equally between three people (approximately 66.67 each)
  - Payment total (200) = Expense amount (200) ✓
  - Split total (200) = Expense amount (200) ✓

---

#### 19. **Valid Multiple Payers - Percentage Split** ✅
- **Test Name:** `valid multiple payers with percentage split`
- **Scenario:** Expense with multiple payers, split by percentage
- **Input:**
  - Title: `"Hotel"`
  - Amount: `"300.00"`
  - Payments:
    - User 1: `"150.00"` (50% of expense)
    - User 2: `"100.00"` (33.33% of expense)
    - User 3: `"50.00"` (16.67% of expense)
    - Total: 150 + 100 + 50 = 300 ✓
  - Splits:
    - User 1: `"150.00"` (SplitType: `"percentage"`) - 50%
    - User 2: `"100.00"` (SplitType: `"percentage"`) - 33.33%
    - User 3: `"50.00"` (SplitType: `"percentage"`) - 16.67%
    - Total: 150 + 100 + 50 = 300 ✓
- **Expected Result:** ✅ Success (no error)
- **What it tests:**
  - Three people pay for the expense (150 + 100 + 50 = 300)
  - Expense is split by percentage (50% / 33.33% / 16.67%)
  - Payment total (300) = Expense amount (300) ✓
  - Split total (300) = Expense amount (300) ✓

---

## 🌐 Handler Layer Tests

**Test Function:** `TestCreateExpenseHandler`  
**File:** `internal/http/handlers/expenses_test.go`

These tests verify HTTP request/response handling and validation.

#### 1. **Invalid Date Format**
- **Test Name:** `invalid date format`
- **Request:**
  - Date: `"invalid-date"` (not in YYYY-MM-DD format)
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: Valid payment
  - Splits: Valid split
- **Expected Status:** `400 Bad Request`
- **What it tests:** Date must be in YYYY-MM-DD format

---

#### 2. **Invalid Payment User ID**
- **Test Name:** `invalid payment user_id`
- **Request:**
  - Date: Valid date
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: UserID = `"invalid-uuid"` (not a valid UUID)
  - Splits: Valid split
- **Expected Status:** `422 Unprocessable Entity`
- **What it tests:** Payment user_id must be a valid UUID (validated by middleware)

---

#### 3. **Service Error - Not Group Member**
- **Test Name:** `service error - not group member`
- **Request:**
  - All fields valid
  - User is not a member of the group
- **Mock:** Service returns `ErrNotGroupMember`
- **Expected Status:** `403 Forbidden`
- **What it tests:** Handler properly maps service errors to HTTP status codes

---

#### 4. **Successful Creation with User Info** ✅
- **Test Name:** `successful creation with user info`
- **Request:**
  - All fields valid
  - Title: `"Test Expense"`
  - Amount: `"100.00"`
  - Payments: 1 payment
  - Splits: 1 split
- **Mock:** 
  - Service returns successful expense creation
  - `GetExpensePayments` returns payment with user email/name
  - `GetExpenseSplits` returns split with user email/name
- **Expected Status:** `201 Created`
- **What it tests:**
  - Handler successfully creates expense
  - Response includes full user information (email, name) for payments and splits
  - Verifies the fix for empty user email issue

---

## 📊 Summary

### Total Test Cases: **23**

- **Service Layer:** 19 test cases
  - **Error Cases:** 14 tests
  - **Success Cases:** 5 tests
- **Handler Layer:** 4 test cases
  - **Error Cases:** 3 tests
  - **Success Cases:** 1 test

### Coverage

✅ **All 4 requested scenarios are tested:**
1. ✅ Split by equally
2. ✅ Split by percentage
3. ✅ Split by fixed (unequal amounts)
4. ✅ Expense with multiple payers

✅ **Comprehensive validation coverage:**
- Empty/invalid fields
- Amount validation (zero, negative, invalid format)
- Payment validation (totals, individual amounts)
- Split validation (totals, individual amounts)
- Group membership validation

✅ **HTTP layer coverage:**
- Request validation
- Error handling
- Response formatting with user info

---

## 🎯 Key Test Scenarios Explained

### Scenario 1: Equal Split
**Example:** Dinner for $100, split equally between 2 people
- **Who paid:** Person A paid $100
- **Who owes:** Person A owes $50, Person B owes $50
- **Result:** Person B owes Person A $50

### Scenario 2: Percentage Split
**Example:** Trip for $100, split 60/40
- **Who paid:** Person A paid $60, Person B paid $40
- **Who owes:** Person A owes $60, Person B owes $40
- **Result:** Balanced (each person paid what they owe)

### Scenario 3: Fixed Unequal Split
**Example:** Groceries for $150, split $100/$50
- **Who paid:** Person A paid $150
- **Who owes:** Person A owes $100, Person B owes $50
- **Result:** Person B owes Person A $50

### Scenario 4: Multiple Payers
**Example:** Restaurant for $200, 3 people pay, split equally
- **Who paid:** Person A paid $100, Person B paid $50, Person C paid $50
- **Who owes:** Each person owes $66.67
- **Result:** Person A is owed $33.33, Person B owes $16.67, Person C owes $16.67

---

## 🔍 Running the Tests

```bash
# Run all service tests
go test -v ./internal/service/... -run TestExpenseService_CreateExpense

# Run all handler tests
go test -v ./internal/http/handlers/... -run TestCreateExpenseHandler

# Run all tests
go test ./internal/service/... ./internal/http/handlers/...
```

---

## ✅ All Tests Passing

All 23 test cases are currently passing and provide comprehensive coverage of the expense creation functionality.
