//go:build migrate

package app

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 20
	_defaultTimeout  = time.Second
)

const DB_PATH_ENV = "STORAGE_PATH"
const DB_MIGRATIONS_PATH_ENV = "STORAGE_MIGRATIONS_PATH"

func init() {
	databasePath, ok := os.LookupEnv(DB_PATH_ENV)
	if !ok || len(databasePath) == 0 {
		log.Fatalf("migrate: environment variable not declared: %s", DB_PATH_ENV)
	}
	migrationsPath, ok := os.LookupEnv(DB_MIGRATIONS_PATH_ENV)
	if !ok || len(migrationsPath) == 0 {
		log.Fatalf("migrate: environment variable not declared: %s", DB_MIGRATIONS_PATH_ENV)
	}

	var (
		attempts = _defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New(
			fmt.Sprintf("file://%s", migrationsPath),
			fmt.Sprintf("sqlite://%s", databasePath),
		)
		if err == nil {
			break
		}

		log.Printf("Migrate: can't connect, attempts left: %d", attempts)
		time.Sleep(_defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Fatalf("Migrate: postgres connect error: %s", err)
	}

	err = m.Up()
	defer m.Close()

	if errors.Is(err, migrate.ErrNoChange) {
		log.Printf("Migrate: no change")
		return
	}

	if err != nil {
		log.Fatalf("Migrate: up error: %s", err)
	}

	log.Printf("Migrate: up success")
}
