package appconfig

type AppConfig struct {
	Dir         string
	Metadata    *Metadata
	Inputs      *Inputs
	Runner      *Runner
	Sandbox     *Sandbox
	Permissions *Permissions
	BreakGlass  *BreakGlass
	Components  []Component
	Actions     []Action
	Runbooks    []Runbook

	// PermissionRoles holds individual role files from permissions/ directory.
	// These are an alternative to the single permissions.toml file.
	PermissionRoles []PermissionRole

	// BoundaryFiles holds JSON boundary files from permissions/boundaries/.
	BoundaryFiles []JSONPolicyFile

	// PolicyFiles holds JSON policy files from permissions/policies/.
	PolicyFiles []JSONPolicyFile

	// OPAPolicies holds OPA/Rego policy files from policies/ directory.
	OPAPolicies []OPAPolicy
}

type Metadata struct {
	Version     string `toml:"version"`
	Description string `toml:"description"`
	DisplayName string `toml:"display_name"`
}

type Inputs struct {
	Groups []InputGroup `toml:"group"`
	Inputs []Input      `toml:"input"`
}

type InputGroup struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	DisplayName string `toml:"display_name"`
}

type Input struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	DisplayName string `toml:"display_name"`
	Group       string `toml:"group"`
	Type        string `toml:"type"`
	Default     string `toml:"default"`
	Required    bool   `toml:"required"`
	Sensitive   bool   `toml:"sensitive"`
}

type Runner struct {
	RunnerType    string            `toml:"runner_type"`
	HelmDriver    string            `toml:"helm_driver"`
	InitScriptURL string            `toml:"init_script_url"`
	EnvVars       map[string]string `toml:"env_vars"`
}

type Sandbox struct {
	TerraformVersion string            `toml:"terraform_version"`
	Type             string            `toml:"type"`
	PublicRepo       *RepoRef          `toml:"public_repo"`
	ConnectedRepo    *RepoRef          `toml:"connected_repo"`
	Vars             map[string]string `toml:"vars"`
	VarFile          []VarFileRef      `toml:"var_file"`
}

type RepoRef struct {
	Repo      string `toml:"repo"`
	Directory string `toml:"directory"`
	Branch    string `toml:"branch"`
	Tag       string `toml:"tag"`
}

type VarFileRef struct {
	Contents string `toml:"contents"`
}

type Permissions struct {
	ProvisionRole   *Role  `toml:"provision_role"`
	DeprovisionRole *Role  `toml:"deprovision_role"`
	MaintenanceRole *Role  `toml:"maintenance_role"`
	CustomRoles     []Role `toml:"custom_roles"`
}

type BreakGlass struct {
	Roles []Role `toml:"role"`
}

type Role struct {
	Name                string   `toml:"name"`
	Description         string   `toml:"description"`
	DisplayName         string   `toml:"display_name"`
	PermissionsBoundary string   `toml:"permissions_boundary"`
	Policies            []Policy `toml:"policies"`
}

type Policy struct {
	ManagedPolicyName string `toml:"managed_policy_name"`
	Name              string `toml:"name"`
	Contents          string `toml:"contents"`
}

type Component struct {
	File          string            // source file path (not from TOML)
	Name          string            `toml:"name"`
	Type          string            `toml:"type"`
	Labels        map[string]string `toml:"labels"`
	Dependencies  []string          `toml:"dependencies"`
	PublicRepo    *RepoRef          `toml:"public_repo"`
	ConnectedRepo *RepoRef          `toml:"connected_repo"`
	Source        string            `toml:"source"`
	ChartName     string            `toml:"chart_name"`
	Namespace     string            `toml:"namespace"`
}

type Action struct {
	File    string            // source file path (not from TOML)
	Name    string            `toml:"name"`
	Labels  map[string]string `toml:"labels"`
	Timeout string            `toml:"timeout"`
}

type Runbook struct {
	File        string            // source file path (not from TOML)
	Name        string            `toml:"name"`
	Description string            `toml:"description"`
	Labels      map[string]string `toml:"labels"`
}

// PermissionRole is a role defined as an individual file in permissions/ directory
// (e.g., permissions/provision.toml, permissions/maintenance.toml).
type PermissionRole struct {
	File                string   `toml:"-"`
	Type                string   `toml:"type"` // provision, deprovision, maintenance
	Name                string   `toml:"name"`
	Description         string   `toml:"description"`
	DisplayName         string   `toml:"display_name"`
	PermissionsBoundary string   `toml:"permissions_boundary"`
	Policies            []Policy `toml:"policies"`
}

// JSONPolicyFile represents a JSON IAM policy or boundary document on disk.
type JSONPolicyFile struct {
	Path string // file path relative to app config dir
}

// OPAPolicy represents an OPA/Rego policy file.
type OPAPolicy struct {
	File    string // path to the .rego file
	Name    string // derived from filename
	HasTest bool   // whether a corresponding _test.rego exists
}
