package plugin

import (
	"encoding/json"
	"testing"
)

func TestConfig(t *testing.T) {
	var p Plugin
	p.setSharedBytes([]byte(`{"plugin":{"langVersion":"go1.27","modulePath":"example.com/mod","extraRules":true}}`))
	p.RegisterConfig()

	var resolved config
	p.ResolvedConfig()
	if err := json.Unmarshal(p.takeFromSharedBytes(), &resolved); err != nil {
		t.Fatal(err)
	}
	if resolved.LangVersion != "go1.27" || resolved.ModulePath != "example.com/mod" || !resolved.ExtraRules {
		t.Fatalf("unexpected resolved config: %#v", resolved)
	}
	p.ConfigDiagnostics()
	if diagnostics := p.takeFromSharedBytes(); string(diagnostics) != "[]" {
		t.Fatalf("unexpected diagnostics: %s", diagnostics)
	}

	p.ReleaseConfig()
	p.ResolvedConfig()
	if resolved := p.takeFromSharedBytes(); string(resolved) != `{"langVersion":"","modulePath":"","extraRules":false,"extra":{"groupParams":false,"clotheReturns":false,"balanceCalls":false}}` {
		t.Fatalf("unexpected released config: %s", resolved)
	}
}

func TestConfigDiagnostics(t *testing.T) {
	var p Plugin
	p.setSharedBytes([]byte(`{"plugin":{"langVersion":"invalid"}}`))
	p.RegisterConfig()

	var diagnostics []configDiagnostic
	p.ConfigDiagnostics()
	if err := json.Unmarshal(p.takeFromSharedBytes(), &diagnostics); err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 1 ||
		diagnostics[0].PropertyName != "langVersion" ||
		diagnostics[0].Message != `invalid Go version: "invalid"` {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}

	p.setSharedBytes([]byte(`{`))
	p.RegisterConfig()
	p.ConfigDiagnostics()
	if err := json.Unmarshal(p.takeFromSharedBytes(), &diagnostics); err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 1 || diagnostics[0].PropertyName != "gofumpt" {
		t.Fatalf("unexpected malformed JSON diagnostics: %#v", diagnostics)
	}
}
