module github.com/besscroft/bophotos/backend

go 1.24.0

require github.com/mattn/go-sqlite3 v1.14.32

require github.com/golang-migrate/migrate/v4 v4.19.1

require (
	github.com/alexedwards/argon2id v1.0.0
	github.com/alexedwards/scs/sqlite3store v0.0.0-20251002162104-209de6e426de
	github.com/alexedwards/scs/v2 v2.9.0
	github.com/go-chi/chi/v5 v5.3.0
)

require (
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
)
