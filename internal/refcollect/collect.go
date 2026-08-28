package refcollect

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
)

// Context describes where a templated value lives for scope validation.
type Context string

const (
	ContextStack      Context = "stack"
	ContextSandbox    Context = "sandbox"
	ContextComponent  Context = "component"
	ContextAction     Context = "action"
	ContextRunbook    Context = "runbook"
	ContextPermission Context = "permission"
	ContextOther      Context = "other"
)

// Site is a config location with already-parsed template references.
type Site struct {
	File    string
	Entity  string
	Context Context
	Refs    []refs.Ref
}

// StackField is a stack.toml string field checked for early-bound template scope.
type StackField struct {
	File  string
	Field string
	Value string
}

// Collect returns parsed reference sites from a Nuon AppConfig.
func Collect(cfg *config.AppConfig) ([]Site, []StackField) {
	if cfg == nil {
		return nil, nil
	}

	var sites []Site
	add := func(file, entity string, ctx Context, rs []refs.Ref) {
		if len(rs) == 0 {
			return
		}
		sites = append(sites, Site{
			File:    file,
			Entity:  entity,
			Context: ctx,
			Refs:    rs,
		})
	}

	for _, c := range cfg.Components {
		if c == nil {
			continue
		}
		add(c.SourceFile, "component:"+c.Name, ContextComponent, c.References)
	}

	for _, a := range cfg.Actions {
		if a == nil {
			continue
		}
		file := fmt.Sprintf("actions/%s", a.Name)
		add(file, "action:"+a.Name, ContextAction, a.References)
		for _, tr := range a.Triggers {
			if tr != nil && tr.ComponentName != "" {
				add(file, "action:"+a.Name+":trigger:"+tr.Type, ContextAction, []refs.Ref{{
					Type:  refs.RefTypeComponents,
					Name:  tr.ComponentName,
					Input: tr.ComponentName,
				}})
			}
		}
	}

	if cfg.Sandbox != nil {
		add("sandbox.toml", "sandbox", ContextSandbox, cfg.Sandbox.References)
	}

	for _, rb := range cfg.Runbooks {
		if rb == nil {
			continue
		}
		add("", "runbook:"+rb.Name, ContextRunbook, rb.References)
	}

	if cfg.Permissions != nil {
		for _, role := range cfg.Permissions.Roles {
			if role == nil {
				continue
			}
			rs, err := refs.Parse(role)
			if err == nil {
				add("permissions/"+role.Name+".toml", "permission:"+role.Name, ContextPermission, rs)
			}
		}
	}

	if cfg.BreakGlass != nil {
		for _, role := range cfg.BreakGlass.Roles {
			if role == nil {
				continue
			}
			rs, err := refs.Parse(role)
			if err == nil {
				add("break_glass.toml", "break_glass:"+role.Name, ContextPermission, rs)
			}
		}
	}

	// Safety net for refs on types without per-entity References (e.g. component health).
	if allRefs, err := refs.Parse(cfg); err == nil {
		known := refSet(sites)
		var extra []refs.Ref
		for _, r := range allRefs {
			if !known[refKey(r)] {
				extra = append(extra, r)
			}
		}
		add("", "app", ContextOther, extra)
	}

	var stackFields []StackField
	if cfg.Stack != nil {
		stackFields = collectStackFields(cfg.Stack)
	}

	return sites, stackFields
}

func collectStackFields(stack *config.StackConfig) []StackField {
	var fields []StackField
	add := func(field, value string) {
		if value == "" {
			return
		}
		fields = append(fields, StackField{File: "stack.toml", Field: field, Value: value})
	}

	add("name", stack.Name)
	add("description", stack.Description)
	add("vpc_nested_template_url", stack.VPCNestedTemplateURL)
	add("runner_nested_template_url", stack.RunnerNestedTemplateURL)

	for i, ns := range stack.CustomNestedStacks {
		add(fmt.Sprintf("custom_nested_stacks[%d].template_url", i), ns.TemplateURL)
		for param, val := range ns.Parameters {
			add(fmt.Sprintf("custom_nested_stacks[%d].parameters.%s", i, param), val)
		}
	}

	return fields
}

func refSet(sites []Site) map[string]bool {
	out := make(map[string]bool)
	for _, s := range sites {
		for _, r := range s.Refs {
			out[refKey(r)] = true
		}
	}
	return out
}

func refKey(r refs.Ref) string {
	return string(r.Type) + "|" + r.Name + "|" + r.Input
}
