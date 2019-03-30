package test_data

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

func PopulateUpsertReq(url string, pageURL string) (*http.Request, error) {
	samplePagePath := path.Join(os.Getenv("GOPATH"), "src/github.com/getupandgo/gocache/mocks/Example Domain.html")

	content, err := ioutil.ReadFile(samplePagePath)
	if err != nil {
		return nil, err
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	urlField, err := writer.CreateFormField("url")
	if err != nil {
		return nil, err
	}
	_, err = urlField.Write([]byte(pageURL))
	if err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("content", "example_file_name")
	if err != nil {
		return nil, err
	}
	_, err = part.Write(content)
	if err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	newReq, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	newReq.Header.Add("Content-Type", writer.FormDataContentType())

	return newReq, nil
}
