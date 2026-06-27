# DevOps, Automation & Operational Excellence

DevOps/DevSecOps practice, CI/CD, IaC, deployment strategies, continuous testing, and operations. Read this when wiring up delivery automation or designing for operation.

## Contents
- What DevOps is and why
- Components of DevOps
- CI/CD
- Infrastructure as code (IaC)
- Configuration management
- DevSecOps and CI/CD
- Continuous deployment strategies
- Choosing a deployment strategy
- Continuous testing and A/B testing
- DevOps tools (the pipeline)
- DevOps best practices and KPIs
- Operational excellence

## What DevOps is and why

DevOps is a culture + practices that breaks the silo between development and operations (and security, in DevSecOps), with shared responsibility and continuous feedback, to deliver faster and more reliably. It involves the whole org (management, business owners, dev, QA, release, ops, sysadmin). Benefits: **speed**, **fast delivery**, **reliability**, **scalability**, **collaboration**, and **security**. Especially valuable with cloud/distributed systems.

## Components of DevOps

CI/CD, continuous monitoring and improvement, infrastructure as code, and configuration management — with **automation** as the common thread.

## CI/CD

- **Continuous Integration (CI)** — developers commit frequently to a shared repo; each commit triggers an automated build and unit/integration tests, catching defects early. Use hooks (e.g. post-receive) to trigger CI builds; pull requests for review before merge.
- **Continuous Delivery/Deployment (CD)** — extends CI to deploy builds to test/staging/prod. In continuous *delivery*, every change is *ready* for production but a human approves the final deploy (a business decision, still automated by tools). In continuous *deployment*, that final step is automated too. A robust pipeline also provisions test/prod infrastructure and stores binaries in an artifact repository (e.g. JFrog).

## Infrastructure as code (IaC)

Define infrastructure as version-controlled templates so environments are reproducible, auditable, and consistent — eliminating manual, error-prone provisioning and config drift. Tools: Terraform, AWS CloudFormation, AWS CDK, Azure Resource Manager, Google Cloud Deployment Manager, Ansible, Chef, Puppet.

```yaml
# Minimal AWS CloudFormation: an S3 bucket with a name parameter, retained on stack delete
AWSTemplateFormatVersion: '2010-09-09'
Description: 'Create an S3 bucket with a parameterized name.'
Parameters:
  BucketNameParam:
    Type: String
    Default: 'my-app-storage'
    Description: 'Enter the S3 bucket name'
    MinLength: '5'
    MaxLength: '63'
Resources:
  Bucket:
    Type: 'AWS::S3::Bucket'
    DeletionPolicy: Retain          # keep data even if the stack is torn down
    Properties:
      BucketName: !Ref BucketNameParam
      Tags:
        - Key: 'Name'
          Value: 'MyBucket'
Outputs:
  BucketName:
    Description: 'The created bucket name'
    Value: !Ref BucketNameParam
```

## Configuration management

Automation to standardize resource configuration across infrastructure and apps, ensuring consistency and enabling bulk changes. Tools compared:
- **Ansible** — any server can be the controller; agentless (SSH); YAML; playbooks/roles; sequential.
- **Puppet** — centralized Puppet master; Ruby DSL; manifests/modules; non-sequential.
- **Chef** — centralized Chef server + client agents; Ruby; recipes/cookbooks; sequential.
CM tools provide version control and audit of configuration. (Managed options exist, e.g. AWS OpsWorks for Chef/Puppet.)

## DevSecOps and CI/CD

Embed security at every pipeline stage without slowing delivery ("shift left"):
- **Code** — scan for hardcoded secrets/keys.
- **Build** — manage and tag security artifacts (keys, tokens).
- **Test** — scan configuration against security standards.
- **Deploy/Provision** — register security components; checksum to verify file integrity.
- **Monitor** — continuous audit and validation; automated remediation (e.g. auto-close an exposed SSH port; revert unauthorized admin/firewall changes).
Application security testing categories: **SCA** (dependency vulnerabilities/licensing), **SAST** (static, pre-compile, white-box), **DAST** (dynamic, running app, black-box), **IAST** (interactive, during functional tests). Centralize findings (e.g. AWS Security Hub) and automate response.

## Continuous deployment strategies

- **In-place** — update the app on existing servers in one action; some downtime; cheap and fast; rollback = redeploy.
- **Rolling** — update the fleet in subgroups; zero downtime; cost-neutral (no extra infra); a failed deploy affects only a subset; slightly longer.
- **Blue-green** — stand up an identical "green" environment with the new version alongside live "blue"; shift traffic (DNS or auto-scaling-group swap), often gradually with canary analysis; instant rollback by reverting traffic. Zero downtime but ~2x resources during the cutover.
- **Red-black (dark launch)** — like blue-green but a *sudden* DNS cutover after canary testing (vs gradual). Rollback = point DNS back. Can combine with feature flags for beta testing.
- **Immutable** — roll out a brand-new set of servers and terminate the old; best for unknown dependencies / avoiding config drift; requires efficient infra provisioning.
Always factor downtime tolerance and cost (instances to replace × deploy frequency).

## Choosing a deployment strategy

- **In-place** — small/internal apps where simplicity matters and brief downtime is acceptable; always have a rollback plan.
- **Rolling** — apps needing minimal downtime without extra resources, that can run two versions at once.
- **Blue-green** — critical apps needing zero downtime and quick rollback; needs robust load balancing / DNS switching and ~2x resources.
- **Red-black** — fast cutover, often containerized; thorough testing of the new version is essential.
- **Immutable** — consistency/reliability in cloud, complex dependencies; needs efficient provision/decommission.
Weigh application complexity, scale, user base, downtime impact, resource availability, and cost.

## Continuous testing and A/B testing

Bake testing into the pipeline. The testing pyramid: ~70% **unit tests** (fastest, cheapest, on the dev machine) plus static analysis/code coverage; then **integration/system tests** (own environments); then **performance/load/stress**, **UAT**, and **compliance** tests in a production-like staging environment. Smaller, dependency-free unit tests give fast feedback.

**A/B testing** (a production-phase technique) routes traffic across two+ versions (e.g. 90% V1.1, 7% V1.2, 3% V1.3) on isolated server fleets sharing a backend DB, gathering usage metrics to decide which version wins. **Canary analysis** routes ~1% of traffic to a new version to detect issues before full rollout.

## DevOps tools (the pipeline)

- **Code editor/IDE** — VS Code, Eclipse, Cloud9, Ace.
- **Source control** — Git via GitHub/Bitbucket or managed (e.g. CodeCommit); set auth/authz; encrypt in transit (HTTPS/SSH) and at rest.
- **CI/build server** — Jenkins (most popular; can auto-scale agents) or managed build (e.g. CodeBuild). Hooks trigger builds; PRs gate merges.
- **Deploy** — CodeDeploy, Elastic Beanstalk, Chef/Puppet, Jenkins. Deployment configs: OneAtATime, HalfAtATime, AllAtOnce, Custom. Lifecycle events: ApplicationStop → DownloadBundle → BeforeInstall → Install → AfterInstall → ApplicationStart → ValidateService.
- **Test** — Jenkins, BlazeMeter, Ghost Inspector.
- **Pipeline orchestration** — CodePipeline or Jenkins; action categories: Source, Build, Deploy, Test, Invoke, Approval.
- **Config externalization** — AWS Systems Manager Parameter Store, Kubernetes ConfigMaps/Secrets, Docker Swarm secrets, HashiCorp Consul/Vault.

## DevOps best practices and KPIs

Design the pipeline deliberately: number of stages (dev/integration/system/UAT/prod), test types per stage, sequential vs parallel tests, monitoring/reporting, infra provisioning per stage, and a rollback strategy. Automate to avoid slow manual intervention; externalize build/environment config (don't bury it in code) for consistent, scalable builds; design to "fail fast." Consider the **Twelve-Factor App** methodology.

**CI/CD KPIs**: deployment frequency, lead time for changes, change failure rate, mean time to recovery (MTTR), and automated test pass rate. Other DevOps metrics: change volume, % failed deployments, availability/SLA violations, customer complaint volume, % change in user volume.

## Operational excellence

Design for operation: plan logging, monitoring, and alerting to capture every incident and act fast; automate deployment and remediation to avoid human error; include security/compliance (which may change over time). Maintenance can be proactive (modernize on a new OS release) or reactive (wait for end-of-life), but always change in small increments with a rollback strategy via CI/CD and blue-green/canary launches. For operational readiness, maintain a **runbook** (routine activities) and a **playbook** (guidance through issues), and use **root cause analysis** for post-incident reporting so issues don't recur. Continuous monitoring (e.g. CloudWatch) feeds self-healing automation. Every failure is an opportunity to improve.
