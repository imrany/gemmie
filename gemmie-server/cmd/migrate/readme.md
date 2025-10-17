### 1. Manual Control with CLI
```bash
# Check current version
go run cmd/migrate/main.go -command=version

# Apply all pending migrations
go run cmd/migrate/main.go -command=up

# Rollback last migration
go run cmd/migrate/main.go -command=down

# Rollback 2 migrations
go run cmd/migrate/main.go -command=steps -steps=-2

# Go to specific version (e.g., version 1)
go run cmd/migrate/main.go -command=goto -target=1

# Fix dirty state (force to version 2)
go run cmd/migrate/main.go -command=force -version=2
```