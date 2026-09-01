package plugin

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAppendJSONString(t *testing.T) {
	for _, value := range []string{
		"",
		"plain text",
		"quotes: \"\\",
		"controls: \x00\b\f\n\r\t\x1f",
		"html: <>&",
		"unicode: \u2028\u2029 \U0001f642",
		"invalid: \xff",
	} {
		got := appendJSONString(nil, value)
		var decoded string
		if err := json.Unmarshal(got, &decoded); err != nil {
			t.Errorf("appendJSONString(%q) produced invalid JSON %s: %v", value, got, err)
			continue
		}
		want := strings.ToValidUTF8(value, "\ufffd")
		if decoded != want {
			t.Errorf("appendJSONString(%q) decoded to %q, want %q", value, decoded, want)
		}
	}
}
