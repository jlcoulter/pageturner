package types
// TODO Move to sqlc
// BookEntry represents a book reading entry
type BookEntry struct {
	Book       string `json:"book"`
	Rating     int    `json:"rating"`
	StartDate  string `json:"startDate"`
	FinishDate string `json:"finishDate"`
	Pages      string `json:"pages"`
	Thoughts   string `json:"thoughts"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	SearchTerm string `json:"searchTerm"`
}