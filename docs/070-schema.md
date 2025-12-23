# Schema Design

## Core Tables

### `shen_user`

| Field           | Type      | Unique | Index | Description                                          |
|:----------------|:----------|:-------|:------|:-----------------------------------------------------|
| id              | PK        | Y      | -     | Primary key                                          |
| username        | string    | Y      | Y     | User identifier (enforced lowercase)                 |
| hashed_password | string    | N      | N     | Hashed password using Argon2 (nullable - NULL for service accounts)|
| active          | bool      | N      | N     | Account active status (default: true)                |
| role            | FK        | N      | Y     | Foreign key to `shen_user_role` (default: 'user')   |
| created_at      | timestamp | N      | N     | User creation timestamp                              |
| updated_at      | timestamp | N      | N     | User last update timestamp                           |

**Important:** Service accounts (role=`service`) must have `hashed_password = NULL`. These accounts cannot authenticate to Shen's management API.

**Password Hashing:** User passwords are hashed using Argon2id with recommended parameters for password storage.

**Foreign key constraints:**
- `role` REFERENCES `shen_user_role(id)` ON DELETE RESTRICT

### `shen_user_role`

| Field      | Type      | Unique | Index | Description                         |
|:-----------|:----------|:-------|:------|:------------------------------------|
| id         | PK        | Y      | -     | Primary key                         |
| name       | string    | Y      | N     | Role name (enforced lowercase)      |
| created_at | timestamp | N      | N     | Role creation timestamp             |
| updated_at | timestamp | N      | N     | Role last update timestamp          |

**Available roles:**
- `service` - Service account, cannot login to Shen, token-only access
- `user` - Regular user, can manage own PATs and view own groups
- `admin` - Administrator, can manage all Shen resources

### `shen_group`

| Field      | Type      | Unique | Index | Description                                |
|:-----------|:----------|:-------|:------|:-------------------------------------------|
| id         | PK        | Y      | -     | Primary key                                |
| name       | string    | Y      | Y     | Group name (enforced lowercase)            |
| active     | bool      | N      | N     | Group active status (default: true)        |
| created_at | timestamp | N      | N     | Group creation timestamp                   |
| updated_at | timestamp | N      | N     | Group last update timestamp                |

### `shen_user_group_member`

| Field      | Type      | Unique | Index | Description                      |
|:-----------|:----------|:-------|:------|:---------------------------------|
| id         | PK        | Y      | -     | Primary key                      |
| user_id    | FK        | N      | Y     | Foreign key to `shen_user`       |
| group_id   | FK        | N      | Y     | Foreign key to `shen_group`      |
| created_at | timestamp | N      | N     | Assignment creation timestamp    |
| updated_at | timestamp | N      | N     | Assignment last update timestamp |

**Composite unique constraint:** `(user_id, group_id)` - A user can only be assigned to a group once

**Foreign key constraints:**
- `user_id` REFERENCES `shen_user(id)` ON DELETE CASCADE
- `group_id` REFERENCES `shen_group(id)` ON DELETE CASCADE

### `shen_user_group_manager`

| Field      | Type      | Unique | Index | Description                           |
|:-----------|:----------|:-------|:------|:--------------------------------------|
| id         | PK        | Y      | -     | Primary key                           |
| user_id    | FK        | N      | Y     | Foreign key to `shen_user`            |
| group_id   | FK        | N      | Y     | Foreign key to `shen_group`           |
| created_at | timestamp | N      | N     | Manager assignment creation timestamp |
| updated_at | timestamp | N      | N     | Manager assignment last update timestamp |

**Composite unique constraint:** `(user_id, group_id)` - A user can only be a manager of a group once

**Foreign key constraints:**
- `user_id` REFERENCES `shen_user(id)` ON DELETE CASCADE
- `group_id` REFERENCES `shen_group(id)` ON DELETE CASCADE

This table defines which users are managers of which groups. Group managers can add/remove members from groups they manage, but cannot modify group settings or assign other managers (admin-only operations).

### `shen_application`

| Field      | Type      | Unique | Index | Description                                   |
|:-----------|:----------|:-------|:------|:----------------------------------------------|
| id         | PK        | Y      | -     | Primary key                                   |
| name       | string    | Y      | Y     | Application name (enforced lowercase)         |
| active     | bool      | N      | N     | Application active status (default: true)     |
| created_at | timestamp | N      | N     | Application creation timestamp                |
| updated_at | timestamp | N      | N     | Application last update timestamp             |

### `shen_permission`

| Field      | Type      | Unique | Index | Description                                 |
|:-----------|:----------|:-------|:------|:--------------------------------------------|
| id         | PK        | Y      | -     | Primary key                                 |
| priority   | integer   | Y      | Y     | Permission priority                         |
| name       | string    | Y      | N     | Permission name (enforced lowercase)        |
| created_at | timestamp | N      | N     | Permission creation timestamp               |
| updated_at | timestamp | N      | N     | Permission last update timestamp            |

**Available permissions:** `authenticated`, `viewer`, `auditor`, `operator`, `admin`

### `shen_group_app_permission`

| Field         | Type      | Unique | Index | Description                           |
|:--------------|:----------|:-------|:------|:--------------------------------------|
| id            | PK        | Y      | -     | Primary key                           |
| group_id      | FK        | N      | Y     | Foreign key to `shen_group`           |
| application_id| FK        | N      | Y     | Foreign key to `shen_application`     |
| permission_id | FK        | N      | Y     | Foreign key to `shen_permission`      |
| created_at    | timestamp | N      | N     | Assignment creation timestamp         |
| updated_at    | timestamp | N      | N     | Assignment last update timestamp      |

**Composite unique constraint:** `(group_id, application_id)` - A group can only have one specific permission per application

**Foreign key constraints:**
- `group_id` REFERENCES `shen_group(id)` ON DELETE CASCADE
- `application_id` REFERENCES `shen_application(id)` ON DELETE CASCADE
- `permission_id` REFERENCES `shen_permission(id)` ON DELETE RESTRICT

### `shen_tokens`

| Field          | Type      | Unique | Index | Description                                       |
|:---------------|:----------|:-------|:------|:--------------------------------------------------|
| id             | PK        | Y      | -     | Primary key                                       |
| name           | string    | N      | Y     | Token name/identifier (enforced lowercase)        |
| token          | string    | Y      | Y     | Hashed token value                                |
| user_id        | FK        | N      | Y     | Foreign key to `shen_user`                        |
| application_id | FK        | N      | Y     | Foreign key to `shen_application`                 |
| created_at     | timestamp | N      | Y     | Token creation timestamp                          |
| expires_at     | timestamp | N      | Y     | Token expiration timestamp                        |
| revoked        | bool      | N      | Y     | Token revocation status                           |
| revoked_at     | timestamp | N      | N     | Token revocation timestamp (nullable)             |

**Composite unique constraint:** `(user_id, application_id, name)` - A user can only have one token with the same name per application

**Foreign key constraints:**
- `user_id` REFERENCES `shen_user(id)` ON DELETE CASCADE
- `application_id` REFERENCES `shen_application(id)` ON DELETE CASCADE

This table stores PATs and service tokens. These long lived tokens can be submitted to obtain a short-lived stateless JWT which can be used to authenticate to a specific application.

### `shen_sessions`

| Field          | Type      | Unique | Index | Description                                       |
|:---------------|:----------|:-------|:------|:--------------------------------------------------|
| id             | PK        | Y      | -     | Primary key                                       |
| token          | string    | Y      | Y     | Hashed session token value (SHA-256)              |
| user_id        | FK        | N      | Y     | Foreign key to `shen_user`                        |
| created_at     | timestamp | N      | Y     | Session creation timestamp                        |
| expires_at     | timestamp | N      | Y     | Session expiration timestamp                      |
| revoked        | bool      | N      | Y     | Session revocation status                         |
| revoked_at     | timestamp | N      | N     | Session revocation timestamp (nullable)           |

This table stores session tokens used for authenticating users to the Shen management API (not application PATs).

**Foreign key constraints:**
- `user_id` REFERENCES `shen_user(id)` ON DELETE CASCADE
