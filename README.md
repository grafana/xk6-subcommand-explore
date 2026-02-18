# xk6-subcommand-explore

**Explore k6 extensions for Automatic Resolution**

This k6 extension adds an `explore` subcommand that helps you discover available k6 extensions from the official extension registry. It's designed to work with k6's Automatic Extension Resolution feature, allowing you to quickly find compatible extensions, filter them by type or tier, and view detailed information.

The extension provides both human-readable table output and machine-readable JSON output, making it suitable for both interactive use and CI/CD automation.

## Usage

The `explore` subcommand lists available k6 extensions from the extension registry. You can filter, format, and customize the output.

**Flags:**

- `--brief` – Only show module and description columns in table output
- `--json` – Output as JSON (ignores --brief)
- `--tier` – Filter by extension tier (`official`, `community`)
- `--type` – Filter by extension type (`javascript`, `output`, `subcommand`)

**Examples:**

List all extensions (table output):
```shell
k6 x explore
```

Show only module and description columns (brief output):
```shell
k6 x explore --brief
```

Output as JSON (for CI/CD integration):
```shell
k6 x explore --json
```

Filter by tier or type:
```shell
k6 x explore --tier official --type javascript
```

## JSON Output

When using the `--json` flag, the output is an array of extension objects. Each extension object contains the following properties:

- `module` (string) – The Go module path of the extension
- `tier` (string) – Extension tier: `official` or `community`
- `description` (string) – Brief description of the extension's functionality
- `latest` (string) – Latest version tag (e.g., `v0.1.0`)
- `versions` (array of strings) – All available version tags
- `imports` (array of strings) – JavaScript module import paths (for JavaScript extensions)
- `outputs` (array of strings) – Output type names (for output extensions)
- `subcommands` (array of strings) – Subcommand names (for subcommand extensions)

**Example JSON:**

```json
[
  {
    "module": "github.com/grafana/xk6-faker",
    "tier": "official",
    "description": "Generate fake data in your tests",
    "latest": "v0.4.4",
    "versions": ["v0.4.4","v0.4.3","v0.4.2","v0.4.1","v0.4.0"],
    "imports": ["k6/x/faker"]
  },
  {
    "module": "github.com/grafana/xk6-tls",
    "tier": "community",
    "description": "TLS certificates validation and inspection",
    "latest": "v0.1.0",
    "versions": ["v0.1.0"],
    "imports": ["k6/x/tls"]
  },
  {
    "module": "github.com/grafana/xk6-subcommand-httpbin",
    "tier": "community",
    "description": "Run a local httpbin server from k6",
    "latest": "v1.0.0",
    "versions": ["v1.0.0"],
    "subcommands": ["httpbin"]
  }
]
```

## Build

Currently, you need to build a custom k6 binary with this extension to use the `explore` subcommand. Use the [xk6](https://github.com/grafana/xk6) tool to build k6 with the `xk6-subcommand-explore` extension. Refer to the [xk6 documentation](https://github.com/grafana/xk6) for more information.

In the near future, k6 will introduce subcommand support to the Automatic Extension Resolution feature, making this extension available automatically without requiring a custom build.

## Contribute

If you wish to contribute to this project, please start by reading the [Contributing Guidelines](CONTRIBUTING.md).
