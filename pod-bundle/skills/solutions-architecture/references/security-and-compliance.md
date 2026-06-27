# Security & Compliance

Security design principles, identity, web/network/data protection, threat modeling, compliance, and DevSecOps. Read this for any security or compliance question.

## Contents
- Security design principles
- Identity and access management
- Federation, SSO, and protocols (Kerberos, AD, SAML, OAuth/OIDC, JWT)
- Web security: attacks and mitigations
- Application and OS hardening
- Network security
- Intrusion detection/prevention (IDS/IPS)
- Data security: classification, encryption, key management
- Compliance frameworks
- Shared responsibility model
- DevSecOps (see also devops-and-operations.md)

## Security design principles

- **Authentication and authorization control** — centralize user management; start authorization from **least privilege** (no access by default, grant only what the role needs); use role/group-based access; enable SSO and MFA; rotate credentials; deactivate inactive users.
- **Apply security everywhere (defense in depth)** — don't rely on the perimeter alone. Layer controls: edge/DNS protection, WAF and load-balancer filtering, per-instance firewalls, OS hardening/antivirus, IDS/IPS. Secure every layer and component.
- **Reduce the blast radius** — isolate the system into small segments (separate networks for load balancer, web, app, DB); least privilege; MFA; temporary credentials and frequent key rotation for programmatic access.
- **Monitor and audit everything, all the time** — log every activity, transaction, and API call to centralized, tamper-resistant storage; alert and respond proactively; many compliance regimes require audit logging.
- **Automate everything** — automated remediation reverts undesired changes (e.g. an open firewall port, a new admin user); security as version-controlled code scales operations.
- **Protect data** — classify by sensitivity and protect accordingly; minimize direct human access via tooling; encrypt and restrict access. (See Data security below.)
- **Be ready for incidents** — define an incident-management process appropriate to data sensitivity (PII demands tighter response); simulate it; automate detection/investigation/response; do RCA to prevent recurrence.

## Identity and access management

Differentiate corporate users (employees/contractors/vendors — special privileges, access to ERP/HR/payroll, hundreds–thousands of users) from end users (customers — thousands–millions, internet-facing, higher threat exposure). Centralize policy (strong passwords, rotation, MFA). MFA providers: Google Authenticator, RSA SecurID, Duo, YubiKey, Microsoft Authenticator. Use **role-based access (RBA)**: create groups per role (admin/developer/tester) with matching policies; assign users to groups (e.g. devs get full dev access, read-only prod).

## Federation, SSO, and protocols

- **Federated Identity Management (FIM)** — the user authenticates to a trusted identity provider (IdP); the service provider trusts the IdP rather than the user's credentials directly.
- **SSO** — one set of credentials accesses multiple services.
- **Kerberos** — ticket-based SSO authentication via a Key Distribution Center (Authentication Server + Ticket-Granting Server); issues a TGT, then service tickets.
- **Microsoft Active Directory** — LDAP-based identity service (AD DS domain controller); related services: AD LDS (LDAP interface), AD CS (certificates/PKI), AD FS (federation for external/web logins). Cloud directory services connect on-prem AD to the cloud.
- **SAML 2.0** — XML-based assertion establishing trust between an IdP and a service provider; popular for enterprise SSO/federation.
- **OAuth 2.0 / OIDC** — for large consumer user bases (social/e-commerce). OAuth 2.0 is an *authorization* protocol delegating access via tokens (no password sharing). **OIDC** layers *authentication* on top of OAuth 2.0 (verifies user identity and returns profile info). Use OAuth/OIDC over SAML for very large user bases.
- **JWT** — compact, self-contained, digitally signed JSON token (header.payload.signature); smaller than SAML; ideal for passing identity between microservices and as access tokens in web/mobile apps.

Packaged solutions: Amazon Cognito, Okta, Ping Identity reduce the implementation burden.

## Web security: attacks and mitigations

Common attacks:
- **DoS / DDoS** — flood the target to deny service to legitimate users; often via botnets. Application-layer: DNS floods, SSL-negotiation attacks. Infrastructure-layer: UDP reflection, SYN floods.
- **SQL injection (SQLi)** — inject malicious SQL to read/modify data (e.g. `... WHERE loanId = 117 or '1=1'`).
- **Cross-site scripting (XSS)** — inject client-side script into a trusted site to steal cookies/tokens.
- **Cross-site request forgery (CSRF)** — trick an authenticated user into a state-changing request (password change, money transfer).
- **Buffer overflow / memory corruption** — overwrite memory to execute attacker code.

Mitigations:
- **WAF** — applies rules to HTTP/HTTPS traffic (IP, headers, body, URI, geolocation); blocks XSS/SQLi; rate-limiting; allow/deny lists; reusable rule sets. Can be staged between load balancers ("WAF sandwich").
- **DDoS mitigation** — reduce the attack surface (don't expose what needn't be public; open the load balancer, not web servers; hide/restrict entry points); mitigate at the CDN edge (CloudFront blocks malformed connections, isolates geographically); scale horizontally/vertically with sane maximum limits to cap cost.
- Keep up with security patches; follow OWASP secure-coding practices; enforce authentication/authorization.

## Application and OS hardening

Limit attacks to the application level: harden permissions on files/folders/partitions; avoid root privileges for the app or its users; restrict process memory/CPU to prevent DoS; isolate each app to a directory with only the required access. Automate app restart (systemd/System V, DAEMON Tools, Supervisord). Apply the latest OS security patches (automate where possible; use a CI/CD pipeline with automated testing so patches don't break the app). Managed cloud services offload patching.

## Network security

Apply security at every layer with trusted boundaries and minimal access (AWS example, generalizable):
- **VPC** — logically isolated network; define IP range via CIDR (e.g. `10.0.0.0/16` = 65,535 addresses).
- **Subnets** — organize by internet access (public vs private), not by tier. Put internet-facing resources (public load balancers, NAT, bastion) in **public subnets**; app and DB in **private subnets**. Allocate more addresses to private subnets.
- **Route tables** — per-subnet custom routes.
- **Security groups** — stateful virtual firewalls per instance; deny-all by default; allow specific protocols/ports.
- **NACLs** — stateless subnet-level firewall; must define inbound *and* outbound rules explicitly.
- **Internet gateway (IGW)** — makes a subnet public; **NAT gateway** — lets private-subnet instances reach the internet outbound (e.g. for patches) without inbound exposure.
- **Bastion host** — hardened jump server into private subnets; use public-key auth.
- **VPC flow logs** — capture accepted/rejected traffic for diagnostics, security, and policy review; set alarms on anomalies.

## Intrusion detection/prevention (IDS/IPS)

- **IDS** — detects attacks by recognizing patterns; monitoring/detection only.
- **IPS** — detects *and* blocks malicious traffic; sits behind the firewall. Detection methods: **signature-based** (known-exploit patterns) and **statistical anomaly-based** (deviation from a baseline).
- **Host-based** (agent per host) — deep per-host inspection, scales per host, but adds config-management overhead and struggles with coordinated attacks. **Network-based** (appliance in the traffic path) — single component, single view, but adds a network hop and must decrypt/re-encrypt to inspect (perf hit + risk).

## Data security

**Three states**: at rest (storage — encrypt; access controls; audits), in transit (network — TLS), in use (in memory — hardest; emerging TEEs and homomorphic encryption). Also use masking and tokenization for at-rest protection.

**Classification** (balance usability vs access; avoid direct human access):
- **Restricted** — direct harm if compromised (PII: SSN, passport, credit card, payment). Encrypt, tightly restrict.
- **Private** — sensitive, usable to plan an attack (email, phone, full name, address).
- **Public** — minimal protection (ratings, reviews, public username).

**Encryption**:
- **Symmetric** — same key encrypts/decrypts; AES (128/192/256-bit) is standard (DES is legacy).
- **Asymmetric (public-key)** — public key encrypts, private key decrypts; RSA is common; used in TLS handshakes.
- Encryption adds processing overhead — encrypt where required (sensitive data), weighing the perf trade-off.

**Encryption in transit** — use TLS/SSL (HTTPS) to defend against eavesdropping and man-in-the-middle attacks; certificates issued by a CA, secured via PKI; the TLS handshake exchanges a symmetric session key using asymmetric encryption.

**Key management** — control creation, storage, rotation, deletion, and access. **Envelope encryption**: a data key encrypts the data; a master key encrypts the data key. Managed services: AWS KMS, GCP Cloud KMS, Azure Key Vault. For stricter/regulatory needs, use a **hardware security module (HSM)** (e.g. CloudHSM) — tamper-responsive, access-controlled; deploy multiple for HA.

## Compliance frameworks

Know the major regimes and what they protect: **PCI-DSS** (payment card data), **HIPAA** (US healthcare/patient data), **GDPR** (EU data privacy), **SOC** (service-org data management controls). Comply per industry and region; design comprehensive logging and monitoring to meet audit requirements. Retention rules vary (e.g. PCI-DSS may require multi-year retention — archive to cheap cold storage).

## Shared responsibility model

In the cloud, the provider secures the physical infrastructure; the customer secures the application, data, identity, and configuration. Lock down your environment and use cloud-native monitoring/alerting/automation.

## DevSecOps

Integrate security early and throughout the SDLC ("shift left"), breaking silos between dev, ops, and security, without slowing delivery. Automate continuous security testing in the CI/CD pipeline; monitor for drift from the desired state with automated remediation. Application security testing categories: **SCA** (open-source dependency vulnerabilities/licensing), **SAST** (static analysis of source pre-compile, white-box), **DAST** (dynamic testing of the running app, black-box), **IAST** (interactive, during functional testing). See `devops-and-operations.md` for pipeline integration.
