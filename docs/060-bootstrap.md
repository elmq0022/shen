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

Keys are stored in the filesystem at:
- Private key: `$SHEN_DATA_DIR/keys/jwt-private.pem` (PEM format, RSA PKCS#8)
- Public key: `$SHEN_DATA_DIR/keys/jwt-public.pem` (PEM format)

Default `SHEN_DATA_DIR` is `./data` (relative to Shen binary).

**Security:**
- Private key file permissions: `0600` (owner read/write only)
- Public key file permissions: `0644` (world readable)
- Keys directory permissions: `0700` (owner access only)

**Key Rotation:**

Administrators can rotate keys using:
```bash
shenctl keys rotate
```

This generates a new key pair and updates the JWKS endpoint. Applications will automatically fetch the new public key on their next verification.

## Design Rationale

### Key Storage Format: PEM vs JWK

**Chosen approach:** Store keys as PEM files on disk, convert to JWK format at runtime for the JWKS endpoint.

**Alternatives considered:**

1. **PEM files (chosen)**
   - **Pros:**
     - Industry standard for storing cryptographic keys on filesystem
     - Easy to inspect and debug with standard tools (`openssl rsa -text -in jwt-private.pem`)
     - Portable across different JWT libraries and tools
     - Clear separation between storage format and wire format (JWKS)
     - Simpler key rotation workflow (swap files, conversion happens at runtime)
   - **Cons:**
     - Requires runtime conversion from PEM → JWK for the `/.well-known/jwks.json` endpoint

2. **JWK files**
   - **Pros:**
     - No conversion needed for JWKS endpoint
     - More "modern" web-focused approach
   - **Cons:**
     - Less standard for file storage
     - Fewer operational tools for inspection and debugging
     - May still need PEM format for certain signing libraries
     - Couples storage format to HTTP API format

**Why PEM:** The operational benefits of using the industry-standard PEM format outweigh the minor cost of runtime conversion to JWK. The separation of concerns (storage vs serving) is cleaner, and PEM provides better tooling support for key management tasks like inspection, backup, and rotation.

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
