package generator

import (
	"context"
	"fmt"

	"github.com/schraf/assistant/pkg/generators"
	"github.com/schraf/assistant/pkg/models"
	"github.com/schraf/newspaper-assistant/internal/newspaper"
)

func init() {
	generators.MustRegister("newspaper", factory)
}

func factory(generators.Config) (models.ContentGenerator, error) {
	return &generator{}, nil
}

type generator struct{}

func (g *generator) Generate(ctx context.Context, request models.ContentRequest, assistant models.Assistant) (*models.Document, error) {
	daysBack, ok := toInt(request.Body["days_back"])
	if !ok {
		return nil, fmt.Errorf("no 'days_back' provided (expected positive integer)")
	}

	if daysBack <= 0 {
		return nil, fmt.Errorf("invalid 'days_back' %d (must be positive)", daysBack)
	}

	maxLength, ok := toInt(request.Body["max_length"])
	if !ok {
		return nil, fmt.Errorf("no 'max_length' provided (expected positive integer)")
	}

	if maxLength <= 0 {
		return nil, fmt.Errorf("invalid 'max_length' %d (must be positive)", maxLength)
	}

	location, _ := request.Body["location"].(string)

	options := newspaper.NewspaperOptions{
		DaysBack:  daysBack,
		Location:  location,
		MaxLength: maxLength,
	}

	return newspaper.CreateNewspaper(ctx, assistant, options)
}

func toInt(value any) (int, bool) {
	switch typedValue := value.(type) {
	case int:
		return typedValue, true
	case int64:
		return int(typedValue), true
	case float64:
		return int(typedValue), true
	default:
		return 0, false
	}
}
