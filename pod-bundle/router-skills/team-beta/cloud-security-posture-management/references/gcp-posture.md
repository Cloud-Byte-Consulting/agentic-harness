# GCP Posture Assessment

Security Command Center as the CSPM platform, its detector services, IAM right-sizing, Organization Policy, and common GCP misconfigurations with `gcloud` fixes.

## Security Command Center (SCC) — the platform

SCC is GCP's native CSPM/CNAPP. Enable it at the **organization** node so every folder and project is covered (project-level activation exists but leaves gaps). Tiers:

- **Standard** — free. Includes **Security Health Analytics** (a subset of detectors), **Web Security Scanner** (custom scans), and basic asset inventory.
- **Premium** — paid. Full **Security Health Analytics** detectors, **Event Threat Detection**, **Container Threat Detection**, **Virtual Machine Threat Detection**, **compliance reporting** (CIS, PCI DSS, NIST, ISO), **attack path simulation / attack exposure scores**, and **Cloud Infrastructure Entitlement Management** features.
- **Enterprise** — the CNAPP tier: multicloud (adds AWS/Azure connectors), integrated SIEM/SOAR (Chronicle/Mandiant), case management, and broader threat intel.

Key SCC components:
- **Security Health Analytics (SHA)** — the misconfiguration scanner (public buckets, open firewall rules, no-MFA, over-broad IAM, unencrypted resources, etc.). Findings are the core posture signal.
- **Cloud Asset Inventory** — the live inventory and config history underneath SCC.
- **Security posture service** — define and deploy a **posture** (a set of policies/constraints) as code, with drift detection.
- **Attack path simulation** (Premium+) — computes **attack exposure scores** showing how an external attacker could reach high-value resources; prioritize the path, not raw findings.

## IAM right-sizing (CIEM)

- **IAM Recommender** (part of **Active Assist**) analyzes 90 days of usage and recommends removing unused/over-broad roles, moving from **basic roles** (`roles/owner`, `roles/editor`, `roles/viewer`) to **predefined or custom roles** at least privilege.
- **Policy Analyzer** answers "who can access what."
- **Policy Intelligence** flags over-granted access.
SCC Premium surfaces excessive permissions as findings. See `cnapp-and-ciem.md`.

## Organization Policy Service — preventive guardrails

Org Policy applies **constraints** across the resource hierarchy to *prevent* misconfigurations rather than just detect them. High-value constraints:
- `constraints/storage.publicAccessPrevention` — block public buckets org-wide.
- `constraints/compute.requireOsLogin`, `constraints/compute.vmExternalIpAccess` (deny external IPs).
- `constraints/iam.disableServiceAccountKeyCreation` — stop long-lived SA keys.
- `constraints/sql.restrictPublicIp` — no public IP on Cloud SQL.
Pair with **VPC Service Controls** to build a service perimeter around sensitive data (anti-exfiltration).

## Common GCP misconfigurations and fixes

**Public GCS bucket**
```bash
# Enforce org/bucket-level public access prevention:
gcloud storage buckets update gs://my-bucket --public-access-prevention
# Remove an allUsers/allAuthenticatedUsers grant if present:
gcloud storage buckets remove-iam-policy-binding gs://my-bucket \
  --member=allUsers --role=roles/storage.objectViewer
```

**Firewall rule open to `0.0.0.0/0`** (SSH/RDP)
```bash
gcloud compute firewall-rules list --format="table(name,sourceRanges.list(),allowed[].map().firewall_rule().list())"
gcloud compute firewall-rules update allow-ssh --source-ranges=203.0.113.0/24
# Prefer Identity-Aware Proxy (IAP) for SSH/RDP instead of public ingress.
```

**VM with external IP** — remove it / enforce via org policy:
```bash
gcloud compute instances delete-access-config my-vm \
  --access-config-name "External NAT" --zone us-central1-a
```

**Over-privileged IAM (basic role at project)**
```bash
gcloud projects get-iam-policy my-project --format=json
# Replace owner/editor with a scoped predefined role:
gcloud projects remove-iam-policy-binding my-project \
  --member=user:dev@example.com --role=roles/editor
gcloud projects add-iam-policy-binding my-project \
  --member=user:dev@example.com --role=roles/compute.viewer
```

**Service-account key sprawl** — disable key creation via org policy; use **Workload Identity Federation** / attached service accounts instead of downloaded JSON keys.

**Cloud SQL public IP**
```bash
gcloud sql instances patch my-instance --no-assign-ip
# Use Private IP / Private Service Connect.
```

**Audit logging gaps** — ensure **Cloud Audit Logs** Admin Activity (always on) plus **Data Access** logs are enabled for sensitive services; export to a logging sink / BigQuery / SCC; alert on critical changes.

**CMEK for sensitive data** — GCP encrypts at rest by default; for regulated data use **customer-managed encryption keys (CMEK)** in Cloud KMS with rotation.

## Compliance reporting

SCC Premium maps findings to **CIS GCP Benchmark, PCI DSS, NIST 800-53, ISO 27001** in the compliance dashboard; export reports for auditors. See `compliance-frameworks.md`.

## Shift-left on GCP

Scan **Terraform** (the dominant GCP IaC) with Checkov / Trivy / terrascan; **Deployment Manager** (YAML/Python) and **Config Connector** manifests with KICS/Checkov. Use the SCC **security posture service** to deploy posture-as-code and detect drift. See `remediation-and-iac-scanning.md`.
