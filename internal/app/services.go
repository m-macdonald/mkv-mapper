package app

import (
	"errors"
	"fmt"
	"io"

	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/makemkv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Services struct {
	Engine *engine.Engine
	Logger *zap.SugaredLogger

	closers []io.Closer
}

func BuildServices(cfg config.Config) (*Services, error) {
	logger, err := initLogger(cfg)
	if err != nil {
		return nil, err
	}

	makemkvClient := makemkv.NewClient(
		cfg.MakeMkvPath,
		logger.Named("makemkv"),
	)

	cache, err := discdb.NewSQLiteCache(cfg.CachePath)
	if err != nil {
		return nil, err
	}

	remoteClient := discdb.NewRemoteClient()

	discdbClient, err := discdb.NewCachedClient(cache, remoteClient)
	if err != nil {
		return nil, err
	}

	engine := engine.New(
		makemkvClient,
		discdbClient,
		logger.Named("engine"))

	return &Services{
		Engine: engine,
		Logger:  logger,
		closers: []io.Closer{cache},
	}, nil
}

func (s *Services) Close() error {
	errs := make([]error, 0, len(s.closers))
	for i := len(s.closers) - 1; i >= 0; i-- {
		closer := s.closers[i]
		if closer != nil {
			err := closer.Close()
			errs = append(errs, err)
		}
	}

	if s.Logger != nil {
		err := s.Logger.Sync()
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func initLogger(config config.Config) (*zap.SugaredLogger, error) {
	logLevel, err := zapcore.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("parse log level: %s", config.LogLevel)
	}
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = zap.NewAtomicLevelAt(logLevel)

	// Sugaring the logger by default as this is code is not performance critical
	return zap.Must(loggerConfig.Build()).Sugar(), nil
}
