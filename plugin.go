//go:build tinygo

// dprint wasm plugin (schema version 4) for gofumpt.
// See: https://github.com/dprint/dprint/blob/main/docs/wasm-plugin-development.md
package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"unsafe"

	gofumpt "mvdan.cc/gofumpt/format"
)

func main() {}

var (
	sharedBytes []byte

	formattedText []byte
	errorText     string

	config      pluginConfig
	diagnostics []configDiagnostic
)

type pluginConfig struct {
	LangVersion string `json:"langVersion"`
	ModulePath  string `json:"modulePath"`
	ExtraRules  bool   `json:"extraRules"`
}

type configDiagnostic struct {
	PropertyName string `json:"propertyName"`
	Message      string `json:"message"`
}

func setSharedBytes(data []byte) uint32 {
	sharedBytes = data
	return uint32(len(data))
}

func takeFromSharedBytes() []byte {
	result := sharedBytes
	sharedBytes = nil
	return result
}

//go:wasmexport get_shared_bytes_ptr
func get_shared_bytes_ptr() uint32 {
	if len(sharedBytes) == 0 {
		return 0
	}
	return uint32(uintptr(unsafe.Pointer(&sharedBytes[0])))
}

//go:wasmexport clear_shared_bytes
func clear_shared_bytes(size uint32) uint32 {
	sharedBytes = make([]byte, size)
	return uint32(uintptr(unsafe.Pointer(&sharedBytes[0])))
}

//go:wasmexport dprint_plugin_version_4
func dprint_plugin_version_4() uint32 {
	return 4
}

//go:embed LICENSE
var licenseText string

//go:wasmexport get_license_text
func get_license_text() uint32 {
	return setSharedBytes([]byte(licenseText))
}

//go:wasmexport get_plugin_info
func get_plugin_info() uint32 {
	info := struct {
		Name            string   `json:"name"`
		Version         string   `json:"version"`
		ConfigKey       string   `json:"configKey"`
		FileExtensions  []string `json:"fileExtensions"`
		FileNames       []string `json:"fileNames"`
		HelpURL         string   `json:"helpUrl"`
		ConfigSchemaURL string   `json:"configSchemaUrl"`
	}{
		Name:            "dprint-plugin-gofumpt",
		Version:         "0.0.1",
		ConfigKey:       "gofumpt",
		FileExtensions:  []string{"go"},
		FileNames:       []string{},
		HelpURL:         "https://github.com/jakebailey/dprint-plugin-gofumpt",
		ConfigSchemaURL: "https://plugins.dprint.dev/jakebailey/gofumpt/v0.0.1/schema.json",
	}
	data, _ := json.Marshal(info)
	return setSharedBytes(data)
}

//go:wasmexport register_config
func register_config(_ uint32) {
	data := takeFromSharedBytes()
	if data == nil {
		return
	}

	diagnostics = nil
	config = pluginConfig{}

	var raw struct {
		Plugin pluginConfig `json:"plugin"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		diagnostics = []configDiagnostic{{
			PropertyName: "gofumpt",
			Message:      err.Error(),
		}}
		return
	}
	config = raw.Plugin
}

//go:wasmexport release_config
func release_config(_ uint32) {
	config = pluginConfig{}
	diagnostics = nil
}

//go:wasmexport get_config_diagnostics
func get_config_diagnostics(_ uint32) uint32 {
	if len(diagnostics) == 0 {
		return setSharedBytes([]byte("[]"))
	}
	data, _ := json.Marshal(diagnostics)
	return setSharedBytes(data)
}

//go:wasmexport get_resolved_config
func get_resolved_config(_ uint32) uint32 {
	data, _ := json.Marshal(config)
	return setSharedBytes(data)
}

//go:wasmexport get_config_file_matching
func get_config_file_matching(_ uint32) uint32 {
	return setSharedBytes([]byte(`{"fileExtensions":["go"],"fileNames":[]}`))
}

//go:wasmexport set_file_path
func set_file_path() {}

//go:wasmexport set_override_config
func set_override_config() {}

//go:wasmexport format
func format(_ uint32) uint32 {
	input := takeFromSharedBytes()
	if input == nil {
		return 0
	}

	opts := gofumpt.Options{
		LangVersion: config.LangVersion,
		ModulePath:  config.ModulePath,
		ExtraRules:  config.ExtraRules,
	}

	output, err := gofumpt.Source(input, opts)
	if err != nil {
		errorText = err.Error()
		return 2
	}

	if bytes.Equal(output, input) {
		return 0
	}

	formattedText = output
	return 1
}

//go:wasmexport get_formatted_text
func get_formatted_text() uint32 {
	result := formattedText
	formattedText = nil
	return setSharedBytes(result)
}

//go:wasmexport get_error_text
func get_error_text() uint32 {
	result := errorText
	errorText = ""
	return setSharedBytes([]byte(result))
}
