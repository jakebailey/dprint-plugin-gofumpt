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
        // Copy these from your go.mod file!
        "langVersion": "go1.24",
        "modulePath": "github.com/myorg/myproject"
    },
    "plugins": [
        "https://plugins.dprint.dev/jakebailey/gofumpt-v0.0.15.wasm"
    ]
}
```

Alternatively, the plugin is published to npm and can be referenced with an
[`npm:` specifier](https://dsherret.dev/posts/dprint-0.55/) (requires dprint
0.55 or newer):

```sh
dprint add npm:@jakebailey/dprint-plugin-gofumpt
```

```jsonc
{
    // ...
    "plugins": [
        "npm:@jakebailey/dprint-plugin-gofumpt@0.0.15"
    ]
}
```

## Configuration

It is recommended to set both `langVersion` and `modulePath` for consistent
formatting results, as the plugin cannot infer this information from `go.mod`.

| Property              | Type      | Default | Description                                                                                                                                                    |
| --------------------- | --------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `langVersion`         | `string`  | `""`    | The Go language version to target (e.g., `"go1.24"`). Must be a valid [Go version string](https://pkg.go.dev/go/version#IsValid). If empty, defaults to `go1`. |
| `modulePath`          | `string`  | `""`    | The module path of the package being formatted. Used for import sorting.                                                                                       |
| `extra`               | `object`  | `{}`    | Enable individual extra rules.                                                                                                                                 |
| `extra.groupParams`   | `boolean` | `false` | Group adjacent function parameters with repeated types.                                                                                                        |
| `extra.clotheReturns` | `boolean` | `false` | Replace naked returns in functions with named results with explicit return values.                                                                             |
| `extra.balanceCalls`  | `boolean` | `false` | Place a multi-line call's closing parenthesis on its own line when the opening parenthesis ends a line.                                                        |
| `extraRules`          | `boolean` | `false` | Enable all extra formatting rules beyond the default gofumpt rules. Deprecated: use `extra` to enable individual rules instead.                                |

## Versions

This plugin is versioned separately from `gofumpt`. Below is a table of which
plugin versions correspond to which `gofumpt` versions.

| Plugin Version | gofumpt Version                       |
| -------------- | ------------------------------------- |
| v0.0.15+       | v0.11.1-0.20260820074422-a2bc6805583d |
| v0.0.14+       | v0.11.0                               |
| v0.0.12+       | v0.10.1-0.20260531213040-cc84354298ea |
| v0.0.10+       | v0.10.0                               |
| v0.0.2+        | v0.9.3-0.20251215221355-d3e4b13ef7fa  |
| v0.0.1         | v0.9.2                                |
