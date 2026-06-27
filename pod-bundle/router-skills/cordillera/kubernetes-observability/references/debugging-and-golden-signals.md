# Debugging & Golden Signals

The production debugging workflow: the four golden signals, the `kubectl` toolkit, ephemeral debug
containers, correlating the three pillars, and reading common failure signatures.

## Contents
- [The mental model: symptoms → cause](#the-mental-model-symptoms--cause)
- [Golden signals, RED, USE](#golden-signals-red-use)
- [Impact vs diagnostic metrics](#impact-vs-diagnostic-metrics)
- [The kubectl toolkit](#the-kubectl-toolkit)
- [Ephemeral debug containers](#ephemeral-debug-containers)
- [Correlating metrics, logs, and traces](#correlating-metrics-logs-and-traces)
- [Common failure signatures](#common-failure-signatures)
- [A worked debugging flow](#a-worked-debugging-flow)

## The mental model: symptoms → cause

Start at the **symptom the user feels** (errors, slowness), confirm it on a dashboard, then drill toward
the **cause** using traces (where) and logs (what). Don't start from a random low-level metric. This
keeps you from chasing CPU graphs while checkouts are failing for an unrelated reason.

Note that a pod can look healthy while its app is broken: the container is "running" but the binary
inside is wedged, or a **sidecar** (mesh proxy, Vault agent) is down while the main app is up. Always
look inside the pod, not just at its phase.

## Golden signals, RED, USE

- **Four golden signals**: **latency** (how long requests take — split success vs error latency),
  **traffic** (demand, e.g. req/s), **errors** (failed-request rate), **saturation** (how full the most
  constrained resource is — CPU throttling, memory vs limit, queue depth).
- **RED** for request services: **R**ate, **E**rrors, **D**uration. Dashboard every service this way.
- **USE** for resources (nodes, disks, CPU): **U**tilization, **S**aturation, **E**rrors.

Alert on golden-signal **symptoms**; use everything else for diagnosis. Queries for all of these are in
`promql-and-alerting.md`.

## Impact vs diagnostic metrics

Split your metrics by role so you can triage:
- **Impact metrics** — user/business experience: latency, error rate, request failures. These drive
  alerts and SLOs.
- **Diagnostic metrics** — system internals for root cause: CPU, memory, network, GC, DB query time,
  queue depth. These you read *after* an impact alert fires.

An alert on an impact metric tells you to act; diagnostic metrics tell you why.

## The kubectl toolkit

```bash
# What just happened (state changes, scheduling, evictions, probe failures)
kubectl get events -n <ns> --sort-by=.lastTimestamp
kubectl get events -n <ns> --field-selector involvedObject.name=<pod>

# Pod state: restarts, last state (OOMKilled?), probe config, image, node, conditions
kubectl describe pod <pod> -n <ns>

# Live resource usage (needs Metrics Server)
kubectl top pod <pod> -n <ns> --containers
kubectl top nodes
kubectl top pod -n <ns> --sort-by=memory

# Logs
kubectl logs <pod> -n <ns> -c <container>
kubectl logs <pod> -n <ns> -c <container> --previous      # crashed instance
kubectl logs -f deploy/<name> -n <ns> --max-log-requests=10
kubectl logs <pod> -n <ns> --since=15m --tail=200

# Is it scheduled / why pending
kubectl get pod <pod> -n <ns> -o wide
kubectl describe pod <pod> -n <ns> | grep -A10 Events     # "0/3 nodes available..." etc.

# Run a one-off shell in a running pod (if it has a shell)
kubectl exec -it <pod> -n <ns> -c <container> -- sh
```

`describe` and `events` answer most "why won't it start / why did it restart" questions before you ever
open a dashboard.

## Ephemeral debug containers

When the app container has no shell or tools (distroless/scratch), attach an **ephemeral container** that
shares the target's namespaces — no rebuild, no restart:

```bash
# Attach a busybox into a running pod, sharing the target container's process namespace
kubectl debug -it <pod> -n <ns> --image=busybox:1.36 --target=<container>

# Richer toolbox
kubectl debug -it <pod> -n <ns> --image=nicolaka/netshoot --target=<container>
# inside: ss -tlnp, curl localhost:8080/healthz, nslookup, tcpdump, ps aux, cat /proc/1/...

# Debug a node by scheduling a privileged pod onto it
kubectl debug node/<node> -it --image=busybox

# Copy a crashing pod with a changed command/image to poke at it without traffic
kubectl debug <pod> -n <ns> --copy-to=<pod>-debug --container=<container> -- sleep 1d
```

`--target` shares the PID namespace so you can see the app's processes and `/proc`. Requires the
`EphemeralContainers` feature (GA since Kubernetes 1.25). The ephemeral container can't be removed from
the pod, only the pod deleted — fine for a throwaway debug.

## Correlating metrics, logs, and traces

The payoff of all three pillars is fast root cause:
1. **Metric** (Grafana) shows the symptom — p95 latency or error rate spiked at 10:02.
2. **Trace** (Jaeger/Tempo) for a slow/failed request in that window shows the span that's slow — e.g.
   the call to `payments-service`, or a third-party API.
3. **Logs** (Loki/OpenSearch), filtered by that service and the request's `trace_id`, show the actual
   error ("connection timeout to payment gateway").

Carry the `trace_id` in structured logs so step 2→3 is one click (see `logging.md` and
`tracing-opentelemetry.md`). Without correlation IDs you're grepping disjoint logs and the real culprit
(often an unmonitored dependency) stays hidden — the classic "fragmentation" failure.

## Common failure signatures

| Symptom | Where you see it | Likely cause / next step |
|---|---|---|
| `CrashLoopBackOff` | `get pods`, `describe` | App exits on start. `logs --previous`; check config/secrets/probes. |
| `OOMKilled` (in `describe` Last State) | `describe pod` | Memory limit too low or leak. Raise limit or fix leak; check `container_memory_working_set_bytes` vs limit. |
| CPU throttling, latency up but CPU "fine" | throttle ratio query | CPU **limit** throttling. Raise/remove CPU limit (see autoscaling skill). |
| Pod `Pending` | `describe` Events | No node fits requests, or PVC/affinity/taint. "0/N nodes available." |
| `ImagePullBackOff` | `describe` Events | Bad image ref or missing pull secret. |
| Readiness failing, no traffic | `describe` probes, endpoints | Probe path/port wrong or app slow to start; check `kubectl get endpoints`. |
| Healthy pod, broken app | `exec`/`debug`, app logs | Wedged binary or a down sidecar; look inside the pod. |
| Metric disappeared | Prometheus Targets `DOWN` | Scrape broke (selector/RBAC/port) or app crashed — `absent()` alert. |

CPU is **compressible** (throttled when over limit, keeps running); memory is **non-compressible** (over
limit → process killed → container restart). That distinction explains a lot of incidents. (Resource
requests/limits and right-sizing: see **kubernetes-autoscaling-scheduling**.)

## A worked debugging flow

Checkout latency alert fires:
1. `kubectl get events` + Grafana RED dashboard for the orders service → confirm error/latency spike,
   note the start time.
2. `kubectl describe pod` + `kubectl top pod --containers` → any restarts, OOMKilled, or saturation?
3. If pods look fine, open a **trace** in the spike window → the slow span is the `payments-service`
   call.
4. `kubectl logs deploy/payments --since=20m | <filter by trace_id>` (or Loki/OpenSearch) → "timeout to
   payment gateway."
5. Remediation per runbook: if post-deploy, roll back (workload rollout: see **kubernetes-workloads**);
   if a dependency, check/scale it and add a monitor for that third-party so it's not invisible next
   time; verify recovery on the dashboard before closing.

After the incident, review: did the alert fire early enough and carry enough context? Update the rule
and runbook. Periodically simulate failures to confirm alerts reach the right channel and responders can
follow the runbook — observability is a living discipline, not a one-time setup.
