# Future Enhancements

This document captures features and improvements planned for future releases but not required for the initial v1 implementation.

## Token Lifecycle Notifications

### Expiration Reminders

**Goal:** Proactively notify users before their PATs expire to prevent service disruptions.

**Features:**
- Configurable notification schedule (e.g., 7 days, 3 days, 1 day before expiration)
- Multiple notification channels:
  - Email notifications
  - Webhook callbacks
  - CLI warnings when listing tokens
- Per-user notification preferences
- Batch processing to avoid overwhelming the notification system

**Implementation Considerations:**
- Background worker/cron job to scan `shen_tokens.expires_at` daily
- Notification delivery service (SMTP for email, HTTP client for webhooks)
- Track last notification sent to avoid duplicates
- Add `shen_user.email` field to schema
- Add `shen_user_notification_preferences` table

**Schema Additions:**
```sql
-- Add email to users
ALTER TABLE shen_user ADD COLUMN email VARCHAR(255) UNIQUE;

-- Track notification preferences
CREATE TABLE shen_notification_preferences (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES shen_user(id),
    channel VARCHAR(50), -- 'email', 'webhook'
    enabled BOOLEAN DEFAULT true,
    webhook_url VARCHAR(512),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Track sent notifications to avoid duplicates
CREATE TABLE shen_notification_log (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES shen_user(id),
    token_id INT REFERENCES shen_tokens(id),
    notification_type VARCHAR(50), -- 'expiration_reminder', 'token_created', 'token_revoked'
    sent_at TIMESTAMP,
    channel VARCHAR(50)
);
```

---

### Security Notifications

**Goal:** Alert users to token lifecycle events for security awareness and anomaly detection.

**Notification Events:**
1. **Token Created**
   - Token name, application, expiration date
   - IP address, user agent, timestamp
   - Action: "If this wasn't you, revoke immediately"

2. **Token Revoked**
   - Token name, application
   - Revoked by (user or admin)
   - IP address, timestamp

3. **Failed Authorization Attempts** (optional)
   - Track repeated failed attempts with invalid PATs
   - Could indicate token compromise or misconfiguration

**Implementation Considerations:**
- Trigger notifications synchronously on token CREATE/DELETE operations
- Include request metadata (IP, user agent) in notification
- Add middleware to capture request context
- Consider rate limiting to prevent notification spam
- Make security notifications opt-out (default enabled) for safety

**API Additions:**
```
POST /api/v1/user/notifications/preferences
GET  /api/v1/user/notifications/history
```

---

## Audit Logging

**Goal:** Comprehensive audit trail for compliance and security investigations.

**Events to Log:**
- Authentication attempts (success/failure)
- Token creation/revocation
- User/group/application management operations
- Permission changes (group role assignments)
- Admin actions

**Schema:**
```sql
CREATE TABLE shen_audit_log (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES shen_user(id),
    event_type VARCHAR(100), -- 'login', 'token_created', 'user_created', etc.
    resource_type VARCHAR(50), -- 'token', 'user', 'group', 'application'
    resource_id INT,
    action VARCHAR(50), -- 'create', 'update', 'delete', 'revoke'
    metadata JSONB, -- Store IP, user agent, old/new values, etc.
    created_at TIMESTAMP
);

CREATE INDEX idx_audit_log_user ON shen_audit_log(user_id);
CREATE INDEX idx_audit_log_event ON shen_audit_log(event_type);
CREATE INDEX idx_audit_log_created ON shen_audit_log(created_at);
```

**Features:**
- Query audit logs via API
- Export audit logs for external SIEM systems
- Retention policies (e.g., keep 90 days)
- Tamper-proof logging (write-only, signed entries)

---

## Key Rotation

**Goal:** Securely rotate RSA keys used for JWT signing without service disruption.

**Strategy:**
- Support multiple active keys simultaneously
- Use `kid` (key ID) in JWT header to identify which key signed it
- Rotation process:
  1. Generate new key pair
  2. Publish both old and new public keys in JWKS
  3. Start signing new JWTs with new key
  4. After grace period (e.g., 24 hours), remove old public key from JWKS
  5. Archive old private key securely

**Schema:**
```sql
CREATE TABLE shen_jwt_keys (
    id SERIAL PRIMARY KEY,
    kid VARCHAR(50) UNIQUE, -- Key ID (e.g., timestamp or UUID)
    private_key TEXT, -- PEM-encoded private key (encrypted at rest)
    public_key TEXT, -- PEM-encoded public key
    algorithm VARCHAR(10) DEFAULT 'RS256',
    active BOOLEAN DEFAULT true, -- Currently signing new JWTs
    retired_at TIMESTAMP, -- When key stopped signing new JWTs
    created_at TIMESTAMP
);
```

**CLI Commands:**
```bash
shenctl keys list                    # Show all keys (active/retired)
shenctl keys rotate                  # Generate new key, mark current as retiring
shenctl keys retire <kid>            # Remove key from JWKS (after grace period)
shenctl keys export-public <kid>     # Export public key for manual verification
```

---

## Multi-Factor Authentication (MFA)

**Goal:** Add TOTP-based MFA for enhanced account security.

**Features:**
- TOTP app support (Google Authenticator, Authy, etc.)
- Backup codes for account recovery
- Enforce MFA for admin accounts
- Optional MFA for regular users

**Implementation:**
- Generate MFA secret on enrollment
- Validate TOTP codes on login
- Store backup codes (hashed)
- Add MFA status to session tokens

**Schema Additions:**
```sql
ALTER TABLE shen_user ADD COLUMN mfa_enabled BOOLEAN DEFAULT false;
ALTER TABLE shen_user ADD COLUMN mfa_secret VARCHAR(32); -- Base32-encoded TOTP secret

CREATE TABLE shen_user_backup_codes (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES shen_user(id),
    code_hash VARCHAR(255), -- Hashed backup code
    used BOOLEAN DEFAULT false,
    used_at TIMESTAMP,
    created_at TIMESTAMP
);
```

---

## Rate Limiting

**Goal:** Protect against brute force attacks and API abuse.

**Limits:**
- Login attempts: 5 failed attempts per IP per 15 minutes
- Token authorization: 100 requests per PAT per minute
- Token creation: 10 tokens per user per hour
- Session creation: 10 sessions per user per hour

**Implementation:**
- Use Redis for distributed rate limiting
- Return `429 Too Many Requests` with `Retry-After` header
- Track violations for security monitoring

---

## Webhooks for Application Events

**Goal:** Allow applications to subscribe to Shen events (permission changes, user updates).

**Use Cases:**
- Application invalidates cached permissions when user's group changes
- Trigger onboarding workflow when new user is created
- Alert when service account token is about to expire

**Events:**
- `user.group.added` / `user.group.removed`
- `group.role.assigned` / `group.role.revoked`
- `token.expiring` / `token.revoked`

**Schema:**
```sql
CREATE TABLE shen_webhooks (
    id SERIAL PRIMARY KEY,
    application_id INT REFERENCES shen_application(id),
    url VARCHAR(512),
    events TEXT[], -- Array of event types to subscribe to
    secret VARCHAR(64), -- HMAC secret for signature verification
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP
);

CREATE TABLE shen_webhook_deliveries (
    id SERIAL PRIMARY KEY,
    webhook_id INT REFERENCES shen_webhooks(id),
    event_type VARCHAR(100),
    payload JSONB,
    status_code INT,
    attempts INT DEFAULT 0,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP
);
```

---

## OAuth 2.0 / OIDC Support

**Goal:** Support standard OAuth 2.0 flows for third-party integrations.

**Flows to Support:**
- Authorization Code Flow (with PKCE)
- Client Credentials Flow (for service-to-service)

**Benefits:**
- Interoperability with standard OAuth clients
- No custom SDK needed for integrations
- Industry-standard security practices

**Complexity:**
- Significant scope increase
- Would require consent screens, redirect URIs, client management
- Possibly out of scope for "learning project"

**Recommendation:** Only pursue if you want to learn OAuth 2.0 deeply. Otherwise, the current PAT design is simpler and sufficient.

---

## Priority Ranking

**High Priority (consider for v2):**
1. Token lifecycle notifications (expiration + security alerts)
2. Audit logging
3. Rate limiting

**Medium Priority:**
4. Key rotation
5. MFA

**Lower Priority:**
6. Webhooks
7. OAuth 2.0 (optional learning goal)
