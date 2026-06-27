# Azure Posture Assessment

Microsoft Defender for Cloud as the CSPM platform, Secure Score, MCSB, Azure Policy, entitlement management, and common Azure misconfigurations with `az`/CLI fixes.

## Microsoft Defender for Cloud (MDC) — the platform

MDC is Microsoft's unified cloud-security solution and acts as **CSPM + CWPP + CNAPP** across Azure, AWS, GCP, and hybrid. Two posture tiers:

- **Foundational CSPM** — free, on by default for Azure subscriptions. Provides asset inventory, **Secure Score**, security recommendations assessed against the **Microsoft Cloud Security Benchmark (MCSB)**, and basic compliance.
- **Defender CSPM** (paid plan) — adds **agentless scanning** (machines, containers, sensitive-data awareness), **attack path analysis**, the **Cloud Security Explorer** (graph query over the security graph), **governance rules** (assign/track remediation with owners and SLAs), regulatory-compliance dashboards, code-to-cloud DevOps posture, and **EASM** (external attack surface). This is the tier you want for real prioritization.

Enable MDC at the **Management Group root** so all current and future subscriptions inherit it. To onboard **AWS/GCP**, add a **multicloud connector** in MDC (agentless, role-based).

## Secure Score and MCSB

- **Secure Score** is MDC's posture metric — a percentage rolled up from recommendations, each weighted by impact. Drive it up by remediating recommendations grouped into **security controls**.
- **MCSB** is the default benchmark (the evolution of the old Azure Security Benchmark, rebranded 2022). It consolidates **CIS, NIST, and PCI DSS** guidance into Azure-specific **controls** (e.g., the IAM control family) and per-service **security baselines**. MDC recommendations are aligned to MCSB out of the box.

## Azure Policy — the enforcement engine

MDC recommendations are implemented as **Azure Policy** definitions. Use Policy to:
- **Audit** non-compliant resources (default for posture checks),
- **Deny** non-compliant deployments (preventive guardrail),
- **DeployIfNotExists / Modify** to auto-remediate (e.g., auto-enable diagnostic settings or encryption).
Assign **initiatives** (policy sets) like the MCSB initiative or regulatory initiatives at management-group scope.

## Entitlement management (CIEM)

**Microsoft Entra Permissions Management** (formerly CloudKnox) is the CIEM offering — multicloud (Azure/AWS/GCP). It computes a **Permissions Creep Index (PCI)** (granted vs used), surfaces over-permissioned identities, and right-sizes roles. Pair with **Entra ID PIM** (Privileged Identity Management) for **Just-in-Time** elevation and **Just-Enough-Administration**. See `cnapp-and-ciem.md`.

## Common Azure misconfigurations and fixes

**Storage account allows public blob access**
```bash
az storage account update -n mystorage -g my-rg --allow-blob-public-access false
# Also enforce: require HTTPS only
az storage account update -n mystorage -g my-rg --https-only true \
  --min-tls-version TLS1_2
```

**NSG rule open to the internet** (SSH/RDP)
```bash
# Inspect, then tighten or delete the permissive rule:
az network nsg rule list --nsg-name myNSG -g my-rg -o table
az network nsg rule update --nsg-name myNSG -g my-rg -n allow-ssh \
  --source-address-prefixes 203.0.113.0/24 --access Allow
# Prefer Azure Bastion / JIT VM access instead of any inbound 22/3389.
```
Defender for Cloud's **Just-in-Time (JIT) VM access** keeps management ports closed and opens them on-demand, time-boxed.

**Unencrypted / customer-key gaps** — Azure encrypts at rest by default with platform keys; for sensitive data use **customer-managed keys (CMK)** in **Key Vault** and enable Key Vault **purge protection** + soft delete.

**SQL Database public network access**
```bash
az sql server update -n mysqlserver -g my-rg --enable-public-network false
# Use Private Endpoints / VNet rules for connectivity.
```

**No diagnostic/activity logging** — route the **Activity Log** and resource **diagnostic settings** to a Log Analytics workspace; enable for Key Vault, NSGs, storage, SQL. Connect MDC to **Microsoft Sentinel** (SIEM/SOAR) for detection and automated response.

**Over-privileged role assignment**
```bash
az role assignment list --assignee user@contoso.com --all -o table
# Replace Owner/Contributor at broad scope with a least-privilege built-in or custom role,
# scoped to the resource group / resource:
az role assignment create --assignee user@contoso.com \
  --role "Reader" --scope /subscriptions/<sub>/resourceGroups/my-rg
```
Enforce MFA via **Conditional Access**; remove standing admin via PIM.

## Compliance dashboard

MDC's **Regulatory compliance** dashboard maps your posture to **MCSB, CIS Azure, PCI DSS, NIST 800-53, ISO 27001, SOC 2, HIPAA HITRUST**, and more — add/remove standards per subscription. Export to PDF/CSV for auditors. See `compliance-frameworks.md`.

## Shift-left on Azure

Scan **ARM/Bicep** with Checkov / KICS / Microsoft's `template-analyzer`. MDC **DevOps security** connects Azure DevOps and GitHub to surface IaC and code findings (code-to-cloud). See `remediation-and-iac-scanning.md`.
