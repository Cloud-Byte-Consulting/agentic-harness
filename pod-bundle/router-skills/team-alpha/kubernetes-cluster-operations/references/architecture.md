# Kubernetes Cluster Architecture

Deep reference on what a cluster is made of, how a request flows, and how to make it
highly available.

## Contents
- [Node roles](#node-roles)
- [Control-plane components](#control-plane-components)
- [Node components](#node-components)
- [The watch/reconcile model and request flow](#the-watchreconcile-model-and-request-flow)
- [Add-ons](#add-ons)
- [HA topologies](#ha-topologies)
- [Stacked vs external etcd](#stacked-vs-external-etcd)
- [Managed control planes](#managed-control-planes)

## Node roles

A Kubernetes cluster is a set of Linux (or, for workers, optionally Windows) machines called
**nodes**, split into two roles:

- **Control-plane nodes** (formerly "master") run the components that hold and reconcile cluster
  state. You interact with these.
- **Worker / compute nodes** run your application Pods. You don't talk to them directly; the control
  plane delegates work to them.

Kubernetes is not one binary — it's a set of small Go programs that must communicate. For dev you
can colocate everything on one host; for production you spread components across machines for
availability and scale. A Windows worker can only run Windows containers and vice versa; the
control plane runs on Linux only.

## Control-plane components

### kube-apiserver
The hub. A stateless REST API exposing all Kubernetes resources (Pods, Deployments, Services,
etc.) over HTTPS (default **6443**). It is the **single entry point** for kubectl, controllers,
kubelets, and the dashboard, and the **only** component permitted to read/write etcd. Because it's
stateless, you scale it horizontally and front it with a load balancer. It also enforces
authentication, authorization (RBAC), and admission control on every request.

### etcd
A distributed, consistent key-value store (a separate CNCF project) holding the entire cluster
state on disk. Ports: **2379** client, **2380** peer. It uses the **Raft** consensus algorithm and
elects a leader; writes go to the leader and replicate to followers. HA requires an **odd** number
of members (3 tolerates 1 failure, 5 tolerates 2) to maintain quorum. etcd is more critical than
the API server: a crashed API server restarts, but lost/corrupted etcd without a backup means the
cluster cannot be recovered. Some distros swap it (k3s defaults to SQLite, can use MySQL/Postgres).

### kube-scheduler
Watches for Pods with no node assigned (empty `.spec.nodeName`) and selects a node by **filtering**
(does the node satisfy resource requests, node selectors, taints/tolerations, affinity?) then
**scoring** the survivors. It writes the binding back through the API server. It can be replaced
with a custom scheduler. (Advanced scheduling — affinity, taints, custom schedulers — is detailed
in kubernetes-autoscaling-scheduling.)

### kube-controller-manager
A single binary running many **controllers**, each a reconciliation loop driving actual → desired:
Node, ReplicaSet/Replication, Deployment, StatefulSet, DaemonSet, Job, Endpoints, ServiceAccount,
Namespace, and the Pod garbage collector, among others. If actual state drifts (a Pod dies), the
relevant controller asks the API server to fix it.

### cloud-controller-manager
Holds cloud-provider-specific loops so that core Kubernetes stays vendor-neutral: Node controller
(detect deleted cloud VMs), Route controller (program cloud routes), Service controller (provision
cloud LoadBalancers). Absent on bare-metal, local, and on-prem clusters.

## Node components

Run on every node, control-plane nodes included.

### kubelet
The node agent. It runs as a **host systemd service**, never inside a container. On start it reads
its config (e.g. `/var/lib/kubelet/config.yaml`, bootstrap creds in `/etc/kubernetes/kubelet.conf`)
which tells it the API server endpoint and the container-runtime CRI socket. It watches the API for
Pods bound to its node and translates each Pod spec into containers via the CRI. It does not read
etcd and isn't aware etcd exists. Requires HTTPS connectivity to the API server on 6443.

### kube-proxy
Runs on each node and programs the local **iptables** or **IPVS** rules that implement the Service
abstraction (virtual IPs, load balancing to Pod endpoints). Some CNIs (e.g. Cilium in
kube-proxy-replacement mode) eliminate it.

### Container runtime (via CRI)
Executes containers. Kubernetes talks to it through the **Container Runtime Interface (CRI)**.
Common runtimes and their Unix sockets:

| Runtime | CRI socket |
|---|---|
| containerd | `unix:///var/run/containerd/containerd.sock` |
| CRI-O | `unix:///var/run/crio/crio.sock` |
| Docker Engine (via cri-dockerd) | `unix:///var/run/cri-dockerd.sock` |

Docker's built-in `dockershim` was **removed in Kubernetes v1.24**. Prefer containerd or CRI-O.
Images built by Docker still run everywhere because they follow the **OCI** image spec.
`RuntimeClass` lets you assign different runtime configs to Pods (e.g. a sandboxed runtime for
untrusted workloads).

## The watch/reconcile model and request flow

Nothing is "pushed" to nodes. Components **watch** the API server and **reconcile** toward the
desired state. Example — `kubectl run nginx --image nginx`:

1. kubectl builds an authenticated HTTPS POST to kube-apiserver (target from kubeconfig).
2. API server runs authn → authz (RBAC) → admission, then persists the Pod object to **etcd**.
3. kube-scheduler (watching) sees an unscheduled Pod, picks a node, writes the binding via the API.
4. The chosen node's **kubelet** (watching) sees a Pod bound to it, pulls the image, and starts the
   container via its CRI socket.
5. Controllers keep reconciling (e.g. if the Pod is owned by a ReplicaSet, the controller recreates
   it on failure).

Connectivity summary:

| Component | Talks to | Purpose |
|---|---|---|
| kube-apiserver | etcd, all others, kubectl | REST API; only writer of etcd |
| etcd | kube-apiserver only | cluster state |
| kube-scheduler | kube-apiserver | bind unscheduled Pods |
| kube-controller-manager | kube-apiserver | reconciliation loops |
| kubelet | kube-apiserver + CRI socket | run/stop Pods on its node |
| kube-proxy | kube-apiserver | program Service networking |

## Add-ons

Cluster features implemented as Kubernetes resources (DaemonSets/Deployments), usually in the
`kube-system` namespace: **CoreDNS** (cluster DNS / service discovery), the **CNI network plugin**
(Calico, Cilium, Flannel, kindnet, …), **metrics-server** (resource metrics for `kubectl top` and
autoscaling), the optional **dashboard**, and logging agents. CNI/NetworkPolicy specifics belong to
kubernetes-networking.

## HA topologies

1. **Single-node** — control plane + worker on one machine (kind/minikube default). Dev/edge only;
   no availability.
2. **Single control-plane, N workers** — workloads are HA, but the lone control-plane node and its
   single etcd are a SPOF. If it dies, running Pods become orphans you can only manage by SSHing to
   workers and using `crictl`. Better than single-node but not production-grade. Also, every worker's
   kubelet polls the API; many workers against one API server can overload it.
3. **Multi control-plane (3+) + load balancer** — the production target. Replicate both control
   plane and workers. Put an L4/L7 load balancer with a stable FQDN in front of the API servers;
   kubeconfig points at the LB, not a single host. Each control-plane node typically runs an etcd
   member forming a Raft quorum (use 3 or 5 members).

## Stacked vs external etcd

- **Stacked etcd** (default with kubeadm): each control-plane node also runs a local etcd member.
  Simplest to deploy; failure of a node loses both an API server and an etcd member together.
- **External etcd**: a dedicated etcd cluster on separate hosts. More machines and operational work,
  but isolates the etcd failure domain from the control-plane components — preferred for demanding
  reliability requirements.

In any multi-control-plane setup the etcd members form one internal Raft cluster, electing a leader
that accepts writes and replicates them; a failed leader triggers re-election. Quorum (a majority)
must survive, which is why member counts are odd.

## Managed control planes

With EKS/AKS/GKE/LKE/DigitalOcean/OpenShift the provider installs, scales, patches, and backs up the
control plane and etcd, exposing only an API endpoint. You manage worker node pools, workloads, and
observability. "Serverless" tiers go further and abstract the nodes too: **GKE Autopilot**,
**EKS Fargate**, **AKS Virtual Kubelet → ACI**. See `managed-clusters.md`.
