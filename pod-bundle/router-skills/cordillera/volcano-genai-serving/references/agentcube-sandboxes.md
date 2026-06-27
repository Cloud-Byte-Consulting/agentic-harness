# AgentCube: Serverless Agent Sandboxes on Kubernetes

Orchestrating interactive AI Agents or running untrusted code interpreters (e.g., executing Python script tools on behalf of an LLM) requires execution environments with sub-second creation times, precise inactivity timeouts, and strict security isolation. AgentCube provides this serverless agent infrastructure on Volcano.

---

## 1. Split-Plane Architecture

AgentCube decouples the control plane from the request path to maintain high responsiveness:

```
┌──────────────────────────────────────────────────────────────┐
│                    CLIENT APPLICATION                        │
└──────────────────────────────┬───────────────────────────────┘
                               │ HTTP/gRPC via Session ID
                               ▼
┌──────────────────────────────────────────────────────────────┐
│                  DATA PLANE: AgentCube Router                │
│   • Intercepts calls, validates JWT tokens                   │
│   • Resolves Session ID -> Sandbox Pod Endpoint (via Redis)  │
│   • Dynamic routing with sub-millisecond overhead           │
└──────────────────────────────┬───────────────────────────────┘
                               │
                               ▼
┌──────────────────────────────────────────────────────────────┐
│             CONTROL PLANE: AgentCube Workload Manager        │
│   • Manages warm pools (Pre-warmed Sandbox Pods)             │
│   • Spawns Sandboxes on-demand (lazy creation)               │
│   • Garbage Collector: evicts idle sessions based on TTL     │
└──────────────────────────────────────────────────────────────┘
```

---

## 2. Defining Sandboxes: `CodeInterpreter` and `AgentRuntime`

AgentCube provides two CRDs under the API group `runtime.agentcube.volcano.sh/v1alpha1`.

### A. `CodeInterpreter`
Strictly optimized for secure, short-lived code-execution containers:

```yaml
apiVersion: runtime.agentcube.volcano.sh/v1alpha1
kind: CodeInterpreter
metadata:
  name: python-sandbox-runner
  namespace: default
spec:
  ports:
    - pathPrefix: "/"
      port: 8080
      protocol: "HTTP"
  template:
    image: ghcr.io/volcano-sh/picod:latest # Inside-pod daemon
    args: ["--workspace=/root"]
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
      limits:
        cpu: 500m
        memory: 512Mi
    runtimeClassName: kata            # VM-level kernel isolation (CRITICAL)
  sessionTimeout: "15m"               # Hibernate/kill if idle for 15 minutes
  maxSessionDuration: "8h"            # Hard TTL regardless of activity
  warmPoolSize: 2                     # Keeps 2 pre-warmed idle pods ready
```
- `warmPoolSize: 2`: The Workload Manager maintains 2 ready, idle pods. When a user requests a session, they are allocated a pre-warmed pod instantly, reducing cold-start latency from ~5 seconds to **under 100 milliseconds**.

### B. `AgentRuntime`
Designed for stateful, long-running agent containers that require standard Pod configurations (such as volume mounts, sidecars, or specific security contexts):

```yaml
apiVersion: runtime.agentcube.volcano.sh/v1alpha1
kind: AgentRuntime
metadata:
  name: stateful-browser-agent
  namespace: default
spec:
  targetPort:
    - pathPrefix: "/"
      port: 8080
      protocol: "HTTP"
  podTemplate:
    spec:
      containers:
        - name: browser-use
          image: playright-browser:latest
          command: ["python", "agent_node.py"]
          volumeMounts:
            - name: session-data
              mountPath: /data
      volumes:
        - name: session-data
          persistentVolumeClaim:
            claimName: agent-pvc
  sessionTimeout: "30m"
  maxSessionDuration: "12h"
```

---

## 3. End-to-End Python SDK Usage

Developers interact with AgentCube sandboxes via an intuitive, imperative SDK:

```python
from agentcube import CodeInterpreterClient

# 1. Establish context with a pre-warmed runner
with CodeInterpreterClient(name="python-sandbox-runner", namespace="default") as client:
    
    # 2. Write local variables/files to the sandbox workspace
    client.write_file("data.csv", "name,score\nAlice,95\nBob,87")
    
    # 3. Execute python code inside the sandbox (executed by PicoD daemon)
    code = """
import pandas as pd
df = pd.read_csv('data.csv')
print(df['score'].mean())
"""
    result = client.run_code(language="python", code=code)
    print(f"Result from sandbox: {result}") # Returns: 91.0
    
    # 4. Download generated artifacts
    client.download_file("output.png", "./local_output.png")

# Sandbox pod is automatically recycled/released back to the warm pool upon context exit
```

---

## 4. Integrating with Model Context Protocol (MCP)

AgentCube can act as a standard Model Context Protocol (MCP) server. When an LLM client (such as Cursor or Claude Desktop) requests code execution, the MCP gateway delegates the container lifecycle to AgentCube, isolating the execution securely inside the Kubernetes cluster.
