package plugin

import (
	"bytes"
	"unsafe"

	gofumpt "mvdan.cc/gofumpt/format"
)

const (
	formatNoChange uint32 = iota
	formatChanged
	formatError
)

type Plugin struct {
	sharedBytes   []byte
	config        config
	diagnostics   []configDiagnostic
	formattedText []byte
	errorText     string
}

func (p *Plugin) setSharedBytes(data []byte) uint32 {
	p.sharedBytes = data
	return uint32(len(data))
}

func (p *Plugin) takeFromSharedBytes() []byte {
	result := p.sharedBytes
	p.sharedBytes = nil
	return result
}

func (p *Plugin) SharedBytesPtr() uint32 {
	if len(p.sharedBytes) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&p.sharedBytes[0])))
}

func (p *Plugin) ClearSharedBytes(size uint32) uint32 {
	if size == 0 {
		p.sharedBytes = nil
		return 0
	}
	p.sharedBytes = make([]byte, size)
	return uint32(uintptr(unsafe.Pointer(&p.sharedBytes[0])))
}

func (p *Plugin) LicenseText(text string) uint32 {
	return p.setSharedBytes([]byte(text))
}

func (p *Plugin) Info(version string) uint32 {
	var data []byte
	data = append(data, `{"name":"dprint-plugin-gofumpt","version":`...)
	data = appendJSONString(data, version)
	data = append(data, `,"configKey":"gofumpt","fileExtensions":`...)
	data = appendJSONStringArray(data, []string{"go"})
	data = append(data, `,"fileNames":[],"helpUrl":"https://github.com/jakebailey/dprint-plugin-gofumpt","configSchemaUrl":`...)
	data = appendJSONString(data, "https://plugins.dprint.dev/jakebailey/gofumpt/v"+version+"/schema.json")
	data = append(data, `,"updateUrl":"https://plugins.dprint.dev/jakebailey/gofumpt/latest.json"}`...)
	return p.setSharedBytes(data)
}

func (p *Plugin) SetFilePath() {
	_ = p.takeFromSharedBytes()
}

func (p *Plugin) Format() uint32 {
	input := p.takeFromSharedBytes()
	if len(input) == 0 {
		return formatNoChange
	}

	opts := gofumpt.Options{
		LangVersion: p.config.LangVersion,
		ModulePath:  p.config.ModulePath,
		Extra: gofumpt.Extra{
			GroupParams:   p.config.ExtraRules || p.config.Extra.GroupParams,
			ClotheReturns: p.config.ExtraRules || p.config.Extra.ClotheReturns,
			BalanceCalls:  p.config.ExtraRules || p.config.Extra.BalanceCalls,
		},
	}

	output, err := gofumpt.Source(input, opts)
	if err != nil {
		p.errorText = err.Error()
		return formatError
	}

	if bytes.Equal(output, input) {
		return formatNoChange
	}

	p.formattedText = output
	return formatChanged
}

func (p *Plugin) FormattedText() uint32 {
	result := p.formattedText
	p.formattedText = nil
	return p.setSharedBytes(result)
}

func (p *Plugin) ErrorText() uint32 {
	result := []byte(p.errorText)
	p.errorText = ""
	return p.setSharedBytes(result)
}
