# Backend Deployment Guide - Environment Variables

## Overview

The backend application loads configuration from:
1. **Environment variables** (highest priority)
2. **`.env` file** (loaded via `godotenv.Load()`)
3. **Default values** (fallback)

The application looks for a `.env` file in the **current working directory** when it starts.

## Current Setup Analysis

### Dockerfile.dev
- Builds the Go application
- Copies binary and migrations to final image
- **Does NOT include `.env` file** (correct - .env should not be baked into image)

### CI/CD Workflow (dev_ci.yml)
- Builds Docker image
- Pushes to GitHub Container Registry
- Deploys with volume mapping: `/opt/taxifleet_dev` (host) ↔ `/opt/taxifleet_dev` (container)

## Solution: Mount .env File via Volume

Since your deployment uses volume mapping (`v_map: true`), you can mount the `.env` file in the container.

### Option 1: Mount .env File in Working Directory (Recommended)

The application runs from `/root/` (WORKDIR in Dockerfile), so the `.env` file should be placed there.

**On your deployment server:**
```bash
# Create the directory if it doesn't exist
sudo mkdir -p /opt/taxifleet_dev

# Create/update the .env file
sudo nano /opt/taxifleet_dev/.env
```

**Update the deployment API call** to ensure the .env file is accessible:
```json
{
  "name": "taxifleet-backend-dev",
  "image": "ghcr.io/kowalski2019/taxifleet-backend:dev_latest",
  "network": "host",
  "v_map": true,
  "volume_ex": "/opt/taxifleet_dev",
  "volume_in": "/opt/taxifleet_dev",
  "opts": "--restart=always -v /opt/taxifleet_dev/.env:/root/.env"
}
```

### Option 2: Update Dockerfile to Use Mounted Volume Location

Modify the Dockerfile to change WORKDIR to the mounted volume:

```dockerfile
WORKDIR /opt/taxifleet_dev
```

Then mount the .env file:
```json
"opts": "--restart=always -v /opt/taxifleet_dev/.env:/opt/taxifleet_dev/.env"
```

### Option 3: Use Environment Variables Only (No .env file)

Pass all environment variables directly in the deployment:

```json
{
  "name": "taxifleet-backend-dev",
  "image": "ghcr.io/kowalski2019/taxifleet-backend:dev_latest",
  "network": "host",
  "v_map": true,
  "volume_ex": "/opt/taxifleet_dev",
  "volume_in": "/opt/taxifleet_dev",
  "opts": "--restart=always -e DATABASE_URL=postgres://... -e JWT_SECRET=... -e ENVIRONMENT=production"
}
```

## Required Environment Variables

### Database Configuration
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=taxifleet
DB_PASSWORD=your-secure-password
DB_NAME=taxifleet
DB_SSL_MODE=disable
```

### Server Configuration
```env
SERVER_PORT=8880
SERVER_HOST=0.0.0.0
ENVIRONMENT=production
```

### JWT Configuration
```env
JWT_SECRET=your-very-secure-secret-key-min-32-chars
JWT_EXPIRATION=15m
JWT_REFRESH_EXPIRATION=168h
```

### CORS Configuration
```env
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
```

### Logging Configuration
```env
LOG_LEVEL=info
LOG_FORMAT=json
```

## Recommended Approach

**For your deployment setup, I recommend:**

1. **Create `.env` file on deployment server:**
   ```bash
   sudo mkdir -p /opt/taxifleet_dev
   sudo nano /opt/taxifleet_dev/.env
   ```

2. **Update CI workflow** to mount the .env file explicitly (see updated workflow below)

3. **Ensure the application can read from the mounted volume**

## Security Best Practices

1. **Never commit `.env` files to Git**
2. **Use secrets management** (GitHub Secrets, HashiCorp Vault, etc.)
3. **Restrict file permissions:**
   ```bash
   sudo chmod 600 /opt/taxifleet_dev/.env
   sudo chown root:root /opt/taxifleet_dev/.env
   ```
4. **Use environment variables for sensitive data** in production
5. **Rotate secrets regularly**

## Troubleshooting

### Application can't find .env file

1. Check if file exists in container:
   ```bash
   docker exec taxifleet-backend-dev ls -la /root/.env
   ```

2. Check application working directory:
   ```bash
   docker exec taxifleet-backend-dev pwd
   ```

3. Verify volume mount:
   ```bash
   docker inspect taxifleet-backend-dev | grep -A 10 Mounts
   ```

### Environment variables not loading

1. Check if variables are set:
   ```bash
   docker exec taxifleet-backend-dev env | grep DB_
   ```

2. Verify .env file format (no spaces around `=`)
   ```env
   DB_HOST=localhost  # Correct
   DB_HOST = localhost  # Wrong (spaces)
   ```

3. Check application logs:
   ```bash
   docker logs taxifleet-backend-dev
   ```

