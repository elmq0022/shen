# DESIGN AND REQUIRMENTS

## Database Selection and Tooling

postgres for db and docker / docker compose for local dev
golang-migrate for versioned db migrations
sqlc for auto generating go code from sql queries

## Schema Design



## CLI Design

### Commands and Description

shenctl - the name of the cli program; will list help when invoked

shenctl user list
shenctl user create <user-name> <role: one of service-account, user, admin>
shenctl user update <user-name> <role: one of service-account, user, admin>
shenctl user delete <user-name> (soft delete mark as inactive)
shenctl user add-groups <user-name> <group1> <group2> ...

shenctl group list
shenctl group create <group-name>
shenctl group delete <group-name>
shenctl group add-users <group-name> <user1> <user2> ...
shenctl group assign-role <group-name> <app1=rbac1> <app2=rbac2> ...

shenctl app list
shenctl app create <app-name>
shenctl app delete <app-name: soft delete mark inactive>

shenctl token list <user: optional for admin use only>
shenctl token create <token-name> <app> <user: optional for admin use only>
shenctl token revoke <token-name>

## RBAC Roles

No custom RBAC roles for now as this will help narrow the scope of the project

Roles that will be available:

- viewer
- service
- auditor
- operator
- admin
