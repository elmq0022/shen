# Database Integration Tests

This directory contains integration tests for the database layer using sqlc-generated queries.

## Running Tests

These tests require a running PostgreSQL database with the proper schema. They use the `integration` build tag:

```bash
go test -v -tags=integration ./db/tests
```

Make sure your environment is configured with the necessary database connection settings (direnv recommended).

## Test Structure

All tests follow a common pattern:

```go
func TestSomeFunction(t *testing.T) {
    tdb := SetupTestDB(t)           // Creates isolated test database
    f := CreateStandardFixtures(t, tdb)  // Creates common test fixtures (optional)

    // Test logic here
}
```

## Helper Functions

Common test helpers are available in `helpers.go` and `setup.go`. Use these when applicable:
- Fixture creation functions (users, groups, applications)
- Bulk operation helpers (adding users to groups, setting permissions)
- Sorting functions for predictable assertions

## Variable Naming Conventions

To maintain consistency across all tests, follow these naming patterns:

| Variable Purpose                   | Convention           | Example                                             |
| ---------------------------------- | -------------------- | --------------------------------------------------- |
| Creation result                    | `created`            | `created := CreateTestUser(...)`                    |
| Simple fetch                       | `fetched`            | `fetched, err := GetByID(...)`                      |
| Fetch by specific field            | `fetchedBy<Field>`   | `fetchedByID`, `fetchedByName`, `fetchedByUsername` |
| Verification after update          | `updated`            | `updated, err := GetByID(...)` (after update)       |
| Verification after state change    | `<state>`            | `deactivated`, `activated`                          |
| Return from update operation       | `updatedRecord`      | When you need both update return and verification   |
| Paginated results                  | `page1`, `page2`     | Cursor-based pagination pattern                     |
| Complete result sets               | `all<Plural>`        | `allGroups`, `allMembers`, `allApps`                |
| Contextual descriptive names       | `<context><Entity>`  | `group1Managers`, `user1Groups`                     |

### Examples

```go
// Creation
created := CreateTestUser(t, tdb, "alice", RoleUser)

// Fetch by ID
fetchedByID, err := tdb.Queries.GetUserByID(tdb.Ctx, created.ID)

// Fetch by other field
fetchedByUsername, err := tdb.Queries.GetUserByUsername(tdb.Ctx, "alice")

// Update verification pattern
err := tdb.Queries.UpdateUser(tdb.Ctx, updateParams)
updated, err := tdb.Queries.GetUserByID(tdb.Ctx, userID)

// State change verification
err := tdb.Queries.DeactivateUser(tdb.Ctx, userID)
deactivated, err := tdb.Queries.GetUserByID(tdb.Ctx, userID)

// Pagination (cursor-based)
page1, err := tdb.Queries.ListUsers(tdb.Ctx, db.ListUsersParams{Limit: 2, Cursor: ""})
page2, err := tdb.Queries.ListUsers(tdb.Ctx, db.ListUsersParams{Limit: 2, Cursor: page1.NextCursor})

// Complete results
allUsers, err := tdb.Queries.ListUsers(tdb.Ctx, db.ListUsersParams{Limit: 100, Cursor: ""})

// Contextual naming
group1Managers := []db.ShenUser{f.User1, f.User2}
addManagersToGroup(t, tdb, group1Managers, f.Group1)
```

## Best Practices

1. **Use proper variable names** - Follow the naming conventions table above
2. **Use helper functions** - Leverage bulk helpers from `helpers.go` for setting up test data
3. **Test isolation** - Each test gets a fresh database via `SetupTestDB(t)`
4. **Clear assertions** - Use descriptive error messages in `require` and `assert`
5. **Test both success and failure** - Include error case testing
6. **Use proper struct types** - Prefer `db.SomeParams` over anonymous structs

## Constants

Common permission IDs used in tests:

```go
const (
    PermissionAuthenticated int32 = 1
    PermissionViewer        int32 = 2
    PermissionAuditor       int32 = 3
    PermissionOperator      int32 = 4
    PermissionAdmin         int32 = 5
)
```

Common roles:

```go
const (
    RoleUser  = "user"
    RoleAdmin = "admin"
)
```
