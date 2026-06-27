# Kthena: LLM Inference Orchestration

Kthena is Volcano's specialized subproject for LLM inference orchestration and serving. It abstracts model deployment, hardware-aware scaling, and webhook-driven admission, acting as a complete control plane for model runtimes like vLLM, SGLang, and Triton.

---

## 1. Core Architecture

The Kthena control plane operates inside the `kthena-system` namespace, communicating directly with Volcano:

```
                  ┌──────────────────────────────────────────────┐
                  │                 KUBE API SERVER              │
                  └──────▲──────────────────▲───────────────▲────┘
                         │                  │               │
        Read/Write CRDs/ │                  │ Mutate/       │ Direct Pod
        ServingGroups    │                  │ Validate      │ Bindings
                         │                  │               │
┌────────────────────────┴─┐       ┌────────┴─────────┐     │ ┌────────────────────────┐
│ kthena-controller-mgr    │       │  kthena-webhook  │     │ │   Volcano Scheduler    │
│ (pkg/autoscaler/)        │       │  (pkg/webhook/)  │     │ │   (Gang, Topology,     │
│                          │       │                  │     │ │    Proportion)         │
│ • Reconciles             │       │ • Schema checks  │     │ └────────────────────────┘
│   ModelServing/ModelBooster│     │ • Automatic CA   │
│ • Manages serving groups │       │   certificates   │
└──────────────────────────┘       └──────────────────┘
```

- **kthena-controller-manager**: Reconciles Kthena CRDs, orchestrating the rollout and health-checks of LLM serving pods. It maps complex distributed serving layouts (like prefill-decode groupings) into Volcano-native Pods.
- **kthena-router**: The high-throughput data-plane proxy that intercepts LLM client requests and dispatches them intelligently based on token length, model routing configurations, and cache residency.

---

## 2. High-Level Serving: `ModelBooster`

For standard, single-model deployments, `ModelBooster` provides an all-in-one abstraction:

```yaml
apiVersion: workload.serving.volcano.sh/v1alpha1
kind: ModelBooster
metadata:
  name: qwen-instruct
  namespace: default
spec:
  name: "Qwen-2.5-7B"
  owner: "ml-ops-team"
  backend:
    name: "vllm-backend"
    type: vLLM                # vLLM, SGLang, Triton, TensorRT
    modelURI: "hf://Qwen/Qwen2.5-7B-Instruct" # HuggingFace source
    cacheURI: "pvc://model-cache/qwen"        # PV mount path for weights caching
    env:
      - name: "HF_ENDPOINT"
        value: "https://huggingface.co"
  autoscalingPolicy:
    tolerancePercent: 10
    metrics:
      - name: "gpu_memory_utilization"
        targetValue: "80"
```

---

## 3. Distributed Serving: `ModelServing`

For highly customized distributed layouts, `ModelServing` defines specific roles (like separate prefill and decode tasks) and rollout behaviors:

```yaml
apiVersion: workload.serving.volcano.sh/v1alpha1
kind: ModelServing
metadata:
  name: deepseek-pd-cluster
  namespace: default
spec:
  schedulerName: volcano
  replicas: 2                 # Runs 2 independent ServingGroups (e.g. 2P4D each)
  template:
    restartGracePeriodSeconds: 60
    roles:
      - name: prefill         # Prefill computing role
        replicas: 2
        entryTemplate:
          metadata:
            labels:
              app: deepseek-pd
              modelserving.volcano.sh/rolename: "P-instance"
          spec:
            containers:
              - name: prefill-engine
                image: vllm/vllm:latest
                args: ["--role", "prefill", "--model", "deepseek-ai/DeepSeek-R1-Distill-Qwen-7B"]
                ports:
                  - containerPort: 8000
                resources:
                  requests:
                    nvidia.com/gpu: "1"
                    cpu: "4"
                    memory: "16Gi"
      - name: decode          # Decode token generation role
        replicas: 4
        entryTemplate:
          metadata:
            labels:
              app: deepseek-pd
              modelserving.volcano.sh/rolename: "D-instance"
          spec:
            containers:
              - name: decode-engine
                image: vllm/vllm:latest
                args: ["--role", "decode", "--model", "deepseek-ai/DeepSeek-R1-Distill-Qwen-7B"]
                ports:
                  - containerPort: 8000
                resources:
                  requests:
                    nvidia.com/gpu: "1"
                    cpu: "4"
                    memory: "16Gi"
  rolloutStrategy:
    type: RoleRollingUpdate   # Update role-by-role or group-by-group
    rollingUpdateConfiguration:
      maxUnavailable: 1
      partition: 0
```

---

## 4. Autoscaling & Multi-Hardware Heterogeneous Pools

Standard HPA fails on LLMs because GPU cores stay at 100% load during generation regardless of whether the system is healthy or saturated. Kthena introduces model-aware metric-scaling:

```yaml
apiVersion: workload.serving.volcano.sh/v1alpha1
kind: AutoscalingPolicy
metadata:
  name: llm-autoscaling-policy
  namespace: default
spec:
  tolerancePercent: 15
  metrics:
    - name: gpu_utilization
      targetValue: "75"
    - name: queue_depth       # Scales up if requests are queuing at the Router
      targetValue: "15"
  behavior:
    scaleUp:
      stablePolicy:
        instances: 2
        period: "20s"
      panicPolicy:            # Rapid override during traffic spikes
        instances: 5
        period: "5s"
---
apiVersion: workload.serving.volcano.sh/v1alpha1
kind: AutoscalingPolicyBinding
metadata:
  name: llm-scaling-binding
  namespace: default
spec:
  policyRef:
    name: llm-autoscaling-policy
  heterogeneousTarget:        # Cost-aware heterogeneous scaling across pools
    minReplicas: 1
    maxReplicas: 20
    targets:
      - name: "h100-primary-pool"
        targetRef:
          kind: ModelServing
          name: deepseek-h100
        cost: 33.00           # $/hour weighting
      - name: "a100-secondary-pool"
        targetRef:
          kind: ModelServing
          name: deepseek-a100
        cost: 10.00           # Scales up first on cheap A100s, overflows to H100s
```
- `heterogeneousTarget`: Distributes scale-up across different hardware pools (e.g. A100 vs H100) based on cost and capacity efficiency.
