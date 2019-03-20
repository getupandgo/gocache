package structs

type (
	Page struct {
		URL     string "json:url"
		Content []byte "json:content"
	}

	RemovePageBody struct {
		URL string
	}
)
