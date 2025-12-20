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

## Database Seeding

On startup, Shen will seed the following reference data if not present:

**User Roles** (`shen_user_role`):
- `service`
- `user`
- `admin`

**Application Roles** (`shen_application_role`):
- `authenticated` (priority: 0) - Authentication only, no Shen-managed authorization
- `viewer` (priority: 100)
- `auditor` (priority: 200)
- `operator` (priority: 300)
- `admin` (priority: 400)
