package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func fetchUrl(url string) ([]byte, error) {
	fmt.Printf("fetching '%s'\n", url)

	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("an error has occurred: '%v'\n", err)

		return nil, err
	}

	defer closeIOReader(response.Body)

	if response.StatusCode != http.StatusOK {
		fmt.Printf("the request to fetch '%s' returned a non-success status '%v'\n", url, response.StatusCode)

		return nil, err
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("an error has occurred: '%v'\n", err)

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
