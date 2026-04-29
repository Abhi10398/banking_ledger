package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func main() {
	jdbcURL, user, password := resolveDB()

	cmd := exec.Command(
		"liquibase",
		"--url="+jdbcURL,
		"--username="+user,
		"--password="+password,
		"--changeLogFile=changelog/db.changelog-master.yaml",
		"update",
	)
	cmd.Dir = "migrations/postgres"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
}

// resolveDB derives JDBC connection params from DATABASE_URL (Docker / CI) or
// falls back to individual PG_* env vars for local development.
func resolveDB() (jdbcURL, user, password string) {
	if raw := os.Getenv("DATABASE_URL"); raw != "" {
		u, err := url.Parse(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid DATABASE_URL: %v\n", err)
			os.Exit(1)
		}
		port := u.Port()
		if port == "" {
			port = "5432"
		}
		dbName := strings.TrimPrefix(u.Path, "/")
		pw, _ := u.User.Password()
		return fmt.Sprintf("jdbc:postgresql://%s:%s/%s", u.Hostname(), port, dbName),
			u.User.Username(), pw
	}

	host := envOrDefault("PG_HOST", "localhost")
	port := envOrDefault("PG_PORT", "5432")
	dbName := envOrDefault("PG_DBNAME", "banking_ledger")
	return fmt.Sprintf("jdbc:postgresql://%s:%s/%s", host, port, dbName),
		os.Getenv("PG_USER"),
		os.Getenv("PG_PASSWORD")
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
