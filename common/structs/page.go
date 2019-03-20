package structs

type (
	Page struct {
		Url     string "json:url"
		Content []byte "json:content"
	}

	RemovePageBody struct {
		Url string
	}
)
