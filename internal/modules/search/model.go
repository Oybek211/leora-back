package search

// Result describes a generic search hit.
type Result struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}
