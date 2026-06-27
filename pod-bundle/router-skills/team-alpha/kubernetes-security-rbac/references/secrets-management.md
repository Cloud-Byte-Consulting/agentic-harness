# Secrets Management

How Kubernetes stores secret data, why "encrypted" is mostly a myth out of the box, and the four real strategies
plus how workloads consume secrets.

## Table of contents
- Secret objects and types
- base64 is encoding, not encryption
- Threat model: at rest, in transit, in use
- Encryption at rest (etcd) + KMS
- The four strategies (plain / Sealed Secrets / external manager / hybrid)
- HashiCorp Vault: TokenReview + sidecar injector
- External Secrets Operator & the CSI driver
- SOPS
- Consuming secrets: volume vs env vs API
- RBAC and the view role

## Secret objects and types

A Secret looks like a ConfigMap but its `data` values are base64-encoded (allowing binary) and it carries a
`type` that the API server may validate:
```yaml
apiVersion: v1
kind: Secret
metadata: { name: db-creds, namespace: app }
type: Opaque                # most common; no structure enforced
data:
  username: YWRtaW4=        # base64("admin")
  password: czNjcjN0        # base64("s3cr3t")
```
Use `stringData:` to supply plaintext that the API server encodes for you. Typed Secrets enforce required keys —
e.g. `type: kubernetes.io/tls` must have `tls.crt` and `tls.key`; `kubernetes.io/dockerconfigjson` for registry
creds; `kubernetes.io/service-account-token` for SA tokens.

## base64 is encoding, not encryption

`data` is base64 so binary/case-sensitive bytes survive the YAML→JSON→etcd round-trip intact — **it provides zero
confidentiality**. Anyone who can read the Secret (or etcd) sees the value with one `base64 -d`. Encryption needs
a key; encoding does not.

## Threat model: at rest, in transit, in use

Model the threat before adding "security" that only adds complexity:
- **At rest (etcd):** Secrets are stored **unencrypted by default**. Encryption at rest helps confidentiality +
  integrity, but the decryption key lives on the API-server host — if the host is compromised, attacker has key +
  data. It satisfies most compliance ("check the box") more than it adds real security. CIA triad: helps C and I;
  badly-done key rotation can hurt A (availability).
- **In transit:** always encrypt. cert-manager + an internal CA make this nearly free, and the availability risk
  of a rotated cert is low (retries mitigate). No reason not to.
- **In use (your app):** the weakest link. If your app has an RCE, no secrets manager saves you — the app holds
  the secret to use it. Don't over-engineer storage at the expense of the runtime; layer runtime policy
  (KubeArmor) to restrict which process can read a mounted secret.

## Encryption at rest (etcd) + KMS

Configure an `EncryptionConfiguration` referenced by the API server's `--encryption-provider-config`:
```yaml
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources: ["secrets"]
    providers:
      - kms:                       # preferred: key in an external KMS/HSM
          apiVersion: v2
          name: my-kms
          endpoint: unix:///var/run/kms.sock
      - aescbc:                    # fallback local key (better than identity, weaker than KMS)
          keys:
            - name: key1
              secret: <base64 32-byte key>
      - identity: {}               # last resort = no encryption; keep last so reads still work
```
List the first provider you want for *writes*; `identity` last so existing plaintext stays readable during
migration. After changing keys, rewrite all Secrets: `kubectl get secrets -A -o json | kubectl replace -f -`.
A **KMS provider** (envelope encryption) keeps the key-encryption-key off the host and is the meaningful upgrade.

## The four strategies

1. **Plain Kubernetes Secrets** — standard API, RBAC-restrictable, easy to mount. Usually *fine* given a sane
   threat model. Turn on encryption at rest. Limit: you can't stop in-namespace mounting (namespace = boundary).
2. **Sealed Secrets** (Bitnami) — encrypt a Secret into a `SealedSecret` you can store in Git; a controller
   decrypts it in-cluster. **Anti-pattern**: it puts (encrypted) secret data in Git, which is trivially forked/
   leaked, and recovery after key loss is brutal (rotate every secret + key, re-seal, re-push). Never put secret
   data in Git, even encrypted. Storing *metadata* about secrets in Git is fine.
3. **External secrets manager** (Vault, cloud SM, CyberArk Conjur) — source of truth outside the cluster, with
   authentication via the workload's own identity, rich policies, and **read** auditing (a key compliance need).
4. **Hybrid** — keep the source of truth in Vault but **sync into native Secret objects** so workloads use the
   ordinary Secrets API. Best of both: GitOps-friendly metadata, central audit, no app coupling to a vault SDK.

## HashiCorp Vault: TokenReview + sidecar injector

Authenticate pods to Vault with their **projected SA token** (not a static key). Vault submits a **TokenReview**
to the API server, which confirms the token is valid *and the backing pod still exists* — so an exfiltrated token
from a dead pod is rejected. Map a SA to a Vault policy:
```bash
vault write auth/kubernetes/role/app \
  bound_service_account_names=app-sa \
  bound_service_account_namespaces=app \
  policies=app-policy ttl=24h
```

Vault Agent Injector (a mutating webhook) injects a sidecar based on pod annotations — the app stays
Vault-unaware:
```yaml
metadata:
  annotations:
    vault.hashicorp.com/agent-inject: "true"
    vault.hashicorp.com/role: "app"
    vault.hashicorp.com/agent-inject-secret-db: "secret/data/app/config"
    vault.hashicorp.com/secret-volume-path-db: "/etc/secrets"
    vault.hashicorp.com/agent-inject-template-db: |
      {{- with secret "secret/data/app/config" -}}
      DB_PASSWORD="{{ .Data.data.password }}"
      {{- end }}
spec:
  serviceAccountName: app-sa            # the identity used to auth to Vault
```
In production, mount the Vault CA instead of `vault.hashicorp.com/tls-skip-verify: "true"`.

## External Secrets Operator (ESO)

Syncs from a provider directly into native Secret objects (no pod mount needed first). Two objects:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata: { name: vault-backend, namespace: app }
spec:
  provider:
    vault:
      server: "https://vault.example.com"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "app"
          serviceAccountRef: { name: app-sa }
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata: { name: db-creds, namespace: app }
spec:
  refreshInterval: 1m
  secretStoreRef: { kind: SecretStore, name: vault-backend }
  target:
    name: db-creds                  # the native Secret ESO will create/own
    creationPolicy: Owner
  data:
    - secretKey: password
      remoteRef:
        key: secret/data/app/config
        property: password
```
The **Secrets Store CSI Driver** is an alternative (SIG project) that mounts secrets as a volume and can sync to a
Secret, but it requires a pod to mount before syncing — ESO doesn't.

## SOPS

Mozilla SOPS encrypts the *values* of YAML/JSON files (leaving keys readable) using age, PGP, or a cloud
KMS/Vault. Common in GitOps via the Flux SOPS integration or a controller that decrypts at apply time. Because
the data key lives in an external KMS (not in Git), SOPS-in-Git is far safer than Sealed Secrets — but the
encrypted blob is still in Git, so the same "rotate on leak" discipline applies.

## Consuming secrets: volume vs env vs API

| Method        | Live updates?         | Leak risk            | Notes |
|---------------|-----------------------|----------------------|-------|
| **Volume**    | Yes (eventually)      | Low                  | **Preferred.** Files; supports full config files; `cat`-able but not in `env` dumps. |
| **Env var**   | No (needs pod restart)| High                 | Leaks in logs/debug dumps; security teams flag it. Avoid if possible. |
| **Secrets API** | On demand           | Medium               | Use the Pod's identity + an SDK; for dynamic needs only. |
| **Vault API** | On demand             | Medium               | Tightly couples you to the vendor; prefer ESO/injector. |

Volume mount example:
```yaml
spec:
  containers:
    - name: app
      image: app:1.0
      volumeMounts:
        - name: creds
          mountPath: /etc/secrets
          readOnly: true
  volumes:
    - name: creds
      secret:
        secretName: db-creds
```
Env var (note: does **not** update when the Secret changes):
```yaml
      env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef: { name: db-creds, key: password }
```

## RBAC and the view role

Secrets are their own kind specifically so RBAC can gate them: the built-in `view` ClusterRole deliberately
**excludes Secrets**, because a Secret can hold a SA token bound to a higher-privilege role — read access would
be a privilege-escalation path. This is also why you should never stuff sensitive data into a ConfigMap.
