# Authentication in Kubernetes

How the API server decides *who you are*, before any authorization. Covers the request flow, every authn method,
ServiceAccount/projected tokens, and private-registry credentials.

## Table of contents
- The authn → authz → admission flow
- There are no User objects
- x509 client certificates (+ CSR API)
- OpenID Connect (OIDC)
- ServiceAccount tokens (legacy, projected/bound, TokenRequest)
- Static token files (anti-pattern)
- Webhook & authenticating proxy, bootstrap tokens
- Private registry credentials (imagePullSecrets)

## The authn → authz → admission flow

Every API request hits the API server and passes through, in order:

1. **Authentication** — a chain of authenticator modules runs until one succeeds, producing a username, UID, and
   groups. If none succeed → `401 Unauthorized` (or mapped to `system:anonymous` / group `system:unauthenticated`
   if anonymous access is enabled).
2. **Authorization** — modules (RBAC, Node, Webhook, …) decide allow/deny from the request attributes. Deny →
   `403 Forbidden`.
3. **Admission control** — mutating webhooks (may rewrite the object), then validating webhooks + built-in
   controllers (may reject). Does not apply to read verbs.

`401` means the server can't identify you (expired/missing token, bad cert). `403` means it identified you but
RBAC denies the action. This distinction drives almost all access debugging.

Verify what the server thinks of you:
```bash
kubectl auth whoami
# ATTRIBUTE   VALUE
# Username    jane@example.com
# Groups      [devs system:authenticated]
```

## There are no User objects

Kubernetes stores **no User or Group objects**. Identity is asserted at authn time:
- **Certificates**: username = certificate Subject CN; groups = the O (Organization) fields.
- **OIDC**: username/groups come from configured token claims.
- **ServiceAccounts**: the only real identity objects; username is
  `system:serviceaccount:<namespace>:<name>`.

System groups are assigned by the API server (`system:authenticated`, `system:serviceaccounts`,
`system:serviceaccounts:<ns>`). User-asserted groups have no naming rules and exist only for the duration of the
request.

## x509 client certificates (+ CSR API)

Strong, industry-standard, good for break-glass and high-privilege accounts — but the API server **cannot revoke
a cert** (no CRL/OCSP support), so a leaked cert is valid until expiry. Use short lifetimes and tight storage.

The API server validates client certs against `--client-ca-file`. Generate and approve a cert via the
Certificates API:

```bash
openssl genrsa -out jane.key 2048
openssl req -new -key jane.key -out jane.csr -subj "/CN=jane/O=devs"   # CN=user, O=group(s)
```

```yaml
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: jane
spec:
  request: <base64 of jane.csr, e.g. `cat jane.csr | base64 -w 0`>
  signerName: kubernetes.io/kube-apiserver-client
  expirationSeconds: 86400        # 1 day — keep client certs short-lived
  usages:
    - client auth
```

```bash
kubectl apply -f csr.yaml
kubectl certificate approve jane
kubectl get csr jane -o jsonpath='{.status.certificate}' | base64 -d > jane.crt
kubectl config set-credentials jane --client-key=jane.key --client-certificate=jane.crt
```

The user can authenticate immediately but does nothing until RBAC binds the CN/groups.

## OpenID Connect (OIDC)

The recommended method for human users and SSO. Built on OAuth2; `kubectl` supports it natively (no plugin
required). Benefits: short-lived `id_token`s (limit blast radius if leaked), group claims for RBAC, refresh
tokens scoped to your idle-timeout policy, MFA via the IdP.

### The three tokens
- `access_token` — for the IdP's own APIs; **not used by Kubernetes**, discard.
- `id_token` — a signed JWT carrying your identity; **this is sent to the API server** in the `Authorization:
  Bearer` header on every request. Self-contained: the API server verifies the signature against the IdP's public
  key and checks `iss`, `aud`, and expiry — no callback to the IdP.
- `refresh_token` — used by `kubectl` to silently get a new `id_token` from the IdP when the old one expires;
  never sent to Kubernetes, one-time use.

### Key id_token claims
`iss` (issuer, must match API server config), `aud` (client ID), `sub` (immutable unique ID — **use this as the
username claim, never `mail`/email**), `groups` (non-standard but used for RBAC), `exp`/`iat`/`nbf` (expiry/issued/
not-before, in epoch seconds; nbf gives clock-skew tolerance).

### API server flags
```
--oidc-issuer-url=https://idp.example.com/dex
--oidc-client-id=kubernetes
--oidc-username-claim=sub
--oidc-groups-claim=groups
--oidc-ca-file=/etc/kubernetes/pki/oidc-ca.pem
```
(Newer clusters can use structured `AuthenticationConfiguration` instead of flags.) The IdP must support the
discovery endpoint (`<issuer>/.well-known/openid-configuration`), serve TLS, and present a CA-flagged cert.

### Configuring kubectl
```bash
kubectl config set-credentials jane --auth-provider=oidc \
  --auth-provider-arg=idp-issuer-url=https://idp.example.com/dex \
  --auth-provider-arg=client-id=kubernetes \
  --auth-provider-arg=id-token=$ID_TOKEN \
  --auth-provider-arg=refresh-token=$REFRESH_TOKEN
```
Prefer a public client + **PKCE** over a `client_secret` (a shared client secret has no real security value and
becomes a rotation headache). Managed clusters (EKS/AKS/GKE) wire up their own OIDC; self-managed clusters use
Dex, Keycloak, OpenUnison, or UAA. For managed clusters that don't expose API-server flags, use an
**impersonating reverse proxy** (it authenticates the user, then sends `Impersonate-User`/`Impersonate-Group`
headers using its own privileged identity — keep that identity locked down and short-lived).

## ServiceAccount tokens

ServiceAccounts give in-cluster workloads an identity. A Pod's SA is set via `spec.serviceAccountName`; if unset,
the namespace `default` SA is used. The token is mounted at
`/var/run/secrets/kubernetes.io/serviceaccount/token` (alongside `ca.crt` and `namespace`).

### Projected / bound tokens (1.22+ default; static tokens disabled by default 1.24+)
Modern tokens are obtained via the **TokenRequest API** and *projected* into the pod as a short-lived, audience-
scoped volume. They are bound to the pod's lifetime — when the pod dies, the token is invalid even if not yet
expired. This is far more secure than the old static `Secret`-backed tokens (which never expired and stored the
token in plain text).

Generate a short-lived token on demand (good for bootstrapping/testing, **not** for normal user auth):
```bash
TOKEN=$(kubectl create token my-sa -n app --duration=1h)
curl -H "Authorization: Bearer $TOKEN" --cacert ca.crt https://<apiserver>:6443/api
```

Opt a workload out of auto-mounting a token (reduces attack surface for pods that never call the API):
```yaml
apiVersion: v1
kind: ServiceAccount
metadata: { name: my-sa, namespace: app }
automountServiceAccountToken: false   # can also be set per-Pod (Pod wins)
```

Creating a *static* long-lived SA token Secret is a legacy anti-pattern; only do it when a tool genuinely needs a
non-expiring token, and monitor its use:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-sa-token
  annotations: { kubernetes.io/service-account.name: my-sa }
type: kubernetes.io/service-account-token
```

External services (e.g. Vault) validate a projected SA token by submitting a **TokenReview** to the API server,
which also confirms the backing pod still exists — so an exfiltrated token from a dead pod is rejected.

## Static token files (anti-pattern)

A CSV (`token,user,uid,"group1,group2"`) passed via `--token-auth-file`. Easy but insecure: plain-text tokens, no
rotation without an API-server restart, must be replicated to every control-plane node. Dev/learning only.
(HTTP basic-auth files were removed in 1.19.)

## Webhook & authenticating proxy, bootstrap tokens

- **Webhook token authn**: the API server POSTs a `TokenReview` to an external service that returns the
  user/groups. Used by cloud providers' IAM integrations. Don't build your own unless you *are* a platform vendor.
- **Authenticating proxy**: a trusted proxy authenticates the user and passes identity via configured HTTP
  headers. For integrating LDAP/Kerberos or other non-native IdPs.
- **Bootstrap tokens**: short-lived secrets in `kube-system` used to join new nodes (kubelet TLS bootstrapping),
  not for general auth.

## Private registry credentials (imagePullSecrets)

To pull from a private registry, create a `docker-registry` Secret and reference it from the Pod (or the SA):
```bash
kubectl create secret docker-registry my-reg \
  --docker-server=registry.example.com \
  --docker-username=user --docker-password=*** --docker-email=me@example.com
```
```yaml
apiVersion: v1
kind: Pod
metadata: { name: app }
spec:
  containers:
    - name: app
      image: registry.example.com/app:1.4.2
  imagePullSecrets:
    - name: my-reg
```
Attach it to a ServiceAccount (`imagePullSecrets` on the SA) so every pod using that SA inherits it without
per-pod config.
