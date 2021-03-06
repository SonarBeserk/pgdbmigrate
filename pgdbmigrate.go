package pgdbmigrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Database interface needs to be implemented to migrate a new type of database

type Database interface {
	CreateMigrationsTable() error
	HasMigrated(filename string) (bool, error)
	Migrate(filename string, migration string) error
}

// PostgresDatabase migrates Postgresql databases

type PostgresDatabase struct {
	database *sql.DB
}

func (postgres *PostgresDatabase) CreateMigrationsTable() error {
	_, err := postgres.database.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id serial, 
			name text NOT NULL, 
			created_at timestamp with time zone NOT NULL,
			CONSTRAINT "PK_migrations_id" PRIMARY KEY (id)
	);`)
	if err != nil {
		return err
	}
	_, err = postgres.database.Exec("create unique index idx_migrations_name on migrations(name)")
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return err
		}
	}
	return nil
}

func (postgres *PostgresDatabase) HasMigrated(filename string) (bool, error) {
	var count int
	err := postgres.database.QueryRow("select count(1) from migrations where name = $1", filename).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (postgres *PostgresDatabase) Migrate(filename string, migration string) error {
	_, err := postgres.database.Exec(migration)
	if err != nil {
		return err
	}
	_, err = postgres.database.Exec("insert into migrations(name, created_at) values($1, current_timestamp)", filename)
	return err
}

func NewPostgresDatabase(db *sql.DB) *PostgresDatabase {
	return &PostgresDatabase{database: db}
}

// By default, apply Postgresql migrations, as in older versions
func Run(db *sql.DB, migrationsFolder string) error {
	postgres := NewPostgresDatabase(db)
	return ApplyMigrations(postgres, migrationsFolder)
}

// Run applies migrations from migrationsFolder to database.
func ApplyMigrations(database Database, migrationsFolder string) error {
	// Initialize migrations table, if it does not exist yet
	if err := database.CreateMigrationsTable(); err != nil {
		return err
	}

	// Scan migration file names in migrations folder
	d, err := os.Open(migrationsFolder)
	if err != nil {
		return err
	}
	dir, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	// Run migrations
	sqlFiles := make([]string, 0)
	for _, f := range dir {
		ext := filepath.Ext(f.Name())
		if ".sql" == ext || ".cql" == ext {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)
	for _, filename := range sqlFiles {
		// if exists in migrations table, leave it
		// else execute sql
		migrated, err := database.HasMigrated(filename)
		if err != nil {
			return err
		}
		fullpath := filepath.Join(migrationsFolder, filename)
		if migrated {
			fmt.Println("Already migrated", fullpath)
			continue
		}
		b, err := ioutil.ReadFile(fullpath)
		if err != nil {
			return err
		}
		migration := string(b)
		if len(migration) == 0 {
			fmt.Println("Skipping empty file", fullpath)
			continue // empty file
		}
		err = database.Migrate(filename, migration)
		if err != nil {
			return err
		}
		fmt.Println("Migrated", fullpath)
	}

	return nil
}
