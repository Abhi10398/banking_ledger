package postgres

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"awesomeProject/config"
	"awesomeProject/internal/appcontext/database"
)

var gormDb *gorm.DB
var sqlLogger gormLogger.Interface

func SetupDatabase() error {
	sqlLogger = &database.Dblogger{}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  buildDSN(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: sqlLogger,
	})
	if err != nil {
		return err
	}

	gormDb = db
	sqlDb, err := db.DB()
	if err != nil {
		return err
	}

	dbConf := config.PostgresConf()
	sqlDb.SetMaxIdleConns(int(dbConf.MaxIdleConn))
	sqlDb.SetMaxOpenConns(int(dbConf.MaxOpenConn))
	sqlDb.SetConnMaxIdleTime(time.Duration(dbConf.ConnMaxIdleTime) * time.Minute)

	return nil
}

// buildDSN returns DATABASE_URL when set (Docker / 12-factor), otherwise
// assembles the DSN from the static config file (local development).
func buildDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	dbConf := config.PostgresConf()
	return fmt.Sprintf("dbname=%s host=%s port=%d sslmode=disable",
		config.PostgresDbName(), dbConf.Host, dbConf.Port)
}

func Close() {
	if gormDb != nil {
		if sqlDb, err := gormDb.DB(); err == nil {
			sqlDb.Close()
		}
	}
}

func GetDB() *gorm.DB {
	return gormDb
}
