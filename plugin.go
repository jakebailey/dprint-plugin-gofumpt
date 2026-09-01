//go:build tinygo

// dprint wasm plugin (schema version 4) for gofumpt.
// See: https://github.com/dprint/dprint/blob/main/docs/wasm-plugin-development.md
package main

import (
	_ "embed"
	"strings"

	"github.com/jakebailey/dprint-plugin-gofumpt/internal/plugin"
)

func main() {}

var instance plugin.Plugin

//go:wasmexport get_shared_bytes_ptr
func get_shared_bytes_ptr() uint32 {
	return instance.SharedBytesPtr()
}

//go:wasmexport clear_shared_bytes
func clear_shared_bytes(size uint32) uint32 {
	return instance.ClearSharedBytes(size)
}

//go:wasmexport dprint_plugin_version_4
func dprint_plugin_version_4() uint32 {
	return 4
}

//go:embed metadata/LICENSES
var licenseText string

//go:embed metadata/VERSION
var rawVersion string

var version = strings.TrimSpace(rawVersion)

//go:wasmexport get_license_text
func get_license_text() uint32 {
	return instance.LicenseText(licenseText)
}

//go:wasmexport get_plugin_info
func get_plugin_info() uint32 {
	return instance.Info(version)
}

//go:wasmexport register_config
func register_config(_ uint32) {
	instance.RegisterConfig()
}

//go:wasmexport release_config
func release_config(_ uint32) {
	instance.ReleaseConfig()
}

//go:wasmexport get_config_diagnostics
func get_config_diagnostics(_ uint32) uint32 {
	return instance.ConfigDiagnostics()
}

//go:wasmexport get_resolved_config
func get_resolved_config(_ uint32) uint32 {
	return instance.ResolvedConfig()
}

//go:wasmexport get_config_file_matching
func get_config_file_matching(_ uint32) uint32 {
	return instance.ConfigFileMatching()
}

//go:wasmexport check_config_updates
func check_config_updates() uint32 {
	return instance.CheckConfigUpdates()
}

//go:wasmexport set_file_path
func set_file_path() {
	instance.SetFilePath()
}

//go:wasmexport set_override_config
func set_override_config() {
	instance.SetOverrideConfig()
}

//go:wasmexport format
func format(_ uint32) uint32 {
	return instance.Format()
}

//go:wasmexport get_formatted_text
func get_formatted_text() uint32 {
	return instance.FormattedText()
}

//go:wasmexport get_error_text
func get_error_text() uint32 {
	return instance.ErrorText()
}
