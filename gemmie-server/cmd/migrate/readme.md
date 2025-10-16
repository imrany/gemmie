1. First, do a dry run to see what data will be migrated:
```bash
go run cmd/migrate/main.go \
  --json ./gemmie_data.json \
  --db-host localhost \
  --db-port 5432 \
  --db-user gemmie_user \
  --db-password your_password \
  --db-name gemmie \
  --dry-run
```

This will show you:

- How many records of each type exist
- Sample data preview
- No actual data insertion

2. Run Actual Migration

Once you're satisfied, run the real migration:
```bash
go run cmd/migrate/main.go  --json ./gemmie_data.json --db-host localhost --db-port 5432 --db-user gemmie_user --db-password your_password --db-name gemmie
```

Or

```bash
# Create a migration.env file
DB_HOST=localhost
DB_PORT=5432
DB_USER=gemmie_user
DB_PASSWORD=your_password
DB_NAME=gemmie
```

Then run
```bash
export $(cat migration.env | xargs) && go run cmd/migrate/main.go --json ./gemmie_data.json
```