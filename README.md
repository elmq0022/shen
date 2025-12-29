# 神 SHEN

## Status

In Development

## Description

A token based authorization service that facilitates Role-Based Access Control (RBAC), written in go.

## Goals

- Provide a simple but usable authorization system for managing access primarily for internal systems
- Provide username and password based auth for real users
- Provide token based auth for real users and service accounts
- Administrators can register applications with Shenlong
- Administrators can create user groups
- Users can belong to one or more user groups
- User groups can be mapped to one or more combinations of application and RBAC role
- None is a valid RBAC role and implies that the application does not use RBAC or manages RBAC itself
- shen provides a CLI for administration of the above.

## Development Setup

### Prerequisites

- [Go](https://golang.org/doc/install) 1.24+
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [direnv](https://direnv.net/) - Environment variable management
- [Task](https://taskfile.dev/) - Task runner
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations
- [sqlc](https://sqlc.dev/) - SQL to Go code generator

### Getting Started

1. Clone the repository
2. Copy the environment template and configure:
   ```bash
   cp .envrc.example .envrc
   direnv allow
   ```
3. Start the database and run migrations:
   ```bash
   task db:reset
   ```

### Common Tasks

- `task db:up` - Start PostgreSQL database
- `task db:down` - Stop PostgreSQL database
- `task db:migrate:up` - Run pending migrations
- `task db:migrate:down` - Rollback all migrations
- `task db:reset` - Reset database (down + up migrations)
- `task db:shell` - Open PostgreSQL shell
- `task db:sqlc:generate` - Generate Go code from SQL queries
- `task test:integration` - Run integration tests

Run `task --list` to see all available tasks. 
