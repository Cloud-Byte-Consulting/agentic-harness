# PR bundle — TRAPI skills (staged)

These 5 APIM / Azure security / ASP.NET content skills are **staged** in
agent-coordination for review. They are NOT yet written into the production TRAPI repo.

## Promote into production

Open a PR that copies each folder here into:

```
../trapi-agent-system/.claude/skills/<skill-name>/
```

Skills in this bundle:

- api-management
- cloud-security-posture-management
- dotnet-web-development            (trapi-snapshot + batch ASP.NET Web APIs)
- csharp-dotnet-fundamentals
- dotnet-enterprise-architecture

## Rules

- Content skills only — plain `SKILL.md` folders. Do **not** add them to
  `capabilities.json`.
- Loading = read-only knowledge. APIM policy publish / CSPM remediation remain
  **Tier-3** and route through `/mutation-gate`. TRAPI keeps ADO work items + pipelines
  (no GitHub Issues / Actions).
- Governance mapping: `contracts/skill-routing.yaml` → `trapi-router`.
- Architecture: `knowledge-base/41-skill-routing-architecture.md`.
