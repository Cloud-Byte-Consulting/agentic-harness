# Kthena: Advanced Request Routing & PD Disaggregation

Because LLM requests have extremely variable execution times, standard HTTP round-robin load balancing is insufficient. Kthena's data plane proxy (`kthena-router`) implements token-aware rate limiting, fair queuing, Canary rollouts, and **Prefill-Decode Disaggregation (PD)** routing.

---

## 1. Prefill-Decode Disaggregation (PD) Topologies

In standard deployments, a single GPU handles both prefill (compute-intensive prompt processing) and decode (memory-bound token generation). This causes "prefill interference," where a long prompt temporarily spikes generation latency (TPOT) for all other running streams on the GPU.

Kthena disaggregates them, routing prompts to Prefill nodes and subsequent generation tokens to Decode nodes.

```yaml
apiVersion: networking.serving.volcano.sh/v1alpha1
kind: ModelServer
metadata:
  name: deepseek-pd-server
  namespace: default
spec:
  model: "deepseek-ai/DeepSeek-R1"
  inferenceEngine: vLLM
  workloadSelector:
    matchLabels:
      app: deepseek-pd
    pdGroup:                  # Identifies P and D pod boundaries
      groupKey: "modelserving.volcano.sh/group-name"
      prefillLabels:
        modelserving.volcano.sh/rolename: "P-instance"
      decodeLabels:
        modelserving.volcano.sh/rolename: "D-instance"
  workloadPort:
    port: 8000
    protocol: http
  kvConnector:
    type: lmcache             # lmcache, mooncake, nixl, nixl-native
```
- `kvConnector`: Syncs intermediate Key-Value cache arrays from prefill pods to decode pods. Without this, decode pods would have to recalculate the prompt, making the split useless.

---

## 2. Advanced Routing: LoRA and Canary Rollouts

`ModelRoute` provides header, URI, and JSON-body routing matching, which is ideal for dynamic LoRA adapters and multi-version canary releases:

```yaml
apiVersion: networking.serving.volcano.sh/v1alpha1
kind: ModelRoute
metadata:
  name: deepseek-canary-route
  namespace: default
spec:
  modelName: "deepseek-ai/DeepSeek-R1"
  parentRefs:
    - name: kthena-gateway    # Connects to Kthena Router
      namespace: kthena-system
  rules:
    # --- Rule 1: LoRA Adapter Route ---
    - name: finance-lora
      modelMatch:
        headers:
          x-domain:
            exact: "finance"
      targetModels:
        - modelServerName: "deepseek-finance-lora-server"
          weight: 100
    
    # --- Rule 2: Canary Split ---
    - name: canary-traffic
      modelMatch:
        headers:
          user-tier:
            exact: "standard"
      targetModels:
        - modelServerName: "deepseek-v1-stable"
          weight: 90
        - modelServerName: "deepseek-v2-canary"
          weight: 10
```

---

## 3. Token-Level Rate Limiting

Standard rate limiters count raw HTTP requests. Since a single request can consume 10 tokens or 10,000 tokens, standard limiters fail to prevent GPU resource starvation. Kthena tracks real token consumption:

```yaml
apiVersion: networking.serving.volcano.sh/v1alpha1
kind: ModelRoute
metadata:
  name: token-rate-limited-route
  namespace: default
spec:
  modelName: "qwen-7b"
  rules:
    - name: default
      targetModels:
        - modelServerName: "qwen-7b-server"
  rateLimit:
    inputTokensPerUnit: 100000  # Max input prompt tokens
    outputTokensPerUnit: 50000  # Max generated output tokens
    unit: minute
    global:
      redis:
        address: "redis.kthena-system.svc:6379"
```
The router tokenizes incoming prompts on-the-fly (via `tiktoken`) and queries a cluster-wide Redis store to enforce precise token limits per tenant.

---

## 4. Fairness Scheduling

When enabled, Kthena Router runs a sliding-window fair-queuing algorithm. If multiple users query the same model, the proxy prioritizes requests from users who have consumed the fewest total tokens inside the current window, preventing single power-users from monopolizing the GPUs.
- **Weights**: Configure `inputTokenWeight: 1.0` and `outputTokenWeight: 2.0` (since output token generation is significantly more expensive).
