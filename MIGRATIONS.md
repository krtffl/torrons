# Database Migrations Guide

## Automatic Migrations ✨

**NEW:** The application now runs database migrations automatically on startup!

Simply start the server and migrations will run:

```bash
make run
# or
go run cmd/server/main.go
```

You'll see output like:
```
Running database migrations from: migrations
Current migration version: 9
Migrations completed successfully
```

---

## Manual Migration Commands

### Using Make (Recommended)

```bash
# Run all pending migrations
make migrate

# Rollback the last migration
make migrate-down

# Create a new migration file
make migrate-create
# You'll be prompted for a name

# Check current migration version
make migrate-version

# Force migration to specific version (use with caution!)
make migrate-force

# See all available commands
make help
```

### Using golang-migrate Directly

First, ensure you have the `.env` file configured:

```bash
# Run migrations
migrate -path migrations -database "postgresql://user:pass@localhost:5432/torrons?sslmode=disable" up

# Rollback last migration
migrate -path migrations -database "postgresql://user:pass@localhost:5432/torrons?sslmode=disable" down 1

# Create new migration
migrate create -ext sql -dir migrations -seq migration_name

# Check version
migrate -path migrations -database "postgresql://user:pass@localhost:5432/torrons?sslmode=disable" version
```

---

## Skipping Automatic Migrations

In production or specific scenarios, you might want to run migrations separately:

```bash
# Skip automatic migrations on startup
go run cmd/server/main.go --skip-migrations

# Or set custom migrations path
go run cmd/server/main.go --migrations /custom/path/to/migrations
```

---

## Creating New Migrations

### Using Make

```bash
make migrate-create
# Enter migration name when prompted
# Example: add_user_preferences
```

This creates two files:
- `000XXX_add_user_preferences.up.sql` - Applied when migrating up
- `000XXX_add_user_preferences.down.sql` - Applied when rolling back

### Migration File Structure

**Up migration** (`000XXX_name.up.sql`):
```sql
-- Add new table, column, or data
CREATE TABLE "UserPreferences" (
    "Id" UUID PRIMARY KEY,
    "UserId" UUID NOT NULL REFERENCES "Users"("Id"),
    "Theme" VARCHAR(20) DEFAULT 'light',
    "CreatedAt" TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Down migration** (`000XXX_name.down.sql`):
```sql
-- Revert the changes
DROP TABLE IF EXISTS "UserPreferences";
```

---

## Migration Best Practices

### 1. Always Create Both Up and Down
Every migration should be reversible. Always create both `.up.sql` and `.down.sql`.

### 2. Test Locally First
```bash
# Apply migration
make migrate

# Test your application
make run

# Rollback if needed
make migrate-down

# Re-apply
make migrate
```

### 3. Never Modify Applied Migrations
Once a migration has been applied to any environment, never modify it. Create a new migration instead.

### 4. Keep Migrations Atomic
Each migration should do one logical thing:
- ✅ Good: `add_user_email_column`
- ❌ Bad: `add_columns_and_fix_data_and_create_indexes`

### 5. Use Transactions
Most database operations in migrations are automatically wrapped in transactions. For complex changes:

```sql
BEGIN;

-- Your changes here
ALTER TABLE "Users" ADD COLUMN "Email" VARCHAR(255);
UPDATE "Users" SET "Email" = 'unknown@example.com' WHERE "Email" IS NULL;
ALTER TABLE "Users" ALTER COLUMN "Email" SET NOT NULL;

COMMIT;
```

---

## Troubleshooting

### Dirty Migration State

If a migration fails halfway:

```bash
# Check current state
make migrate-version
# Output: version: 5, dirty: true

# Fix the issue in the migration file, then force to last good version
make migrate-force
# Enter the last known good version

# Try again
make migrate
```

### Connection Refused

Check your `.env` file:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=torrons
DB_SSL_MODE=disable
```

Verify PostgreSQL is running:
```bash
pg_isready
```

### Migration Not Found

Ensure you're in the project root directory and the `migrations` folder exists:
```bash
ls -la migrations/
```

### Permission Denied

Check database user permissions:
```sql
-- As postgres superuser
GRANT ALL PRIVILEGES ON DATABASE torrons TO your_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_user;
```

---

## Production Deployment

### Option 1: Automatic on Startup (Recommended for small apps)

The default behavior. Just deploy and start:
```bash
./server
```

### Option 2: Manual Before Deployment (Recommended for large apps)

Run migrations separately before deploying new code:

```bash
# On production server, before starting new version
make migrate

# Then deploy and start with migrations disabled
./server --skip-migrations
```

### Option 3: CI/CD Pipeline

Add to your deployment pipeline:

```yaml
# Example GitHub Actions
- name: Run Migrations
  run: |
    export DB_HOST=${{ secrets.DB_HOST }}
    export DB_USER=${{ secrets.DB_USER }}
    export DB_PASSWORD=${{ secrets.DB_PASSWORD }}
    export DB_NAME=${{ secrets.DB_NAME }}
    export DB_SSL_MODE=require
    make migrate
```

---

## Migration History

Current migrations (as of v1.0.0):

1. `000001_create_classes` - Product categories
2. `000002_create_torrons` - Torron products
3. `000003_create_pairings` - Vote pairings
4. `000004_create_results` - Vote results
5. `000005_insert_classes` - Seed category data
6. `000006_insert_torrons` - Seed torron data
7. `000007_create_users` - User tracking
8. `000008_create_campaigns` - Campaign management
9. `000009_create_user_elo_snapshots` - Personalized ELO

---

## Quick Reference

```bash
# Development workflow
make run                    # Start server (auto-migrates)
make migrate               # Run migrations manually
make migrate-down          # Undo last migration
make migrate-create        # Create new migration
make migrate-version       # Check current version

# Production workflow
make migrate               # Run migrations
./server --skip-migrations # Start without auto-migrate

# Troubleshooting
make migrate-version       # Check state
make migrate-force         # Fix dirty state (careful!)
make help                  # See all commands
```

---

## FAQ

**Q: Do migrations run on every server restart?**
A: Yes, but they only apply new migrations. If there are no new migrations, it's instant.

**Q: Can I disable automatic migrations?**
A: Yes, use `--skip-migrations` flag when starting the server.

**Q: What happens if a migration fails?**
A: The server logs a warning and continues. Fix the migration and run `make migrate` manually.

**Q: Can I run migrations on multiple servers simultaneously?**
A: The migrate library uses database locks to prevent concurrent migrations. Only one will run, others will wait.

**Q: How do I reset my database?**
A: Drop and recreate:
```bash
dropdb torrons
createdb torrons
make run  # Will run all migrations
```

**Q: Where are migrations stored?**
A: In the `migrations/` directory at project root.

---

For more information, see:
- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL documentation](https://www.postgresql.org/docs/)
