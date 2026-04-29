package appcontext

import (
	"awesomeProject/internal/appcontext/database/postgres"
	"awesomeProject/logger"
)

func Initiate() error {
	logger.SetupLogger()
	err := postgres.SetupDatabase()
	if err != nil {
		return err
	}
	return nil
}
