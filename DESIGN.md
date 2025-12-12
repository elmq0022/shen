# DESIGN AND REQUIREMENTS

## Database Selection and Tooling

- **PostgreSQL** - Primary database
- **Docker / Docker Compose** - Local development environment
- **golang-migrate** - Versioned database migrations
- **sqlc** - Auto-generating Go code from SQL queries

## Authorization Flow


### Verify identity with user name and password

User logs in using this endpoint by submitting their username and password

POST /api/v1/auth/login

payload: {
    username: string,
    password: string 
}


In return they get back a long lived jwt to interact with the shen application this would be stored locally for use with the CLI to interact with shen for generating or rejecting tokens this token would not have any specific audience or rbac associated
with it.

status code: 200
response: {
    token: string (encoded jwt)
}

### Obtain a Personal Access Tokens Scoped to an Applications

user submits a request to this endpoint and recieves a PAT with a default exp of 1 month
optionally submits an *exp* paramter as an iso date time
request must include the long userns long lived JWT from the identity verification in the header to verify their identity. 
Token is named so it can be rejected later by a user or an admin
the apppliation name is also submitted to scope the token to a specific application

POST /api/v1/token/:name/:app

headers

authorization-type Bearer: string (encoded jwt)

params :
- exp: iso string in utc optional

response: {
    name: string
    pat: string
}

shen presents the plain text PAT to the user but only stores the hashed value
will use bcrpyt or argon to has the PAT.

### Verify Application Access from PAT

POST /api/v1/authorize

user submits pat and an app name to shen
if they are authorized to use the application
they get back an encoded short live jwt containing

- username
- aud
- exp
- role

this is stateless but forces them to have shen regenerate permission regularly

payload: {
    pat: string
    app: string
}

response: {
    token: short lived jwt
    exp: iso date time utc
}


### Request Application Access

if the token has not expired submit the encoded token to applicaton
in the authorization bearer header of request
the application then decodes the jwt with the shen's public key

the application must verify:

- username
- audience
- exp
- RBAC

from the encoded JWT.

If JWT is expired the user must resubmit their PAT to get a new short lived JWT. 


## Schema Design

### Core Tables

#### `shen_user`

| Field           | Type      | Unique | Index | Description                           |
|:----------------|:----------|:-------|:------|:--------------------------------------|
| pk              | PK        | Y      | -     | Primary key                           |
| username        | string    | Y      | Y     | User identifier (enforced lowercase)  |
| hashed_password | string    | N      | N     | Hashed password                       |
| active          | bool      | N      | N     | Account active status                 |
| role            | FK        | N      | Y     | Foreign key to `shen_user_roles`      |
| created_at      | timestamp | N      | N     | User creation timestamp               |
| updated_at      | timestamp | N      | N     | User last update timestamp            |

#### `shen_user_roles`

| Field      | Type      | Unique | Index | Description                         |
|:-----------|:----------|:-------|:------|:------------------------------------|
| pk         | PK        | Y      | -     | Primary key                         |
| name       | string    | Y      | N     | Role name (enforced lowercase)      |
| created_at | timestamp | N      | N     | Role creation timestamp             |
| updated_at | timestamp | N      | N     | Role last update timestamp          |

**Available roles:** `service`, `user`, `admin`

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

**Available roles:** `none`, `viewer`, `auditor`, `operator`, `admin`

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
| name           | string    | Y      | Y     | Token name/identifier (enforced lowercase)        |
| token          | string    | Y      | Y     | Hashed token value                                |
| user_fk        | FK        | N      | Y     | Foreign key to `shen_user` (nullable)             |
| application_fk | FK        | N      | Y     | Foreign key to `shen_application` (nullable)      |
| created_at     | timestamp | N      | N     | Token creation timestamp                          |
| expires_at     | timestamp | N      | N     | Token expiration timestamp                        |
| revoked        | bool      | N      | N     | Token revocation status                           |
| revoked_at     | timestamp | N      | N     | Token revocation timestamp (nullable)             |


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

- **none** - Service account permissions
- **viewer** - Read-only access
- **auditor** - Audit and compliance access
- **operator** - Operational management
- **admin** - Full administrative access
