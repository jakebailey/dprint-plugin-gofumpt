package plugin

import (
	"fmt"
	goversion "go/version"
	"strconv"

	"github.com/buger/jsonparser"
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

func parseJSONString(value []byte, valueType jsonparser.ValueType, propertyName string) (string, error) {
	if valueType == jsonparser.Null {
		return "", nil
	}
	if valueType != jsonparser.String {
		return "", fmt.Errorf("%s must be a string", propertyName)
	}
	return jsonparser.ParseString(value)
}

func parseJSONBool(value []byte, valueType jsonparser.ValueType, propertyName string) (bool, error) {
	if valueType == jsonparser.Null {
		return false, nil
	}
	if valueType != jsonparser.Boolean {
		return false, fmt.Errorf("%s must be a boolean", propertyName)
	}
	return jsonparser.ParseBoolean(value)
}

func parseExtraConfig(data []byte) (extraConfig, error) {
	var extra extraConfig
	err := jsonparser.EachObject(data, func(key, value []byte, valueType jsonparser.ValueType, _ int) error {
		var err error
		switch string(key) {
		case "groupParams":
			extra.GroupParams, err = parseJSONBool(value, valueType, "groupParams")
		case "clotheReturns":
			extra.ClotheReturns, err = parseJSONBool(value, valueType, "clotheReturns")
		case "balanceCalls":
			extra.BalanceCalls, err = parseJSONBool(value, valueType, "balanceCalls")
		}
		return err
	})
	return extra, err
}

func parseConfig(data []byte) (config, error) {
	var result config
	err := jsonparser.EachObject(data, func(key, value []byte, valueType jsonparser.ValueType, _ int) error {
		var err error
		switch string(key) {
		case "langVersion":
			result.LangVersion, err = parseJSONString(value, valueType, "langVersion")
		case "modulePath":
			result.ModulePath, err = parseJSONString(value, valueType, "modulePath")
		case "extraRules":
			result.ExtraRules, err = parseJSONBool(value, valueType, "extraRules")
		case "extra":
			switch valueType {
			case jsonparser.Null:
				result.Extra = extraConfig{}
			case jsonparser.Object:
				result.Extra, err = parseExtraConfig(value)
			default:
				err = fmt.Errorf("extra must be an object")
			}
		}
		return err
	})
	return result, err
}

func unmarshalConfig(data []byte) (config, error) {
	var result config
	err := jsonparser.EachObject(data, func(key, value []byte, valueType jsonparser.ValueType, _ int) error {
		if string(key) != "plugin" {
			return nil
		}
		switch valueType {
		case jsonparser.Null:
			result = config{}
			return nil
		case jsonparser.Object:
			var err error
			result, err = parseConfig(value)
			return err
		default:
			return fmt.Errorf("plugin must be an object")
		}
	})
	return result, err
}

func appendExtraConfig(data []byte, extra extraConfig) []byte {
	data = append(data, `{"groupParams":`...)
	data = strconv.AppendBool(data, extra.GroupParams)
	data = append(data, `,"clotheReturns":`...)
	data = strconv.AppendBool(data, extra.ClotheReturns)
	data = append(data, `,"balanceCalls":`...)
	data = strconv.AppendBool(data, extra.BalanceCalls)
	return append(data, '}')
}

func marshalConfig(config config) []byte {
	var data []byte
	data = append(data, `{"langVersion":`...)
	data = appendJSONString(data, config.LangVersion)
	data = append(data, `,"modulePath":`...)
	data = appendJSONString(data, config.ModulePath)
	data = append(data, `,"extraRules":`...)
	data = strconv.AppendBool(data, config.ExtraRules)
	data = append(data, `,"extra":`...)
	data = appendExtraConfig(data, config.Extra)
	return append(data, '}')
}

func marshalConfigDiagnostics(diagnostics []configDiagnostic) []byte {
	var data []byte
	data = append(data, '[')
	for i, diagnostic := range diagnostics {
		if i != 0 {
			data = append(data, ',')
		}
		data = append(data, `{"propertyName":`...)
		data = appendJSONString(data, diagnostic.PropertyName)
		data = append(data, `,"message":`...)
		data = appendJSONString(data, diagnostic.Message)
		data = append(data, '}')
	}
	return append(data, ']')
}

func (p *Plugin) RegisterConfig() {
	data := p.takeFromSharedBytes()
	if data == nil {
		return
	}

	p.diagnostics = nil
	p.config = config{}

	parsedConfig, err := unmarshalConfig(data)
	if err != nil {
		p.diagnostics = []configDiagnostic{{
			PropertyName: "gofumpt",
			Message:      err.Error(),
		}}
		return
	}
	p.config = parsedConfig

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
	return p.setSharedBytes(marshalConfigDiagnostics(p.diagnostics))
}

func (p *Plugin) ResolvedConfig() uint32 {
	return p.setSharedBytes(marshalConfig(p.config))
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
