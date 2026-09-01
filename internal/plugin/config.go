package plugin

import (
	"encoding/json"
	"fmt"
	goversion "go/version"
)

type config struct {
	LangVersion string      `json:"langVersion"`
	ModulePath  string      `json:"modulePath"`
	ExtraRules  bool        `json:"extraRules"`
	Extra       extraConfig `json:"extra"`
}

type extraConfig struct {
	GroupParams   bool `json:"groupParams"`
	ClotheReturns bool `json:"clotheReturns"`
	BalanceCalls  bool `json:"balanceCalls"`
}

type configDiagnostic struct {
	PropertyName string `json:"propertyName"`
	Message      string `json:"message"`
}

func (p *Plugin) RegisterConfig() {
	data := p.takeFromSharedBytes()
	if data == nil {
		return
	}

	p.diagnostics = nil
	p.config = config{}

	var raw struct {
		Plugin config `json:"plugin"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		p.diagnostics = []configDiagnostic{{
			PropertyName: "gofumpt",
			Message:      err.Error(),
		}}
		return
	}
	p.config = raw.Plugin

	if p.config.LangVersion != "" && goversion.Lang(p.config.LangVersion) == "" {
		p.diagnostics = append(p.diagnostics, configDiagnostic{
			PropertyName: "langVersion",
			Message:      fmt.Sprintf("invalid Go version: %q", p.config.LangVersion),
		})
		p.config.LangVersion = ""
	}
}

func (p *Plugin) ReleaseConfig() {
	p.config = config{}
	p.diagnostics = nil
}

func (p *Plugin) ConfigDiagnostics() uint32 {
	if len(p.diagnostics) == 0 {
		return p.setSharedBytes([]byte("[]"))
	}
	data, _ := json.Marshal(p.diagnostics)
	return p.setSharedBytes(data)
}

func (p *Plugin) ResolvedConfig() uint32 {
	data, _ := json.Marshal(p.config)
	return p.setSharedBytes(data)
}

func (p *Plugin) ConfigFileMatching() uint32 {
	return p.setSharedBytes([]byte(`{"fileExtensions":["go"],"fileNames":[]}`))
}

func (p *Plugin) CheckConfigUpdates() uint32 {
	_ = p.takeFromSharedBytes()
	return p.setSharedBytes([]byte(`{"kind":"ok","data":[]}`))
}

func (p *Plugin) SetOverrideConfig() {
	_ = p.takeFromSharedBytes()
}
