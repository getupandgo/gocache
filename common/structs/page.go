package structs

type (
	Page struct {
		URL       string
		Content   []byte
		TTL       int64
		TotalSize int
	}

	ScoredPage struct {
		URL  string
		Hits int
	}
)
