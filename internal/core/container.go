package core

import (
	"errors"
	"log/slog"

	"github.com/tkahng/playground/internal/conf"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/di"
)

func getConfig(contianer di.Container) (*conf.EnvConfig, error) {
	opts, ok := contianer.Get(diContextKeyConfig).(*conf.EnvConfig)
	if !ok {
		return nil, errors.New("failed to get config")
	}
	return opts, nil
}
func getSettings(contianer di.Container) (*conf.AppOptions, error) {
	opts, ok := contianer.Get(diContextKeySettings).(*conf.AppOptions)
	if !ok {
		return nil, errors.New("failed to get settings")
	}
	return opts, nil
}
func getLogger(contianer di.Container) (*slog.Logger, error) {
	logger, ok := contianer.Get(diContextKeyLogger).(*slog.Logger)
	if !ok {
		return nil, errors.New("failed to get logger")
	}
	return logger, nil
}

func getDbx(contianer di.Container) (database.Dbx, error) {
	dbx, ok := contianer.Get(diContextKeyDB).(database.Dbx)
	if !ok {
		return nil, errors.New("failed to get db")
	}
	return dbx, nil
}

func getAdapter(contianer di.Container) (stores.StorageAdapterInterface, error) {
	adapter, ok := contianer.Get(diContextKeyStoreAdapter).(stores.StorageAdapterInterface)
	if !ok {
		return nil, errors.New("failed to get store adapter")
	}
	return adapter, nil
}
