# Initial Bootstrap and Setup

## Default Admin Account

On first startup, if no users exist in the database, Shen will automatically create a default admin account:

**Default credentials:**
- Username: `admin`
- Password: `admin`

**Security Warning:** Change these credentials immediately after first login.

**Configuration via Environment Variables:**
- `SHEN_ADMIN_USERNAME` - Override default admin username (default: `admin`)
- `SHEN_ADMIN_PASSWORD` - Override default admin password (default: `admin`)

## Public/Private Key Generation

On first startup, if no JWT signing keys exist, Shen will automatically generate an RSA key pair (2048-bit):
- Private key: Used to sign JWTs
- Public key: Exposed via `/.well-known/jwks.json` for applications to verify JWTs

**Key Storage:**

Keys are stored in the PostgreSQL database in the `shen_jwt_keys` table:
- Private key: PEM format (RSA PKCS#8), encrypted at rest using a Key Encryption Key (KEK)
- Public key: PEM format, stored in plaintext (public by design)
- Key ID (`kid`): Unique identifier for JWKS endpoint (format: YYYY-MM-DD timestamp)

**Database Schema:**
```sql
CREATE TABLE shen_jwt_keys (
    id SERIAL PRIMARY KEY,
    kid TEXT NOT NULL UNIQUE,
    encrypted_private_key BYTEA NOT NULL,
    public_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active_for_signing BOOLEAN NOT NULL DEFAULT false,
    active_for_verification BOOLEAN NOT NULL DEFAULT true
);
```

**Key Encryption Key (KEK):**

The KEK is used to encrypt private keys before storing them in the database. The KEK must be provided to Shen via environment variable or Kubernetes Secret:

```bash
# Via environment variable (32-byte base64-encoded random key)
SHEN_KEY_ENCRYPTION_KEY=<base64-encoded-32-bytes>

# Generate KEK
openssl rand -base64 32
```

**For Kubernetes deployments:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: shen-kek
type: Opaque
data:
  key-encryption-key: <base64-random-32-bytes>
```

**Security:**
- Private keys encrypted with AES-256-GCM using KEK before storage
- KEK never stored in database
- Public keys stored in plaintext (intended for public distribution)
- Multiple keys supported simultaneously (for zero-downtime rotation)

**Key Rotation:**

Administrators can rotate keys using:
```bash
shenctl keys rotate
```

**Rotation process:**
1. Generate new RSA key pair
2. Encrypt new private key with KEK
3. Insert new key into database with `active_for_signing=true`, `active_for_verification=true`
4. Mark old key as `active_for_signing=false` (keep `active_for_verification=true`)
5. JWKS endpoint immediately exposes both keys (old and new)
6. Applications automatically use correct key based on JWT `kid` header
7. After 10+ minutes (JWT TTL + buffer), old key can be disabled for verification

**Zero-downtime rotation:**
- New JWTs signed with new key immediately
- Old JWTs still verified using old key (up to 7 min TTL)
- No pod restarts required
- No user re-authentication required

## Design Rationale

### Key Storage Format: PEM vs JWK

**Chosen approach:** Store keys in PEM format (in database), convert to JWK format at runtime for the JWKS endpoint.

**Alternatives considered:**

1. **PEM format (chosen)**
   - **Pros:**
     - Industry standard for storing cryptographic keys
     - Easy to inspect and debug with standard tools (`openssl rsa -text`)
     - Portable across different JWT libraries and tools
     - Clear separation between storage format and wire format (JWKS)
     - Works with Go's `crypto/rsa` and `crypto/x509` standard library
   - **Cons:**
     - Requires runtime conversion from PEM → JWK for the `/.well-known/jwks.json` endpoint

2. **JWK format**
   - **Pros:**
     - No conversion needed for JWKS endpoint
     - More "modern" web-focused approach
   - **Cons:**
     - Less standard for key storage
     - Fewer operational tools for inspection and debugging
     - May still need PEM format for certain signing libraries
     - Couples storage format to HTTP API format

**Why PEM:** The operational benefits of using the industry-standard PEM format outweigh the minor cost of runtime conversion to JWK. The separation of concerns (storage vs serving) is cleaner, and PEM provides better tooling support for key management tasks like inspection, backup, and rotation.

### Key Storage Location: Database vs Filesystem vs Kubernetes Secret

**Chosen approach:** Store keys in PostgreSQL database, encrypted at rest with a Key Encryption Key (KEK).

**Alternatives considered:**

1. **Database with KEK (chosen)**
   - **Pros:**
     - **High Availability** - All replicas share same keys automatically (no sync needed)
     - **Zero-downtime rotation** - Update database, all pods see new keys immediately
     - **Simple operations** - Single source of truth, no file distribution
     - **`shenctl` integration** - Easy key rotation via CLI (`shenctl keys rotate`)
     - **Multi-key support** - Store multiple keys for graceful rotation
     - **Audit trail** - Track key creation and rotation in database
     - **Works everywhere** - Not Kubernetes-specific
   - **Cons:**
     - Keys in database backups (mitigated: encrypted with KEK, KEK not in backup)
     - Database compromise exposes encrypted keys (mitigated: need both DB + KEK)
     - Additional table to manage

2. **Kubernetes Secrets**
   - **Pros:**
     - Native K8s pattern for secrets
     - etcd encryption at rest
     - RBAC for access control
     - No database storage concerns
   - **Cons:**
     - **Replica consistency** - All pods mount same secret (works, but requires K8s)
     - **Key rotation complexity** - Update secret + wait for propagation + rolling restart
     - **Downtime during rotation** - Outstanding JWTs become invalid when pods restart
     - **K8s-specific** - Doesn't work for non-K8s deployments
     - **External management** - Can't rotate via `shenctl` easily

3. **Filesystem (PersistentVolume)**
   - **Pros:**
     - Simple file-based approach
     - Standard tools work (openssl)
   - **Cons:**
     - **ReadWriteMany required** - Not all storage classes support it
     - **Cloud-specific** - Need NFS/EFS/Filestore (expensive for 4KB of data)
     - **Operational complexity** - Network filesystem for small static files
     - **Statefulness** - Makes Shen harder to scale/deploy
     - **Sync complexity** - Hard to coordinate rotation across replicas

4. **Database without encryption**
   - **Pros:**
     - Simple implementation
     - Same HA benefits as encrypted approach
   - **Cons:**
     - **Security risk** - Private keys in plaintext in backups
     - **Compliance issues** - Exposing signing keys in database dumps
     - **Insider threat** - DBAs can read private keys

**Why Database + KEK:**

Shen is designed for high availability - multiple replicas are expected, not optional. The database is already the single source of truth for all other data (users, groups, tokens), so storing keys there maintains consistency. Zero-downtime key rotation is critical for production auth systems.

**Key principle:** "Don't write this twice" - supporting both HA and easy key rotation (`shenctl keys rotate`) requires shared storage. The database is the natural choice, and KEK encryption addresses the security concerns of storing keys in the database.

**Security model:**
- **Database compromise alone** - Attacker gets encrypted keys (useless without KEK)
- **KEK compromise alone** - Attacker has nothing to decrypt (keys in database)
- **Both compromised** - Same as any key storage method (infrastructure fully compromised)

**Backup strategy:**
- Database backups contain encrypted keys only
- KEK stored separately (environment variable / K8s Secret)
- KEK never in database or backups
- Restore requires both backup + KEK

### Key Encryption: KEK Approach

**Chosen approach:** Encrypt private keys with AES-256-GCM using a Key Encryption Key before storing in database.

**Why encrypt keys at rest:**
1. **Defense in depth** - Database compromise alone doesn't expose signing keys
2. **Backup security** - Database dumps don't contain plaintext private keys
3. **Compliance** - Sensitive cryptographic material encrypted at rest
4. **Separation of concerns** - KEK managed separately from application data

**KEK management:**
- Provided via environment variable (`SHEN_KEY_ENCRYPTION_KEY`)
- Typically stored in Kubernetes Secret (separate from application data)
- 32-byte random key (256-bit security)
- Never stored in database
- Rotatable independently of JWT signing keys

**Encryption algorithm (AES-256-GCM):**
- Authenticated encryption (prevents tampering)
- Industry standard (NIST approved)
- Fast performance (hardware acceleration on modern CPUs)
- Go standard library support (`crypto/aes`, `crypto/cipher`)

## Database Migrations

Shen uses database migrations to set up the schema and seed reference data. Run migrations before first startup:

```bash
migrate -database $DATABASE_URL -path ./db/migrations up
```

**User Roles** (`shen_user_role`) - seeded via migration:
- `service`
- `user`
- `admin`

**Application Roles** (`shen_application_role`) - seeded via migration:
- `authenticated` (priority: 100) - Authentication only, no Shen-managed authorization
- `viewer` (priority: 200)
- `auditor` (priority: 300)
- `operator` (priority: 400)
- `admin` (priority: 500)
