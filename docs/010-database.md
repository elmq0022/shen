# Database Selection and Tooling

- **PostgreSQL** - Primary database
- **Docker / Docker Compose** - Local development environment
- **golang-migrate** - Versioned database migrations
- **sqlc** - Auto-generating Go code from SQL queries

## Design Rationale

### Database Type: Relational vs NoSQL

**Chosen approach:** PostgreSQL (relational database)

**Why relational databases are required:**

1. **ACID compliance is mandatory**
   - Authorization systems cannot tolerate eventual consistency
   - Examples of unacceptable scenarios:
     - Revoked token still showing as active
     - User removed from group but JWT still claims membership
     - Race conditions during token creation/revocation
   - Strong consistency guarantees are non-negotiable for security systems

2. **Structured, relational data model**
   - Clear relationships: users → groups → roles → applications
   - Many-to-many relationships (users in groups, groups assigned to applications)
   - Referential integrity via foreign keys prevents orphaned data
   - Data normalization eliminates duplication

3. **Complex queries with joins**
   - Permission resolution requires joining multiple tables
   - Example: "What roles does user X have for application Y?" requires traversing user → group memberships → group role assignments
   - NoSQL databases lack efficient join support

4. **Transactional requirements**
   - Token revocation + audit logging must be atomic
   - Group membership changes affecting multiple users need transaction guarantees

**Why NoSQL is ruled out:**
- **Eventual consistency** - Unacceptable for authorization (stale permissions = security vulnerability)
- **No joins** - Permission model requires multi-table queries
- **No transactions** - Can't guarantee atomic operations across entities
- **Schema flexibility not needed** - Data model is well-defined and stable

### Database Choice: PostgreSQL

**Chosen approach:** PostgreSQL

**Alternatives considered:**

1. **PostgreSQL (chosen)**
   - **Pros:**
     - Industry standard for ACID-compliant relational databases
     - Battle-tested and mature with extensive documentation
     - Excellent indexing capabilities (B-tree, hash, partial indexes)
     - Rich ecosystem and tooling (migrations, ORMs, monitoring)
     - Strong performance for both reads and writes
     - JSONB support for future flexible metadata needs
   - **Cons:**
     - Requires careful index management for optimal performance
     - Single-node deployment (but appropriate for this scale)

2. **Distributed SQL (CockroachDB, YugabyteDB)**
   - **Pros:**
     - Horizontal scalability
     - Geographic distribution
   - **Cons:**
     - **Overkill for this use case** - Shen is a centralized auth service, not a global-scale system
     - Added complexity (distributed transactions, CAP theorem tradeoffs)
     - Scale requirements don't justify distributed architecture
     - Harder to operate and debug

3. **MySQL/MariaDB**
   - **Pros:**
     - Also mature and widely used
   - **Cons:**
     - Less familiarity (developer expertise matters)
     - PostgreSQL has better advanced features (JSONB, partial indexes, extensions)

**Why PostgreSQL:** It's the right-sized solution that provides ACID guarantees, excellent performance, and industry-standard tooling. The developer's familiarity with PostgreSQL reduces operational risk and increases development velocity. Distributed databases would add unnecessary complexity without meaningful benefits at this scale.

### Migration Tool: golang-migrate

**Chosen approach:** golang-migrate

**Alternatives considered:**

1. **golang-migrate (chosen)**
   - **Pros:**
     - Separate CLI tool - decoupled from application code
     - Can run as standalone CLI, Kubernetes init container/job, or embedded in application
     - Linear migration history with clear ordering
     - Migrations are actual SQL files (no abstraction layer)
     - Up/down migration support for rollbacks
     - Database-agnostic interface (works with Postgres, MySQL, etc.)
     - Community standard in Go ecosystem, well-maintained
   - **Cons:**
     - Requires external tool installation (but can also be embedded as library)

2. **goose**
   - **Pros:**
     - Similar feature set to golang-migrate
     - Supports both SQL and Go migrations
   - **Cons:**
     - Less widely adopted than golang-migrate
     - Go-based migrations add unnecessary complexity for this use case

3. **Manual SQL scripts**
   - **Pros:**
     - No dependencies
     - Direct control
   - **Cons:**
     - No version tracking
     - No rollback support
     - Error-prone (no automatic ordering)
     - Difficult to coordinate across environments

4. **Embedded in application (custom)**
   - **Pros:**
     - No external tooling needed
   - **Cons:**
     - Loses operational flexibility (can't run migrations independently)
     - Application must run to apply migrations
     - Harder to test migrations in isolation

**Why golang-migrate:** Provides operational flexibility (can run as k8s job before deployment), clear migration history, and uses plain SQL files that match exactly what runs on the database. The ability to run migrations independently of the application is valuable for deployment workflows.

### Database Access: sqlc vs ORMs vs Raw SQL

**Chosen approach:** sqlc (SQL-to-Go code generator)

**Philosophy:** "Just write SQL" - SQL is the lingua franca of databases, abstractions add complexity without meaningful benefits for this use case.

**Alternatives considered:**

1. **sqlc (chosen)**
   - **Pros:**
     - Write actual SQL, get type-safe Go functions generated at compile time
     - Compile-time SQL validation catches errors early
     - Zero runtime overhead (code generation, not reflection)
     - Generated code is readable and debuggable (no magic)
     - Explicit queries prevent N+1 problems and hidden performance issues
     - No query builder DSL to learn - just PostgreSQL SQL
   - **Cons:**
     - Requires code generation step in build process
     - PostgreSQL-specific (but database portability is not a requirement)

2. **ORMs (GORM, Ent, Bun)**
   - **Pros:**
     - Less SQL to write
     - Relationships handled automatically
     - Database portability (irrelevant - committed to PostgreSQL)
   - **Cons:**
     - **N+1 query problems** - Easy to accidentally trigger hidden queries
     - **Abstraction overhead** - Harder to debug and optimize
     - **Learning curve** - Must learn ORM-specific DSL
     - **Complex queries** - Permission resolution requires multi-table joins, clearer in SQL
     - **Runtime overhead** - Reflection and dynamic query building
     - **Implicit behavior** - "Magic" makes it harder to reason about performance in security-critical system

3. **Query builders (squirrel, goqu)**
   - **Pros:**
     - Programmatic query construction
     - Some compile-time type safety
   - **Cons:**
     - **Unnecessary abstraction** - Go code representing SQL is harder to read than SQL itself
     - **Still requires SQL knowledge** - Not hiding complexity, just adding a layer
     - **Static queries** - Most queries in Shen are static, builder adds no value
     - **Harder to optimize** - Performance tuning requires understanding generated SQL anyway

4. **database/sql + raw SQL strings**
   - **Pros:**
     - No dependencies or code generation
     - Complete control
     - Direct SQL visibility
   - **Cons:**
     - **No type safety** - Manual struct scanning is error-prone
     - **Boilerplate** - Repetitive `Scan()` calls for every query
     - **Column/field mismatches** - Easy to make mistakes, caught only at runtime
     - **No compile-time validation** - SQL errors discovered when code runs

**Why sqlc:** Combines the clarity and performance of raw SQL with compile-time type safety. Complex join queries (like permission resolution) are more readable in SQL than in ORM abstractions. Since database portability is not a requirement (committed to PostgreSQL), the main selling point of ORMs doesn't apply. For a security-critical auth system, explicit queries are better than implicit ORM magic.

**Key principle:** "Explicit is better than implicit" - knowing exactly what queries run and when they run is critical for security and performance.
