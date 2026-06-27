---
name: volcano-genai-serving
description: >-
  Orchestrate and route generative AI serving and interactive agent workloads on Kubernetes using Kthena and AgentCube.
  Use for Kthena LLM orchestration (ModelServing, ModelServer, ModelRoute, AutoscalingPolicy),
  prefill-decode disaggregation, KV cache connectors, LoRA adapter routing, and AgentCube serverless agent sandboxes.
  Trigger when users discuss LLM serving on Kubernetes, vLLM / SGLang orchestration, prefix/KV cache-aware routing,
  LoRA adapter routing, secure code interpreter sandboxes, AgentCube python SDK, or serverless AI agents.
---

# Volcano GenAI & serving Orchestration (Kthena & AgentCube)

This skill equips Claude with cutting-edge expertise in next-generation Generative AI orchestration, serving, and interactive agent execution using Volcano's specialized subprojects: **Kthena** and **AgentCube**. It covers high-throughput low-latency LLM routing, prefill-decode disaggregation, LoRA adapter-aware routing, and secure, high-density serverless code interpreter sandboxes.

## When to use this skill

- Deploying and scaling LLM serving runtimes (vLLM, SGLang, Triton, TensorRT-LLM) using Kthena's `ModelServing` and `ModelBooster` CRDs.
- Configuring advanced LLM request routing (canary rollouts, premium-tier traffic splitting, and LoRA adapter-specific routing) via `ModelRoute` and `ModelServer`.
- Optimizing LLM throughput and latency (TTFT, TPOT) through **Prefill-Decode Disaggregation (PD)** and KV cache state-transfer connectors (LMCache, MoonCake, NIXL).
- Enforcing token-level rate limiting, request-level priority, or fair-queuing scheduling at the ingress proxy layer via Kthena Router.
- Provisioning secure, multi-tenant execution sandboxes for AI agents or interactive python sessions using AgentCube's `CodeInterpreter` and `AgentRuntime` CRDs.
- Integrating AgentCube with agent-frameworks like LangChain or Model Context Protocol (MCP) servers.

## Core concepts

- **Kthena**: An ecosystem subproject designed specifically for LLM inference orchestration and traffic control. Unlike traditional general-purpose ingress proxies, Kthena understands the internal semantics of LLM requests (tokens, prompt lengths, KV cache hits, adapter paths) and coordinates with Volcano to schedule and route requests with high accuracy.
- **Prefill-Decode Disaggregation**: LLM inference has two distinct steps: the compute-bound *prefill* phase (processing the input prompt) and the memory-bound *decode* phase (generating tokens one-by-one). Running them on the same GPU leads to performance interference. Kthena splits them into dedicated `prefill` and `decode` pods, transferring the intermediate KV cache state via highly optimized connectors.
- **KV Cache Connector**: Specialized protocols (like `lmcache` or `mooncake`) that sync the key-value cache between prefill and decode instances to reduce TTFT (Time To First Token) and prevent redundant prompt recalculations.
- **AgentCube**: A serverless execution platform for interactive AI Agents and code interpreter tools. It provides ultra-fast sandbox creation (sub-second cold starts via pre-warmed pools), dynamic idle-timeout hibernation (garbage collection), and a secure imperative API (PicoD/AgentD sidecars) for running untrusted code on behalf of LLMs.

## Workflow / how to approach tasks

1. **Deploy LLM Models**: Use `ModelBooster` for simple, one-command deployments that abstract serving and routing setup, or `ModelServing` for complex, multi-role (prefill-decode) clustered runtimes.
2. **Setup Gateway Routing**: Deploy the `ModelServer` to expose your inference pods, and write a `ModelRoute` with `modelMatch` rules to handle canary routing, headers filtering, or LoRA adapter targets.
3. **Bind Autoscaling Policies**: Rather than using standard CPU-based HPA, write an `AutoscalingPolicy` targeting `gpu_utilization` or `queue_depth`. Bind it using `AutoscalingPolicyBinding` to heterogeneous hardware pools (e.g., scale up on cheap A100s first, overflow to H100s).
4. **Implement Sandbox Execution**: For AI agents requiring code execution, deploy AgentCube. Define `CodeInterpreter` specs with `warmPoolSize: 2` to maintain pre-warmed containers, and use the AgentCube Python SDK to securely execute Python scripts, read/write workspace files, and enforce timeouts.

## Common pitfalls & anti-patterns

- **Using Standard Ingress / HPA for LLMs**: Traditional HTTP ingress controllers load-balance based on raw connections, which leads to GPU starvation or VRAM overflow because LLM requests have vastly different token counts and computation lengths. Standard HPA scales on CPU%, which doesn't reflect active GPU cores or VRAM pressure. Always use Kthena and `AutoscalingPolicy`.
- **Ignoring KV Cache Decay**: Deploying a Prefill-Decode disaggregated architecture without setting up a `kvConnector` forces the decode pods to recalculate prompts locally anyway, completely negating the latency benefit of the split.
- **Allowing Unlimited Sandbox TTL**: Setting `sessionTimeout` to 0 or omitting `maxSessionDuration` on AgentCube sandboxes allows idle containers to stay alive indefinitely, saturating cluster resources. Always enforce short idle timeouts (e.g., `15m`).
- **Running Untrusted Sandbox Code without Kata Containers**: Running AgentCube code interpreter containers on standard shared Linux kernels opens the node to host-escape security vulnerabilities. For production setups, always define a secure `runtimeClassName` (like Kata or Kuasar) in your templates to enforce VM-level isolation.

## Reference files

- `references/kthena-llm-orchestration.md` — Kthena architecture, `ModelServing`, `ModelServer`, `ModelBooster` specs, control plane lifecycles, and Helm configuration.
- `references/kthena-advanced-routing.md` — Prefill-decode split topologies, KV cache connectors, LoRA adapter routing, Canary rollouts, token rate limits, and request flows.
- `references/agentcube-sandboxes.md` — AgentCube split-plane architecture, `CodeInterpreter` and `AgentRuntime` CRDs, PicoD daemon, Python SDK integration, and MCP deployment.
