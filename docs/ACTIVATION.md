# License Activation System

## Overview

CKS Weight Room uses a license activation system to control access to the application. The system supports both online and offline activation with secure encryption and device fingerprinting.

## Architecture

### Client (CKS Weight Room App)
- Machine fingerprinting using hostname + MAC address
- AES-256-GCM encryption for storing license keys and tokens
- Automatic activation check on startup
- Redirect to `/activate` page if not activated

### Server (Activation Backend)
- License key validation
- Device tracking (max 3 devices per license)
- Activation token generation
- Periodic validation (every 7 days with 30-day grace period)

## Development vs Production

### Development Mode (Current)

By default, the application runs in **mock mode** which accepts any license key in the correct format for local testing.

**Enable mock mode:**
```bash
export ACTIVATION_MOCK=true
./cks-weight-room
```

**Valid test license keys:**
- `CKSWT-ABCDE-12345-FGHIJ-67890`
- Any key matching `CKSWT-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}`

### Production Mode

To enable production activation, you need to:

1. **Set up the Activation Server**
2. **Configure certificate pinning**
3. **Disable mock mode**

## Production Activation Server

### Required Endpoints

#### POST /api/v1/activate

**Request:**
```json
{
  "licenseKey": "CKSWT-XXXXX-XXXXX-XXXXX-XXXXX",
  "machineId": "ABCD-1234-EFGH-5678",
  "appVersion": "0.1.0"
}
```

**Response (Success):**
```json
{
  "success": true,
  "activationToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expiresAt": null,
  "message": "License activated successfully"
}
```

**Response (Device Limit Reached):**
```json
{
  "success": false,
  "error": "DEVICE_LIMIT_REACHED",
  "message": "This license is activated on 3 machines (maximum). Please deactivate a device first."
}
```

**Response (Invalid License):**
```json
{
  "success": false,
  "error": "INVALID_LICENSE",
  "message": "License key not found or has been revoked"
}
```

#### POST /api/v1/validate

**Request:**
```json
{
  "activationToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "machineId": "ABCD-1234-EFGH-5678"
}
```

**Response:**
```json
{
  "valid": true,
  "expiresAt": null
}
```

### Database Schema (Activation Server)

```sql
-- Licenses table
CREATE TABLE licenses (
    license_key VARCHAR(29) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    max_devices INT DEFAULT 3,
    status VARCHAR(20) DEFAULT 'active' -- active, revoked, expired
);

-- Activations table
CREATE TABLE activations (
    id SERIAL PRIMARY KEY,
    license_key VARCHAR(29) REFERENCES licenses(license_key),
    machine_id VARCHAR(19) NOT NULL,
    activation_token TEXT NOT NULL,
    activated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_validated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    app_version VARCHAR(20),
    UNIQUE(license_key, machine_id)
);

-- Index for quick lookups
CREATE INDEX idx_activations_license_key ON activations(license_key);
CREATE INDEX idx_activations_machine_id ON activations(machine_id);
```

### Sample Activation Server (Node.js/Express)

```javascript
const express = require('express');
const jwt = require('jsonwebtoken');
const { Pool } = require('pg');

const app = express();
app.use(express.json());

const JWT_SECRET = process.env.JWT_SECRET || 'your-secret-key';
const db = new Pool({
  connectionString: process.env.DATABASE_URL
});

// POST /api/v1/activate
app.post('/api/v1/activate', async (req, res) => {
  const { licenseKey, machineId, appVersion } = req.body;

  // Validate license key format
  if (!/^CKSWT-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}$/.test(licenseKey)) {
    return res.json({
      success: false,
      error: 'INVALID_FORMAT',
      message: 'Invalid license key format'
    });
  }

  // Check if license exists and is valid
  const license = await db.query(
    'SELECT * FROM licenses WHERE license_key = $1',
    [licenseKey]
  );

  if (!license.rows.length || license.rows[0].status !== 'active') {
    return res.json({
      success: false,
      error: 'INVALID_LICENSE',
      message: 'License key not found or has been revoked'
    });
  }

  // Check if license has expired
  if (license.rows[0].expires_at && new Date(license.rows[0].expires_at) < new Date()) {
    return res.json({
      success: false,
      error: 'LICENSE_EXPIRED',
      message: 'This license has expired'
    });
  }

  // Check device limit
  const activations = await db.query(
    'SELECT COUNT(*) FROM activations WHERE license_key = $1',
    [licenseKey]
  );

  if (parseInt(activations.rows[0].count) >= license.rows[0].max_devices) {
    // Check if this machine is already activated
    const existing = await db.query(
      'SELECT * FROM activations WHERE license_key = $1 AND machine_id = $2',
      [licenseKey, machineId]
    );

    if (!existing.rows.length) {
      return res.json({
        success: false,
        error: 'DEVICE_LIMIT_REACHED',
        message: `This license is activated on ${license.rows[0].max_devices} machines (maximum). Please deactivate a device first.`
      });
    }
  }

  // Generate activation token (JWT)
  const activationToken = jwt.sign(
    { licenseKey, machineId, appVersion },
    JWT_SECRET,
    { expiresIn: license.rows[0].expires_at ? '365d' : '10y' }
  );

  // Store activation
  await db.query(
    `INSERT INTO activations (license_key, machine_id, activation_token, app_version)
     VALUES ($1, $2, $3, $4)
     ON CONFLICT (license_key, machine_id)
     DO UPDATE SET last_validated_at = CURRENT_TIMESTAMP, activation_token = $3`,
    [licenseKey, machineId, activationToken, appVersion]
  );

  res.json({
    success: true,
    activationToken,
    expiresAt: license.rows[0].expires_at,
    message: 'License activated successfully'
  });
});

// POST /api/v1/validate
app.post('/api/v1/validate', async (req, res) => {
  const { activationToken, machineId } = req.body;

  try {
    // Verify JWT token
    const decoded = jwt.verify(activationToken, JWT_SECRET);

    // Check if machine ID matches
    if (decoded.machineId !== machineId) {
      return res.json({ valid: false, error: 'Machine ID mismatch' });
    }

    // Update last validated timestamp
    await db.query(
      'UPDATE activations SET last_validated_at = CURRENT_TIMESTAMP WHERE activation_token = $1',
      [activationToken]
    );

    res.json({ valid: true });
  } catch (err) {
    res.json({ valid: false, error: 'Invalid or expired token' });
  }
});

app.listen(3001, () => console.log('Activation server running on port 3001'));
```

## Certificate Pinning

To enable certificate pinning for enhanced security:

1. **Get your activation server's public key hash:**
```bash
openssl s_client -connect activation.cks-weight-room.com:443 -showcerts \
  | openssl x509 -pubkey -noout \
  | openssl pkey -pubin -outform DER \
  | openssl dgst -sha256 -binary \
  | openssl enc -base64
```

2. **Update the hash in `internal/activation/client.go`:**
```go
const ExpectedPublicKeyHash = "YOUR_ACTUAL_HASH_HERE"
```

3. **Uncomment certificate verification in the TLS config:**
```go
TLSClientConfig: &tls.Config{
    MinVersion: tls.VersionTLS12,
    VerifyPeerCertificate: verifyCertificatePin,  // Uncomment this line
}
```

## Enabling Production Mode

**Option 1: Environment variable**
```bash
unset ACTIVATION_MOCK  # or set to "false"
./cks-weight-room
```

**Option 2: Build with production defaults**

Edit `internal/activation/client.go` line 58:
```go
useMock := os.Getenv("ACTIVATION_MOCK") == "true"  // Current (mock by default in dev)
```

Change to:
```go
useMock := os.Getenv("ACTIVATION_MOCK") != "false"  // Production (real validation by default)
```

## Testing Production Activation Locally

You can run the activation server locally for testing:

1. Start the activation server on `localhost:3001`
2. Update `internal/activation/client.go`:
```go
const ActivationServerURL = "http://localhost:3001/api/v1"
```
3. Disable certificate pinning (comment out `VerifyPeerCertificate`)
4. Run without `ACTIVATION_MOCK=true`

## Security Considerations

### Implemented
- ✅ AES-256-GCM encryption for sensitive data storage
- ✅ Machine fingerprinting (hostname + MAC address)
- ✅ HTTPS-only activation requests
- ✅ Certificate pinning support
- ✅ TLS 1.2+ minimum version

### Recommended for Production
- 🔒 Use JWT tokens with expiration for activation tokens
- 🔒 Rate limiting on activation endpoints (prevent brute force)
- 🔒 Logging and monitoring of activation attempts
- 🔒 Admin panel for license management (revoke, transfer, etc.)
- 🔒 Email notifications for new activations
- 🔒 Deactivation endpoint for users to free up device slots

## Troubleshooting

### "Network error" during activation
- Check that `ACTIVATION_MOCK=true` is set for development
- Verify activation server is reachable
- Check firewall/proxy settings

### "Invalid license key format"
- Ensure key matches: `CKSWT-XXXXX-XXXXX-XXXXX-XXXXX`
- All characters must be uppercase letters or numbers

### "Certificate pin mismatch"
- Verify the public key hash is correct
- Ensure you're connecting to the right server
- Check if certificate has been renewed (update pin)

## Next Steps

See the following stories in Epic 6:
- **Story 6.2**: Offline Activation with Activation File
- **Story 6.3**: Periodic License Validation (every 7 days, 30-day grace)
