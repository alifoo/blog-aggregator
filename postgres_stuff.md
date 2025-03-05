postgres://postgres:postgres@localhost:5432/gator

## connection string
psql "postgres://postgres:postgres@localhost:5432/gator"

run migrations (must be inside sql/schema to work)
goose postgres "postgres://postgres:postgres@localhost:5432/gator" up
