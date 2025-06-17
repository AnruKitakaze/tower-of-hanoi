# Tower of Hanoi

A puzzle game in development.

## How to play

1. Setup container with PostgreSQL for persistant storage:

```bash
docker run --name container_name -e POSTGRES_USER=user_here \
-e POSTGRES_PASSWORD=password_here -e POSTGRES_DB=db_name -p 5432:5432 -d postgres
```

2. Make sure settings match DSN for database here (later will move to env)
   `internal/infrastructure/persistance/postgresql/repository.go`

3. Run app itself `go run ./cmd/cli/main.go`
