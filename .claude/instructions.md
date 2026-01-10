# Shen Project Context

## Overview

Shen is a centralized authentication and authorization service that issues short-lived JWTs for applications. It manages users, groups, roles, and Personal Access Tokens (PATs).

## Important: Always Reference Documentation

**CRITICAL:** Always check the [docs/](docs/) folder for design decisions, architecture patterns, and implementation details before making changes to the codebase.

## Key Documentation

### Core Architecture
- [docs/010-technology-stack.md](docs/010-technology-stack.md) - Technology stack (PostgreSQL, sqlc, golang-migrate, Echo)
- [docs/020-authentication.md](docs/020-authentication.md) - Auth flows and session tokens
- [docs/030-authorization.md](docs/030-authorization.md) - PATs, JWTs, and token exchange
- [docs/060-bootstrap.md](docs/060-bootstrap.md) - Bootstrap process, key generation, KEK encryption
- [docs/090-rbac.md](docs/090-rbac.md) - Multi-role RBAC model and group-based permissions

### Implementation Details
- [docs/040-service-accounts.md](docs/040-service-accounts.md) - Service account design
- [docs/050-token-revocation.md](docs/050-token-revocation.md) - Token revocation patterns
- [docs/070-schema.md](docs/070-schema.md) - Database schema documentation
- [docs/080-cli.md](docs/080-cli.md) - CLI tool design
- [docs/100-future-enhancements.md](docs/100-future-enhancements.md) - Planned features

## Technology Stack (DO NOT CHANGE)

### Go Module
- **Module path**: `github.com/elmq0022/shen` - ALWAYS use this import path for internal packages

### Database & Tooling
- **PostgreSQL** - Primary database (ACID compliance mandatory for auth system)
- **sqlc** - SQL-to-Go code generator (NO ORMs - write actual SQL)
- **golang-migrate** - Versioned database migrations
- **Docker/Docker Compose** - Local development

### Web Framework
- **Echo** (`github.com/labstack/echo/v4`) - HTTP framework for API server
  - Battle-tested and production-ready
  - Rich middleware ecosystem (CORS, rate limiting, logging, recovery)
  - Clean interface that doesn't obscure implementation
  - Security-focused features (automatic escaping, secure headers)
  - Excellent for authentication services requiring robust middleware

#### Error Response Standard
- **ALWAYS** use `handlers.NewErrorResponse(message)` for JSON error responses
- Located in `internal/handlers/errors.go`
- Provides consistent error format: `{"error": "message"}`
- Example: `return c.JSON(http.StatusBadRequest, handlers.NewErrorResponse("Invalid input"))`
- **NEVER** use inline maps like `map[string]string{"error": "..."}`
- If the standard `ErrorResponse` type doesn't fit the use case, **STOP and ask** before creating a different error format

### Security
- **Argon2id** - Password and PAT hashing (OWASP recommended)
- **RS256** - JWT signing algorithm
- **AES-256-GCM** - Private key encryption (KEK-based)

## Core Design Principles

### 1. "Just Write SQL" Philosophy
- Use sqlc to generate type-safe Go code from SQL
- NO ORMs (GORM, Ent, Bun) - See [docs/010-technology-stack.md](docs/010-technology-stack.md#L128-L184)
- NO query builders (squirrel, goqu)
- Complex queries (permission resolution) are clearer in SQL than ORM abstractions
- Explicit queries prevent N+1 problems and hidden performance issues

### 2. ACID Compliance is Mandatory
- Authorization systems cannot tolerate eventual consistency
- Strong consistency guarantees are non-negotiable for security
- All critical operations must be transactional
- Example: Token revocation + audit logging must be atomic

### 3. Security-First Design
- Short-lived JWTs (7 min default) limit exposure window
- PATs serve as refresh tokens (no separate refresh token flow)
- Private keys encrypted at rest with KEK
- Zero-downtime key rotation supported
- All timestamps must be timezone-aware (use `TIMESTAMPTZ`)

### 4. High Availability by Design
- Multiple replicas expected (not optional)
- JWT signing keys stored in database for replica consistency
- Zero-downtime key rotation via database updates
- No filesystem dependencies for HA-critical data

### 5. Explicit Over Implicit
- Knowing exactly what queries run is critical for security and performance
- No ORM "magic" - explicit queries only
- Clear separation of concerns (Shen manages identity, apps handle fine-grained authz)

## Authentication & Authorization Flow

### Session Tokens (Shen UI/CLI)
- User logs in with username/password
- Receives session token (30 day default, configurable via `SHEN_SESSION_EXPIRY_DAYS`)
- Session token used to manage PATs, groups, etc.

### Personal Access Tokens (Application Access)
1. User creates PAT scoped to specific application (via session token)
2. PAT exchanged for short-lived JWT (7 min default, configurable via `SHEN_JWT_SECONDS_TO_EXPIRY`)
3. JWT contains `roles` (array) and `groups` (array)
4. Application verifies JWT using Shen's public key from `/.well-known/jwks.json`

### RBAC Model
- **Multi-role**: Users can have multiple roles per application
- **Standard roles**: `authenticated`, `viewer`, `auditor`, `operator`, `admin`
- **Groups**: Organizational units (e.g., `engineering`, `data-science`)
- **Role assignment**: Groups map to roles per application
- **Fine-grained permissions**: Applications define what each group can do

## Key Storage & Encryption

### Bootstrap Process
1. First startup generates default admin account (`admin`/`admin`)
2. Generates RSA-2048 key pair for JWT signing
3. Private key encrypted with KEK before database storage
4. Public key stored in plaintext (public by design)

### Key Encryption Key (KEK)
- Provided via `SHEN_KEY_ENCRYPTION_KEY` environment variable
- 32-byte random key (base64 encoded)
- Never stored in database
- Required to decrypt private keys at runtime

### Key Rotation
- `shenctl keys rotate` generates new key pair
- Zero-downtime: old keys remain valid for verification
- JWKS endpoint exposes all active verification keys
- Applications automatically use correct key based on JWT `kid` header

## Database Schema Conventions

- Table prefix: `shen_`
- Primary keys: `id SERIAL PRIMARY KEY` or `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- Timestamps: Always use `TIMESTAMPTZ` (timezone-aware)
- Foreign keys: Enforce referential integrity
- Indexes: B-tree for exact matches, partial indexes for filtered queries

### Pagination Patterns: Cursor-Based

**CRITICAL RULE:** NEVER sort pagination queries by `timestamp` or `timestamptz` fields.

#### Sort Field Selection Priority

Choose the sort field in this order:
1. **Unique orderable index** - If a unique constraint exists on a text field (e.g., `username`, `name`), use it
2. **Auto-incrementing ID** - Use `id` (integer SERIAL) for simple cursor-based pagination
3. **ULID/UUID** - Use ULID fields like `kid` (for JWT keys) which are naturally sortable

**Why avoid timestamps for sorting:**
- Timestamps are not unique - multiple records can have identical values
- Clock skew and concurrent inserts cause unpredictable ordering
- Composite cursors `(timestamp, id)` add unnecessary complexity
- Auto-incrementing IDs provide stable, guaranteed ordering

#### Single-Column Cursor (Unique Text Field)

**Use when table has a unique text constraint** (e.g., `username`, `name`):

```sql
-- name: ListUsers :many
SELECT id, username, created_at
FROM shen_user
WHERE
    ($1::text = '' OR username > $1)
ORDER BY username
LIMIT $2;
```

#### Single-Column Cursor (Auto ID)

**Use for tables without a unique orderable text field:**

```sql
-- name: ListSessionsByUser :many
SELECT id, hashed_token, user_id, created_at
FROM shen_session
WHERE
    user_id = sqlc.arg(user_id)
    AND (sqlc.arg(cursor_id) = 0 OR id > sqlc.arg(cursor_id))
ORDER BY id ASC
LIMIT $1;
```

#### Single-Column Cursor (ULID)

**Use for ULID fields** (naturally sortable, time-ordered):

```sql
-- name: ListJWTKeys :many
SELECT id, kid, created_at
FROM shen_jwt_keys
WHERE
    (sqlc.arg(cursor_kid)::text = '' OR kid < sqlc.arg(cursor_kid))
ORDER BY kid DESC
LIMIT $1;
```

#### Composite Cursor (Multiple Unique Fields)

**Only use for multi-column unique constraints** (e.g., composite keys for join tables):

```sql
-- name: ListAllGroupMembers :many
SELECT m.id, g.name AS group_name, u.username AS username
FROM shen_user_group_member m
JOIN shen_user u ON m.user_id = u.id
JOIN shen_group g ON m.group_id = g.id
WHERE
    ($1::text = '' OR g.name > $1 OR (g.name = $1 AND u.username > $2))
ORDER BY g.name, u.username
LIMIT $3;
```

#### Pagination Conventions

1. **Named parameters**: Use `sqlc.arg()` for filter parameters ONLY
2. **Cursor parameters**: Use positional parameters for simplicity (e.g., `$1`, `$2`)
3. **Empty string check**: For text cursors, use `$1::text = ''` to detect initial page
4. **Zero check**: For ID cursors, use `sqlc.arg(cursor_id) = 0` to detect initial page
5. **Comparison operator**: Use `>` for ASC, `<` for DESC
6. **LIMIT**: Always use positional parameter (e.g., `LIMIT $1`)

#### Examples in Codebase

**Auto ID cursor**:
- [db/queries/session.sql](db/queries/session.sql) - `ListSessionsByUser`, `ListActiveSessions`
- [db/queries/token.sql](db/queries/token.sql) - `ListTokensByUser`, `ListActiveTokensByUser`

**Unique text cursor**:
- [db/queries/user.sql](db/queries/user.sql) - `ListUsers` (username)
- [db/queries/group.sql](db/queries/group.sql) - `ListGroups` (name)
- [db/queries/application.sql](db/queries/application.sql) - `ListApplications` (name)

**ULID cursor**:
- [db/queries/jwt_keys.sql](db/queries/jwt_keys.sql) - `ListJWTKeys` (kid)

**Composite cursor**:
- [db/queries/group_member.sql](db/queries/group_member.sql) - `ListAllGroupMembers` (group_name, username)
- [db/queries/group_application_role.sql](db/queries/group_application_role.sql) - Multi-column sorting

## Development Guidelines

### Go Code Style
- **Prefer `any` over `interface{}`** - Use the modern `any` alias introduced in Go 1.18+ for empty interfaces
  - Example: `func process(data any)` instead of `func process(data interface{})`
  - More concise and readable
  - Semantically identical to `interface{}`

### Before Making Changes
1. Read relevant documentation in [docs/](docs/)
2. Understand existing patterns and rationale
3. Check if technology choice is documented (don't change without discussion)
4. Use sqlc for database access (no ORMs)
5. Write actual SQL (no query builders)

### Testing
- Unit tests for business logic
- Integration tests for database queries
- Test timezone handling for all timestamp operations

### Migrations
- Use golang-migrate for all schema changes
- Include both up and down migrations
- Test migrations in isolation before deployment

### Security Checklist
- Use Argon2id for password/token hashing
- Encrypt private keys with KEK before database storage
- Use transactions for multi-step security operations
- Validate expiration timestamps against current time
- Use timezone-aware timestamps (`TIMESTAMPTZ`)

## Configuration via Environment Variables

- `SHEN_ADMIN_USERNAME` - Default admin username (default: `admin`)
- `SHEN_ADMIN_PASSWORD` - Default admin password (default: `admin`)
- `SHEN_SESSION_EXPIRY_DAYS` - Session token TTL (default: 30)
- `SHEN_JWT_SECONDS_TO_EXPIRY` - JWT TTL (default: 420 = 7 min)
- `SHEN_KEY_ENCRYPTION_KEY` - 32-byte KEK for private key encryption (required)

## When in Doubt

1. Check the docs first - the answer is probably there
2. Follow existing patterns in the codebase
3. Write SQL directly, not via ORM or query builder
4. Prioritize security and correctness over convenience
5. Ask for clarification if design rationale isn't documented
