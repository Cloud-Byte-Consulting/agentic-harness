# GenAI & GPU Workloads on Kubernetes

Running generative-AI and ML workloads on Kubernetes: GPU scheduling (the NVIDIA device plugin, `nvidia.com/gpu` requests, MIG/time-slicing), model serving (vLLM, KServe, Triton), deploying RAG/agent apps, and the self-host-vs-managed decision.

## Contents
- Self-host on K8s vs a managed inference API
- GPU scheduling fundamentals (device plugin, requests/limits)
- Sharing GPUs: time-slicing, MPS, MIG
- GPU node pools, taints, and the GPU Operator
- Serving an LLM (vLLM, KServe, Triton)
- Deploying a GenAI app (chat / RAG / agent) on K8s
- RAG and vector stores
- Pitfalls

## Self-host on K8s vs a managed inference API

GenAI on K8s splits into two very different jobs:
1. **Deploying the *application*** (a chat UI, RAG service, or agent) that *calls* a model — this is ordinary stateless K8s (Deployment + Service + config/secrets) and needs no GPU if the model lives behind an API like Amazon Bedrock, OpenAI, Anthropic, or a separate inference cluster.
2. **Self-hosting the *model***, which needs GPUs, careful scheduling, and real inference-serving infrastructure.

**Call a managed inference API** when you can: GPUs are scarce, expensive, and operationally heavy; an API removes capacity planning, driver/CUDA management, and autoscaling-of-GPUs problems. The book's approach — a Streamlit app on K8s calling **Amazon Bedrock** foundational models (Claude, Titan embeddings) — is exactly this: no GPU in the cluster, just a Deployment with cloud credentials.

**Self-host the model on K8s** only when you need: data residency / on-prem / air-gapped inference, a private or fine-tuned model the APIs don't offer, predictable cost at sustained high volume, or ultra-low latency without egress. Then you take on GPU scheduling — covered below.

## GPU scheduling fundamentals

GPUs are **extended resources**, not core CPU/memory. The chain to make them schedulable:

1. **GPU nodes** with the NVIDIA driver installed (managed node groups can do this, or use the GPU Operator below).
2. The **NVIDIA device plugin** DaemonSet running on GPU nodes — it discovers GPUs and advertises `nvidia.com/gpu` as an allocatable resource. **Without it, `nvidia.com/gpu` doesn't exist and GPU pods stay `Pending` forever.**
3. Pods request the resource under `limits` (extended resources require `limits`; you cannot overcommit a whole GPU):

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: llm-server
spec:
  replicas: 1
  selector: { matchLabels: { app: llm-server } }
  template:
    metadata: { labels: { app: llm-server } }
    spec:
      runtimeClassName: nvidia          # if the node uses the nvidia container runtime via RuntimeClass
      containers:
        - name: server
          image: vllm/vllm-openai:latest
          args: ["--model", "meta-llama/Llama-3.1-8B-Instruct"]
          resources:
            limits:
              nvidia.com/gpu: 1          # whole GPUs only (integer)
              memory: 24Gi
              cpu: "4"
          ports:
            - containerPort: 8000
```

Key facts:
- `nvidia.com/gpu` must be an **integer** in `limits` — a pod gets *whole* GPUs by default (no fractional sharing without the techniques below).
- A GPU allocated to a pod is **not shared** with other pods by default — exclusive access.
- The scheduler only places the pod on a node advertising enough `nvidia.com/gpu`. Combine with `nodeSelector`/affinity on a GPU label and tolerations for the GPU taint.

## Sharing GPUs: time-slicing, MPS, MIG

Whole-GPU-per-pod wastes expensive hardware for small models / dev. Three sharing strategies (configured via the device plugin / GPU Operator):
- **Time-slicing** — the device plugin advertises N "replicas" of each GPU; pods interleave on the GPU in time. Simple, no isolation/memory guarantees — good for dev, bursty, or many small workloads. Configured in the device-plugin ConfigMap (`sharing.timeSlicing.replicas`).
- **MPS (Multi-Process Service)** — concurrent kernels from multiple processes on one GPU with better utilization than time-slicing; limited isolation.
- **MIG (Multi-Instance GPU)** — hardware partitioning on A100/H100-class GPUs into isolated slices (e.g. `nvidia.com/mig-1g.5gb`), each with dedicated memory/compute. The only option giving true isolation; best for multi-tenant serving. Pods request the MIG profile resource name.

Pick time-slicing for dev/throughput, MIG for isolated multi-tenant production.

## GPU node pools, taints, and the GPU Operator

**Taint GPU nodes** so only GPU workloads land on them (GPUs are wasted by CPU-only pods):

```bash
kubectl taint nodes <gpu-node> nvidia.com/gpu=present:NoSchedule
```

Tolerate it on GPU pods plus select the pool:

```yaml
spec:
  tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
  nodeSelector:
    cloud.google.com/gke-accelerator: nvidia-tesla-a100   # or your cloud's GPU label
```

The **NVIDIA GPU Operator** is the recommended way to manage the whole GPU stack on K8s — it deploys and lifecycle-manages the driver, container toolkit, device plugin, DCGM metrics exporter, MIG manager, and node feature discovery, so you don't hand-install drivers per node:

```bash
helm repo add nvidia https://helm.ngc.nvidia.com/nvidia
helm repo update
helm install gpu-operator nvidia/gpu-operator -n gpu-operator --create-namespace
```

It auto-labels GPU nodes and exposes DCGM GPU metrics for Prometheus (utilization, memory, temperature) — wire into your monitoring (see **kubernetes-observability**).

## Serving an LLM

Don't hand-roll a Flask wrapper around a model. Use a purpose-built inference server:

- **vLLM** — high-throughput LLM serving with paged-attention and continuous batching; exposes an OpenAI-compatible API. Run as a Deployment requesting `nvidia.com/gpu` (example above), front with a Service. The simplest path to self-host an open-weights LLM.
- **KServe** — Kubernetes-native model-serving CRD (`InferenceService`) with autoscaling (including **scale-to-zero** via Knative), canary rollouts, and multi-framework support. Best when you want managed serving semantics:

```yaml
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: llama-3-8b
spec:
  predictor:
    minReplicas: 0          # scale to zero when idle — saves GPU cost
    maxReplicas: 3
    model:
      modelFormat:
        name: huggingface
      args: ["--model_id=meta-llama/Llama-3.1-8B-Instruct"]
      resources:
        limits:
          nvidia.com/gpu: "1"
          memory: 24Gi
```

- **NVIDIA Triton Inference Server** — multi-framework (TensorRT, ONNX, PyTorch, TF), dynamic batching, model ensembles; strong for non-LLM and mixed model fleets.

Serving design notes: large model weights need fast load — bake into the image or pull from object storage / a PVC with a readiness probe that gates traffic until loaded; size GPU memory to the model (a 70B model needs multiple GPUs / tensor parallelism); set `readinessProbe` so the Service doesn't route to a still-loading replica; autoscale on GPU utilization/queue depth (KServe/Knative or custom metrics HPA — see **kubernetes-autoscaling-scheduling**).

## Deploying a GenAI app on K8s

The application layer (chat UI, RAG service, agent backend) is a normal stateless workload — **no GPU needed if it calls a model API**. Standard Deployment + Service, with model/endpoint config in env, secrets for credentials, non-sensitive IDs in a ConfigMap:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chat-app
  namespace: genai
spec:
  replicas: 2
  selector: { matchLabels: { app: chat-app } }
  template:
    metadata: { labels: { app: chat-app } }
    spec:
      containers:
        - name: chat-app
          image: myrepo/chat-app:v1          # e.g. a Streamlit/FastAPI front end
          ports: [{ containerPort: 8501 }]
          env:
            - name: MODEL_ENDPOINT            # point at vLLM/KServe Service or an external API
              value: "http://llm-server.genai.svc:8000/v1"
            - name: AWS_ACCESS_KEY_ID         # only if calling a cloud API like Bedrock
              valueFrom: { secretKeyRef: { name: ai-credentials, key: access_key_id } }
            - name: KB_ID                     # non-secret config (e.g. RAG knowledge-base id)
              valueFrom: { configMapKeyRef: { name: app-config, key: kb_id } }
          resources:
            requests: { cpu: "250m", memory: 512Mi }
            limits:   { cpu: "1",    memory: 1Gi }
---
apiVersion: v1
kind: Service
metadata:
  name: chat-app
  namespace: genai
spec:
  type: ClusterIP                # expose via authenticated ingress, not a naked public LoadBalancer
  selector: { app: chat-app }
  ports: [{ port: 8501, targetPort: 8501 }]
```

Patterns worth keeping from the book's Bedrock examples (provider-agnostic):
- **Secrets vs ConfigMap split** — credentials in a `Secret` (env via `secretKeyRef`); non-sensitive runtime config (knowledge-base IDs, agent IDs, model names) in a `ConfigMap`. Changing a `ConfigMap` value avoids rebuilding the image.
- **Never bake credentials into the image** (especially public images). Inject at runtime.
- **Front the app with an authenticated ingress** — public LoadBalancers with no auth are demo-only.

## RAG and vector stores

**Retrieval-Augmented Generation** grounds a model in your data: retrieve relevant chunks from a knowledge base, inject them as context, then generate — reducing hallucination and adding fresh/private knowledge without retraining. On K8s the moving parts:
- An **embeddings model** (self-hosted on GPU, or an embeddings API) turns documents and queries into vectors.
- A **vector store** for similarity search — **OpenSearch/Elasticsearch** (kNN dense vectors; you may already run ECK — see `data-lake-and-query.md`), or dedicated stores like **Qdrant**, **Weaviate**, **Milvus**, **pgvector** (Postgres). Run these as StatefulSets/operators with persistent storage.
- A **retrieval + prompt-assembly service** (LangChain/LlamaIndex or your own) that queries the vector store and calls the LLM.

Keep the vector store close to the retrieval service (latency), back it with fast persistent storage, and treat embedding (re)generation as a batch/streaming pipeline (Spark/Airflow) when the corpus is large.

**Agents** extend this: the model calls tools/functions (defined by a schema) to take actions. The agent backend is again a normal K8s service; the action implementations are your own services/functions. Same deployment shape, plus RBAC/network controls around whatever the tools can touch.

## Pitfalls

- **No device plugin → GPU pods `Pending` forever.** Install the device plugin / GPU Operator first; confirm `kubectl describe node <gpu-node>` shows `nvidia.com/gpu` under Allocatable.
- **Requesting `nvidia.com/gpu` in `requests` only / as a fraction.** Extended resources go in `limits` as integers; fractional sharing needs time-slicing/MIG.
- **CPU-only pods landing on GPU nodes.** Taint GPU nodes and tolerate only on GPU workloads, or you waste GPUs.
- **Driver/CUDA/runtime mismatch.** The container's CUDA version must be compatible with the node driver; let the GPU Operator manage versions rather than hand-installing.
- **Self-hosting when an API would do.** The biggest cost/ops mistake — only self-host for a concrete residency/cost/customization reason.
- **Unbounded GPU memory.** A model larger than GPU memory crashes at load; size the GPU (or use tensor parallelism across GPUs) to the model.
- **Treating a GenAI app as special.** The *app* is ordinary K8s — apply normal Deployment/Service/Secret/ingress hygiene; reserve the special handling for the GPU *serving* tier.
