package auth

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

const (
	sessionCookieName = "bophotos_session"
	sessionAdminKey   = "administrator"
)

func NewSessionManager(db *sql.DB, secure bool) (*scs.SessionManager, func()) {
	store := sqlite3store.New(db)
	manager := scs.New()
	manager.Store = store
	manager.Lifetime = 24 * time.Hour
	manager.IdleTimeout = 30 * time.Minute
	manager.HashTokenInStore = true
	manager.Cookie.Name = sessionCookieName
	manager.Cookie.Path = "/"
	manager.Cookie.HttpOnly = true
	manager.Cookie.Persist = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = secure
	return manager, store.StopCleanup
}
