# Token Revocation

## Overview

Administrators can revoke any user or service account tokens. Regular users can only revoke their own tokens. Revoking a token immediately prevents it from being used to obtain new JWTs.

**Important:** Revoking a PAT does not invalidate already-issued JWTs. Applications must wait for the JWT to expire naturally (7 minutes by default). This is a deliberate trade-off for stateless JWT verification.

---

## Revoke Token

**Endpoint:**
```
DELETE /api/v1/token/:id
```

**Path Parameters:**
- `:id` - Token ID (primary key from `shen_tokens` table)

**Headers:**
```
Authorization: Bearer <session-token>
```

**Process:**

1. Validate session token
2. Lookup token by ID
3. Check authorization:
   - If user is admin → allow
   - If user owns the token (`shen_tokens.user_id == current_user.id`) → allow
   - Otherwise → 403 Forbidden
4. Update token record:
   - Set `revoked = true`
   - Set `revoked_at = NOW()` (UTC timestamp)

**Response:**

```
Status: 204 No Content
```

No response body on successful revocation.

**Error Responses:**

- `401 Unauthorized` - Invalid or expired session token
- `403 Forbidden` - User does not have permission to revoke this token
- `404 Not Found` - Token ID does not exist

**Example:**

```bash
# Revoke token with ID 42
curl -X DELETE https://shen.example.com/api/v1/token/42 \
  -H "Authorization: Bearer shen_session_abc123..."
```

---

## Revoke All User Tokens

Administrators can revoke all tokens for a specific user (useful for account compromise or offboarding).

**Endpoint:**
```
DELETE /api/v1/tokens/user/:username
```

**Path Parameters:**
- `:username` - Username of the account

**Headers:**
```
Authorization: Bearer <session-token>
```

**Authorization:**
- Admin only

**Process:**

1. Validate session token and verify user is admin
2. Find all tokens for the specified user
3. Update all token records:
   - Set `revoked = true`
   - Set `revoked_at = NOW()`
4. Return count of revoked tokens

**Response:**

```
Status: 200 OK
```

```json
{
  "username": "alice",
  "tokens_revoked": 5
}
```

**Error Responses:**

- `401 Unauthorized` - Invalid or expired session token
- `403 Forbidden` - User is not an administrator
- `404 Not Found` - Username does not exist

**Example:**

```bash
# Revoke all tokens for user 'alice'
curl -X DELETE https://shen.example.com/api/v1/tokens/user/alice \
  -H "Authorization: Bearer shen_session_admin456..."
```

---

## List User Tokens

To find token IDs for revocation, users can list their tokens.

**Endpoint:**
```
GET /api/v1/tokens
```

**Headers:**
```
Authorization: Bearer <session-token>
```

**Query Parameters:**
- `user` *(optional, admin only)* - Filter tokens by username

**Authorization:**
- Regular users: Can only list their own tokens
- Admins: Can list any user's tokens by providing `?user=<username>`

**Response:**

```
Status: 200 OK
```

```json
{
  "tokens": [
    {
      "id": 42,
      "name": "my-prod-token",
      "application": "my-app",
      "created_at": "2025-11-15T10:30:00Z",
      "expires_at": "2025-12-15T10:30:00Z",
      "revoked": false
    },
    {
      "id": 43,
      "name": "old-token",
      "application": "my-app",
      "created_at": "2025-10-01T08:00:00Z",
      "expires_at": "2025-11-01T08:00:00Z",
      "revoked": true,
      "revoked_at": "2025-10-15T12:00:00Z"
    }
  ]
}
```

**Note:** The actual PAT value is never returned, only metadata.

**Error Responses:**

- `401 Unauthorized` - Invalid or expired session token
- `403 Forbidden` - Regular user attempting to list another user's tokens

**Example:**

```bash
# List your own tokens
curl https://shen.example.com/api/v1/tokens \
  -H "Authorization: Bearer shen_session_abc123..."

# Admin lists tokens for specific user
curl https://shen.example.com/api/v1/tokens?user=alice \
  -H "Authorization: Bearer shen_session_admin456..."
```

---

## Revocation Behavior

### PAT Revocation

When a PAT is revoked:

1. **Immediate Effect:** The next call to `POST /api/v1/authorize` with this PAT will fail with `401 Unauthorized`
2. **Existing JWTs:** Already-issued JWTs remain valid until expiration (max 7 minutes by default)
3. **Database State:** Token record remains in `shen_tokens` table with `revoked=true` (soft delete, not hard delete)

### Why Not Invalidate Existing JWTs?

Shen uses **stateless JWT verification** - applications verify JWTs using the public key without calling back to Shen. This is a performance optimization.

**Trade-off:**
- ✅ **Performance:** Applications don't need to query Shen for every request
- ✅ **Scalability:** No central bottleneck for token validation
- ⚠️ **Security:** Up to 7 minutes delay before revocation takes full effect

**Mitigation:** The short JWT lifetime (7 minutes) limits the window of exposure.

---

## Session Token Revocation

Session tokens can also be revoked to force users to re-authenticate.

**Endpoint:**
```
DELETE /api/v1/session/:id
```

**Authorization:**
- Admins can revoke any session
- Regular users can revoke their own sessions

**Process:**

Same as PAT revocation but operates on `shen_sessions` table:
- Set `revoked = true`
- Set `revoked_at = NOW()`

**Use Cases:**
- User logs out (revokes their own session)
- Admin forces logout (security incident, account compromise)
- Admin disables account (revokes all sessions for that user)

---

## CLI Commands

```bash
# List your tokens
shenctl token list

# List tokens for specific user (admin only)
shenctl token list --user alice

# Revoke token by ID
shenctl token revoke 42

# Revoke all tokens for a user (admin only)
shenctl token revoke-all <username>

# Cleanup expired tokens (admin only)
shenctl token cleanup

# List active sessions
shenctl session list

# Revoke session by ID
shenctl session revoke 15
```

---

## Security Considerations

### Revoked Token Storage

Revoked tokens are kept in the database (soft delete) for:
- Audit trail
- Forensic analysis
- Preventing token reuse

**Recommendation:** Implement a cleanup job to purge revoked tokens after a retention period (e.g., 90 days).

### Emergency Revocation

For account compromise scenarios, use the revoke-all endpoint or CLI command to immediately invalidate all tokens for a user:

```bash
# Revoke all tokens for compromised user
shenctl token revoke-all alice

# Also revoke all sessions to force re-login
shenctl session revoke-all alice
```

This ensures the user cannot create new JWTs with any existing PATs and must re-authenticate to Shen.

