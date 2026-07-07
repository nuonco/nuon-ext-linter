# nuon-ext-linter

Lint Nuon app config directories for best practices, security issues, and common errors.

## Installation

```bash
nuon ext install nuonco/nuon-ext-linter
```

## Usage

### Lint an app config directory

```bash
nuon linter lint [app-config-dir]
```

If no directory is given, the current directory is used.

```bash
# Lint the current directory
nuon linter lint

# Lint a specific directory
nuon linter lint ./byoc-nuon

# JSON output
nuon linter lint ./byoc-nuon --format json

# Only show errors (skip warnings and info)
nuon linter lint --severity error

# Run a specific rule
nuon linter lint --rule no-admin-permissions
```

**Exit codes:** `0` — no errors, `1` — error-level findings detected.

### Initialize a lint config

```bash
nuon linter init [app-config-dir]
```

Generates a `lint.toml` with all built-in rules listed and commented. Won't overwrite an existing file.

### List available rules

```bash
nuon linter rules
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--config` | `-c` | Path to lint.toml (default: `<dir>/lint.toml`) |
| `--format` | `-f` | Output format: `text`, `json` (default: `text`) |
| `--severity` | `-s` | Minimum severity: `info`, `warning`, `error` (default: `warning`) |
| `--rule` | `-r` | Run only specific rule(s), comma-separated |

## Built-in Rules

All rules are enabled by default.

| Rule | Severity | Description |
|------|----------|-------------|
| `require-labels` | warning | Components, actions, and runbooks must have labels |
| `no-admin-permissions` | error | No AdministratorAccess, roles/owner, or overly broad IAM policies |
| `component-nuon-toml` | warning | Components should have a nuon.toml in their source directory |
| `expected-directories` | warning | Directories exist for referenced resources (components/, actions/, etc.) |
| `sandbox-use-tag` | warning | Sandbox repo should use a pinned tag instead of a branch |
| `runner-init-script` | warning | Runner init script URL should match the platform (AWS/GCP/Azure) |
| `permissions-boundary-scope` | error | Permissions boundary JSON must not grant Action:\* Resource:\* |
| `no-wildcard-actions` | warning | Policy/boundary JSON must not contain service-wide wildcards (iam:\*, s3:\*) |
| `require-permissions-boundary` | warning | Permission roles must reference a permissions_boundary |
| `require-policy-tests` | warning | OPA/Rego policies must have corresponding \_test.rego files |

## Configuration

Create a `lint.toml` in your app config directory (or run `nuon linter init`):

```toml
[settings]
# Additional directory to search for custom rule executables
custom_rules_path = "./lint-rules"

# Minimum severity to report: "info", "warning", "error"
min_severity = "warning"

[rules]

# Disable a rule
[rules.no-admin-permissions]
enabled = false

# Require specific labels
[rules.require-labels]
required_labels = ["team", "env"]
```

Rules omitted from `lint.toml` use their defaults (enabled).

## Custom Rules

Custom rules are external executables named `nuon-lint-<name>`. They are auto-discovered in `$PATH` and in the directory specified by `custom_rules_path` in `lint.toml`.

### Writing a custom rule

Create an executable named `nuon-lint-<name>` (any language):

```bash
#!/usr/bin/env bash
# nuon-lint-check-dns — example custom rule

APP_DIR="$1"
INPUT=$(cat)  # JSON: {"settings": {...}, "platform": "aws"}

# Your check logic here...
if grep -rq 'nuon.run' "$APP_DIR"; then
  cat <<'EOF'
{"findings": []}
EOF
else
  cat <<'EOF'
{"findings": [
  {
    "rule_id": "custom:check-dns",
    "severity": "warning",
    "message": "No nuon.run domain references found",
    "file": "sandbox.toml"
  }
]}
EOF
fi
```

### Custom rule protocol

1. **Input:** `argv[1]` = absolute path to app config directory. `stdin` = JSON with `settings` (from lint.toml) and `platform` (detected from runner.toml).
2. **Output:** JSON to stdout: `{"findings": [{"rule_id": "...", "severity": "info|warning|error", "message": "...", "file": "..."}]}`
3. **Exit code:** `0` = success (findings may still be present). Non-zero = execution error.
4. **Timeout:** 30 seconds per rule.

### Disabling a custom rule

```toml
[rules."custom:check-dns"]
enabled = false
```

### Passing settings to a custom rule

```toml
[rules."custom:check-dns"]
[rules."custom:check-dns".settings]
required_domain = "nuon.run"
```

The settings object is passed to the executable as part of the stdin JSON.

## Development

```bash
# Build
go build .

# Run locally
./nuon-ext-linter lint /path/to/app-config

# Run with Go
go run . lint /path/to/app-config
```
