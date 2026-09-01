package plugin

import (
	"strings"

	"github.com/buger/jsonparser"
)

func appendJSONString(dst []byte, value string) []byte {
	value = strings.ToValidUTF8(value, "\ufffd")
	return append(dst, jsonparser.Escape(value)...)
}

func appendJSONStringArray(dst []byte, values []string) []byte {
	dst = append(dst, '[')
	for i, value := range values {
		if i != 0 {
			dst = append(dst, ',')
		}
		dst = appendJSONString(dst, value)
	}
	return append(dst, ']')
}
