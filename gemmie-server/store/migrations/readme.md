# Database Migration Management Guide

## Understanding Migrations

### What are UP and DOWN migrations?

- **UP Migration** (`*.up.sql`): Applies changes to move the database schema forward
  - Example: Creating tables, adding columns, creating indexes
  - File: `000001_initial_schema.up.sql`

- **DOWN Migration** (`*.down.sql`): Reverses the changes made by UP migration
  - Example: Dropping tables, removing columns, dropping indexes
  - File: `000001_initial_schema.down.sql`

### Migration Versions

Migrations are numbered sequentially:

- `000001_initial_schema` - Version 1
- `000002_add_user_transaction_relationship` - Version 2
- `000003_your_next_migration` - Version 3

The system tracks which version your database is currently at.

---

## Migration Operations

### 1. **UP** - Apply Migrations Forward

**What it does:** Runs all pending migrations to bring your database to
the latest version.

```go
// Automatically runs in InitStorage()
store.InitStorage(connString)

// Or manually:
store.MigrateUp()
```

**Example:**

- Current version: 1
- Available versions: 1, 2, 3
- Result: Applies versions 2 and 3

---

### 2. **DOWN** - Rollback One Migration

**What it does:** Rolls back the most recent migration (one step back).

```go
err := store.MigrateDown()
```

**Example:**

- Current version: 3
- After down: version 2
- The `000003_*.down.sql` file is executed

---

### 3. **STEPS** - Rollback Multiple Migrations

**What it does:** Rolls back a specific number of migrations.

```go
// Roll back 2 migrations
err := store.MigrateSteps(-2)
```

**Example:**

- Current version: 5
- After `MigrateSteps(-2)`: version 3
- Executes `000005_*.down.sql` then `000004_*.down.sql`

---

### 4. **TO** - Migrate to Specific Version

**What it does:** Migrates to an exact version (can go up or down).

```go
// Go to version 2
err := store.MigrateTo(2)
```

**Examples:**

- Current: 5, Target: 2 → Rolls back versions 5, 4, 3
- Current: 1, Target: 3 → Applies versions 2, 3

---

### 5. **VERSION** - Check Current Version

**What it does:** Shows which migration version the database is at.

```go
version, dirty, err := store.GetMigrationVersion()
fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)
```

---

### 6. **FORCE** - Fix Dirty State

**What it does:** Forces the migration version when database is in "dirty" state.

A "dirty" state means a migration failed halfway through.

```go
// Force to version 2
err := store.ForceMigrationVersion(2)
```

---

## Common Scenarios

### Scenario 1: Normal Development

```bash
# Start app - migrations run automatically
go run main.go

# Database goes from version 0 → 1 → 2 automatically
```

### Scenario 2: Rollback Last Change

```bash
# Use CLI tool to rollback
go run cmd/migrate/main.go --command=down

# Or in code:
store.MigrateDown()
```

### Scenario 3: Test a Migration

```bash
# Apply migration
go run cmd/migrate/main.go --command=up

# Test it...

# Rollback if there's an issue
go run cmd/migrate/main.go --command=down

# Fix migration file, then re-apply
go run cmd/migrate/main.go --command=up
```

### Scenario 4: Production Rollback

```bash
# Check current version
go run cmd/migrate/main.go --command=version
# Output: Current version: 5

# Rollback to version 3
go run cmd/migrate/main.go --command=goto --target=3

# Verify
go run cmd/migrate/main.go --command=version
# Output: Current version: 3
```

### Scenario 5: Migration Failed (Dirty State)

```bash
# Check status
go run cmd/migrate/main.go --command=version
# Output: Current version: 3, Dirty: true

# Force to last known good version
go run cmd/migrate/main.go --command=force --version=2

# Then re-run migrations
go run cmd/migrate/main.go --command=up
```

---

## Best Practices

### ✅ DO

1. **Always create both UP and DOWN migrations** together
2. **Test migrations in development** before production
3. **Backup database** before running migrations in production
4. **Make migrations reversible** when possible
5. **Keep migrations small** and focused on one change

### ❌ DON'T

1. **Never edit a migration** that's already been run in production
2. **Don't delete migration files** after they've been applied
3. **Don't skip migration versions** - they must be sequential
4. **Don't manually edit the `schema_migrations` table**

---

## Creating New Migrations

### Step 1: Create migration files

```bash
# In store/migrations/ directory
touch 000003_add_user_settings.up.sql
touch 000003_add_user_settings.down.sql
```

### Step 2: Write UP migration

```sql
-- 000003_add_user_settings.up.sql
ALTER TABLE users ADD COLUMN settings JSONB DEFAULT '{}';
CREATE INDEX idx_users_settings ON users USING gin(settings);
```

### Step 3: Write DOWN migration

```sql
-- 000003_add_user_settings.down.sql
DROP INDEX IF EXISTS idx_users_settings;
ALTER TABLE users DROP COLUMN IF EXISTS settings;
```

### Step 4: Restart app or run manually

```bash
# Automatic when app starts
go run main.go

# Or manual
go run cmd/migrate/main.go --command=up
```

---

## Migration States

| State       | Description                               | Action            |
| ----------- | ----------------------------------------- | ----------------- |
| **Clean**   | All migrations applied successfully       | Normal operation  |
| **Pending** | New migration files exist but not applied | Run UP            |
| **Dirty**   | Migration failed halfway                  | FORCE then retry  |
| **Down**    | Need to rollback                          | Run DOWN or STEPS |

---

## Troubleshooting

### Problem: "Migration is dirty"

```bash
# Solution: Force to last good version
go run cmd/migrate/main.go --command=force --version=1
go run cmd/migrate/main.go --command=up
```

### Problem: "No change" error

```bash
# This is normal - means database is already at latest version
# No action needed
```

### Problem: Migration failed with SQL error

```bash
# 1. Check the error message
# 2. Fix the SQL in the migration file
# 3. Force to version before the failed one
go run cmd/migrate/main.go --command=force --version=X
# 4. Re-run migrations
go run cmd/migrate/main.go --command=up
```
