package main

import (
	"strings"
)

// image formats and magic numbers
//https://en.wikipedia.org/wiki/Magic_number_(programming)

var magicTable = map[string]string{
	"\xff\xd8\xff":      "image/jpeg",
	"\x89PNG\r\n\x1a\n": "image/png",
	"GIF87a":            "image/gif",
	"GIF89a":            "image/gif",
}

func isImageMime(file []byte) string {
	for magic, mime := range magicTable {
		if strings.HasPrefix(string(file), magic) {
			return mime
		}
	}

	return ""
}
