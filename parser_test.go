package main

import (
	"reflect"
	"testing"
)

func Test_parseLinks(t *testing.T) {
	webPage := `<a href="/abc/foo/img.jpg"></a>
				<a href="https://example.com/baz/page.html"></a>
				<a href="https://another.domain/abc/page2.html"></a>
				<a href="/baz/media.mp3"></a>
				<a href="https://domain.com/abc/music/video.mp3"></a>
				`

	t.Run("fetch all valid children links", func(t *testing.T) {
		baseUrl := "https://example.com"

		expected := make([]string, 0)
		expected = append(expected, "https://example.com/abc/foo/img.jpg")
		expected = append(expected, "https://example.com/baz/page.html")
		expected = append(expected, "https://example.com/baz/media.mp3")

		result, err := findChildrenLinks([]byte(webPage), baseUrl)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected: '%v' but got: '%v'", expected, result)
		}
	})

	t.Run("fetch all valid children links", func(t *testing.T) {
		baseUrl := "https://example.com/abc"

		expected := make([]string, 0)
		expected = append(expected, "https://example.com/abc/abc/foo/img.jpg")
		expected = append(expected, "https://example.com/abc/baz/media.mp3")

		result, err := findChildrenLinks([]byte(webPage), baseUrl)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected: '%v' but got: '%v'", expected, result)
		}
	})
}
