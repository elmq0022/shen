# Service Accounts

Service accounts are user records with `role=service` that exist solely for token-based application access. They cannot authenticate to Shen's management API and have no password.

## Characteristics

- **No Shen Access:** Service accounts cannot login to Shen or use any Shen management endpoints
- **No Password:** Service accounts have `hashed_password = NULL` and cannot use `/api/v1/auth/login`
- **Token-Only:** Tokens are created for service accounts by administrators
- **Group-Based Permissions:** Service accounts are added to groups and inherit application permissions the same way human users do

## Service Account Workflow

**1. Administrator creates service account:**
```bash
shenctl user create ci-deploy service
# Creates user with role=service, no password
```

**2. Administrator adds service account to groups:**
```bash
shenctl user add-groups ci-deploy ops-team
# Service account now has permissions via ops-team's group roles
```

**3. Administrator creates token for service account:**
```bash
shenctl token create deploy-token my-application ci-deploy
# Token is created and linked to ci-deploy service account
```

**4. Service account uses token to access applications:**
- Token is used with `POST /api/v1/authorize` (same as user PATs)
- Permissions derived from group memberships
- Short-lived JWT returned with application role

## Authorization Logic

When any request is made to Shen's management API:
```
if user.role == "service":
    return 403 Forbidden
```

Service accounts are blocked from all Shen management operations but can have tokens used for application access via the `/api/v1/authorize` endpoint.
