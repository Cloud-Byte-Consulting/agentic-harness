# Misconfiguration Catalog, Remediation & IaC Scanning

Lead with the fix. The misconfiguration classes below cause the large majority of cloud breaches; each entry gives the risk and the remediation. Then: shift-left IaC scanning and Policy as Code.

## Contents
- Network misconfigurations
- Identity (IAM) misconfigurations
- Data-protection misconfigurations
- Logging & monitoring misconfigurations
- Lateral-movement enablers
- Suspicious activity CSPM should surface
- Remediate in code, not the console
- IaC scanners (Checkov, Trivy/tfsec, KICS, cfn-nag, cfn-guard, terrascan)
- Policy as Code
- Shift-left CI/CD integration

## Network misconfigurations

- **Unrestricted inbound access (`0.0.0.0/0`)** to DBs, admin ports (22/3389), APIs, buckets. *Fix*: whitelist specific CIDRs; close unused ports; front SSH/RDP with a bastion, AWS SSM Session Manager, Azure Bastion, or GCP IAP; default-deny.
- **Inadequate segmentation.** *Fix*: separate web/app/db tiers into subnets/VPCs/VNets; allow only required cross-tier traffic; isolate sensitive resources in private subnets.
- **Weak NACL/firewall rules.** *Fix*: least privilege per rule, deny-by-default, centralize rule management, control egress as tightly as ingress (blocks exfiltration).
- **Unused security groups/rules.** *Fix*: routine audits and cleanup; naming conventions; automate the security-group lifecycle.
- **No encryption in transit.** *Fix*: enforce TLS 1.2+ everywhere, disable SSLv2/SSLv3 and weak ciphers, enable PFS, HTTPS for web, monitor/rotate certs.
- **Missing network monitoring/logging.** *Fix*: enable VPC Flow Logs / NSG Flow Logs / VPC Flow Logs (GCP); centralize; IDS/IPS; real-time alerts.
- **Improper VPN / interconnect config; DNS not secured** (spoofing/hijacking). *Fix*: limit access points; secure DNS (DNSSEC, locked zones).

## Identity (IAM) misconfigurations

These are the most critical class — IAM is the cloud perimeter.
- **Excessive permissions** (wildcards, broad managed roles). *Fix*: least privilege; replace `"Action":"*"`/`"Resource":"*"` and `roles/owner`/`Contributor`/`AdministratorAccess` with scoped grants. Use access-usage data (AWS Access Advisor / IAM Access Analyzer, Azure Entra Permissions Management, GCP IAM Recommender) to strip unused permissions. See `cnapp-and-ciem.md`.
- **Unused/stale users, roles, keys.** *Fix*: review and disable/delete dormant principals and old access keys.
- **Missing MFA** on root/admin/sensitive actions. *Fix*: enforce MFA, prefer phishing-resistant factors.
- **Shared credentials / API keys.** *Fix*: individual identities; federate via SSO; no shared secrets.
- **Privilege-escalation paths** (e.g., a role that can edit IAM policies). *Fix*: close `iam:*`/policy-edit grants on low-trust principals.
- **Unmonitored IAM changes.** *Fix*: alert on role/policy/user changes.
- **Weak role segregation / default privileges.** *Fix*: separation of duties; disable/modify permissive default roles.

## Data-protection misconfigurations

- **Public object storage** (S3/Blob/GCS). *Fix*: enable account-wide public-access blocks — **S3 Block Public Access**, Azure storage **"allow public access = disabled"**, GCP **public access prevention** — not per-bucket toggles.
- **Unencrypted data at rest.** *Fix*: enable default encryption (S3 SSE/KMS, EBS encryption-by-default, Azure SSE/CMK, GCP CMEK).
- **Insecure storage/database access controls.** *Fix*: scope IAM/ACLs; no public DB endpoints; private connectivity.
- **Exposed secrets in code/config.** *Fix*: use Secrets Manager / Key Vault / Secret Manager / HashiCorp Vault; never hardcode; scan for secrets in CI.
- **Unprotected backups; missing data classification; bad retention.** *Fix*: encrypt and access-control backups; classify by sensitivity; set retention/disposal policy.
- **Data residency violations.** *Fix*: pin resources to compliant regions.
- **Key management.** *Fix*: customer-managed keys (CMK/CMEK) for sensitive data; automate rotation and lifecycle.

## Logging & monitoring misconfigurations

Foundational CIS controls. *Fix*: enable and centralize **AWS CloudTrail (all regions, multi-account)**, **Azure Activity Logs + diagnostic settings**, **GCP Cloud Audit Logs (Admin Activity + Data Access)**; protect log integrity; alert on critical changes (root login, policy edits, new admin, security-group changes); retain per compliance.

## Lateral-movement enablers

Weak segmentation, excessive inter-resource trust, shared privileges, unrestricted internal comms, over-scoped roles, unpatched resources, missing monitoring, unsecured SSH/RDP, undetected persistence. *Fix*: micro-segmentation, tight trust relationships, least privilege, patching, and monitoring — exactly the controls above, viewed through the lens of "if one box is owned, how far can the attacker go?"

## Suspicious activity CSPM should surface

Anomalous access (impossible-travel logins), brute force, account-takeover behavior, unusual data access/exfiltration, suspicious API calls, unexpected config changes (security groups/firewall/policies), privilege escalation (vertical/horizontal), suspicious network traffic, instance/service hijacking, data manipulation/injection. CSPM flags these or feeds a SIEM/SOAR that does.

## Remediate in code, not the console

A console fix is overwritten on the next IaC apply (**config drift**). Workflow:
1. Identify the misconfiguration class above.
2. **Stopgap (live fix)** via CLI when urgent:
   - AWS: `aws s3api put-public-access-block --bucket NAME --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true`
   - AWS SG: `aws ec2 revoke-security-group-ingress --group-id sg-xxx --protocol tcp --port 22 --cidr 0.0.0.0/0`
   - Azure: `az storage account update -n NAME -g RG --allow-blob-public-access false`
   - GCP: `gcloud storage buckets update gs://NAME --no-public-access-prevention` is the *wrong* direction; use `--public-access-prevention` to enforce.
3. **Backport the fix into the template** so it sticks.
4. **Re-scan** to verify; add a policy so it can't recur.

## IaC scanners

Run these locally and in CI; fail the build on violations. All are current/maintained as of 2024–2025.

| Tool | Covers | Notes |
|---|---|---|
| **Checkov** (Bridgecrew/Prisma) | Terraform, CloudFormation, ARM/Bicep, Kubernetes, Helm, Dockerfile, Serverless | Broadest coverage; custom policies in Python or YAML; `checkov -d .` |
| **Trivy** (Aqua) | Terraform, CloudFormation, Kubernetes, Dockerfile + image/SCA/secret scan | **tfsec is now merged into Trivy** — prefer `trivy config .`; tfsec still works but is in maintenance. |
| **KICS** (Checkmarx) | Terraform, CloudFormation, ARM, Kubernetes, Ansible, Docker, OpenAPI | Large query set; `kics scan -p .` |
| **terrascan** (Tenable) | Terraform, Kubernetes, Helm | OPA/Rego-based policies. |
| **cfn-nag** | CloudFormation | Long-standing CFN linter for insecure patterns. |
| **cfn-guard** (AWS) | CloudFormation, plus general JSON/YAML | AWS's policy-as-code DSL; pairs with CloudFormation Hooks/Guard rules. |
| **kube-score / kubesec** | Kubernetes manifests | Security + reliability scoring. |

Example minimal Terraform check (Checkov flags this — public ingress + default-encryption gap):
```hcl
resource "aws_security_group" "bad" {
  ingress {                    # CKV: SSH open to the world
    from_port = 22
    to_port   = 22
    protocol  = "tcp"
    cidr_blocks = ["0.0.0.0/0"]   # fix: restrict to a known CIDR
  }
}
```

## Policy as Code

Codify org rules so they're enforced consistently and version-controlled:
- **OPA / Rego** (used by terrascan, conftest, Gatekeeper for Kubernetes admission).
- **Checkov custom policies** (Python/YAML).
- **AWS cfn-guard** rules; **Azure Policy** definitions; **GCP Organization Policy** constraints.
Treat policies with the same rigor as code: DRY/modular, peer-reviewed, tested.

## Shift-left CI/CD integration

The cheapest fix is pre-deploy. Pattern:
1. **Pre-commit / local**: developers run `checkov`/`trivy config` before pushing.
2. **PR gate**: CI job (GitHub Actions, GitLab CI, Azure Pipelines, Jenkins) runs scanners; **fail on HIGH/CRITICAL**; post findings inline.
3. **Pre-apply**: scan the `terraform plan`/rendered template; a CSPM/CNAPP with IaC scanning can correlate to the same policies used at runtime ("code-to-cloud").
4. **Drift detection**: periodically compare live state to IaC; alert/auto-correct.
5. **Guardrails for auto-remediation**: scope tightly, test, provide rollback, start in alert-only mode before enabling auto-fix.

Embrace **immutable infrastructure** (replace, don't patch in place) and **SaC (Security as Code)** so security policy, RBAC, encryption settings, and secrets handling live in the same versioned pipeline as the infrastructure. This closes the loop: detect → fix in code → prevent recurrence.
