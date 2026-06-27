// Package personas scaffolds the persona packs (replaces scaffold-personas.sh):
// for each persona it stamps the 5-file Pod template and writes a populated
// persona.yaml selecting skills, sources, workflows, and the mutation tier.
package personas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Spec is one persona's selection.
type Spec struct {
	ID, Name, Role string
	Shared         []string
	Router         []string
	Add            []string
	Sources        []string
	Workflows      []string
	Tier           string
}

// Specs is the canonical persona set.
var Specs = []Spec{
	{"software-engineer", "Software Engineer", "SWE",
		[]string{"clean-code-typescript", "test-driven-development", "code-review-and-quality", "frontend-ui-engineering", "building-ai-agents"},
		nil, nil,
		[]string{"repo", "git", "tests", "ci"},
		[]string{"feature", "bug-fix", "refactor"}, "tier2_elevated"},
	{"technical-program-manager", "Technical Program Manager", "TPM",
		[]string{"technical-program-management", "planning-and-task-breakdown", "documentation-and-adrs", "co-operating-model", "idea-refine"},
		nil, nil,
		[]string{"tickets", "roadmaps", "status-docs", "issue-tracker"},
		[]string{"program-sync", "dependency-map", "risk-review", "status-rollup"}, "tier1_low"},
	{"lead-program-manager", "Lead Program Manager", "LPM",
		[]string{"technical-program-management", "solutions-architecture", "planning-and-task-breakdown", "documentation-and-adrs"},
		nil, nil,
		[]string{"portfolio-dashboards", "okrs", "multi-program-status"},
		[]string{"portfolio-review", "exec-brief", "cross-program-prioritization"}, "tier1_low"},
	{"engineering-manager", "Engineering Manager", "EM",
		[]string{"planning-and-task-breakdown", "co-operating-model", "technical-program-management", "documentation-and-adrs"},
		nil,
		[]string{"delivery-metrics", "one-on-one-prep", "hiring-loop"},
		[]string{"delivery-metrics", "team-docs", "pr-throughput", "incident-counts"},
		[]string{"team-health-review", "sprint-planning", "growth-planning", "hiring"}, "tier1_low"},
	{"site-reliability-engineer", "Site Reliability Engineer", "SRE",
		[]string{"security-operations-mitre-attack", "evidence-grounded-investigation"},
		[]string{"cordillera"},
		[]string{"incident-response", "slo-management", "runbook-authoring"},
		[]string{"metrics", "logs", "alerts", "dashboards", "runbooks"},
		[]string{"incident-triage", "slo-review", "postmortem", "capacity-planning"}, "tier3_high_exposure"},
	{"devops-engineer", "DevOps Engineer", "DevOps",
		[]string{"powershell-scripting", "evidence-grounded-investigation"},
		[]string{"cordillera", "trapi"},
		[]string{"infrastructure-as-code", "ci-cd-pipeline-engineering", "release-and-artifact-management"},
		[]string{"pipelines", "ci-logs", "iac-repos", "deploy-manifests", "registries", "environments"},
		[]string{"pipeline-build", "iac-change", "deploy-rollback", "env-provisioning"}, "tier3_high_exposure"},
	{"data-engineer", "Data Engineer", "DataEng",
		[]string{"test-driven-development", "evidence-grounded-investigation", "solutions-architecture"},
		[]string{"cordillera"},
		[]string{"data-pipeline-engineering", "data-modeling-and-schemas", "sql-and-warehousing", "data-quality-and-validation"},
		[]string{"data-catalogs", "schemas", "pipeline-dags", "warehouse-logs", "dq-reports"},
		[]string{"pipeline-build", "schema-migration", "data-quality-check", "backfill"}, "tier3_high_exposure"},
	{"business-analyst", "Business Analyst", "BA",
		[]string{"idea-refine", "planning-and-task-breakdown", "documentation-and-adrs", "co-operating-model", "solutions-architecture"},
		nil,
		[]string{"requirements-elicitation", "process-modeling", "stakeholder-analysis", "reporting-and-insights"},
		[]string{"requirements-docs", "stakeholder-interviews", "tickets", "bi-reports", "process-maps"},
		[]string{"requirements-gathering", "process-mapping", "gap-analysis", "report-definition"}, "tier1_low"},
	{"cybersecurity-engineer", "Cybersecurity Engineer", "SecEng",
		[]string{"security-operations-mitre-attack", "evidence-grounded-investigation"},
		[]string{"trapi", "cordillera"},
		[]string{"threat-modeling", "vulnerability-management", "secure-code-review", "detection-and-response"},
		[]string{"siem-alerts", "cve-feeds", "sast-dast-sca-reports", "asset-inventory", "audit-logs", "threat-intel"},
		[]string{"threat-model", "vuln-triage", "security-review", "incident-response", "pentest"}, "tier3_high_exposure"},
}

var podFiles = []string{"README.md", "pod.md", "behavior.md", "sources.md", "workflows.md"}

// Scaffold stamps every persona pack from templateDir into outDir/<id>/.
func Scaffold(templateDir, outDir string) ([]string, error) {
	tmpls := map[string]string{}
	for _, f := range podFiles {
		b, err := os.ReadFile(filepath.Join(templateDir, f))
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", f, err)
		}
		tmpls[f] = string(b)
	}
	var made []string
	for _, s := range Specs {
		dir := filepath.Join(outDir, s.ID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		for _, f := range podFiles {
			body := strings.ReplaceAll(tmpls[f], "__POD_ID__", s.ID)
			body = strings.ReplaceAll(body, "__POD_NAME__", s.Name)
			if err := os.WriteFile(filepath.Join(dir, f), []byte(body), 0o644); err != nil {
				return nil, err
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "persona.yaml"), []byte(s.YAML()), 0o644); err != nil {
			return nil, err
		}
		made = append(made, s.ID)
	}
	return made, nil
}

// YAML renders the persona.yaml content.
func (s Spec) YAML() string {
	arr := func(xs []string) string { return "[" + strings.Join(xs, ", ") + "]" }
	var b strings.Builder
	fmt.Fprintf(&b, "# Persona pack — selects skills, evidence, workflows, and gate posture for the %s role.\n", s.Role)
	b.WriteString("# The 5 pod files alongside this carry the editable Pod definition.\n")
	fmt.Fprintf(&b, "id: %s\n", s.ID)
	fmt.Fprintf(&b, "role: %s\n", s.Role)
	fmt.Fprintf(&b, "name: %q\n", s.Name)
	b.WriteString("pod: ./pod.md\n")
	b.WriteString("skills:\n")
	fmt.Fprintf(&b, "  shared: %s\n", arr(s.Shared))
	fmt.Fprintf(&b, "  router: %s\n", arr(s.Router))
	if len(s.Add) > 0 {
		fmt.Fprintf(&b, "  add: %s   # net-new, author under pod-bundle/skills/\n", arr(s.Add))
	} else {
		b.WriteString("  add: []\n")
	}
	fmt.Fprintf(&b, "sources: %s\n", arr(s.Sources))
	fmt.Fprintf(&b, "workflows: %s\n", arr(s.Workflows))
	fmt.Fprintf(&b, "defaultMutationTier: %s\n", s.Tier)
	b.WriteString("surfaces: [role-router, cachy, token-dashboard]\n")
	return b.String()
}
