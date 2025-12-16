# DESIGN AND REQUIREMENTS

## Database Selection and Tooling

- **PostgreSQL** - Primary database
- **Docker / Docker Compose** - Local development environment
- **golang-migrate** - Versioned database migrations
- **sqlc** - Auto-generating Go code from SQL queries


## Authorization Flow for Users Personal Access Tokens (PAT)

### 1. Verify Identity with Username and Password

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
- 1 month expiration (configurable via `SHEN_SESSION_EXPIRY_SECONDS`)
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

---

### 2. Obtain a Personal Access Token Scoped to an Application

User submits a request to create a PAT with a default expiration of 1 month. The PAT is scoped to a specific application and can be optionally configured with a custom expiration date.

**Endpoint:**
```
POST /api/v1/token/:name/:app
```

**Path Parameters:**
- `:name` - Token identifier (enforced lowercase)
- `:app` - Application name to scope the token to

**Headers:**
```
Authorization: Bearer <session-token>
```

**Query Parameters:**
- `exp` *(optional)* - ISO 8601 datetime in UTC for custom expiration

**Response:**

Shen presents the plaintext PAT to the user **only once**. The hashed value is stored using bcrypt or argon2.

```
Status: 200 OK
```

```json
{
    "name": "my-prod-token",
    "pat": "shen_pat_a1b2c3d4e5f6..."
}
```

**Error Responses:**
- `401 Unauthorized` - Invalid or expired session token
- `403 Forbidden` - User not authorized for the requested application
- `404 Not Found` - Application not found
- `409 Conflict` - Token name already exists for this user/application combination

---

### 3. Verify Application Access from PAT

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
2. The application scope is determined from the token record (`application_fk`)
3. User's role is resolved via group memberships (see Role Resolution below)
4. A short-lived JWT is generated and returned

**Short-lived JWT contains:**
- `username` - User identifier
- `aud` - Application name (from the PAT record)
- `exp` - Expiration (7 minutes by default, configurable via `SHEN_JWT_SECONDS_TO_EXPIRY`)
- `role` - Effective application role (determined by group memberships)

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

**Role Resolution:**

When a user is a member of multiple groups with different roles for the same application, Shen determines the effective role by selecting the **highest priority** role across all groups.

**Resolution Process:**
1. Lookup all groups the user belongs to (via `shen_user_group`)
2. For each group, find the role assigned for the target application (via `shen_group_application_role`)
3. Select the role with the highest priority value (via `shen_application_role.priority`)
4. Include this role in the JWT

**Example:**
- User is in `developers` group → role: `viewer` (priority: 100)
- User is in `leads` group → role: `operator` (priority: 300)
- **Effective role:** `operator` (higher priority wins)

**Error Responses:**
- `401 Unauthorized` - Invalid, expired, or revoked PAT
- `403 Forbidden` - User no longer has access to the application (not in any groups with roles for this app)
- `404 Not Found` - Application not found

---

### 4. Request Application Access

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
- `role` - User's RBAC role for authorization decisions

**Token Expiration:**

If the JWT has expired, the user or client must resubmit the PAT to `/api/v1/authorize` to obtain a new short-lived JWT. 


## Service Accounts

Service accounts are user records with `role=service` that exist solely for token-based application access. They cannot authenticate to Shen's management API and have no password.

### Characteristics

- **No Shen Access:** Service accounts cannot login to Shen or use any Shen management endpoints
- **No Password:** Service accounts have `hashed_password = NULL` and cannot use `/api/v1/auth/login`
- **Token-Only:** Tokens are created for service accounts by administrators
- **Group-Based Permissions:** Service accounts are added to groups and inherit application permissions the same way human users do

### Service Account Workflow

**1. Administrator creates service account:**
```bash
shenctl user create ci-deploy service
# Creates user with role=service, no password
```

**2. Administrator adds service account to groups:**
```bash
shenctl user add-groups ci-deploy ops-team
# Service account now has permissions via ops-team's group roles
```

**3. Administrator creates token for service account:**
```bash
shenctl token create deploy-token my-app ci-deploy
# Token is created and linked to ci-deploy service account
```

**4. Service account uses token to access applications:**
- Token is used with `POST /api/v1/authorize` (same as user PATs)
- Permissions derived from group memberships
- Short-lived JWT returned with application role

### Authorization Logic

When any request is made to Shen's management API:
```
if user.role == "service":
    return 403 Forbidden
```

Service accounts are blocked from all Shen management operations but can have tokens used for application access via the `/api/v1/authorize` endpoint.


## Token Revocation

TODO: Design token revocation API and workflow. Key questions to address:
- API endpoint for token revocation (e.g., `DELETE /api/v1/token/:name`)
- Can users revoke their own tokens or admin only?
- How are revoked long-lived JWTs tracked?
- Should there be a token revocation list or blacklist?


## Initial Bootstrap and Setup

### Default Admin Account

On first startup, if no users exist in the database, Shen will automatically create a default admin account:

**Default credentials:**
- Username: `admin`
- Password: `admin`

**Security Warning:** Change these credentials immediately after first login.

**Configuration via Environment Variables:**
- `SHEN_ADMIN_USERNAME` - Override default admin username (default: `admin`)
- `SHEN_ADMIN_PASSWORD` - Override default admin password (default: `admin`)

### Public/Private Key Generation

On first startup, if no JWT signing keys exist, Shen will automatically generate an RSA key pair:
- Private key: Used to sign JWTs
- Public key: Exposed via `/.well-known/jwks.json` for applications to verify JWTs

Keys are stored securely and can be rotated by administrators.

### Database Seeding

On startup, Shen will seed the following reference data if not present:

**User Roles** (`shen_user_roles`):
- `service`
- `user`
- `admin`

**Application Roles** (`shen_application_role`):
- `authenticated` (priority: 0) - Authentication only, no Shen-managed authorization
- `viewer` (priority: 100)
- `auditor` (priority: 200)
- `operator` (priority: 300)
- `admin` (priority: 400)


## Schema Design

### Core Tables

#### `shen_user`

| Field           | Type      | Unique | Index | Description                                          |
|:----------------|:----------|:-------|:------|:-----------------------------------------------------|
| pk              | PK        | Y      | -     | Primary key                                          |
| username        | string    | Y      | Y     | User identifier (enforced lowercase)                 |
| hashed_password | string    | N      | N     | Hashed password (nullable - NULL for service accounts)|
| active          | bool      | N      | N     | Account active status                                |
| role            | FK        | N      | Y     | Foreign key to `shen_user_roles`                     |
| created_at      | timestamp | N      | N     | User creation timestamp                              |
| updated_at      | timestamp | N      | N     | User last update timestamp                           |

**Important:** Service accounts (role=`service`) must have `hashed_password = NULL`. These accounts cannot authenticate to Shen's management API.

#### `shen_user_roles`

| Field      | Type      | Unique | Index | Description                         |
|:-----------|:----------|:-------|:------|:------------------------------------|
| pk         | PK        | Y      | -     | Primary key                         |
| name       | string    | Y      | N     | Role name (enforced lowercase)      |
| created_at | timestamp | N      | N     | Role creation timestamp             |
| updated_at | timestamp | N      | N     | Role last update timestamp          |

**Available roles:**
- `service` - Service account, cannot login to Shen, token-only access
- `user` - Regular user, can manage own PATs and view own groups
- `admin` - Administrator, can manage all Shen resources

#### `shen_group`

| Field      | Type      | Unique | Index | Description                         |
|:-----------|:----------|:-------|:------|:------------------------------------|
| pk         | PK        | Y      | -     | Primary key                         |
| name       | string    | Y      | Y     | Group name (enforced lowercase)     |
| created_at | timestamp | N      | N     | Group creation timestamp            |
| updated_at | timestamp | N      | N     | Group last update timestamp         |

#### `shen_user_group`

| Field      | Type      | Unique | Index | Description                      |
|:-----------|:----------|:-------|:------|:---------------------------------|
| pk         | PK        | Y      | -     | Primary key                      |
| user_fk    | FK        | N      | Y     | Foreign key to `shen_user`       |
| group_fk   | FK        | N      | Y     | Foreign key to `shen_group`      |
| created_at | timestamp | N      | N     | Assignment creation timestamp    |
| updated_at | timestamp | N      | N     | Assignment last update timestamp |

**Composite unique constraint:** `(user_fk, group_fk)` - A user can only be assigned to a group once

#### `shen_application`

| Field      | Type      | Unique | Index | Description                            |
|:-----------|:----------|:-------|:------|:---------------------------------------|
| pk         | PK        | Y      | -     | Primary key                            |
| name       | string    | Y      | Y     | Application name (enforced lowercase)  |
| created_at | timestamp | N      | N     | Application creation timestamp         |
| updated_at | timestamp | N      | N     | Application last update timestamp      |

#### `shen_application_role`

| Field      | Type      | Unique | Index | Description                                 |
|:-----------|:----------|:-------|:------|:--------------------------------------------|
| pk         | PK        | Y      | -     | Primary key                                 |
| priority   | integer   | N      | Y     | Role priority                               |
| name       | string    | Y      | N     | Role name (enforced lowercase)              |
| created_at | timestamp | N      | N     | Application role creation timestamp         |
| updated_at | timestamp | N      | N     | Application role last update timestamp      |

**Available roles:** `authenticated`, `viewer`, `auditor`, `operator`, `admin`

#### `shen_group_application_role`

| Field               | Type      | Unique | Index | Description                           |
|:--------------------|:----------|:-------|:------|:--------------------------------------|
| pk                  | PK        | Y      | -     | Primary key                           |
| group_fk            | FK        | N      | Y     | Foreign key to `shen_group`           |
| application_fk      | FK        | N      | Y     | Foreign key to `shen_application`     |
| application_role_fk | FK        | N      | Y     | Foreign key to `shen_application_role`|
| created_at          | timestamp | N      | N     | Assignment creation timestamp         |
| updated_at          | timestamp | N      | N     | Assignment last update timestamp      |

**Composite unique constraint:** `(group_fk, application_fk)` - A group can only have one specific role per application

#### `shen_tokens`

| Field          | Type      | Unique | Index | Description                                       |
|:---------------|:----------|:-------|:------|:--------------------------------------------------|
| pk             | PK        | Y      | -     | Primary key                                       |
| name           | string    | N      | Y     | Token name/identifier (enforced lowercase)        |
| token          | string    | Y      | Y     | Hashed token value                                |
| user_fk        | FK        | N      | Y     | Foreign key to `shen_user` (nullable)             |
| application_fk | FK        | N      | Y     | Foreign key to `shen_application` (nullable)      |
| created_at     | timestamp | N      | N     | Token creation timestamp                          |
| expires_at     | timestamp | N      | N     | Token expiration timestamp                        |
| revoked        | bool      | N      | N     | Token revocation status                           |
| revoked_at     | timestamp | N      | N     | Token revocation timestamp (nullable)             |

**Composite unique constraint:** `(user_fk, application_fk, name)` - A user can only have one token with the same name per application

#### `shen_sessions`

| Field          | Type      | Unique | Index | Description                                       |
|:---------------|:----------|:-------|:------|:--------------------------------------------------|
| pk             | PK        | Y      | -     | Primary key                                       |
| token          | string    | Y      | Y     | Hashed session token value                        |
| user_fk        | FK        | N      | Y     | Foreign key to `shen_user`                        |
| created_at     | timestamp | N      | N     | Session creation timestamp                        |
| expires_at     | timestamp | N      | N     | Session expiration timestamp                      |
| revoked        | bool      | N      | N     | Session revocation status                         |
| revoked_at     | timestamp | N      | N     | Session revocation timestamp (nullable)           |

This table stores session tokens used for authenticating users to the Shen management API (not application PATs).


## CLI Design

The CLI program is named **`shenctl`** and will display help information when invoked without arguments.

### Configuration Management

```bash
shenctl config show              # Display current configuration
shenctl config set key=value     # Set a configuration value
shenctl config del key           # Delete a configuration key
```

### User Management

```bash
shenctl user list                                           # List all users
shenctl user create <user-name> <role>                      # Create a new user
                                                            # Role: service-account, user, or admin
shenctl user update <user-name> <role>                      # Update user role
shenctl user delete <user-name>                             # Soft delete (mark as inactive)
shenctl user add-groups <user-name> <group1> <group2> ...   # Add user to groups
```

### Group Management

```bash
shenctl group list                                          # List all groups
shenctl group create <group-name>                           # Create a new group
shenctl group delete <group-name>                           # Delete a group
shenctl group add-users <group-name> <user1> <user2> ...    # Add users to group
shenctl group assign-role <group-name> <app1=rbac1> ...     # Assign roles to group
```

### Application Management

```bash
shenctl app list                # List all applications
shenctl app create <app-name>   # Create a new application
shenctl app delete <app-name>   # Soft delete (mark as inactive)
```

### Token Management

```bash
shenctl token list [user]                    # List tokens (user optional, admin only)
shenctl token create <token-name> <app> [user]  # Create token (user optional, admin only)
shenctl token revoke <token-name>            # Revoke a token
```

## RBAC Roles

Custom RBAC roles are not supported in the initial implementation to help narrow the project scope.

### Available Roles

- **authenticated** - Authentication only. User identity is verified by Shen, but the application handles its own authorization logic. Useful for applications with custom permission systems or during migration to Shen-managed RBAC.
- **viewer** - Read-only access
- **auditor** - Audit and compliance access
- **operator** - Operational management
- **admin** - Full administrative access
