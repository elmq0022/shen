# Shen Project Context

## Overview

Shen is a centralized authentication and authorization service that issues short-lived JWTs for applications. It manages users, groups, roles, and Personal Access Tokens (PATs).

## Important: Always Reference Documentation

**CRITICAL:** Always check the [docs/](docs/) folder for design decisions, architecture patterns, and implementation details before making changes to the codebase.

## Key Documentation

### Core Architecture
- [docs/010-database.md](docs/010-database.md) - Database design (PostgreSQL, sqlc, golang-migrate)
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
- **github.com/elmq0022/kami** - HTTP router/framework for API server

### Security
- **Argon2id** - Password and PAT hashing (OWASP recommended)
- **RS256** - JWT signing algorithm
- **AES-256-GCM** - Private key encryption (KEK-based)

## Core Design Principles

### 1. "Just Write SQL" Philosophy
- Use sqlc to generate type-safe Go code from SQL
- NO ORMs (GORM, Ent, Bun) - See [docs/010-database.md](docs/010-database.md#L128-L184)
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

## Development Guidelines

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
