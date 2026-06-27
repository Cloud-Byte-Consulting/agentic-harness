# Secrets in GitOps

GitOps wants *everything* in Git, but plaintext secrets must never be committed (base64 is encoding, not
encryption). Three industry-standard approaches let you keep secrets in (or referenced from) Git safely. Pick one
based on whether you have an external vault. Read this for full manifests and tradeoffs.

## Decision

| Tool | What's in Git | Decryption | Best when |
|---|---|---|---|
| **Sealed Secrets** | Encrypted `SealedSecret` | In-cluster controller (asymmetric key) | On-prem / no external vault; simplest, industry default |
| **External Secrets Operator (ESO)** | A *reference* (`ExternalSecret`) — no secret data | Operator fetches from vault at runtime | You already run AWS/Azure/GCP secret manager or Vault |
| **SOPS** | Encrypted file (values encrypted in place) | Flux native / Argo plugin (age/KMS/PGP) | You want encrypted *files* in Git and use Flux |

## Sealed Secrets

A controller in the cluster holds a private key; `kubeseal` encrypts with the matching public cert. Only that
controller can decrypt, so the resulting `SealedSecret` is safe to commit. On apply, the controller decrypts it
into a normal `Secret`.

```bash
# 1. Install the controller (via GitOps) + the kubeseal CLI
helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
helm install sealed-secrets sealed-secrets/sealed-secrets -n kube-system
brew install kubeseal        # or download the release binary

# 2. Fetch the controller's public cert (so you can seal offline / in CI)
kubeseal --fetch-cert > sealed-secret.crt

# 3. Seal a secret: build a normal Secret with --dry-run, pipe through kubeseal
kubectl create secret generic my-secret \
  --from-literal=password='myStrongPassword' \
  --dry-run=client -o json \
  | kubeseal --cert sealed-secret.crt -o yaml > mysealedsecret.yaml

# 4. Commit mysealedsecret.yaml to Git. The GitOps agent applies it; the controller
#    decrypts it into a real Secret named my-secret.
```

The committed file looks like (encrypted, scoped to namespace+name by default):
```yaml
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: my-secret
  namespace: my-app
spec:
  encryptedData:
    password: AgB...<ciphertext>...==
```

Notes: sealing is scoped to namespace+name by default (a `SealedSecret` can't be moved to another namespace
without re-sealing, unless you relax scope). Back up the controller's private key — losing it means re-sealing
every secret. Rotation = re-seal with a new key.

## External Secrets Operator (ESO)

Keep secrets in an external manager (AWS Secrets Manager, Azure Key Vault, GCP Secret Manager, HashiCorp Vault).
Git holds only an `ExternalSecret` describing *what to fetch*; the operator pulls it at runtime and materializes a
Kubernetes `Secret`. Nothing sensitive ever touches Git.

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets --create-namespace
```

```yaml
# A SecretStore/ClusterSecretStore (configured once) connects ESO to the backend.
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: my-external-secret
  namespace: my-app
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: my-secret-store
    kind: SecretStore          # or ClusterSecretStore
  target:
    name: my-kubernetes-secret # the Secret ESO will create/keep in sync
  data:
    - secretKey: db-password    # key in the resulting K8s Secret
      remoteRef:
        key: prod/db-password   # path/name in the external store
```

Tradeoffs: requires an external vault. The store connection itself needs auth — **prefer workload/managed
identities** (AKS workload identity, EKS IRSA) so no static ID/secret has to be shipped to the cluster; otherwise
you have a bootstrapping problem (securely getting the store credential in via CI). Big win: secrets are fetched
at runtime and can be rotated centrally without touching Git, and ESO can run independently of cluster state.

## SOPS (Mozilla SOPS)

Encrypts the *values* inside a YAML/JSON file (keys stay readable) using age, cloud KMS, or PGP. Commit the
encrypted file; the agent decrypts at apply time. **Flux decrypts SOPS natively**; Argo CD needs a config
management plugin.

```bash
age-keygen -o age.agekey                              # generate an age keypair
sops --encrypt --age <age-public-key> secret.yaml > secret.enc.yaml   # commit secret.enc.yaml
kubectl create secret generic sops-age -n flux-system --from-file=age.agekey   # give Flux the private key
```

Flux Kustomization decryption:
```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata: { name: app, namespace: flux-system }
spec:
  # ... source/path ...
  decryption:
    provider: sops
    secretRef:
      name: sops-age
```

## GitOps-at-scale pattern

Combine **ESO + a policy engine (Kyverno)** to distribute a single secret fleet-wide: the platform team creates
one secret (e.g. a registry pull secret) and Kyverno replicates it into every namespace, so every team can pull
from a central registry. This also lets you scan every image and inventory what's running.

## Pitfalls

- **Committing a base64 `Secret`.** That's plaintext; anyone can decode it. Use one of the tools above.
- **Losing the Sealed Secrets controller key** with no backup — irrecoverable; you must re-seal everything.
- **ESO store credential shipped insecurely** — use managed/workload identity instead of a static client secret.
- **Never-expiring tokens** (Argo CD `argocd-manager`, cluster bearer tokens) stored as plain Secrets — rotate
  them; deleting the Secret usually forces regeneration. Prefer identity over long-lived tokens.
- **Secrets in CI logs** — mask them; pass via the CI's secret store (GitHub Actions Secrets, Azure variable
  groups), never echo to stdout.
