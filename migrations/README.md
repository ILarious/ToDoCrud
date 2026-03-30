# Migrations

This directory uses `golang-migrate` file format:

- `*.up.sql` for applying changes
- `*.down.sql` for rollback

Current migration:

- `000001_init_schema`

## Run with local migrate binary

```bash
migrate -path ./migrations \
  -database "postgres://postgres:postgres@localhost:5432/todo?sslmode=disable" up
```

Rollback one step:

```bash
migrate -path ./migrations \
  -database "postgres://postgres:postgres@localhost:5432/todo?sslmode=disable" down 1
```

## Run with Docker (no local install)

```bash
docker run --rm \
  -v "$PWD/migrations:/migrations" \
  --add-host=host.docker.internal:host-gateway \
  migrate/migrate:v4.18.1 \
  -path=/migrations \
  -database "postgres://postgres:postgres@host.docker.internal:5432/todo?sslmode=disable" up
```
