# Cost Optimization & FinOps

Design principles and techniques for cost-aware architecture. Read this for any cost/FinOps question. Cost optimization maximizes ROI and reduces risk — it is *not* blind cost cutting that harms the customer experience.

## Contents
- Cost design principles
- Total cost of ownership (TCO)
- Budget vs forecast
- Managing demand and service catalogs
- Tracking expenditure: show-back vs charge-back
- Continuous cost optimization and right-sizing
- Reducing architectural complexity and increasing IT efficiency
- Standardization, governance, tagging, account structure
- Cloud pricing models
- Green IT

## Cost design principles

Cost is everyone's responsibility across the whole lifecycle (planning → post-production). Core principles:
- **Calculate TCO**, not just upfront price.
- **Plan budget and forecast**; monitor against both.
- **Manage demand and service catalogs** to align supply with need.
- **Keep track of expenditure** and tie it to owners (show-back/charge-back).
- **Continuously optimize** — never stop until the cost of finding savings exceeds the savings.

## Total cost of ownership (TCO)

TCO = **CapEx** (upfront acquisition: software/licenses, hardware, implementation, migration) + **OpEx** (ongoing: maintenance/support, patching/updates, customization, data-center cost, security, license renewals, admin/IT staff, consultants, training/tools). Decide build-vs-buy and on-prem-vs-cloud on TCO and ROI, not upfront price — like buying an energy-efficient appliance: higher upfront, lower total cost. SaaS (per-user subscription) can win for moderate user counts; IaaS + off-the-shelf for larger; build only if nothing fits.

## Budget vs forecast

- **Budget** — a detailed, longer-term (e.g. annual) financial plan of expected revenue/expense/allocation; adjusted infrequently; used for strategic planning and performance evaluation (planned vs actual).
- **Forecast** — a dynamic, frequently updated (monthly/quarterly) projection based on current trends; used for tactical operational decisions.
Use the forecast to act now (e.g. "at this rate you'll exceed the $450 monthly budget by end of November — adjust").

## Managing demand and service catalogs

Aggregate demand across business units for economies of scale and better pricing (e.g. cloud **private pricing agreements / enterprise discount programs** for committed spend). Two approaches:
- **Demand management** — for existing environments with overspend; analyze historical data, find overprovisioning, and streamline.
- **Service catalog management** — for new services without history; offer a catalog of common, pre-priced building blocks (e.g. "small Linux + MySQL dev environment") so teams self-serve within limits.

## Tracking expenditure: show-back vs charge-back

Link costs to systems/owners for transparency and accountability.
- **Show-back** — inform each unit of its spend without billing it. Start here as the org matures.
- **Charge-back** — bill each unit for its consumption under a master payee account. Adopt as maturity grows (often charge-back at department/BU level, show-back at team level).
Configure budget/forecast alerts so teams are notified as they approach thresholds.

## Continuous cost optimization and right-sizing

Continuously hunt for idle/underused resources (e.g. shut down dev instances nights/weekends — up to ~70% workspace savings; spin up batch systems only to run jobs). Avoid biased utilization metrics: don't size for peak-only data (Black Friday) — analyze in context to prevent overprovisioning. Apply archival policies to control storage; check DB deployment needs (multi-AZ? provisioned IOPS?). 

**Right-sizing best practices**:
- Ensure monitoring reflects the end-user experience; use p99, not averages.
- Pick the right monitoring cycle (hourly/daily/weekly) to catch periodic peaks.
- Assess the cost of change against the saving (testing/effort).
- Match utilization to business requirements (expected requests at month-end/peak).
Use monitoring tools (CloudWatch, Splunk) and custom metrics (CPU, RAM, network, connections) to identify over/under-utilization. Set measurable goals (e.g. reduce cost per transaction by 10% per quarter) aligned across org/team levels.

## Reducing architectural complexity and increasing IT efficiency

Decentralized BUs build duplicate systems and inconsistent data, raising cost and risk. Reduce complexity with **standardization, reusable architecture patterns, and shared services** (a service catalog), plus automation. Eliminate duplication; reuse via RESTful/service-oriented and microservice design (e.g. one payment service reused by e-commerce and vendor payments). A centralized IT architecture team aligns BUs to the company vision and reduces tech debt.

Increase IT efficiency: optimize/retire unused software licenses; negotiate bulk discounts; re-evaluate high-cost projects against business value; cancel non-compliant low-value projects; decommission unused apps; modernize legacy to cut maintenance; consolidate data, vendors, and duplicate systems (payment, access management); eliminate overprovisioned waste. Move to the cloud's pay-as-you-go model and automate provisioning/monitoring/processing. **Trade off carefully** — cost cuts that degrade the customer experience add business risk (the theme-park-with-fewer-rides example) and are the wrong kind of optimization.

## Standardization, governance, tagging, account structure

Set resource limits across the org; use a service catalog with **IaC** (Terraform, CloudFormation, Ansible) to prevent overprovisioning and config drift, version-control infrastructure, and ensure consistency. Attribute cost via **resource tagging** (project, environment, department, cost center, owner) and **account/organization-unit structure** (e.g. OUs for HR and Finance, each with department accounts) — enabling granular cost visibility, charge-back, and consolidated vendor spend, plus consistent security/compliance. Engage all stakeholders (CFO, app/department owners, vendors) in usage and cost discussions; require vendors to provide cost analysis aligned to your financial goals.

## Cloud pricing models

Public clouds trade CapEx for variable OpEx (economies of scale, continued price reductions). Use:
- **On-demand / pay-as-you-go** — agility for variable workloads.
- **Savings plans / reserved instances** — commit to usage/spend for steep discounts on predictable, steady workloads. Analyze usage data to size commitments.
- **Spot** — deeply discounted spare capacity for interruptible workloads.
- **Managed services** — eliminate infra maintenance and monitoring overhead, lowering TCO as adoption grows.
Set service limits per account (e.g. dev capped at 10 servers) for governance. Tools surface savings: AWS Trusted Advisor (idle resources, e.g. an idle load balancer to shut down), Cost Explorer (analyze spend over time), AWS Budgets (thresholds and alerts). Cost-and-usage and forecast reports plus alerts (at, say, 50%/80% of budget) enable proactive control.

## Green IT

Environmentally sustainable computing that also cuts cost: energy-efficient hardware, virtualization (fewer physical servers, less cooling/space), cloud (more efficient data centers, economies of scale), hardware recycling/reuse and refurbished purchases, telecommuting, electronic document management, sustainable procurement (durable/long-warranty), maintenance optimization (extend equipment life), efficient asset disposal, and carbon credits. Cloud providers offer sustainability tooling (e.g. AWS customer carbon footprint tool; the AWS Well-Architected sustainability pillar). Serverless + auto-scaling + right-sized managed services align cost optimization with sustainability (pay only for what you use, on efficient infrastructure).
