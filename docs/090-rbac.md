# RBAC Design

Shen implements a **multi-role RBAC model** where users can have multiple application roles, and applications interpret group membership for fine-grained permissions.

## Architecture Overview

### Two Dimensions of Authorization

1. **Roles** - Coarse-grained access levels for UI/endpoint access
   - Multiple roles per user per application
   - Standard set of predefined roles: `authenticated`, `viewer`, `auditor`, `operator`, `admin`
   - Managed by Shen
   - Included in JWT `roles` claim as an array

2. **Groups** - User's Shen group memberships
   - Organizational groups (e.g., `engineering`, `data-science`, `ops`)
   - Managed by Shen (user membership, group-to-role mappings)
   - Included in JWT `groups` claim as an array
   - **Applications define what each group can do** - Shen just passes the group names

### JWT Structure

When a user authenticates to an application, the JWT contains:

```json
{
  "sub": "user-uuid",
  "app": "application-name",
  "roles": ["operator", "viewer"],
  "groups": ["engineering", "data-science"],
  "iat": 1234567890,
  "exp": 1234567890
}
```

## Available Application Roles

These are the standard roles available for all applications. Custom roles are not supported.

- **authenticated** - Authentication only. User identity is verified by Shen, but the application handles its own authorization logic. Useful for applications with custom permission systems or during migration to Shen-managed RBAC.
- **viewer** - Read-only access
- **auditor** - Audit and compliance access
- **operator** - Operational management
- **admin** - Full administrative access

### Multi-Role Assignment

Users can have multiple roles for an application based on their group memberships. For example:
- A user in both `engineering` and `security` groups might get `operator` + `auditor` roles
- A user in `ops` group might get only `operator` role
- Applications check for specific roles to determine UI/endpoint access

## Groups

Groups are organizational units in Shen (e.g., `engineering`, `data-science`, `finance`).

### How Groups Work

1. **Shen manages:**
   - Group definitions
   - User membership in groups
   - Which roles each group grants per application

2. **Applications interpret:**
   - What permissions each group has
   - What actions group members can perform
   - Fine-grained authorization logic

### Group Membership in JWT

The JWT includes the names of all Shen groups the user belongs to. The application uses these group names to make authorization decisions based on its own permission model.

## Authorization Model

### How Shen Assigns Roles and Groups

1. **Users** are assigned to **Shen groups** (e.g., `engineering`, `data-science`)
2. Each **Shen group** is mapped to **application roles** for each application
3. When a user authenticates to an application:
   - Shen collects all groups the user belongs to
   - Shen collects all roles granted by those groups for that application (deduplicated)
   - Both are included in the JWT

### Example

**Setup in Shen:**
- User `alice` belongs to groups: `engineering`, `data-science`
- Group `engineering` for app `idp` → roles: `[viewer, operator]`
- Group `data-science` for app `idp` → roles: `[operator]`

**Alice's JWT for `idp`:**
```json
{
  "sub": "alice-uuid",
  "app": "idp",
  "roles": ["viewer", "operator"],          // Deduplicated union from both groups
  "groups": ["engineering", "data-science"] // Just the group names
}
```

**IDP application defines what each group can do:**
```go
// In IDP's configuration/code
var permissions = map[string][]string{
    "engineering":   {"can_deploy_model", "can_run_ab_test"},
    "data-science":  {"can_deploy_model", "can_view_datasets"},
}

func (h *Handler) DeployModel(w http.ResponseWriter, r *http.Request) {
    jwt := parseJWT(r)

    // Check if user's groups have this permission
    allowed := false
    for _, group := range jwt.Groups {
        if perms, ok := permissions[group]; ok {
            if contains(perms, "can_deploy_model") {
                allowed = true
                break
            }
        }
    }

    if !allowed {
        return http.StatusForbidden
    }

    // Deploy the model
}
```

### Application Authorization Patterns

**Role-based checks (for UI/endpoint access):**
```go
// Check if user has specific role
if contains(jwt.Roles, "operator") {
    // Show operator dashboard/endpoints
}

// Check if user has any of several roles
if hasAny(jwt.Roles, []string{"operator", "admin"}) {
    // Allow operational actions
}
```

**Group-based checks (for fine-grained permissions):**
```go
// Application defines its own group permissions
allowedGroups := []string{"engineering", "data-science"}
if hasAny(jwt.Groups, allowedGroups) {
    // Allow specific action
}

// Or use a permission mapping
if hasPermission(jwt.Groups, "can_deploy_model") {
    // Allow deployment
}
```

**Combined checks:**
```go
// Require both role and group membership
if contains(jwt.Roles, "operator") && contains(jwt.Groups, "engineering") {
    // Allow production deployment
}
```

## Design Rationale

### Why Multi-Role Instead of Single Highest Role?

A user might belong to multiple groups that grant different roles:
- `security` group → `auditor` role (for compliance dashboard)
- `engineering` group → `operator` role (for operations dashboard)

Multi-role allows the application to show both dashboards/endpoints when the user has both roles. This is cleaner than trying to model orthogonal concerns with a single hierarchical role.

### Why Pass Group Names Instead of Expanded Permissions?

1. **Separation of concerns** - Shen handles identity and group membership; applications handle authorization logic
2. **Flexibility** - Applications can change what groups mean without touching Shen
3. **Simplicity** - Shen doesn't need to know about application-specific permissions
4. **JWT size** - Group names are typically short; expanded permissions could be numerous
5. **Security** - Applications maintain control over their permission model

### Why Fixed Role Set?

1. **Simplicity** - Five roles cover most common access patterns (authenticated, read, audit, operate, admin)
2. **Consistency** - Same role names mean the same thing across all applications
3. **UI paradigm** - Roles map to broad UI access levels; groups handle fine-grained permissions
4. **Scope management** - Limited role set keeps the model simple

### Why Per-Application Role Mappings?

Different applications may have different security requirements:
- `engineering` group might be `admin` for internal tools
- `engineering` group might be `operator` for production systems
- Each application gets to decide what access each group should have

## Summary

**Shen is a token vending machine:**
- ✅ Authenticates users
- ✅ Manages group membership
- ✅ Maps groups to roles per application
- ✅ Issues JWTs with roles and group names
- ❌ Does NOT define fine-grained permissions
- ❌ Does NOT enforce application-specific authorization

**Applications using Shen:**
- ✅ Receive JWTs with roles and groups
- ✅ Define what each group can do
- ✅ Implement their own permission model
- ✅ Enforce authorization based on roles and groups
