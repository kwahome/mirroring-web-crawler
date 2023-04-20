package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func getUrl(url string) ([]byte, error) {
	fmt.Printf("Fetching: '%s'\n", url)

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("An error has occurred: '%v'\n", err)

		return nil, err
	}

	defer closeIOReader(response.Body)

	if response.StatusCode != http.StatusOK {
		fmt.Printf("The request to fetch '%s' returned a non-success status '%s'\n", url, response.StatusCode)

		return nil, err
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("An error has occurred: '%v'\n", err)

		return nil, err
	}

	return content, nil
}

func closeIOReader(reader io.ReadCloser) {
	err := reader.Close()
	if err != nil {
		log.Fatalf("an error has occurred while closing io reader: %v\n", err)
	}
}
