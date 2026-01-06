# Shen Development Milestones

This document tracks the implementation milestones for the Shen authentication and authorization service. Each milestone represents a functional increment that can be tested end-to-end.

---

## Milestone 1: Bootstrap & Admin Authentication

**Goal:** Set up the foundation - database, bootstrap process, and admin login via CLI.

### Server/Backend
- [x] Database schema migrations (users, sessions, applications, groups, tokens, JWT keys)
- [x] Bootstrap process - create default admin account on first startup
- [x] Bootstrap process - generate RSA key pair for JWT signing
- [x] Store JWT keys in database (PEM format, encrypted with KEK)
- [x] Password hashing with Argon2id
- [ ] `POST /api/v1/auth/login` endpoint - username/password authentication
- [ ] Session token generation (SHA-256 hashed, database-backed)
- [ ] Session token validation middleware
- [ ] `GET /.well-known/jwks.json` endpoint - expose public keys in JWK format

### CLI (shenctl)
- [ ] `shenctl auth login` - authenticate with username/password
- [ ] Store session token in local config file (`~/.shen/config` or similar)
- [ ] Load session token from config for authenticated requests
- [ ] `shenctl config show` - display current configuration
- [ ] `shenctl config set` - set configuration values
- [ ] `shenctl config del` - delete configuration keys

### Testing
- [ ] Integration test: bootstrap creates admin user and JWT keys
- [ ] Integration test: login with default admin credentials
- [ ] Integration test: session token stored and reused by CLI
- [ ] Integration test: invalid credentials return 401
- [ ] Integration test: JWKS endpoint returns valid JWK

---

## Milestone 2: User Management (Admin CRUD)

**Goal:** Admin can manage users (create, list, update, delete) via CLI.

### Server/Backend
- [ ] `GET /api/v1/users` - list all users (admin only)
- [ ] `POST /api/v1/users` - create new user (admin only)
- [ ] `PATCH /api/v1/users/:username` - update user role (admin only)
- [ ] `DELETE /api/v1/users/:username` - soft delete user (admin only)
- [ ] Authorization middleware - check user role for admin endpoints
- [ ] Validate user roles: service, user, admin

### CLI (shenctl)
- [ ] `shenctl user list` - list all users
- [ ] `shenctl user create <username> <role>` - create new user
- [ ] `shenctl user update <username> <role>` - update user role
- [ ] `shenctl user delete <username>` - soft delete user

### Testing
- [ ] Integration test: admin creates new user
- [ ] Integration test: admin lists users
- [ ] Integration test: admin updates user role
- [ ] Integration test: admin deletes user
- [ ] Integration test: non-admin cannot access user management endpoints
- [ ] Integration test: service accounts cannot login to Shen

---

## Milestone 3: Application Management

**Goal:** Admin can register and manage applications via CLI.

### Server/Backend
- [ ] `GET /api/v1/applications` - list all applications (admin only)
- [ ] `POST /api/v1/applications` - create new application (admin only)
- [ ] `DELETE /api/v1/applications/:name` - soft delete application (admin only)
- [ ] Validate application names (lowercase enforcement)

### CLI (shenctl)
- [ ] `shenctl application list` - list all applications
- [ ] `shenctl application create <name>` - create new application
- [ ] `shenctl application delete <name>` - soft delete application

### Testing
- [ ] Integration test: admin creates application
- [ ] Integration test: admin lists applications
- [ ] Integration test: admin deletes application
- [ ] Integration test: application names are normalized to lowercase

---

## Milestone 4: Group Management

**Goal:** Admin can create groups and manage group memberships via CLI.

### Server/Backend
- [ ] `GET /api/v1/groups` - list all groups (admin only)
- [ ] `POST /api/v1/groups` - create new group (admin only)
- [ ] `DELETE /api/v1/groups/:name` - delete group (admin only)
- [ ] `POST /api/v1/groups/:name/members` - add users to group (admin or group manager)
- [ ] `DELETE /api/v1/groups/:name/members` - remove users from group (admin or group manager)
- [ ] `GET /api/v1/groups/:name/members` - list group members
- [ ] Authorization middleware - check admin or group manager privileges

### CLI (shenctl)
- [ ] `shenctl group list` - list all groups
- [ ] `shenctl group create <name>` - create new group
- [ ] `shenctl group delete <name>` - delete group
- [ ] `shenctl group add-users <group> <user1> <user2> ...` - add users to group
- [ ] `shenctl group remove-users <group> <user1> <user2> ...` - remove users from group
- [ ] `shenctl user add-groups <username> <group1> <group2> ...` - add user to groups

### Testing
- [ ] Integration test: admin creates group
- [ ] Integration test: admin adds users to group
- [ ] Integration test: admin removes users from group
- [ ] Integration test: list group members
- [ ] Integration test: delete group removes all memberships

---

## Milestone 5: RBAC - Group Role Assignments

**Goal:** Admin can assign application roles to groups, enabling RBAC.

### Server/Backend
- [ ] Seed application roles in database (authenticated, viewer, auditor, operator, admin)
- [ ] `POST /api/v1/groups/:name/roles` - assign role to group for application (admin only)
- [ ] `DELETE /api/v1/groups/:name/roles` - remove role from group for application (admin only)
- [ ] `GET /api/v1/groups/:name/roles` - list roles for group (optionally filtered by application)
- [ ] Validate application roles against seeded values
- [ ] Role priority logic (highest priority wins when user is in multiple groups)

### CLI (shenctl)
- [ ] `shenctl group add-role <group> <application> <role>` - assign role to group
- [ ] `shenctl group remove-role <group> <application> <role>` - remove role from group
- [ ] `shenctl group list-roles <group> [application]` - list roles for group

### Testing
- [ ] Integration test: admin assigns role to group for application
- [ ] Integration test: admin removes role from group
- [ ] Integration test: list roles for group
- [ ] Integration test: list roles filtered by application
- [ ] Integration test: validate role priority resolution (highest wins)

---

## Milestone 6: Personal Access Tokens (PAT)

**Goal:** Users can create PATs for applications and exchange them for JWTs.

### Server/Backend
- [ ] `POST /api/v1/token/:name/:application` - create PAT (user or admin)
- [ ] `GET /api/v1/tokens` - list tokens for authenticated user
- [ ] `GET /api/v1/tokens?user=<username>` - list tokens for specific user (admin only)
- [ ] PAT generation (cryptographically secure random, 32 bytes)
- [ ] PAT hashing with Argon2id before storage
- [ ] Token expiration validation (default 30 days, configurable)
- [ ] `POST /api/v1/authorize` - exchange PAT for short-lived JWT
- [ ] Role resolution via group memberships
- [ ] JWT generation with claims: username, aud, exp, roles, groups, iat
- [ ] JWT signing with RSA private key from database
- [ ] Validate PAT is not expired or revoked
- [ ] Validate application is active

### CLI (shenctl)
- [ ] `shenctl token list` - list your own tokens
- [ ] `shenctl token list --user <username>` - list tokens for specific user (admin only)
- [ ] `shenctl token create <name> <application>` - create token for yourself
- [ ] `shenctl token create <name> <application> <user>` - create token for specific user (admin only)

### Testing
- [ ] Integration test: user creates PAT for application
- [ ] Integration test: PAT is returned only once (plaintext)
- [ ] Integration test: PAT is hashed before database storage
- [ ] Integration test: exchange PAT for JWT
- [ ] Integration test: JWT contains correct claims (username, roles, groups, aud, exp)
- [ ] Integration test: JWT is signed with RSA key
- [ ] Integration test: role resolution - user in multiple groups gets highest priority role
- [ ] Integration test: expired PAT cannot be exchanged
- [ ] Integration test: PAT for inactive application returns 404
- [ ] Integration test: user without group membership cannot get JWT

---

## Milestone 7: Token & Session Revocation

**Goal:** Admin can revoke tokens and sessions for security management.

### Server/Backend
- [ ] `DELETE /api/v1/tokens/:id` - revoke specific token by ID
- [ ] `DELETE /api/v1/tokens?user=<username>` - revoke all tokens for user (admin only)
- [ ] `DELETE /api/v1/tokens/cleanup` - remove expired and revoked tokens (admin only)
- [ ] `DELETE /api/v1/sessions/:id` - revoke specific session by ID
- [ ] `DELETE /api/v1/sessions?user=<username>` - revoke all sessions for user (admin only)
- [ ] Validate authorization for revocation (own tokens/sessions or admin)

### CLI (shenctl)
- [ ] `shenctl token revoke <id>` - revoke specific token
- [ ] `shenctl token revoke-all <username>` - revoke all tokens for user (admin only)
- [ ] `shenctl token cleanup` - remove expired and revoked tokens (admin only)
- [ ] `shenctl session list` - list your own sessions
- [ ] `shenctl session list --user <username>` - list sessions for user (admin only)
- [ ] `shenctl session revoke <id>` - revoke specific session
- [ ] `shenctl session revoke-all <username>` - revoke all sessions for user (admin only)

### Testing
- [ ] Integration test: user revokes own token
- [ ] Integration test: admin revokes token for other user
- [ ] Integration test: revoked token cannot be exchanged for JWT
- [ ] Integration test: admin revokes all tokens for user
- [ ] Integration test: cleanup removes expired and revoked tokens
- [ ] Integration test: user revokes own session
- [ ] Integration test: admin revokes session for other user
- [ ] Integration test: revoked session cannot access protected endpoints

---

## Milestone 8: Service Accounts

**Goal:** Support token-only service accounts for automated systems.

### Server/Backend
- [ ] Validate service accounts cannot login (`POST /api/v1/auth/login` returns 403)
- [ ] Validate service accounts cannot access Shen management API (403 for all endpoints)
- [ ] Service accounts can be added to groups (existing endpoint)
- [ ] Service accounts can have tokens created by admin (existing endpoint)
- [ ] Service accounts can exchange PAT for JWT (existing endpoint)

### CLI (shenctl)
- [ ] Ensure `shenctl user create <name> service` creates service account (no password)
- [ ] Ensure `shenctl token create <name> <app> <service-account>` works for service accounts

### Testing
- [ ] Integration test: service account cannot login
- [ ] Integration test: service account cannot access user management endpoints
- [ ] Integration test: admin creates service account
- [ ] Integration test: admin adds service account to group
- [ ] Integration test: admin creates token for service account
- [ ] Integration test: service account PAT can be exchanged for JWT
- [ ] Integration test: service account JWT contains correct roles from groups

---

## Milestone 9: Group Managers

**Goal:** Enable delegated group management via group managers.

### Server/Backend
- [ ] `POST /api/v1/groups/:name/managers` - assign group managers (admin only)
- [ ] `DELETE /api/v1/groups/:name/managers` - remove group managers (admin only)
- [ ] `GET /api/v1/groups/:name/managers` - list group managers
- [ ] Authorization middleware - allow group managers to add/remove members
- [ ] Validate group managers cannot assign roles (admin only)

### CLI (shenctl)
- [ ] `shenctl group add-managers <group> <user1> <user2> ...` - assign group managers
- [ ] `shenctl group remove-managers <group> <user1> ...` - remove group managers
- [ ] `shenctl group list-managers <group>` - list group managers

### Testing
- [ ] Integration test: admin assigns group manager
- [ ] Integration test: group manager adds users to their group
- [ ] Integration test: group manager removes users from their group
- [ ] Integration test: group manager cannot assign roles to group
- [ ] Integration test: group manager cannot manage other groups
- [ ] Integration test: list group managers

---

## Milestone 10: JWT Key Rotation

**Goal:** Support zero-downtime JWT key rotation for security.

### Server/Backend
- [ ] Key rotation endpoint - generate new RSA key pair
- [ ] Encrypt new private key with KEK before storage
- [ ] Mark old key as `active_for_signing=false`, keep `active_for_verification=true`
- [ ] JWKS endpoint returns all active verification keys
- [ ] JWT signing uses newest `active_for_signing=true` key
- [ ] JWT verification supports multiple keys (via `kid` header)

### CLI (shenctl)
- [ ] `shenctl keys rotate` - rotate JWT signing keys (admin only)
- [ ] `shenctl keys list` - list all keys and their status

### Testing
- [ ] Integration test: rotate JWT keys
- [ ] Integration test: new JWTs signed with new key
- [ ] Integration test: old JWTs still verified with old key
- [ ] Integration test: JWKS endpoint returns multiple keys after rotation
- [ ] Integration test: zero downtime during rotation (no failed JWT verifications)

---

## Milestone 11: Security Hardening

**Goal:** Production-ready security features and validation.

### Server/Backend
- [ ] Rate limiting on login endpoint (prevent brute force)
- [ ] Account lockout after N failed login attempts
- [ ] Password complexity validation
- [ ] HTTPS/TLS enforcement (reject HTTP in production)
- [ ] CORS configuration
- [ ] Security headers (Content-Security-Policy, X-Frame-Options, etc.)
- [ ] Input validation and sanitization
- [ ] SQL injection prevention (verify sqlc query safety)
- [ ] Audit logging for sensitive operations (user creation, role changes, token creation)

### Testing
- [ ] Security test: rate limiting prevents brute force
- [ ] Security test: account lockout after failed attempts
- [ ] Security test: weak passwords rejected
- [ ] Security test: SQL injection attempts blocked
- [ ] Security test: XSS attempts sanitized
- [ ] Security test: CORS policy enforced

---

## Milestone 12: End-to-End Integration

**Goal:** Complete end-to-end workflow testing with real applications.

### Sample Application
- [ ] Build sample application that integrates with Shen
- [ ] Fetch JWKS from Shen (`GET /.well-known/jwks.json`)
- [ ] Verify JWT signature using Shen's public key
- [ ] Validate JWT claims (aud, exp, roles, groups)
- [ ] Implement role-based authorization in sample app
- [ ] Implement group-based authorization in sample app

### End-to-End Tests
- [ ] E2E test: admin creates user, group, application
- [ ] E2E test: admin assigns role to group for application
- [ ] E2E test: admin adds user to group
- [ ] E2E test: user creates PAT for application
- [ ] E2E test: user exchanges PAT for JWT
- [ ] E2E test: sample app verifies JWT and grants access
- [ ] E2E test: role change propagates to new JWT within 7 minutes
- [ ] E2E test: revoked PAT cannot generate new JWT
- [ ] E2E test: service account workflow (create, assign groups, create token, access app)

---

## Future Enhancements

*These are documented in [docs/100-future-enhancements.md](docs/100-future-enhancements.md) and are not part of the initial release.*

- [ ] Token lifecycle notifications (expiration reminders, security alerts)
- [ ] Webhook support for token events
- [ ] Multi-factor authentication (MFA)
- [ ] OAuth 2.0 / OpenID Connect provider support
- [ ] Advanced audit logging with log aggregation
- [ ] Grafana/Prometheus metrics
- [ ] Web UI for administration (alternative to CLI)
