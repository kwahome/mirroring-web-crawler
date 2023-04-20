package crawler

import (
	"fmt"
	"regexp"
	"strings"
)

const linkPattern = `="(.*?)"`

var htmlAttributesToParse = []string{"href", "src"}

func findChildrenLinks(data []byte, baseUrl string) ([]string, error) {
	links := make([]string, 0)

	for _, attribute := range htmlAttributesToParse {

		matches := regexp.MustCompile(attribute+linkPattern).FindAllString(string(data), -1)

		for _, match := range matches {
			link := strings.TrimPrefix(match, attribute+`="`)[:(len(match)-len(attribute+`="`))-1]

			// if the link is relative, make it an absolute url
			if strings.HasPrefix(link, "/") {
				link = baseUrl + link
			}

			// valid links have the baseUrl as their prefix
			if strings.HasPrefix(link, baseUrl) {
				fmt.Printf("Found a child url: %s\n", link)

				links = append(links, link)
			}
		}
	}

	return links, nil
}
