package appconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

func Load(dir string) (*AppConfig, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("app config directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	app := &AppConfig{Dir: dir}

	if err := decodeFileIfExists(filepath.Join(dir, "metadata.toml"), &app.Metadata); err != nil {
		return nil, fmt.Errorf("metadata.toml: %w", err)
	}

	if err := decodeFileIfExists(filepath.Join(dir, "inputs.toml"), &app.Inputs); err != nil {
		return nil, fmt.Errorf("inputs.toml: %w", err)
	}

	if err := decodeFileIfExists(filepath.Join(dir, "runner.toml"), &app.Runner); err != nil {
		return nil, fmt.Errorf("runner.toml: %w", err)
	}

	if err := decodeFileIfExists(filepath.Join(dir, "sandbox.toml"), &app.Sandbox); err != nil {
		return nil, fmt.Errorf("sandbox.toml: %w", err)
	}

	if err := decodeFileIfExists(filepath.Join(dir, "permissions.toml"), &app.Permissions); err != nil {
		return nil, fmt.Errorf("permissions.toml: %w", err)
	}

	if err := decodeFileIfExists(filepath.Join(dir, "break_glass.toml"), &app.BreakGlass); err != nil {
		return nil, fmt.Errorf("break_glass.toml: %w", err)
	}

	components, err := loadDir[Component](filepath.Join(dir, "components"))
	if err != nil {
		return nil, fmt.Errorf("components: %w", err)
	}
	app.Components = components

	actions, err := loadDir[Action](filepath.Join(dir, "actions"))
	if err != nil {
		return nil, fmt.Errorf("actions: %w", err)
	}
	app.Actions = actions

	runbooks, err := loadDir[Runbook](filepath.Join(dir, "runbooks"))
	if err != nil {
		return nil, fmt.Errorf("runbooks: %w", err)
	}
	app.Runbooks = runbooks

	// Load individual permission role files from permissions/ directory
	app.PermissionRoles = loadPermissionRoles(filepath.Join(dir, "permissions"))

	// Load JSON boundary and policy files
	app.BoundaryFiles = loadJSONFiles(filepath.Join(dir, "permissions", "boundaries"))
	app.PolicyFiles = loadJSONFiles(filepath.Join(dir, "permissions", "policies"))

	// Load OPA policies
	app.OPAPolicies = loadOPAPolicies(filepath.Join(dir, "policies"))

	return app, nil
}

func loadPermissionRoles(dir string) []PermissionRole {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "*.toml"))
	var roles []PermissionRole
	for _, path := range matches {
		var role PermissionRole
		if _, err := toml.DecodeFile(path, &role); err != nil {
			continue
		}
		role.File = path
		roles = append(roles, role)
	}
	return roles
}

func loadJSONFiles(dir string) []JSONPolicyFile {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	var files []JSONPolicyFile
	for _, path := range matches {
		files = append(files, JSONPolicyFile{Path: path})
	}
	return files
}

func loadOPAPolicies(dir string) []OPAPolicy {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "*.rego"))
	var policies []OPAPolicy
	for _, path := range matches {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.rego") {
			continue // skip test files, they're tracked via HasTest
		}
		name := strings.TrimSuffix(base, ".rego")
		testPath := filepath.Join(dir, name+"_test.rego")
		_, err := os.Stat(testPath)
		hasTest := err == nil
		policies = append(policies, OPAPolicy{
			File:    path,
			Name:    name,
			HasTest: hasTest,
		})
	}
	return policies
}

func decodeFileIfExists[T any](path string, dst **T) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	val := new(T)
	if _, err := toml.DecodeFile(path, val); err != nil {
		return err
	}
	*dst = val
	return nil
}

type hasFile interface {
	setFile(string)
}

func (c *Component) setFile(f string) { c.File = f }
func (a *Action) setFile(f string)    { a.File = f }
func (r *Runbook) setFile(f string)   { r.File = f }

func loadDir[T any, PT interface {
	*T
	hasFile
}](dir string) ([]T, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.toml"))
	if err != nil {
		return nil, err
	}

	var items []T
	for _, path := range matches {
		var item T
		if _, err := toml.DecodeFile(path, &item); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		PT(&item).setFile(path)
		items = append(items, item)
	}
	return items, nil
}
