# Authentication Flow

## User Login with Username and Password

User logs in by submitting their credentials to the authentication endpoint.

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

In return they receive a session token to interact with the Shen application. This token is stored locally and used to authorize the user/CLI to interact with Shen. Users can then manage their PATs for applications or check their group memberships. Administrators have all user privileges plus the ability to manage users, groups, applications, and RBAC settings.

The session token is a random string stored in the database (in `shen_sessions` table) with:

- 30 days expiration (configurable via `SHEN_SESSION_EXPIRY_DAYS`)
- Can be instantly revoked by administrators
- Validated on every request to Shen

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
