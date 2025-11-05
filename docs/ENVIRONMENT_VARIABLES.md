# Environment Variables Configuration

This document describes how to configure the Torrons application using environment variables.

## Overview

The application supports configuration via:
1. **Environment variables** (highest priority)
2. **`.env` file** (loaded automatically if present)
3. **`config.yaml`** file (fallback)

Environment variables always take precedence over config file values.

## Setup

### Development

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your local values:
   ```bash
   nano .env
   # or
   vim .env
   ```

3. **NEVER commit `.env` to version control!**
   The `.gitignore` file is configured to exclude it.

### Production

Set environment variables directly in your hosting platform:

**Railway.app:**
```
Settings → Variables → Add your environment variables
```

**Fly.io:**
```bash
flyctl secrets set DB_PASSWORD=your_password
flyctl secrets set DB_HOST=your_host
# etc...
```

**Docker:**
```bash
docker run -e DB_HOST=localhost -e DB_PASSWORD=secret ...
```

## Available Variables

### Server Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `PORT` | HTTP server port | `3000` | `8080` |

### Database Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `DB_HOST` | Database hostname | `localhost` | `postgres.railway.internal` |
| `DB_PORT` | Database port | `5432` | `5432` |
| `DB_USER` | Database username | `myUser` | `postgres` |
| `DB_PASSWORD` | Database password | `myPassword` | `super_secret_pass` |
| `DB_NAME` | Database name | `databaseName` | `torrons_production` |
| `DB_SSL_MODE` | SSL mode | `disable` | `require`, `verify-full` |

#### Database SSL Modes

- **`disable`**: No SSL (local development only, NOT for production!)
- **`require`**: SSL required, no certificate verification
- **`verify-full`**: SSL required, verify certificate and hostname (recommended for production)
- **`verify-ca`**: SSL required, verify certificate

### Logger Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `LOGGER_FORMAT` | Log format | `json` | `json`, `common` |
| `LOGGER_LEVEL` | Log level | `info` | `debug`, `info`, `warn`, `error` |
| `LOGGER_PATH` | Log file path | `logs/torro.log` | `/var/log/torrons.log` |

#### Logger Formats

- **`json`**: Structured JSON logging (recommended for production)
- **`common`**: Human-readable text format (good for development)

#### Logger Levels

- **`debug`**: Very verbose, all messages
- **`info`**: General information messages (recommended for production)
- **`warn`**: Warning messages
- **`error`**: Error messages only
- **`trace`**: Most verbose, includes stack traces
- **`fatal`**: Fatal errors only

## Example Configurations

### Local Development

```env
# .env for local development
PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=torrons_dev
DB_SSL_MODE=disable

LOGGER_FORMAT=common
LOGGER_LEVEL=debug
LOGGER_PATH=logs/torro.log
```

### Production (Railway/Fly.io)

```env
# Production environment variables
PORT=3000
DB_HOST=${DATABASE_HOST}  # Provided by platform
DB_PORT=5432
DB_USER=${DATABASE_USER}  # Provided by platform
DB_PASSWORD=${DATABASE_PASSWORD}  # Provided by platform
DB_NAME=${DATABASE_NAME}  # Provided by platform
DB_SSL_MODE=verify-full  # Always use SSL in production!

LOGGER_FORMAT=json
LOGGER_LEVEL=info
LOGGER_PATH=/var/log/torrons.log
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    environment:
      - PORT=3000
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=torrons
      - DB_PASSWORD=secret_password
      - DB_NAME=torrons
      - DB_SSL_MODE=disable
      - LOGGER_FORMAT=json
      - LOGGER_LEVEL=info
    ports:
      - "3000:3000"
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=torrons
      - POSTGRES_PASSWORD=secret_password
      - POSTGRES_DB=torrons
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## Security Best Practices

### 1. Never Commit Secrets

```bash
# ✅ GOOD: Using environment variables
DB_PASSWORD=${DATABASE_PASSWORD}

# ❌ BAD: Hardcoding secrets
DB_PASSWORD=my_super_secret_password_123
```

### 2. Use Strong Passwords

```bash
# ✅ GOOD: Strong, random password
DB_PASSWORD=X7$mK9#pL2@vN4&qR8

# ❌ BAD: Weak password
DB_PASSWORD=password123
```

### 3. Enable SSL in Production

```bash
# ✅ GOOD: SSL enabled
DB_SSL_MODE=verify-full

# ❌ BAD: No SSL in production
DB_SSL_MODE=disable
```

### 4. Use JSON Logging in Production

```bash
# ✅ GOOD: Structured logs for parsing
LOGGER_FORMAT=json

# ❌ BAD: Harder to parse in production
LOGGER_FORMAT=common
```

### 5. Set Appropriate Log Level

```bash
# ✅ GOOD: Info level for production
LOGGER_LEVEL=info

# ❌ BAD: Debug level generates too many logs
LOGGER_LEVEL=debug
```

## Troubleshooting

### "Database connection failed"

1. Check `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
2. Verify database is running: `psql -h $DB_HOST -U $DB_USER -d $DB_NAME`
3. Check SSL mode matches database configuration

### "Config file not found"

This is normal! The app will create a default config file. Environment variables will still work.

### "Environment variables not loading"

1. Ensure `.env` file is in the root directory (next to `go.mod`)
2. Check file has correct format (no spaces around `=`)
3. Restart the application after changing `.env`

### "SSL connection failed"

If using `DB_SSL_MODE=verify-full`, ensure:
1. Database has valid SSL certificate
2. Certificate hostname matches `DB_HOST`
3. Try `DB_SSL_MODE=require` first, then upgrade to `verify-full`

## Testing Configuration

To test your configuration without running the full app:

```bash
# Print current configuration (masked secrets)
go run cmd/server/main.go --config-check

# Or start the app and check logs
make run
# Should see: [Config - Load] - Loaded configuration successfully
```

## Migration Guide

### From config.yaml to .env

If you have an existing `config/config.yaml`, you can migrate like this:

**Old config.yaml:**
```yaml
port: 3000
database:
  host: localhost
  port: 5432
  user: myUser
  password: myPassword
  name: databaseName
  ssl: disable
logger:
  format: json
  level: info
  path: logs/torro.log
```

**New .env:**
```env
PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=myUser
DB_PASSWORD=myPassword
DB_NAME=databaseName
DB_SSL_MODE=disable
LOGGER_FORMAT=json
LOGGER_LEVEL=info
LOGGER_PATH=logs/torro.log
```

The config file will still work as a fallback, but environment variables take precedence.

## Platform-Specific Guides

### Railway.app

1. Go to your project settings
2. Click "Variables" tab
3. Add each environment variable
4. Railway provides `DATABASE_URL` automatically - you can use it or set individual vars
5. Click "Deploy"

### Fly.io

```bash
# Set secrets
flyctl secrets set DB_PASSWORD=your_password
flyctl secrets set DB_HOST=your_host

# Set regular env vars in fly.toml
[env]
  PORT = "3000"
  LOGGER_FORMAT = "json"
  LOGGER_LEVEL = "info"
```

### Docker

```bash
# Using -e flags
docker run \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=secret \
  -e DB_NAME=torrons \
  your-image

# Or using --env-file
docker run --env-file .env your-image
```

---

**Last Updated:** November 5, 2025
**Version:** 1.0
