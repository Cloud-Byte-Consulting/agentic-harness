# Software Engineer Pod Bundle

This bundle is markdown-only so users can define behavior without touching code.

## Files

- `pod.md` - Pod identity, ownership, and routing hints.
- `behavior.md` - Prioritization and triage behavior.
- `sources.md` - Data connectors and required fields.
- `workflows.md` - Task playbooks and escalation paths.

## Packaging

Run from repo root:

```powershell
.\scripts\package-pod-bundles.ps1 -PodsRoot .\pods -PodName "Software Engineer" -Force
```
