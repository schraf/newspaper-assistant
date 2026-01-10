package newspaper

import "fmt"

func CreateSections(options NewspaperOptions) []Section {
	return []Section{
		{
			Title:       "Local News",
			Description: fmt.Sprintf("Local news stories from %s.", options.Location),
		},
		{
			Title:       "US News",
			Description: "Major news stories from across the United States.",
		},
		{
			Title:       "World News",
			Description: "Significant international events and developments.",
		},
		{
			Title:       "Business and Financial",
			Description: "Business, markets, and financial news.",
		},
		{
			Title:       "Technology",
			Description: "Technology industry, innovation, and digital trends.",
		},
		{
			Title:       "Health and Science",
			Description: "Health, medicine, and scientific discoveries.",
		},
	}
}
