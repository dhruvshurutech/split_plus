# P0 Features API Documentation

## Overview
This document covers the API endpoints for Phase 1 P0 features: **Expense Categories**, **Comments**, and **Activity Feed**.

---

## 1. Expense Categories

### Get Category Presets
Get system-defined category presets.

```http
GET /categories/presets
```

**Authentication:** Not required

**Response:**
```json
{
  "status": true,
  "data": [
    {
      "slug": "food",
      "name": "Food & Drink",
      "icon": "🍔",
      "color": "#FF6B6B"
    }
  ]
}
```

---

### List Group Categories
Get all categories for a specific group.

```http
GET /groups/{group_id}/categories
```

**Authentication:** Required  
**Authorization:** User must be a group member

**Response:**
```json
{
  "status": true,
  "data": [
    {
      "id": "uuid",
      "group_id": "uuid",
      "slug": "food",
      "name": "Food",
      "icon": "🍔",
      "color": "#FF6B6B",
      "created_by": "uuid",
      "updated_by": "uuid",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

**Error Responses:**
- `400`: Invalid group_id
- `401`: Unauthorized
- `403`: Not a group member

---

### Create Category
Create a custom category for a group.

```http
POST /groups/{group_id}/categories
```

**Authentication:** Required  
**Authorization:** User must be a group member

**Request Body:**
```json
{
  "name": "Custom Category",
  "icon": "💰",
  "color": "#FF0000"
}
```

**Validation:**
- `name`: Required
- `icon`: Optional
- `color`: Optional

**Response:** `201 Created`
```json
{
  "status": true,
  "data": {
    "id": "uuid",
    "group_id": "uuid",
    "slug": "custom-category",
    "name": "Custom Category",
    "icon": "💰",
    "color": "#FF0000"
  }
}
```

**Error Responses:**
- `400`: Invalid request
- `401`: Unauthorized
- `403`: Not a group member
- `409`: Category already exists
- `422`: Validation error

---

### Create Categories from Presets
Bulk create categories from system presets.

```http
POST /groups/{group_id}/categories/from-presets
```

**Authentication:** Required  
**Authorization:** User must be a group member

**Request Body:**
```json
{
  "preset_slugs": ["food", "transport", "entertainment"]
}
```

**Response:** `201 Created`
```json
{
  "status": true,
  "data": [
    {
      "id": "uuid",
      "slug": "food",
      "name": "Food & Drink",
      "icon": "🍔",
      "color": "#FF6B6B"
    }
  ]
}
```

---

### Update Category
Update an existing category.

```http
PUT /groups/{group_id}/categories/{id}
```

**Authentication:** Required  
**Authorization:** User must be a group member

**Request Body:**
```json
{
  "name": "Updated Name",
  "icon": "🎉",
  "color": "#00FF00"
}
```

**Response:** `200 OK`

**Error Responses:**
- `400`: Invalid request
- `401`: Unauthorized
- `403`: Not a group member
- `404`: Category not found

---

### Delete Category
Soft delete a category.

```http
DELETE /groups/{group_id}/categories/{id}
```

**Authentication:** Required  
**Authorization:** User must be a group member

**Response:** `204 No Content`

**Error Responses:**
- `400`: Invalid request
- `401`: Unauthorized
- `403`: Not a group member
- `404`: Category not found

---

## 2. Expense Comments

### Create Comment
Add a comment to an expense.

```http
POST /groups/{group_id}/expenses/{expense_id}/comments
```

**Authentication:** Required  
**Authorization:** User must be a group member

**Request Body:**
```json
{
  "comment": "Great expense!"
}
```

**Validation:**
- `comment`: Required, min 1 char, max 1000 chars

**Response:** `201 Created`
```json
{
  "status": true,
  "data": {
    "id": "uuid",
    "expense_id": "uuid",
    "user_id": "uuid",
    "comment": "Great expense!",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Error Responses:**
- `400`: Invalid request
- `401`: Unauthorized
- `403`: Not a group member
- `404`: Expense not found
- `422`: Validation error

---

### List Comments
Get all comments for an expense.

```http
GET /groups/{group_id}/expenses/{expense_id}/comments
```

**Authentication:** Not required (public within group context)

**Response:** `200 OK`
```json
{
  "status": true,
  "data": [
    {
      "id": "uuid",
      "expense_id": "uuid",
      "user_id": "uuid",
      "comment": "Great expense!",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "user": {
        "email": "user@example.com",
        "name": "John Doe",
        "avatar_url": "https://..."
      }
    }
  ]
}
```

**Error Responses:**
- `400`: Invalid expense_id

---

### Update Comment
Update an existing comment.

```http
PUT /groups/{group_id}/expenses/{expense_id}/comments/{comment_id}
```

**Authentication:** Required  
**Authorization:** User must be the comment author

**Request Body:**
```json
{
  "comment": "Updated comment"
}
```

**Response:** `200 OK`

**Error Responses:**
- `400`: Invalid request
- `401`: Unauthorized
- `403`: Permission denied (not comment author)
- `404`: Comment not found
- `422`: Validation error

---

### Delete Comment
Delete a comment.

```http
DELETE /groups/{group_id}/expenses/{expense_id}/comments/{comment_id}
```

**Authentication:** Required  
**Authorization:** User must be the comment author

**Response:** `204 No Content`

**Error Responses:**
- `400`: Invalid request
- `401`: Unauthorized
- `403`: Permission denied (not comment author)
- `404`: Comment not found

---

## 3. Activity Feed

### List Group Activities
Get activity feed for a group.

```http
GET /groups/{group_id}/activity?limit=20&offset=0
```

**Authentication:** Not required (implementation may vary)

**Query Parameters:**
- `limit`: Optional, default 20, max items to return
- `offset`: Optional, default 0, pagination offset

**Response:** `200 OK`
```json
{
  "status": true,
  "data": [
    {
      "id": "uuid",
      "group_id": "uuid",
      "user_id": "uuid",
      "action": "expense_created",
      "entity_type": "expense",
      "entity_id": "uuid",
      "metadata": {
        "amount": "25.50",
        "title": "Lunch"
      },
      "created_at": "2024-01-01T00:00:00Z",
      "user": {
        "email": "user@example.com",
        "name": "John Doe",
        "avatar_url": "https://..."
      }
    }
  ]
}
```

**Activity Actions:**
- `expense_created`
- `expense_updated`
- `expense_deleted`
- `comment_added`
- `settlement_created`
- `settlement_completed`
- `member_added`
- `member_joined`

**Error Responses:**
- `400`: Invalid group_id
- `500`: Internal server error

---

### Get Expense History
Get activity history for a specific expense.

```http
GET /groups/{group_id}/expenses/{expense_id}/history
```

**Authentication:** Not required

**Response:** `200 OK`
```json
{
  "status": true,
  "data": [
    {
      "id": "uuid",
      "group_id": "uuid",
      "user_id": "uuid",
      "action": "expense_created",
      "entity_type": "expense",
      "entity_id": "uuid",
      "metadata": {},
      "created_at": "2024-01-01T00:00:00Z",
      "user": {
        "email": "user@example.com",
        "name": "John Doe"
      }
    }
  ]
}
```

**Error Responses:**
- `400`: Invalid expense_id
- `500`: Internal server error

---

## Common Error Response Format

All endpoints return errors in this format:

```json
{
  "status": false,
  "error": {
    "message": "Error description",
    "code": "ERROR_CODE"
  }
}
```

## Authentication

Most endpoints require authentication via JWT token in the Authorization header:

```http
Authorization: Bearer <jwt_token>
```

The token should contain the `user_id` claim for authorization checks.

---

## Notes

1. **UUIDs**: All IDs are UUIDs in string format
2. **Timestamps**: All timestamps are in RFC3339 format (ISO 8601)
3. **Soft Deletes**: Deleted entities have `deleted_at` timestamp set
4. **Pagination**: Activity endpoints support limit/offset pagination
5. **Activity Logging**: Comments and category operations automatically create activity log entries
