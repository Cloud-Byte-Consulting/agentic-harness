# User-defined Pods and zip deliverables

This pattern makes Pods fully user-defined and markdown-driven, with no hardcoded default Pod behavior.

## Design changes to make Pods generic

1. Treat each Pod as a folder under `pods/` (no built-in defaults required).
2. Load behavior only from markdown files inside that Pod folder.
3. Refuse to run a Pod if required markdown files are missing.
4. Package each Pod folder into one zip so it can be shared/imported as a single artifact.

## Required markdown contract per Pod

Each Pod must include:

- `README.md` - human summary and packaging command.
- `pod.md` - metadata, ownership, routing, write policy.
- `behavior.md` - ranking, triage, and escalation rules.
- `sources.md` - source systems and evidence requirements.
- `workflows.md` - playbooks and completion criteria.

This keeps behavior editable by users while remaining deterministic for automation.

## Walkthrough: create your own Pod

From repo root:

```powershell
.\scripts\new-pod-bundle.ps1 -Name "My Operations Pod"
```

This scaffolds:

```text
pods\my-operations-pod\
  README.md
  pod.md
  behavior.md
  sources.md
  workflows.md
```

Then edit each markdown file to define your Pod behavior.

## Walkthrough: package as a single zip

Package one Pod:

```powershell
.\scripts\package-pod-bundles.ps1 -PodsRoot .\pods -PodName my-operations-pod -Force
```

Package all Pods:

```powershell
.\scripts\package-pod-bundles.ps1 -PodsRoot .\pods -Force
```

Output:

- `dist\pods\<pod-name>.zip` for each Pod.
- `dist\pods\manifest.json` with SHA256 hashes for integrity.

## Why this meets your goal

- **No defaults required:** behavior lives in user markdown.
- **Easy customization:** users change markdown, not code.
- **Single deliverable per Pod:** each Pod ships as one zip.
- **Portable and auditable:** manifest hashes support review and provenance.
- **Plugin install support:** packaged plugin can auto-install templates, scripts, docs, and skills into the workspace.

## Source files

- `templates/pod-bundle/*`
- `scripts/new-pod-bundle.ps1`
- `scripts/package-pod-bundles.ps1`
