# Authentication Endpoints

This folder contains Bruno API requests for JWT-based authentication.

## Quick Start

1. **Create a user** (if you haven't already):
   - Go to `users/create-user.yml`
   - Run the request
   - Note the user's email and password

2. **Login**:
   - Run `auth/login.yml`
   - Access token and refresh token are automatically saved to environment variables
   - You're now authenticated for all protected endpoints!

3. **Use protected endpoints**:
   - All protected endpoints require the `Authorization` header with Bearer token
   - Format: `Authorization: Bearer {access_token}`
   - Tokens are automatically included in requests

## Endpoints

### 1. Login

**File**: `login.yml`  
**Method**: `POST /auth/login`

Authenticates with email/password and returns JWT tokens.

**Auto-saves**:

- `access_token` → Used for all authenticated requests
- `refresh_token` → Used to get new access tokens

**Token Lifetimes**:

- Access token: 7 days
- Refresh token: 30 days

---

### 2. Refresh Token

**File**: `refresh-token.yml`  
**Method**: `POST /auth/refresh`

Gets a new access token using your refresh token.

**Use when**:

- Access token expires
- You want to extend your session

**Auto-updates**: `access_token`

---

### 3. Logout

**File**: `logout.yml`  
**Method**: `POST /auth/logout`

Logs out from current device/session.

**What happens**:

- Invalidates the refresh token
- Blacklists the access token (immediate revocation)
- Clears tokens from environment

---

### 4. Logout All

**File**: `logout-all.yml`  
**Method**: `POST /auth/logout-all`

Logs out from ALL devices/sessions.

**Use when**:

- Security concern (unauthorized access)
- Want to force logout everywhere
- After password change

**What happens**:

- Invalidates all refresh tokens
- User logged out from all devices
- Clears tokens from environment

---

## Environment Variables

The following variables are used by auth endpoints:

| Variable        | Description        | Auto-set?              |
| --------------- | ------------------ | ---------------------- |
| `user_email`    | Email for login    | No - set manually      |
| `user_password` | Password for login | No - set manually      |
| `access_token`  | JWT access token   | Yes - by login/refresh |
| `refresh_token` | JWT refresh token  | Yes - by login         |

## Authentication Flow

```
1. Create User (POST /users)
   ↓
2. Login (POST /auth/login)
   ↓
3. Get access_token + refresh_token (auto-saved)
   ↓
4. Use access_token for API requests (via Authorization header)
   ↓
5. When access_token expires (7 days):
   → Refresh token (POST /auth/refresh) to get new access_token
   ↓
6. When done:
   → Logout (single device) or Logout All (all devices)
```

## Using Authentication in Requests

All protected endpoints require the Authorization header:

```yaml
http:
  headers:
    - name: Authorization
      value: "Bearer {{access_token}}"
  auth: none
```

### Public Endpoints (No Auth Required)
- `POST /users` - Create user
- `POST /auth/login` - Login
- `POST /auth/refresh` - Refresh token
- `GET /categories/presets` - Get category presets
- `GET /invitations/{token}` - View invitation details

### Protected Endpoints (Auth Required)
- All `/groups/*` endpoints
- All `/friends/*` endpoints
- All `/expenses/*` endpoints
- All `/settlements/*` endpoints
- All `/balances/*` endpoints
- All `/categories/*` endpoints (except presets)
- All `/recurring-expenses/*` endpoints
- All `/comments/*` endpoints
- All `/activities/*` endpoints

## Testing Tips

1. **First time setup**:

   ```
   1. Create user (users/create-user.yml)
   2. Update user_email and user_password in environment
   3. Run login
   4. Now all protected endpoints will work
   ```

2. **Daily usage**:

   ```
   1. Run login once
   2. All other requests work automatically with Bearer token
   ```

3. **Token expiration testing**:

   ```
   1. Login
   2. Try protected endpoint → works
   3. Wait 7 days (or manually expire token in DB)
   4. Try protected endpoint → fails (401)
   5. Run refresh-token
   6. Try protected endpoint → works again
   ```

4. **Logout testing**:
   ```
   1. Login
   2. Try protected endpoint → works
   3. Logout
   4. Try protected endpoint → fails (401 Unauthorized)
   ```

## Response Format

All auth endpoints return:

**Success**:

```json
{
  "status": true,
  "data": {
    // endpoint-specific data
  }
}
```

**Error**:

```json
{
  "status": false,
  "error": {
    "message": "error description"
  }
}
```

## Common Status Codes

- `200 OK` - Success
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Invalid credentials or missing/invalid token
- `403 Forbidden` - Valid token but insufficient permissions
- `404 Not Found` - Resource not found
- `422 Unprocessable Entity` - Validation error
- `500 Internal Server Error` - Server error

## Notes

- All protected endpoints use `Authorization: Bearer {token}` header
- Tokens are automatically managed by Bruno scripts
- Access tokens expire in 7 days
- Refresh tokens expire in 30 days
- Blacklisted tokens are cleaned up hourly by background worker
- All timestamps are in UTC
