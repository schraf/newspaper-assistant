package newspaper

type ResearchDepth int

const (
	ResearchDepthShort  ResearchDepth = 0
	ResearchDepthMedium ResearchDepth = 1
	ResearchDepthLong   ResearchDepth = 2
)

func (d ResearchDepth) Validate() bool {
	switch d {
	case ResearchDepthShort, ResearchDepthMedium, ResearchDepthLong:
		return true
	default:
		return false
	}
}

// SectionType represents the fixed high-level sections of the newspaper.
type SectionType string

const (
	SectionLocal         SectionType = "local"
	SectionUS            SectionType = "us"
	SectionWorld         SectionType = "world"
	SectionBusiness      SectionType = "business"
	SectionTechnology    SectionType = "technology"
	SectionHealthScience SectionType = "health_science"
)

// NewspaperOptions controls generation of a newspaper edition.
type NewspaperOptions struct {
	// DaysBack is the number of days in the past to start considering news from.
	// The end of the window is always "now" (current date).
	DaysBack int
	// Location for the Local section (typically a US state, but can be more general).
	Location string
	// Depth/length of the newspaper (short/medium/long).
	Depth ResearchDepth
}

// ArticlePlan describes a single news story to be researched and written.
type ArticlePlan struct {
	Headline  string   `json:"headline"`
	Slug      string   `json:"slug"`
	Summary   string   `json:"summary"`
	Questions []string `json:"questions"`
}

// SectionPlan lists the articles to cover for a given section.
type SectionPlan struct {
	Type         SectionType   `json:"section_type"`
	Title        string        `json:"title"`
	ArticlePlans []ArticlePlan `json:"articles"`
}

// NewspaperPlan represents the overall plan for an edition.
type NewspaperPlan struct {
	DateRange string        `json:"date_range"`
	Location  string        `json:"location"`
	Sections  []SectionPlan `json:"sections"`
}

// ArticleResearch contains the knowledge gathered for a single article.
type ArticleResearch struct {
	Plan      ArticlePlan
	Knowledge []Knowledge
	Error     error
}

// SectionResearch aggregates research for a section.
type SectionResearch struct {
	Plan     SectionPlan
	Articles []ArticleResearch
}
