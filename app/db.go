package app

import (
	"database/sql"

	log "github.com/swayrider/swlib/logger"
)

// DB is an interface for database connection wrappers.
// Implementations must provide access to the underlying *sql.DB connection.
type DB interface {
	SqlDB() *sql.DB
}

// DBCtor is a constructor function type for creating database connections.
// It receives the App instance for accessing configuration values.
//
// Example:
//
//	func createDB(a app.App) app.DB {
//	    connStr := fmt.Sprintf(
//	        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
//	        app.GetConfigField[string](a.Config(), app.KeyDBHost),
//	        app.GetConfigField[int](a.Config(), app.KeyDBPort),
//	        app.GetConfigField[string](a.Config(), app.KeyDBUser),
//	        app.GetConfigField[string](a.Config(), app.KeyDBPassword),
//	        app.GetConfigField[string](a.Config(), app.KeyDBName),
//	        app.GetConfigField[string](a.Config(), app.KeyDBSSLMode),
//	    )
//	    db, err := sql.Open("postgres", connStr)
//	    // ... error handling and wrapper creation
//	    return &MyDBWrapper{db: db}
//	}
type DBCtor func(App) DB

func (a *app) initDatabase() {
	lg := a.lg.Derive(log.WithFunction("initDatabase"))
	if a.dbCtor != nil {
		lg.Infoln("Configuring database connection")
		a.dbconn = a.dbCtor(a)
		if a.dbBootstrap != nil {
			if err := a.dbBootstrap(a); err != nil {
				lg.Fatalf("Failed to bootstrap database: %v", err)
			}
		}
	}
}
