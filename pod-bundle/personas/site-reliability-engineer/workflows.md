# Workflows

Define Pod-specific playbooks in markdown.

## Workflow Template

### Name

Add workflow name.

### Trigger

Describe the condition that starts this workflow.

### Steps

1. Read and classify with `/platform-router`.
2. Ground in evidence with `/evidence-grounded-investigation`.
3. Run `/mutation-gate` before any write.
4. Execute through the Omnigent supervisor (`omnigent/examples/role_router/config.yaml`) and require Judge PASS (the ported Judge skill).
5. Record via Scribe to owning work item.

### Exit Criteria

List objective completion conditions.
