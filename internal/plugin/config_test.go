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

func TestConfigJSON(t *testing.T) {
	want := config{
		LangVersion: "go1.27",
		ModulePath:  `example.com/a"b`,
		ExtraRules:  true,
		Extra: extraConfig{
			GroupParams:   true,
			ClotheReturns: false,
			BalanceCalls:  true,
		},
	}

	data := []byte(`{"plugin":{"langVersion":"go1.27","modulePath":"example.com/a\"b","extraRules":true,"extra":{"groupParams":true,"clotheReturns":false,"balanceCalls":true}}}`)
	got, err := unmarshalConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("unmarshalConfig() = %#v, want %#v", got, want)
	}

	gotJSON := marshalConfig(got)
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("marshalConfig() = %s, want %s", gotJSON, wantJSON)
	}
}

func TestConfigJSONTypes(t *testing.T) {
	for _, test := range []struct {
		name string
		data string
	}{
		{name: "plugin", data: `{"plugin":true}`},
		{name: "langVersion", data: `{"plugin":{"langVersion":true}}`},
		{name: "extraRules", data: `{"plugin":{"extraRules":"true"}}`},
		{name: "extra", data: `{"plugin":{"extra":[]}}`},
		{name: "groupParams", data: `{"plugin":{"extra":{"groupParams":1}}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := unmarshalConfig([]byte(test.data)); err == nil {
				t.Fatal("unmarshalConfig() succeeded, want error")
			}
		})
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
