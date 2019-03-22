package structs

type (
	Page struct {
		URL       string
		Content   []byte
		TotalSize int
	}

	ScoredPage struct {
		URL  string
		Hits int
	}

	RemovePageBody struct {
		URL string
	}
)
