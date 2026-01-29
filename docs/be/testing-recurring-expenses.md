# Testing Recurring Expenses

## Overview

This document explains how to test the recurring expenses feature, including unit tests, integration tests, and manual testing.

## Test Structure

### Unit Tests (Service Layer)

Location: `internal/service/recurring_expense_service_test.go`

**What to Test:**
- ✅ Create recurring expense (daily, weekly, monthly, yearly)
- ✅ Validation of interval-specific fields (day_of_month, day_of_week)
- ✅ Date calculation logic for each interval
- ✅ Payment and split template copying
- ✅ Error cases (invalid inputs, not group member, etc.)
- ✅ Generate expense from recurring template
- ✅ Process due recurring expenses (worker job)

**Test Pattern:**
```go
func TestRecurringExpenseService_CreateRecurringExpense(t *testing.T) {
    tests := []struct {
        name          string
        input         CreateRecurringExpenseInput
        mockSetup     func(*MockRecurringExpenseRepository, *MockExpenseRepository)
        expectedError error
        validate      func(*testing.T, sqlc.RecurringExpense)
    }{
        // Test cases...
    }
    // Run tests...
}
```

### Handler Tests (HTTP Layer)

Location: `internal/http/handlers/recurring_expenses_test.go` (to be created)

**What to Test:**
- ✅ HTTP request/response handling
- ✅ Request validation
- ✅ Error status codes mapping
- ✅ Authentication/authorization
- ✅ JSON serialization/deserialization

## Running Tests

### Run All Tests
```bash
just test
```

### Run Specific Test File
```bash
go test ./internal/service -run TestRecurringExpenseService
```

### Run with Coverage
```bash
just test-coverage
```

### Run Specific Test Case
```bash
go test ./internal/service -run TestRecurringExpenseService_CreateRecurringExpense
```

## Test Scenarios

### 1. Create Daily Recurring Expense

**Test Case:**
- Interval: `daily`
- `day_of_month`: `nil`
- `day_of_week`: `nil`
- Should succeed

**Expected:**
- Recurring expense created
- `next_occurrence_date` = `start_date`
- Payments and splits saved

### 2. Create Weekly Recurring Expense

**Test Case:**
- Interval: `weekly`
- `day_of_week`: `1` (Monday)
- `day_of_month`: `nil`
- Should succeed

**Expected:**
- Recurring expense created
- Validation passes

### 3. Create Monthly Recurring Expense

**Test Case:**
- Interval: `monthly`
- `day_of_month`: `3`
- `day_of_week`: `nil`
- Should succeed

**Expected:**
- Recurring expense created
- Next occurrence calculated correctly

### 4. Invalid Interval Configuration

**Test Cases:**
- Daily with `day_of_month` set → Should fail
- Weekly without `day_of_week` → Should fail
- Monthly without `day_of_month` → Should fail

### 5. Generate Expense from Recurring

**Test Case:**
- Recurring expense with `next_occurrence_date` <= today
- Call `GenerateExpenseFromRecurring`
- Should create expense with copied payments/splits
- Should update `next_occurrence_date`

**Expected:**
- New expense created
- Payments copied from template
- Splits copied from template
- `next_occurrence_date` updated

### 6. Date Calculation Edge Cases

**Test Cases:**
- Monthly: Day 31 → Next month with fewer days (should use last day)
- Monthly: Feb 29/30 → Handle leap years
- Yearly: Same month-end handling

## Integration Testing

### Manual Testing with Bruno/Postman

1. **Create Recurring Expense**
   ```
   POST /groups/{group_id}/recurring-expenses
   {
     "title": "Spotify Subscription",
     "amount": "9.99",
     "repeat_interval": "monthly",
     "day_of_month": 3,
     "start_date": "2024-01-03",
     "payments": [...],
     "splits": [...]
   }
   ```

2. **List Recurring Expenses**
   ```
   GET /groups/{group_id}/recurring-expenses
   ```

3. **Get Recurring Expense**
   ```
   GET /groups/{group_id}/recurring-expenses/{id}
   ```

4. **Update Recurring Expense**
   ```
   PUT /groups/{group_id}/recurring-expenses/{id}
   ```

5. **Manually Generate Expense**
   ```
   POST /groups/{group_id}/recurring-expenses/{id}/generate
   ```

6. **Delete Recurring Expense**
   ```
   DELETE /groups/{group_id}/recurring-expenses/{id}
   ```

### Testing Worker Job

1. **Create recurring expense with `next_occurrence_date` = today**
2. **Start worker**: `go run ./cmd/worker`
3. **Wait for processing** (or trigger manually)
4. **Verify**: Expense created, `next_occurrence_date` updated

## Test Coverage Goals

- **Service Layer**: 100% coverage
- **Handler Layer**: >90% coverage
- **Repository Layer**: Tested via service tests

## Common Test Patterns

### Mock Setup Pattern
```go
mockSetup: func(repo *MockRecurringExpenseRepository, expenseRepo *MockExpenseRepository) {
    repo.GetGroupByIDFunc = func(ctx context.Context, id pgtype.UUID) (sqlc.Group, error) {
        return testutil.CreateTestGroup(groupID, "Test Group", userID), nil
    }
    repo.GetGroupMemberFunc = func(ctx context.Context, params sqlc.GetGroupMemberParams) (sqlc.GroupMember, error) {
        return testutil.CreateTestGroupMember(...), nil
    }
    // ... more mocks
}
```

### Validation Pattern
```go
validate: func(t *testing.T, re sqlc.RecurringExpense) {
    if re.Title != "Expected Title" {
        t.Errorf("expected title 'Expected Title', got '%s'", re.Title)
    }
    // ... more validations
}
```

## Next Steps

1. ✅ Service layer tests created (basic structure)
2. ⏳ Add more comprehensive test cases
3. ⏳ Create handler tests
4. ⏳ Add integration tests
5. ⏳ Test worker job execution
