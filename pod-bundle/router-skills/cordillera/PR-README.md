# PR bundle — Cordillera skills (staged)

These 15 Kubernetes / cloud-native content skills are **staged** in agent-coordination
for review. They are NOT yet written into the production Cordillera repo.

## Promote into production

Open a PR that copies each folder here into:

```
../cordillera-agent-system/.claude/skills/<skill-name>/
```

Skills in this bundle:

- container-fundamentals
- kubernetes-workloads
- kubernetes-networking
- kubernetes-storage
- kubernetes-security-rbac
- kubernetes-autoscaling-scheduling
- kubernetes-cluster-operations
- kubernetes-gitops-cicd
- kubernetes-observability
- kubernetes-data-platforms
- cloud-native-platform-engineering
- volcano-core-scheduling
- volcano-advanced-hardware
- volcano-ecosystem-workflows
- volcano-genai-serving

## Rules

- Content skills only — plain `SKILL.md` folders. Do **not** add them to
  `capabilities.json` (that manifest is for thin router wrappers).
- Loading = read-only knowledge. Operational actions (cluster upgrade, RBAC change,
  GitOps promote) remain **Tier-3** and route through `/mutation-gate`.
- Governance mapping: `contracts/skill-routing.yaml` → `cordillera-router`.
- Architecture: `knowledge-base/41-skill-routing-architecture.md`.
