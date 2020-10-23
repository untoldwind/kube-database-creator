package platforms

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"k8s.io/klog/v2"
)

const IllegalChars = " ();,.\"'"

type postgresPlatform struct {
	database *sql.DB
}

func newPostgresPlatform(databaseUrl string) (Platform, error) {
	database, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		return nil, err
	}
	return &postgresPlatform{
		database: database,
	}, nil
}

func (p *postgresPlatform) CheckExists(databaseName string) (bool, error) {
	klog.Infof("Postgres: Checking existence of %s", databaseName)

	var count int
	if err := p.database.QueryRow(`SELECT count(*) from pg_database where datname = $1`, databaseName).Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (p *postgresPlatform) Create(databaseName string, adminUsername string, adminPassword string) error {
	klog.Infof("Postgres: Creating of %s", databaseName)

	// Create statements cannot be prepared, so we have to provide at least a tiny amount of effort to prevent code-injection
	if strings.ContainsAny(databaseName, IllegalChars) {
		return fmt.Errorf("datatbase name contains illegal characters")
	}
	if strings.ContainsAny(adminUsername, IllegalChars) {
		return fmt.Errorf("database admin role name contains illegal characters")
	}

	if err := p.database.QueryRow(fmt.Sprintf(`CREATE DATABASE %s`, databaseName)).Err(); err != nil {
		return err
	}

	if err := p.database.QueryRow(fmt.Sprintf(`CREATE ROLE %s WITH LOGIN PASSWORD '%s'`, adminUsername, adminPassword)).Err(); err != nil {
		return err
	}

	if err := p.database.QueryRow(fmt.Sprintf(`GRANT ALL ON DATABASE %s TO %s`, databaseName, adminUsername)).Err(); err != nil {
		return err
	}

	return nil
}
