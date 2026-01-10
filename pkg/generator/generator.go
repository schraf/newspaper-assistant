package generator

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	section := newspaper.Section{
		Title:       strings.TrimSpace(toString(request.Body["section_title"])),
		Description: strings.TrimSpace(toString(request.Body["section_description"])),
	}

	if section.Title == "" {
		return nil, fmt.Errorf("no 'section_title' provided")
	}

	if section.Description == "" {
		return nil, fmt.Errorf("no 'section_description' provided")
	}

	options := newspaper.NewspaperOptions{
		DaysBack:  daysBack,
		MaxLength: maxLength,
	}

	doc, err := newspaper.CreateNewspaper(ctx, assistant, section, options)
	if err != nil {
		return nil, err
	}

	doc.Title = section.Title + ": " + dateRangeText(daysBack)

	return doc, nil
}

func toString(value any) string {
	valueString, _ := value.(string)
	return valueString
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

func dateRangeText(daysBack int) string {
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -daysBack)

	var dateRange string
	if start.Format("2006-01-02") == end.Format("2006-01-02") {
		dateRange = end.Format("Jan 2, 2006")
	} else {
		dateRange = fmt.Sprintf("%s to %s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
	}

	return dateRange
}
