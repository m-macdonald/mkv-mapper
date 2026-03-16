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
}

type contextKey struct{}

var AppContextKey = contextKey{}

type AppContext struct {
	Config *config.Config
	Logger *zap.SugaredLogger
}

func BuildServices(ctx AppContext) *Services {
	makemkvClient := makemkv.NewClient(
		ctx.Config.MakeMkvPath,
		ctx.Logger.Named("makemkv"),
	)

	discdbClient := discdb.NewClient()

	engine := engine.New(
		makemkvClient,
		discdbClient,
		ctx.Logger.Named("pipeline"))

	ripper := NewRipper(engine, ctx.Logger.Named("ripper"))

	return &Services{
		Ripper: ripper,
	}
}
