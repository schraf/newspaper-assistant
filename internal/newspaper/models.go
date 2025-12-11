package newspaper

type NewspaperOptions struct {
	DaysBack int
	Location string
}

type Section struct {
	Title       string
	Description string
}

type Article struct {
	Section  string
	Headline string
	Summary  string
	Research string
	Body     string
}
