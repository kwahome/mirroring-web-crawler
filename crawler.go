package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

func crawl(targetUrl string, dir string, overwrite bool, target string) ([]string, error) {
	fmt.Printf("starting crawler on: '%s'\n", targetUrl)

	parsedUrl, err := url.Parse(targetUrl)
	if err != nil {
		fmt.Printf("An error has occurred: '%v'\n", err)

		return nil, err
	}

	urlPath := parsedUrl.Path
	if targetUrl == target {
		urlPath = "index.html"
	}

	filePath := filepath.Join(dir, urlPath)
	if filepath.Ext(urlPath) == "" {
		filePath += ".html"
	}

	// checks if resource already exists in which case
	// it is skipped unless the overwrite flag is set
	if _, err := os.Stat(filePath); !overwrite && targetUrl != target && err == nil {
		fmt.Printf("content for target url '%s' found under file path: '%s'\n", targetUrl, filePath)

		return nil, nil
	}

	content, err := fetchUrl(targetUrl)
	if err != nil {
		fmt.Printf("an error has occurred: '%v'\n", err)

		return nil, err
	}

	err = writeToFile(content, filePath)
	if err != nil {
		fmt.Printf("an error has occurred: '%v'\n", err)

		return nil, err
	}

	childrenLinks, err := findChildrenLinks(content, target)
	if err != nil {
		fmt.Printf("An error has occurred: '%v'\n", err)

		return nil, err
	}

	fmt.Printf("crawler has finished crawling: '%s'\n", targetUrl)

	return childrenLinks, nil
}
