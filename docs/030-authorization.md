# Authorization Flow - Personal Access Tokens (PAT)

## 1. Obtain a Personal Access Token Scoped to an Application

User submits a request to create a PAT with a default expiration of 30 days. The PAT is scoped to a specific application and can be optionally configured with a custom expiration date.

**Endpoint:**
```
POST /api/v1/token/:name/:application
```

**Path Parameters:**
- `:name` - Token identifier (enforced lowercase)
- `:application` - Application name to scope the token to

**Headers:**
```
Authorization: Bearer <session-token>
```

**Query Parameters:**
- `exp` *(optional)* - ISO 8601 datetime in UTC for custom expiration

**Response:**

Shen presents the plaintext PAT to the user **only once**. The hashed value is stored using Argon2id.

```
Status: 200 OK
```

```json
{
    "name": "my-prod-token",
    "pat": "shen_pat_a1b2c3d4e5f6...",
    "exp": "2025-12-15T15:30:00Z" 
}
```

**Error Responses:**
- `401 Unauthorized` - Invalid or expired session token
- `403 Forbidden` - User not authorized for the requested application
- `404 Not Found` - Application not found
- `409 Conflict` - Token name already exists for this user/application combination

---

## 2. Verify Application Access from PAT

User exchanges their PAT for a short-lived JWT containing application-specific permissions.

**Endpoint:**
```
POST /api/v1/authorize
```

**Request Payload:**
```json
{
    "pat": "shen_pat_a1b2c3d4e5f6..."
}
```

**Process:**

1. Shen verifies the hashed token is valid, matches an active token record, and has not expired
2. The application scope is determined from the token record (`application_id`)
3. Verify the application is active (`shen_application.active = true`)
4. User's role is resolved via group memberships (see Role Resolution below)
5. A short-lived JWT is generated and returned

**Short-lived JWT contains:**
- `username` - User identifier
- `aud` - Application name (from the PAT record)
- `exp` - Expiration (420 sec, or 7 min, by default, configurable via `SHEN_JWT_SECONDS_TO_EXPIRY`) - NumericDate (Unix timestamp)
- `roles` - Array of application roles the user has (determined by group memberships)
- `groups` - Array of Shen group names the user belongs to
- `iat` - Issued at timestamp - NumericDate (Unix timestamp)

**Why short-lived tokens?**
Forcing frequent JWT regeneration reduces the window for malicious activity and ensures permission changes (group memberships, role assignments) take effect quickly.

**Response:**

```
Status: 200 OK
```

```json
{
    "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "exp": "2025-12-15T15:30:00Z"
}
```

**Role and Group Resolution:**

When a user authenticates, Shen determines their roles and groups for the application:

**Resolution Process:**
1. Lookup all active groups the user belongs to (via `shen_user_group_member`)
2. For each group, find all roles assigned for the target application (via `shen_group_application_role`)
3. Collect the deduplicated union of all roles across all groups
4. Include all role names and all group names in the JWT

**Example:**
- User `alice` is in groups: `engineering`, `data-science`
- `engineering` group → roles: `[viewer, operator]` for `api-server`
- `data-science` group → roles: `[operator]` for `api-server`
- **JWT roles:** `["viewer", "operator"]` (deduplicated)
- **JWT groups:** `["engineering", "data-science"]` (group names)

**Authorization Model:**
- **Roles** determine UI/endpoint access (coarse-grained)
- **Groups** enable fine-grained permissions (applications define what each group can do)

**Error Responses:**
- `401 Unauthorized` - Invalid, expired, or revoked PAT
- `403 Forbidden` - User no longer has access to the application (not in any active groups)
- `404 Not Found` - Application not found or inactive

---

## 3. Request Application Access

Once the client has a valid short-lived JWT, it can access the target application.

**How Applications Verify JWTs:**

1. **Fetch Shen's Public Key:**

Applications decode the JWT using Shen's public key available at:

```
GET https://shen.example.com/.well-known/jwks.json
```

**Response:**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "2024-12-15",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

2. **Send JWT to Application:**

```
GET https://my-app.example.com/api/resource
Authorization: Bearer <short-lived-jwt>
```

3. **Application Verifies:**

The application must verify the following JWT claims:
- `sig` - Signature is valid (using Shen's public key)
- `username` - User identifier
- `aud` - Audience matches the application name
- `exp` - Token has not expired
- `iat` - Issued at timestamp
- `roles` - Array of user's application roles for coarse-grained authorization
- `groups` - Array of user's group memberships for fine-grained authorization

**Authorization in Applications:**

Applications use the JWT claims to make authorization decisions:

```go
// Role-based authorization (coarse-grained)
if contains(jwt.Roles, "operator") {
    // Show operator dashboard
}

// Group-based authorization (fine-grained)
// Application defines what each group can do
if contains(jwt.Groups, "engineering") {
    // Grant engineering team permissions
}
```

**Token Expiration:**

If the JWT has expired, the user or client must resubmit the PAT to `/api/v1/authorize` to obtain a new short-lived JWT. Since the PAT is already long-lived, there's no need for a refresh token.
