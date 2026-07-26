package engine

import (
	"context"
	"fmt"

	"m-macdonald/mkv-mapper/internal/discdb"
	"m-macdonald/mkv-mapper/internal/event"
	"m-macdonald/mkv-mapper/internal/files"
	"m-macdonald/mkv-mapper/internal/makemkv"
	"m-macdonald/mkv-mapper/internal/makemkv/lines"
	"m-macdonald/mkv-mapper/internal/model"
	"m-macdonald/mkv-mapper/internal/signature"

	"go.uber.org/zap"
)

type Engine struct {
	makemkv      *makemkv.Client
	discdb       *discdb.CachedClient
	discResolver *files.Resolver
	selector     model.Selector
	logger       *zap.SugaredLogger
}

type EngineEventSink func(event.Event)

func New(
	makemkv *makemkv.Client,
	discdb *discdb.CachedClient,
	discResolver *files.Resolver,
	logger *zap.SugaredLogger,
	selector model.Selector,
) *Engine {
	return &Engine{
		makemkv:      makemkv,
		discdb:       discdb,
		discResolver: discResolver,
		logger:       logger,
		selector:     selector,
	}
}

func (e *Engine) ResolveTitlesForSource(ctx context.Context, source string, titles []model.TitlePlan) ([]model.TitlePlan, error) {
	disc, err := e.makemkv.ReadDisc(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("scanning rip source for title resolution: %w", err)
	}

	bySignature := make(map[signature.SegmentSignature]makemkv.Title, len(disc.Titles))
	for _, title := range disc.Titles {
		bySignature[title.Signature] = title
	}

	resolved := make([]model.TitlePlan, 0, len(titles))
	for _, titlePlan := range titles {
		title, ok := bySignature[titlePlan.SegmentSignature]
		if !ok {
			return nil, fmt.Errorf("title %q (signature %s) not found when re-scanning %s", titlePlan.FinalName, titlePlan.SegmentSignature, source)
		}

		titlePlan.TitleId = title.TitleId
		resolved = append(resolved, titlePlan)
	}
	return resolved, nil
}

func parsedLineToEvent(line lines.ParsedLine) (event.Event, bool) {
	switch l := line.(type) {
	case lines.ProgressValue:
		return event.ProgressPercentEvent{
			TotalPercent:   l.TotalPercent(),
			CurrentPercent: l.CurrentPercent(),
		}, true
	case lines.ProgressCurrent:
		return event.ProgressCurrentEvent{
			Message: l.Name,
		}, true
	case lines.ProgressTitle:
		return event.ProgressTotalEvent{
			Message: l.Name,
		}, true
	case lines.Message:
		return event.MessageEvent{
			Message: l.Message,
		}, true
	}

	return nil, false
}
