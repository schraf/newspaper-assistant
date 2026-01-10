package newspaper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/schraf/assistant/pkg/models"
)

type contextKey int

var assistantContextKey contextKey = 0
var optionsContextKey contextKey = 1

func withAssistant(ctx context.Context, assistant models.Assistant) context.Context {
	return context.WithValue(ctx, assistantContextKey, assistant)
}

func withOptions(ctx context.Context, options NewspaperOptions) context.Context {
	return context.WithValue(ctx, optionsContextKey, options)
}

func optionsFrom(ctx context.Context) NewspaperOptions {
	options, ok := ctx.Value(optionsContextKey).(NewspaperOptions)
	if !ok {
		return NewspaperOptions{}
	}

	return options
}

func ask(ctx context.Context, persona string, request string) (*string, error) {
	assistant, ok := ctx.Value(assistantContextKey).(models.Assistant)
	if !ok {
		return nil, fmt.Errorf("no assistant in context")
	}

	return assistant.Ask(ctx, persona, request)
}

func structuredAsk(ctx context.Context, persona string, request string, schema map[string]any) (json.RawMessage, error) {
	assistant, ok := ctx.Value(assistantContextKey).(models.Assistant)
	if !ok {
		return nil, fmt.Errorf("no assistant in context")
	}

	return assistant.StructuredAsk(ctx, persona, request, schema)
}
