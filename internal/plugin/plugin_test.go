package plugin

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	var p Plugin
	var info struct {
		Name            string   `json:"name"`
		Version         string   `json:"version"`
		ConfigKey       string   `json:"configKey"`
		FileExtensions  []string `json:"fileExtensions"`
		ConfigSchemaURL string   `json:"configSchemaUrl"`
	}
	if size := p.Info("1.2.3"); size != uint32(len(p.sharedBytes)) {
		t.Fatalf("Info() = %d, want %d", size, len(p.sharedBytes))
	}
	if err := json.Unmarshal(p.takeFromSharedBytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.Name != "dprint-plugin-gofumpt" ||
		info.Version != "1.2.3" ||
		info.ConfigKey != "gofumpt" ||
		len(info.FileExtensions) != 1 ||
		info.FileExtensions[0] != "go" ||
		!strings.Contains(info.ConfigSchemaURL, "/v1.2.3/") {
		t.Fatalf("unexpected plugin info: %#v", info)
	}
}

func TestFormat(t *testing.T) {
	var p Plugin
	input := []byte("package main\n\nfunc  main() {}\n")
	p.setSharedBytes(input)
	if result := p.Format(); result != formatChanged {
		t.Fatalf("Format() = %d, want %d", result, formatChanged)
	}
	p.FormattedText()
	formatted := p.takeFromSharedBytes()
	if string(formatted) != "package main\n\nfunc main() {}\n" {
		t.Fatalf("unexpected formatted text: %q", formatted)
	}
	p.FormattedText()
	if formatted := p.takeFromSharedBytes(); formatted != nil {
		t.Fatalf("TakeFormattedText() did not clear the result: %q", formatted)
	}

	p.setSharedBytes(formatted)
	if result := p.Format(); result != formatNoChange {
		t.Fatalf("Format() = %d, want %d", result, formatNoChange)
	}

	p.setSharedBytes([]byte("not go"))
	if result := p.Format(); result != formatError {
		t.Fatalf("Format() = %d, want %d", result, formatError)
	}
	p.ErrorText()
	if errorText := p.takeFromSharedBytes(); len(errorText) == 0 {
		t.Fatal("TakeErrorText() returned an empty error")
	}
	p.ErrorText()
	if errorText := p.takeFromSharedBytes(); len(errorText) != 0 {
		t.Fatalf("TakeErrorText() did not clear the error: %q", errorText)
	}
}

func TestSharedBytes(t *testing.T) {
	var p Plugin
	if ptr := p.SharedBytesPtr(); ptr != 0 {
		t.Fatalf("SharedBytesPtr() = %d, want 0", ptr)
	}
	if ptr := p.ClearSharedBytes(4); ptr == 0 {
		t.Fatal("ClearSharedBytes() returned a zero pointer")
	}
	if len(p.sharedBytes) != 4 {
		t.Fatalf("shared buffer length = %d, want 4", len(p.sharedBytes))
	}
	copy(p.sharedBytes, "test")
	if data := p.takeFromSharedBytes(); string(data) != "test" {
		t.Fatalf("takeFromSharedBytes() = %q, want test", data)
	}
	if ptr := p.SharedBytesPtr(); ptr != 0 {
		t.Fatalf("SharedBytesPtr() after take = %d, want 0", ptr)
	}
}
