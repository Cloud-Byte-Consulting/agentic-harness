# kubectl & kubeconfig Mastery

kubectl is just an HTTP client for kube-apiserver that handles auth, resource paths, and output
formatting for you. Mastering it = mastering day-to-day cluster operation.

## Contents
- [kubeconfig structure](#kubeconfig-structure)
- [Contexts and namespaces](#contexts-and-namespaces)
- [Getting credentials from managed clusters](#getting-credentials-from-managed-clusters)
- [Imperative vs declarative](#imperative-vs-declarative)
- [Common verbs](#common-verbs)
- [Output formats, jsonpath, custom-columns](#output-formats-jsonpath-custom-columns)
- [kubectl explain](#kubectl-explain)
- [Plugins and krew](#plugins-and-krew)
- [Completion](#completion)

## kubeconfig structure

kubectl loads config in this order: `--kubeconfig` flag → `$KUBECONFIG` env var → default
`~/.kube/config`. The file has three lists that join together:

- **clusters** — each entry: API server `server:` URL + `certificate-authority`(-data) to trust it.
- **users** — credentials: client cert/key, token, or an `exec` auth plugin (cloud CLIs).
- **contexts** — a named `(cluster, user, namespace)` triple. The active one is `current-context`.

```yaml
apiVersion: v1
kind: Config
current-context: prod
clusters:
- name: prod
  cluster:
    server: https://k8s-api.example.com:6443
    certificate-authority-data: <base64 CA>
users:
- name: prod-admin
  user:
    client-certificate-data: <base64>
    client-key-data: <base64>
contexts:
- name: prod
  context:
    cluster: prod
    user: prod-admin
    namespace: default
```
View the merged config (secrets redacted):
```bash
kubectl config view
KUBECONFIG=~/.kube/config:~/.kube/dev.yaml kubectl config view --flatten > merged.yaml  # merge files
```

## Contexts and namespaces

```bash
kubectl config get-contexts                      # list; * marks current
kubectl config current-context                   # show active
kubectl config use-context dev                    # switch cluster/user/namespace
kubectl config set-context --current --namespace=team-a   # change default namespace
kubectl config get-contexts -o name              # just the names
```
Per-command overrides: `kubectl get pods -n kube-system`, `--context dev`, `--all-namespaces`/`-A`.
For heavy multi-cluster work, the `kubectx`/`kubens` plugins (via krew) speed this up.

## Getting credentials from managed clusters

These write/update a context for you:
```bash
aws eks update-kubeconfig --name CLUSTER --region REGION
az aks get-credentials --resource-group RG --name CLUSTER
gcloud container clusters get-credentials CLUSTER --zone ZONE   # or --region
```

## Imperative vs declarative

- **Imperative** — fast, direct, and the only way to *list* resources or do ad-hoc actions:
  ```bash
  kubectl run my-pod --image=busybox:latest --restart=Never
  kubectl get rs -n team-a
  kubectl delete pod my-pod
  ```
- **Declarative** — author YAML, version it in Git (IaC), apply repeatably:
  ```bash
  kubectl apply -f manifest.yaml          # create or update to match the file
  kubectl diff  -f manifest.yaml          # preview changes before applying
  kubectl delete -f manifest.yaml
  ```
  One file can hold multiple resources separated by `---`. Every manifest needs `apiVersion`,
  `kind`, `metadata`, and `spec`. Prefer declarative for anything you want reproducible.
  (Workload object authoring is the domain of kubernetes-workloads; here we care about cluster-level
  use of these verbs.)

## Common verbs

```bash
kubectl get nodes -o wide                      # nodes + IPs, OS, runtime, version
kubectl get pods -A                            # all pods, all namespaces
kubectl describe node <node>                   # conditions, capacity, allocated resources, events
kubectl logs <pod> [-c container] [-f] [--previous]
kubectl exec -it <pod> -- sh
kubectl top nodes / kubectl top pods           # needs metrics-server
kubectl cluster-info                           # control-plane & CoreDNS endpoints
kubectl get --raw='/readyz?verbose'            # API server readiness checks (prefer over componentstatuses)
kubectl api-resources                          # every resource kind + short name + apiVersion
kubectl get events --sort-by=.lastTimestamp -A # recent cluster events
```

## Output formats, jsonpath, custom-columns

```bash
kubectl get pods -o yaml                        # full object
kubectl get pods -o json | jq '.items[].metadata.name'
kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'
kubectl get pods -o custom-columns='NAME:.metadata.name,NODE:.spec.nodeName,STATUS:.status.phase'
kubectl get pods -o wide                        # extra columns (node, IP)
kubectl get pods --field-selector status.phase=Running
kubectl get pods -l app=nginx                   # label selector
```
jsonpath is ideal for scripting (e.g. extract a node's kubelet version or a Service's external IP).

## kubectl explain

Self-documenting schema — no need to memorize fields or hunt the docs:
```bash
kubectl explain pod.spec.containers
kubectl explain deployment.spec.strategy --recursive
kubectl explain node.status
```

## Plugins and krew

**krew** is the kubectl plugin manager. Any executable named `kubectl-<name>` on your PATH becomes
`kubectl <name>`.
```bash
# install krew (see krew.sigs.k8s.io for the current snippet), then:
kubectl krew update
kubectl krew install ctx ns ns tree neat
kubectl ctx          # = kubectx (switch contexts)
kubectl ns team-a    # = kubens (switch namespace)
kubectl tree deploy/web   # ownership tree of a resource
```
Useful plugins: `ctx`/`ns`, `tree`, `neat` (strip server-added clutter from `get -o yaml`),
`view-secret`, `stern` (multi-pod logs), `kubectl-node-shell`.

## Completion

```bash
# Bash
echo 'source <(kubectl completion bash)' >> ~/.bashrc
# Zsh
echo 'source <(kubectl completion zsh)' >> ~/.zshrc
# alias k=kubectl and keep completion:
echo 'alias k=kubectl; complete -o default -F __start_kubectl k' >> ~/.bashrc
```
Also supported for Fish and PowerShell.

## Version-skew note

Keep kubectl within **±1 minor** of the API server (a v1.30 kubectl manages v1.29–v1.31). See
upgrades-and-version-skew.md.
