# dprint-plugin-gofumpt

A [dprint](https://dprint.dev/) plugin for formatting Go code using
[gofumpt](https://github.com/mvdan/gofumpt).

## Getting Started

Run the command below to add the plugin:

```sh
dprint config add jakebailey/gofumpt
```

After adding the plugin, configure the plugin in `dprint.json`:

```jsonc
{
    // ...
    "gofumpt": {
        "langVersion": "go1.24",
        "modulePath": "github.com/myorg/myproject"
    },
    "plugins": [
        "https://plugins.dprint.dev/jakebailey/gofumpt-v0.0.1.wasm"
    ]
}
```

## Configuration

It is recommended to set both `langVersion` and `modulePath` for consistent
formatting results, as the plugin cannot infer this information from `go.mod`.

| Property      | Type      | Default | Description                                                                                                     |
| ------------- | --------- | ------- | --------------------------------------------------------------------------------------------------------------- |
| `langVersion` | `string`  | `""`    | The Go language version to target (e.g., `"go1.24"`). Must start with `go` prefix. If empty, defaults to `go1`. |
| `modulePath`  | `string`  | `""`    | The module path of the package being formatted. Used for import sorting.                                        |
| `extraRules`  | `boolean` | `false` | Enable extra formatting rules beyond the default gofumpt rules.                                                 |
