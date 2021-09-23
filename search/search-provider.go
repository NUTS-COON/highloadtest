package search

type Result struct {
	Error error
	Items []ResourceItem
}

type ResourceItem struct {
	Host string
	Url  string
}

type Provider interface {
	Search(query string, useProxy bool) ([]ResourceItem, error)
}
