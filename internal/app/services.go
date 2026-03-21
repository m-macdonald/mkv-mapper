package app

import (
	"m-macdonald/mkv-mapper/internal/config"
	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/engine"
	"m-macdonald/mkv-mapper/internal/makemkv"

	"go.uber.org/zap"
)

type Services struct {
	Ripper *Ripper

	cache *discdb.SQLiteCache
}

type contextKey struct{}

var AppContextKey = contextKey{}

type AppContext struct {
	Config config.Config
	Logger *zap.SugaredLogger
}

func BuildServices(ctx AppContext) (*Services, error) {
	makemkvClient := makemkv.NewClient(
		ctx.Config.MakeMkvPath,
		ctx.Logger.Named("makemkv"),
	)

	cache, err := discdb.NewSQLiteCache(ctx.Config.CachePath)
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
		ctx.Logger.Named("pipeline"))

	ripper := NewRipper(engine, ctx.Logger.Named("ripper"))

	return &Services{
		Ripper: ripper,
		cache:  cache,
	}, nil
}

func (s *Services) Close() error {
	if s.cache != nil {
		return s.cache.Close()
	}

	return nil
}
