# Debugging & Observing Containers

Diagnosing why a container won't start, crashes, or misbehaves — plus logging and basic monitoring.

## Contents
- [First-response triage](#first-response-triage)
- [Container states and exit codes](#container-states-and-exit-codes)
- [docker logs](#docker-logs)
- [docker inspect](#docker-inspect)
- [docker exec vs docker attach](#docker-exec-vs-docker-attach)
- [Debugging distroless/scratch and ephemeral debug containers](#debugging-distrolessscratch-and-ephemeral-debug-containers)
- [netshoot — network debugging](#netshoot--network-debugging)
- [Live code debugging inside a container](#live-code-debugging-inside-a-container)
- [Logging drivers and rotation](#logging-drivers-and-rotation)
- [Shipping logs to a central system](#shipping-logs-to-a-central-system)
- [Metrics with Prometheus and Grafana](#metrics-with-prometheus-and-grafana)

## First-response triage

When a container is broken, work in this order:

1. **`docker ps -a`** — is it running, exited, restarting? Note the **STATUS** and exit code.
2. **`docker logs --tail 50 -f <c>`** — read what the app printed before it died. Usually the answer.
3. **`docker inspect <c>`** — verify mounts, env, network, published ports, the exact command, restart reason,
   `OOMKilled`.
4. **`docker exec -it <c> sh`** — get a shell in a *running* container to check connectivity, files, processes.
5. If there's no shell (distroless/scratch) or it already exited, use an **ephemeral debug container** or
   **netshoot** sharing its namespaces.

## Container states and exit codes

States: `created`, `restarting`, `running`, `paused`, `exited`, `dead`. In `docker ps -a` a healthy container
shows `Up …`; a finished one shows `Exited (<code>) … ago`.

| Exit code | Meaning |
|---|---|
| 0 | Clean exit (the main process finished successfully — normal for a job/one-shot container). |
| 1 | Generic application error. |
| 2 | Misuse / resource not found. |
| 125 | Docker daemon error (the `docker run` itself failed). |
| 126 | Command found but not executable. |
| 127 | Command/binary not found in the image. |
| 137 | SIGKILL (128+9) — usually **OOM** (memory limit) or `docker kill`/forced stop. |
| 143 | SIGTERM (128+15) — graceful stop. |

A non-zero or unexpected exit + the logs usually pinpoint the cause. `docker stop` sends SIGTERM, waits ~10s,
then SIGKILL — a container that takes the full 10s likely has a shell-form CMD not forwarding signals (see
`dockerfile-authoring.md`).

## docker logs

Containerized apps should log to **STDOUT/STDERR** (not files) so Docker captures them.
```bash
docker logs <c>                 # all logs from the beginning (current json-file)
docker logs --tail 5 <c>        # last 5 lines
docker logs --tail 5 -f <c>     # last 5, then follow live
docker logs --since 10m <c>     # last 10 minutes
docker logs --until 2024-02-23T18:35:13 <c>
```
Note: with the `none` log driver, `docker logs` returns an error (no log to read). With `json-file`, `logs`
reads only the *current* file, not rolled-off older files. (Docker 20.10+ "dual logging" lets `docker logs`
work even with non-file drivers via a local buffer.)

## docker inspect

Low-level JSON metadata: network settings, mounts, env, the start command, state, restart reason.
```bash
docker container inspect <c>
docker container inspect -f '{{json .State}}' <c> | jq .          # just the state
docker container inspect --format '{{json .Mounts}}' <c> | jq .   # mounts
docker container inspect <c> | grep HostPort                       # published host port
```

## docker exec vs docker attach

These are different and people conflate them:

- **`docker exec`** starts a **new** process inside a running container. Safe — it does not touch PID 1.
  ```bash
  docker exec -it <c> /bin/sh            # interactive shell
  docker exec <c> ps                     # one-off command; PID 1 is the container's main process
  docker exec -it -e MY_VAR=hello <c> /bin/sh
  ```
  Exiting the shell (`Ctrl+D`/`exit`) leaves the container running.

- **`docker attach`** connects your terminal to the container's **main process** (PID 1) STDIN/STDOUT/STDERR.
  Useful to watch a foreground app's live output, but **`Ctrl+C` here sends SIGINT to PID 1 and stops the
  container.** To detach without killing it, use **`Ctrl+P` then `Ctrl+Q`** (doesn't work inside VS Code's
  integrated terminal — use a standalone terminal).

Prefer `exec` for diagnostics; reach for `attach` only to interact with the primary process directly.

## Debugging distroless/scratch and ephemeral debug containers

Minimal images (distroless/scratch) have no shell, so `docker exec … sh` fails. Two approaches:

- Run a **debug container that shares the target's namespaces** so you can inspect its network/PID/filesystem
  context without modifying its image (see netshoot below for the network case).
- For inspecting the host/another container's filesystem, a privileged helper with `nsenter` works:
  ```bash
  docker run -it --privileged --pid=host debian nsenter -t 1 -m -u -n -i sh
  ```
  `--pid=host` lets it join host PID 1's namespaces (e.g. to browse Docker's volume backing dirs inside the
  Desktop VM).

(In Kubernetes, the equivalent is `kubectl debug` ephemeral containers — see the **kubernetes-workloads** skill.)

## netshoot — network debugging

`nicolaka/netshoot` bundles curl, dig, tcpdump, traceroute, etc. Attach it to a running container's network
namespace to debug *its* connectivity with no changes to the app image:
```bash
docker run --rm -it --network container:myapp nicolaka/netshoot
# now curl/dig/tcpdump run in myapp's exact network context
```

## Live code debugging inside a container

For an edit-test loop without rebuilding, bind-mount source and use a file-watcher to auto-restart:
- **Node**: `nodemon` (or `node --inspect=0.0.0.0` and attach a debugger on port 9229).
- **Python/Flask**: run with `debug=True` / `--reload`.
- **Java/Spring Boot**: spring-boot-devtools.
- **.NET**: `dotnet watch run`.

```bash
docker run --rm -it --init -v "$(pwd)":/app -p 3000:3000 my-dev-image nodemon index.js
```
Map source into the container (`-v "$(pwd)":/app`), expose the app port and any debug port, and `--init` so
`Ctrl+C` cleanly stops the process. For line-by-line debugging, expose the language's debug port (e.g.
`-p 9229:9229`) and attach your IDE; set `"restart": true` in the VS Code launch config so the debugger
re-attaches after an auto-restart. Line-by-line in-container debugging should be a last resort — prefer unit/
integration tests or debugging on the host.

## Logging drivers and rotation

The default driver is **json-file**. Others: `none`, `journald`, `syslog`, `gelf` (Graylog/Logstash),
`fluentd`, `awslogs` (CloudWatch), `splunk`.

Per-container:
```bash
docker run --log-driver=json-file --log-opt max-size=10m --log-opt max-file=5 nginx
docker run --log-driver none busybox sh -c 'echo hi'   # produces no captured logs
```
Globally (daemon-wide) in `/etc/docker/daemon.json` — **set rotation to avoid filling the disk**:
```json
{
  "log-driver": "json-file",
  "log-opts": { "max-size": "10m", "max-file": "3" }
}
```
Reload the daemon to apply: `sudo kill -SIGHUP $(pidof dockerd)` (reloads config without a full restart), or
restart via Docker Desktop's Settings → Docker Engine.

## Shipping logs to a central system

Containers are ephemeral and spread across hosts, so centralize logs (Elastic Stack: Filebeat → Elasticsearch →
Kibana). On **Linux**, Filebeat tails Docker's JSON files directly:
```yaml
filebeat.inputs:
  - type: container
    paths: ['/var/lib/docker/containers/*/*.log']
output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```
On **macOS/Windows** (`/var/lib/docker/containers` is inside the Desktop VM) use the shared-volume workaround:
apps write logs to a file in a shared volume, and Filebeat tails that volume. Enrich events with a
`service.name` field so you can filter per service in Kibana (`service.name : "node-api" AND message :
"ERROR"`). Useful Kibana/Docker fields: `message`, `container.name`, `log.level`, `@timestamp`.

## Metrics with Prometheus and Grafana

Logs say *what happened*; metrics show *behavior over time*. Prometheus uses a **pull** model — it scrapes a
`/metrics` HTTP endpoint each app exposes (Prometheus format is language-agnostic; client libs exist for Go,
Python, .NET, Node, etc.).

`prometheus.yml`:
```yaml
global:
  scrape_interval: 15s
scrape_configs:
  - job_name: 'app'
    metrics_path: /metrics
    static_configs:
      - targets: ['app:8080']
  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']   # host CPU/mem/disk metrics
```
Run Prometheus + node-exporter + Grafana via Compose; verify scrape targets at `:9090` (Status → Targets, all
`UP`); add Prometheus as a Grafana data source (`http://prometheus:9090`) and build dashboards / basic alerts.
Useful node-exporter queries: `node_cpu_seconds_total`, `node_memory_MemAvailable_bytes`,
`node_filesystem_avail_bytes`.

**Observability = logs + metrics + traces** together. Full cluster-scale observability and runtime security
monitoring (Falco) sit at the orchestrator level — see the **kubernetes-observability** skill; runtime security
tooling is also covered in `container-security.md`.
