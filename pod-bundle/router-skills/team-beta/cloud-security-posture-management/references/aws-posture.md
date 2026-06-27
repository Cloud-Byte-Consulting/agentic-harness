# AWS Posture Assessment

Native AWS services for posture management, how they fit together, and common AWS misconfigurations with CLI fixes.

## The native stack and how the pieces fit

AWS does not have one "CSPM product" — it composes several services:

- **AWS Security Hub** — the aggregation and posture layer. Runs **security standards** (CIS AWS Foundations Benchmark, **AWS Foundational Security Best Practices (FSBP)**, PCI DSS, NIST SP 800-53) as automated controls, produces pass/fail findings with a **security score**, and normalizes findings from GuardDuty, Inspector, Macie, IAM Access Analyzer, and partners into the **AWS Security Finding Format (ASFF)**. This is your single pane of glass.
- **AWS Config** — the configuration recorder and rules engine underneath. Records resource configuration over time, evaluates **Config Rules** (managed + custom Lambda/Guard), supports **conformance packs** (bundles of rules mapped to a framework), and enables **auto-remediation** via SSM Automation documents. Security Hub controls are largely backed by Config rules — **Config must be enabled** for most Security Hub standards to function.
- **IAM Access Analyzer** — finds resources shared externally (buckets, roles, KMS keys, etc.), validates policies, and generates **least-privilege policies from CloudTrail access activity** (CIEM-style).
- **IAM Access Advisor** — per-principal "last accessed" data to identify and remove unused permissions.
- **Amazon GuardDuty** — threat detection (not posture): anomalous API calls, crypto-mining, exfiltration. Pairs with posture data for attack-path/triage context.
- **Amazon Inspector** — vulnerability scanning for EC2, ECR images, and Lambda (CVE/CVSS).
- **Amazon Macie** — sensitive-data discovery in S3 (DSPM-adjacent).

> 2024 note: AWS announced a **next-generation Security Hub** that adds risk prioritization and attack-path-style correlation across these signals. Verify current naming/availability in-console.

## Multi-account onboarding (do this first)

1. Use **AWS Organizations**.
2. Designate a **delegated administrator** account for Security Hub, GuardDuty, Config, and Access Analyzer (keep the management account clean).
3. Enable **Config aggregators** and **Security Hub cross-region/cross-account aggregation** so all accounts/regions roll up.
4. For third-party CSPM, create a **cross-account IAM role** the tool assumes, attached to the AWS-managed `SecurityAudit` (and often `ViewOnlyAccess`) policy — agentless, read-only.

## Picking standards

Enable in Security Hub: **CIS AWS Foundations** (baseline) + **FSBP** (broadest AWS-specific coverage), then **PCI DSS** / **NIST 800-53** as compliance requires. Use Config **conformance packs** for frameworks Security Hub doesn't cover directly. See `compliance-frameworks.md`.

## Common AWS misconfigurations and fixes

**Public S3 bucket**
```bash
# Account-wide guardrail (preferred):
aws s3control put-public-access-block --account-id 111122223333 \
  --public-access-block-configuration \
  BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
# Per-bucket:
aws s3api put-public-access-block --bucket my-bucket \
  --public-access-block-configuration \
  BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
```

**Security group open to the world** (SSH/RDP)
```bash
aws ec2 revoke-security-group-ingress --group-id sg-0abc \
  --protocol tcp --port 22 --cidr 0.0.0.0/0
# Re-add scoped, or use SSM Session Manager instead of opening 22 at all.
```

**Unencrypted EBS** — turn on default encryption per region:
```bash
aws ec2 enable-ebs-encryption-by-default --region us-east-1
```

**Over-permissive IAM** — find unused permissions, then scope:
```bash
aws iam generate-service-last-accessed-details --arn arn:aws:iam::111122223333:role/AppRole
# Use IAM Access Analyzer policy generation to build a least-privilege policy from CloudTrail.
```
Eliminate `"Action":"*"` / `"Resource":"*"`; never attach `AdministratorAccess` to app roles.

**No MFA on root / root access keys exist** — delete root access keys, enable MFA on root, stop using root for daily work (CIS 1.x controls).

**CloudTrail not enabled in all regions**
```bash
aws cloudtrail create-trail --name org-trail --s3-bucket-name my-cloudtrail-logs \
  --is-multi-region-trail --is-organization-trail
aws cloudtrail start-logging --name org-trail
```

**RDS/database publicly accessible**
```bash
aws rds modify-db-instance --db-instance-identifier mydb \
  --no-publicly-accessible --apply-immediately
```

## Auto-remediation pattern

Security Hub finding → EventBridge rule → SSM Automation document (or Lambda) that applies the fix. AWS publishes **Automated Security Response (ASR)** solution playbooks for common FSBP/CIS findings. Gate with scoping and start in alert-only mode (see `remediation-and-iac-scanning.md`).

## Shift-left on AWS

Scan **CloudFormation** with cfn-nag / cfn-guard / Checkov / KICS; scan **Terraform** with Checkov / Trivy. Use **CloudFormation Hooks** or **Guard rules** to block non-compliant deploys. See `remediation-and-iac-scanning.md`.
