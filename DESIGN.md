# DESIGN AND REQUIREMENTS

## Database Selection and Tooling

- **PostgreSQL** - Primary database
- **Docker / Docker Compose** - Local development environment
- **golang-migrate** - Versioned database migrations
- **sqlc** - Auto-generating Go code from SQL queries

## Authorization Flow

*To be documented*

## Schema Design

### Core Tables

#### `shen_user`

| Field    | Type   | Description                      |
|----------|--------|----------------------------------|
| pk       | PK     | Primary key                      |
| username | string | User identifier                  |
| active   | bool   | Account active status            |
| role     | FK     | Foreign key to `shen_user_roles` |

#### `shen_user_roles`

| Field | Type   | Description |
|-------|--------|-------------|
| pk    | PK     | Primary key |
| name  | string | Role name   |

**Available roles:** `server`, `user`, `admin`

#### `shen_group`

| Field | Type   | Description |
|-------|--------|-------------|
| pk    | PK     | Primary key |
| name  | string | Group name  |

#### `shen_application`

| Field | Type   | Description      |
|-------|--------|------------------|
| pk    | PK     | Primary key      |
| name  | string | Application name |

#### `shen_role`

| Field    | Type    | Description      |
|----------|---------|------------------|
| pk       | PK      | Primary key      |
| priority | integer | Role priority    |
| name     | string  | Role name        |


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

- **viewer** - Read-only access
- **service** - Service account permissions
- **auditor** - Audit and compliance access
- **operator** - Operational management
- **admin** - Full administrative access
