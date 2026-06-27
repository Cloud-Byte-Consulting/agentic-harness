# Templates 🧱

Reusable seeds for the artifacts an agentic team produces. Naming: `*.template.md`
with placeholders in `<angle brackets>`. Shared discipline across all work
products: **lead with the outcome/ask, cite primary evidence (`file:line`, links,
test output), and separate fact from inference** — the same evidence-grounded
posture the [mutation gate](../docs/pods_and_skill_routing.md) and Judge expect.

## Harness setup (instruction seeds)
Fanned out to each tool's instruction file by `air agents link`.

- `AGENTS.template.md` · `voice.template.md` · `opinions.template.md`

## Work products by category

| Category | Templates |
| :--- | :--- |
| **decisions/** | [`adr`](decisions/adr.template.md) · [`prd`](decisions/prd.template.md) · [`pdr`](decisions/pdr.template.md) · [`tech-design-doc`](decisions/tech-design-doc.template.md) |
| **reviews/** | [`commit-message`](reviews/commit-message.template.md) · [`pr-description`](reviews/pr-description.template.md) · [`pr-comment`](reviews/pr-comment.template.md) |
| **communications/** | [`email`](communications/email.template.md) |
| **reliability/** | [`incident-postmortem`](reliability/incident-postmortem.template.md) · [`runbook`](reliability/runbook.template.md) · [`slo`](reliability/slo.template.md) |
| **delivery/** | [`change-request`](delivery/change-request.template.md) · [`release-notes`](delivery/release-notes.template.md) |
| **data/** | [`data-contract`](data/data-contract.template.md) · [`data-quality-report`](data/data-quality-report.template.md) |
| **security/** | [`threat-model`](security/threat-model.template.md) · [`security-assessment`](security/security-assessment.template.md) |
| **program/** | [`status-report`](program/status-report.template.md) · [`raid-log`](program/raid-log.template.md) · [`meeting-notes`](program/meeting-notes.template.md) |
| **portfolio/** | [`roadmap`](portfolio/roadmap.template.md) · [`okr`](portfolio/okr.template.md) |
| **people/** | [`one-on-one`](people/one-on-one.template.md) · [`growth-plan`](people/growth-plan.template.md) |
| **analysis/** | [`brd`](analysis/brd.template.md) · [`user-story`](analysis/user-story.template.md) · [`process-map`](analysis/process-map.template.md) |
| **work-tracking/** | [`github`](work-tracking/github-issue.template.md) · [`jira`](work-tracking/jira-issue.template.md) · [`azure-devops`](work-tracking/azure-devops-work-item.template.md) · [`gitlab`](work-tracking/gitlab-issue.template.md) |

## Signature templates by [persona](../docs/packaging_and_personas.md)

| Persona | Primary templates |
| :--- | :--- |
| **Software Engineer** | tech-design-doc, commit-message, pr-description, pr-comment, user-story |
| **Site Reliability Engineer** | incident-postmortem, runbook, slo, change-request |
| **DevOps Engineer** | change-request, release-notes, runbook |
| **Data Engineer** | data-contract, data-quality-report |
| **Technical Program Manager** | status-report, raid-log, meeting-notes |
| **Lead Program Manager** | roadmap, okr, exec email |
| **Engineering Manager** | one-on-one, growth-plan, meeting-notes |
| **Business Analyst** | brd, user-story, process-map |
| **Cybersecurity Engineer** | threat-model, security-assessment, incident-postmortem |
| **Cross-cutting (all)** | adr, prd, email, meeting-notes, work-tracking/* |

> See [docs/packaging_and_personas.md](../docs/packaging_and_personas.md) for the
> persona packs these artifacts belong to.
