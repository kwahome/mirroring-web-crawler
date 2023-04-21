package main

import (
	"fmt"
	"regexp"
	"strings"
)

const linkPattern = `%s="(.*?)"`

// we can add more resource linking attributes such as src
var htmlAttributesToParse = []string{"href", "src"}

var unwantedPrefixes = []string{"font-family:", "#popup:", "'"}

func findChildrenLinks(data []byte, baseUrl string) ([]string, error) {
	links := make([]string, 0)

	for _, attribute := range htmlAttributesToParse {

		regexPattern := fmt.Sprintf(linkPattern, attribute)

		matches := regexp.MustCompile(regexPattern).FindAllString(string(data), -1)

		for _, match := range matches {
			// clean the linked url
			cleanedLink := strings.TrimPrefix(match, attribute+`="`)[:(len(match)-len(attribute+`="`))-1]

			for _, invalidPrefix := range unwantedPrefixes {
				if strings.HasPrefix(cleanedLink, invalidPrefix) {
					continue
				}
			}

			// make the cleanedLink an absolute url if relative
			if strings.HasPrefix(cleanedLink, "/") {
				cleanedLink = strings.TrimSuffix(baseUrl, "/") + cleanedLink
			}

			// valid links have the baseUrl as their prefix
			if !strings.HasPrefix(cleanedLink, baseUrl) {
				continue
			}

			fmt.Printf("extracted valid child url: '%s'\n", cleanedLink)

			links = append(links, cleanedLink)
		}
	}

	return links, nil
}
