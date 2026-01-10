package newspaper

type NewspaperOptions struct {
	DaysBack  int
	MaxLength int
	Location  string
}

type Section struct {
	Title       string
	Description string
}

type Article struct {
	Valid    bool
	Section  Section
	Headline string
	Summary  string
	Research string
	Body     string
}
