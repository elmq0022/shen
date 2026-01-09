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

### `shen_application_role`

| Field      | Type      | Unique | Index | Description                                 |
|:-----------|:----------|:-------|:------|:--------------------------------------------|
| id         | PK        | Y      | -     | Primary key                                 |
| priority   | integer   | Y      | Y     | Role priority (for display/sorting)         |
| name       | string    | Y      | N     | Role name (enforced lowercase)              |
| created_at | timestamp | N      | N     | Role creation timestamp                     |
| updated_at | timestamp | N      | N     | Role last update timestamp                  |

**Available application roles:** `authenticated`, `viewer`, `auditor`, `operator`, `admin`

This table defines the standard set of roles that can be assigned to users for applications. These are separate from `shen_user_role` which controls access to Shen itself.

### `shen_group_application_role`

| Field         | Type      | Unique | Index | Description                           |
|:--------------|:----------|:-------|:------|:--------------------------------------|
| id            | PK        | Y      | -     | Primary key                           |
| group_id      | FK        | N      | Y     | Foreign key to `shen_group`           |
| application_id| FK        | N      | Y     | Foreign key to `shen_application`     |
| role_id       | FK        | N      | Y     | Foreign key to `shen_application_role`|
| created_at    | timestamp | N      | N     | Assignment creation timestamp         |
| updated_at    | timestamp | N      | N     | Assignment last update timestamp      |

**Composite unique constraint:** `(group_id, application_id, role_id)` - A group can have each role only once per application, but can have multiple different roles for the same application

**Foreign key constraints:**
- `group_id` REFERENCES `shen_group(id)` ON DELETE CASCADE
- `application_id` REFERENCES `shen_application(id)` ON DELETE CASCADE
- `role_id` REFERENCES `shen_application_role(id)` ON DELETE RESTRICT

This table implements many-to-many mapping between groups, applications, and roles. A group can have multiple roles for an application (e.g., both `viewer` and `auditor`).

When generating a JWT for a user:
- Shen collects all groups the user belongs to
- For each group, Shen looks up which roles that group grants for the target application
- The JWT includes the deduplicated union of all roles
- The JWT also includes the group names themselves - applications interpret what permissions each group has

### `shen_token`

| Field          | Type      | Unique | Index | Description                                       |
|:---------------|:----------|:-------|:------|:--------------------------------------------------|
| id             | PK        | Y      | -     | Primary key                                       |
| name           | string    | N      | Y     | Token name/identifier (enforced lowercase)        |
| hashed_token   | string    | Y      | Y     | Hashed token value (Argon2id)                     |
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

**Token Hashing:** PAT tokens are hashed using Argon2id with recommended parameters for password storage before being stored in the database.

This table stores PATs and service tokens. These long lived tokens can be submitted to obtain a short-lived stateless JWT which can be used to authenticate to a specific application.

### `shen_session`

| Field          | Type      | Unique | Index | Description                                       |
|:---------------|:----------|:-------|:------|:--------------------------------------------------|
| id             | PK        | Y      | -     | Primary key                                       |
| hashed_token   | string    | Y      | Y     | Hashed session token value (SHA-256)              |
| user_id        | FK        | N      | Y     | Foreign key to `shen_user`                        |
| created_at     | timestamp | N      | Y     | Session creation timestamp                        |
| expires_at     | timestamp | N      | Y     | Session expiration timestamp                      |
| revoked        | bool      | N      | Y     | Session revocation status                         |
| revoked_at     | timestamp | N      | N     | Session revocation timestamp (nullable)           |

This table stores session tokens used for authenticating users to the Shen management API (not application PATs).

**Foreign key constraints:**
- `user_id` REFERENCES `shen_user(id)` ON DELETE CASCADE

## Pagination Design

### Cursor-Based vs Offset-Based Pagination

Shen uses **cursor-based pagination** for all list endpoints. This design choice has significant performance and stability advantages over traditional offset-based pagination.

#### Design Trade-offs

**Cursor-Based Pagination (Used in Shen)**

Advantages:
- **No performance degradation on deep pages** - Queries use indexed columns with simple comparison operators (`WHERE id > cursor_id`), maintaining O(1) seek time regardless of page depth
- **Stable results** - Cursor references specific records, preventing items from being skipped or duplicated when data changes between requests
- **Index-friendly** - Leverages B-tree indexes efficiently with range scans instead of offset skips
- **Scalability** - Performance remains constant even with millions of records

Disadvantages:
- Cannot jump to arbitrary page numbers (no "go to page 5" functionality)
- Slightly more complex client implementation (must track cursor values)

**Offset-Based Pagination (NOT used)**

Advantages:
- Simple implementation (`LIMIT x OFFSET y`)
- Can jump to arbitrary page numbers
- Familiar API pattern

Disadvantages:
- **Performance degrades linearly with page depth** - Database must scan and discard `OFFSET` rows before returning results, making deep pages extremely slow
- **Unstable results** - Insertions/deletions between requests cause items to shift, leading to duplicates or missing records
- **Index inefficiency** - Even with indexes, the database must walk through offset rows
- **Poor scalability** - Large offsets can cause timeouts and high database load

#### Implementation Pattern

Shen's cursor-based pagination uses indexed columns for stable, efficient iteration:

```sql
-- Example: Paginating users by username (unique, indexed)
SELECT id, username, created_at
FROM shen_user
WHERE ($1::text = '' OR username > $1)  -- $1 is cursor
ORDER BY username
LIMIT $2;

-- Example: Paginating sessions by auto-incrementing ID
SELECT id, hashed_token, user_id, created_at
FROM shen_session
WHERE user_id = $1
  AND ($2 = 0 OR id > $2)  -- $2 is cursor_id
ORDER BY id ASC
LIMIT $3;
```

**Key characteristics:**
- Cursor always references an indexed column (unique text field, auto-incrementing ID, or ULID)
- First page uses empty cursor (`''` for text, `0` for IDs)
- Each response returns cursor for next page
- Queries remain fast regardless of dataset size

**Note on Auto-Incrementing IDs:**
Shen currently uses auto-incrementing `SERIAL` IDs for pagination sorting, which is sufficient for this learning exercise. However, production systems should consider **ULID** (Universally Unique Lexicographically Sortable Identifier) or **UUID7** instead because:

- **Distributed system compatibility** - Auto-incrementing IDs require centralized sequence generation, creating a single point of contention and preventing true horizontal scaling across multiple database instances
- **No coordination overhead** - ULIDs/UUID7s can be generated independently by any application instance without database round-trips or lock contention
- **Information leakage prevention** - Sequential IDs expose business metrics (creation rate, total records) and make enumeration attacks trivial
- **Merge/replication safety** - ULIDs eliminate ID collision risks when merging databases or replicating across regions
- **Sortable by creation time** - Unlike UUID4, ULID and UUID7 maintain lexicographic sorting that correlates with creation time, making them ideal cursor values
- **Future-proof architecture** - Switching from auto-increment to distributed IDs later requires complex migration; starting with ULIDs avoids this technical debt

For a centralized authentication service that may eventually need multi-region deployment or high-availability failover, ULIDs provide better architectural runway without sacrificing pagination performance.

#### Performance Comparison

For a table with 1 million records, fetching page 1000 (records 100,000-100,100):

| Method           | Query Pattern                   | Approximate Time | Index Usage                |
|:-----------------|:--------------------------------|:-----------------|:---------------------------|
| Cursor-based     | `WHERE id > 99999 LIMIT 100`    | ~1-2ms           | Efficient seek to position |
| Offset-based     | `LIMIT 100 OFFSET 99900`        | ~500-1000ms      | Must scan 99,900 rows      |

The performance gap widens as the offset increases - cursor-based pagination maintains constant time while offset-based approaches become unusable for deep pages.
