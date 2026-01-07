# Authentication Flow

## User Login with Username and Password

User logs in by submitting their credentials to the authentication endpoint. Passwords are verified against Argon2id hashes stored in the database.

**Endpoint:**
```
POST /api/v1/auth/login
```

**Request Payload:**
```json
{
    "username": "string",
    "password": "string"
}
```

**Response:**

In return they receive a session token to interact with the Shen application. This token is stored locally and used to authorize the user/CLI to interact with Shen. Users can then manage their PATs for applications or check their group memberships. Group managers have user privileges plus the ability to add/remove members from groups they manage. Administrators have all privileges including the ability to manage users, groups, applications, RBAC settings, and assign group managers.

The session token is a cryptographically secure random string (32 bytes), hashed using SHA-256 and stored in the database (in `shen_session` table) with:

- 30 days expiration (configurable via `SHEN_SESSION_EXPIRY_DAYS`)
- Can be instantly revoked by administrators
- Validated on every request to Shen
- Plaintext token returned only once at creation time

```
Status: 200 OK
```

```json
{
    "session_token": "shen_session_a1b2c3d4e5f6..."
}
```

**Error Responses:**
- `401 Unauthorized` - Invalid username or password
- `403 Forbidden` - User account is inactive

---

## User Logout

User logs out by revoking their current session token. The token is extracted from the Authorization header, hashed, and the corresponding session is revoked.

**Endpoint:**
```
POST /api/v1/auth/logout
```

**Headers:**
```
Authorization: Bearer <session-token>
```

**Process:**

1. Extract session token from Authorization header
2. Hash the token using SHA-256
3. Lookup session in `shen_session` table by token hash
4. Update session record:
   - Set `revoked = true`
   - Set `revoked_at = NOW()` (UTC timestamp)

**Response:**

```
Status: 204 No Content
```

No response body on successful logout.

**Error Responses:**
- `401 Unauthorized` - Invalid or expired session token

**Example:**

```bash
curl -X POST https://shen.example.com/api/v1/auth/logout \
  -H "Authorization: Bearer shen_session_abc123..."
```
