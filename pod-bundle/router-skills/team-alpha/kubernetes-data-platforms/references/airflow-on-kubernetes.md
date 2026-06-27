# Apache Airflow on Kubernetes

Deploying Airflow with the official Helm chart, the `KubernetesExecutor`, getting DAGs into the cluster, launching work as pods, and authoring DAGs that scale.

## Contents
- Airflow architecture recap
- Executors — pick KubernetesExecutor on K8s
- Installing via the official Helm chart
- Key Helm values (executor, logging, gitSync, webserver)
- Getting DAGs in: gitSync vs baked image
- Connections & Variables (secrets)
- Running work as pods: KubernetesPodOperator
- Triggering Spark: SparkKubernetesOperator + RBAC
- DAG authoring best practices

## Airflow architecture recap

Airflow orchestrates pipelines as **DAGs** (directed acyclic graphs of tasks). Core components:
- **Metadata database** (Postgres/MySQL in prod; never SQLite beyond local dev) — the single source of truth for DAG/task state, XComs, connections, variables.
- **Scheduler** — examines DAGs and dependencies, decides what to run, enforces state via DB locks (prevents race conditions; scales to multiple schedulers).
- **Webserver** — the UI + REST API.
- **Executor** — decides *how/where* tasks run.
- **Workers** — actually execute task logic.

Airflow's strengths: scheduling, dependency management (DAGs), monitoring (UI, logs, Gantt), and abstraction (pipeline authors write business logic, not orchestration plumbing). **It is an orchestrator, not a data-processing engine** — see the anti-pattern note below.

## Executors — pick KubernetesExecutor on K8s

- **LocalExecutor** — parallel processes on one host. Local dev only; no horizontal scale.
- **CeleryExecutor** — a fixed worker pool + a broker (Redis/RabbitMQ). Horizontal scale, but you run and size a standing worker fleet.
- **KubernetesExecutor** — launches a **fresh worker pod per task**, then tears it down. Excellent scaling, per-task resource isolation, no idle workers, no Redis/broker needed. **This is the right choice on Kubernetes.**
- **CeleryKubernetesExecutor** — hybrid; route small tasks to Celery, heavy ones to K8s pods. Niche.

With `KubernetesExecutor` you do **not** need Redis — disable it in the chart.

## Installing via the official Helm chart

```bash
helm repo add apache-airflow https://airflow.apache.org
helm repo update

helm install airflow apache-airflow/airflow \
  --namespace airflow --create-namespace \
  -f values.yaml
```

The release runs a DB migration job on first install, so it can take a few minutes before the UI is up.

## Key Helm values

A trimmed `values.yaml` covering the important bits:

```yaml
# Pin versions deliberately. defaultAirflowTag must match a real image
# compatible with your chart version.
airflowVersion: "2.9.3"
defaultAirflowTag: "2.9.3"

executor: "KubernetesExecutor"

# KubernetesExecutor doesn't need Redis:
redis:
  enabled: false

# Custom image (e.g. one with the cncf-kubernetes provider pinned for SparkKubernetesOperator):
images:
  airflow:
    repository: "myrepo/airflow"
    tag: "2.9.3-cncf"
    pullPolicy: IfNotPresent

# Remote logging to S3 (best practice: durable, survives pod teardown):
env:
  - name: "AIRFLOW__LOGGING__REMOTE_LOGGING"
    value: "True"
  - name: "AIRFLOW__LOGGING__REMOTE_BASE_LOG_FOLDER"
    value: "s3://my-airflow-logs/airflow-logs/"
  - name: "AIRFLOW__LOGGING__REMOTE_LOG_CONN_ID"
    value: "aws_conn"        # an Airflow Connection you create in the UI

webserver:
  service:
    type: ClusterIP          # expose via Ingress with auth in prod; LoadBalancer only for demos
defaultUser:
  enabled: true
  role: Admin
  username: admin
  email: admin@example.com
  firstName: Admin
  lastName: User
  password: "changeme"       # change in the UI immediately; don't keep it in values long-term

# Ship DAGs from Git:
dags:
  gitSync:
    enabled: true
    repo: "https://github.com/<org>/<repo>.git"
    branch: main
    rev: HEAD
    depth: 1
    subPath: "dags"
    # for private repos, reference a Secret with an SSH key / token:
    # credentialsSecret: airflow-git-credentials
```

Notes:
- **`defaultAirflowTag`/`airflowVersion` must be compatible with the chart version** you install. Mismatches cause migration/runtime errors. Check the chart's docs for the supported app version.
- The remote-logging env vars are why you can read task logs even after the per-task worker pod is gone — without it, KubernetesExecutor logs vanish with the pod.
- `webserver.service.type: LoadBalancer` is convenient for a quick demo but exposes the UI publicly; in production use `ClusterIP` + an authenticated Ingress and keep it in the VPC.

After install, find the UI:
```bash
kubectl get svc -n airflow          # if LoadBalancer, note EXTERNAL-IP (then :8080)
# or port-forward for ClusterIP:
kubectl port-forward svc/airflow-webserver 8080:8080 -n airflow
```

## Getting DAGs in: gitSync vs baked image

- **gitSync** (recommended) — a sidecar continuously pulls a Git repo's `dags/` folder into the scheduler/worker pods. Push to Git → DAG appears in seconds. Decouples DAG changes from image rebuilds.
- **Baked into the image** — copy DAGs into a custom Airflow image. More reproducible/immutable, but every DAG change is an image build+rollout. Use when you want strict provenance or can't reach Git from the cluster.
- **PV-backed `dags.persistence`** — mount a shared volume; needs RWX storage and an upload mechanism. Generally inferior to gitSync.

## Connections & Variables (secrets)

DAGs authenticate to external systems via **Connections** (Admin → Connections) and read parameters/secrets via **Variables** (Admin → Variables). Airflow auto-masks values it detects as secrets in the UI. For AWS, create a Connection of type *Amazon Web Services* named to match `AIRFLOW__LOGGING__REMOTE_LOG_CONN_ID` (e.g. `aws_conn`). In production, back these with a real secrets backend (AWS Secrets Manager / Vault / GCP Secret Manager) via `AIRFLOW__SECRETS__BACKEND` rather than storing creds in the metadata DB.

```python
from airflow.models import Variable
aws_key = Variable.get("aws_access_key_id")
```

## Running work as pods: KubernetesPodOperator

The idiomatic way to run *any* containerized step from Airflow on K8s — Airflow launches a pod, waits, collects logs/exit code:

```python
from airflow.providers.cncf.kubernetes.operators.pod import KubernetesPodOperator

process = KubernetesPodOperator(
    task_id="transform",
    namespace="airflow",
    image="myrepo/transform:1.2.0",
    cmds=["python", "transform.py"],
    arguments=["--date", "{{ ds }}"],
    name="transform-pod",
    get_logs=True,
    is_delete_operator_pod=True,        # clean up after success
    container_resources={
        "requests": {"cpu": "500m", "memory": "1Gi"},
        "limits": {"cpu": "1", "memory": "2Gi"},
    },
)
```

This is how you keep heavy work *out* of the Airflow worker and in a right-sized, isolated pod.

## Triggering Spark: SparkKubernetesOperator + RBAC

To submit `SparkApplication`s from a DAG, the Airflow image needs the **`apache-airflow-providers-cncf-kubernetes`** provider, and the Airflow worker ServiceAccount needs RBAC to create `sparkapplications`. Operator + sensor pattern (launch, then wait):

```python
from airflow.providers.cncf.kubernetes.operators.spark_kubernetes import SparkKubernetesOperator
from airflow.providers.cncf.kubernetes.sensors.spark_kubernetes import SparkKubernetesSensor

submit = SparkKubernetesOperator(
    task_id="tsvs_to_parquet",
    namespace="airflow",
    application_file="spark_tsv_parquet.yaml",   # a SparkApplication manifest in the dags folder
    kubernetes_conn_id="kubernetes_default",
    do_xcom_push=True,
)

wait = SparkKubernetesSensor(
    task_id="tsvs_to_parquet_sensor",
    namespace="airflow",
    application_name="{{ task_instance.xcom_pull(task_ids='tsvs_to_parquet')['metadata']['name'] }}",
    kubernetes_conn_id="kubernetes_default",
)

submit >> wait
```

RBAC so Airflow workers can drive the Spark Operator's CRDs:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: spark-cr
rules:
  - apiGroups: ["sparkoperator.k8s.io"]
    resources: ["sparkapplications", "scheduledsparkapplications"]
    verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: airflow-spark-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: spark-cr
subjects:
  - kind: ServiceAccount
    name: airflow-worker          # the SA the KubernetesExecutor worker pods use
    namespace: airflow
```

Also grant the `spark` ServiceAccount (used by the driver) pod-management rights in the namespace where jobs run:

```bash
kubectl create serviceaccount spark -n airflow
kubectl create clusterrolebinding spark-role-airflow \
  --clusterrole=edit --serviceaccount=airflow:spark
```

> The Spark Operator provider's class/import paths have shifted across provider versions. If `SparkKubernetesOperator` import fails, pin the `cncf-kubernetes` provider version known to work with your Airflow version and check that provider's changelog.

## DAG authoring best practices

Modern TaskFlow-API DAG:

```python
from airflow.decorators import dag, task
from datetime import datetime

@dag(
    schedule="@daily",
    start_date=datetime(2024, 1, 1),
    catchup=False,            # critical: don't backfill every missed run on enable
    tags=["imdb"],
    description="IMDb batch pipeline",
)
def imdb_batch():
    @task
    def acquire() -> str:
        # download to /tmp, upload to S3, return the path/key
        return "s3://bucket/landing/imdb/"

    @task
    def kickoff_processing(landing: str):
        ...   # trigger Spark, etc.

    kickoff_processing(acquire())

imdb_batch()
```

Rules of thumb:
- **`catchup=False`** unless you genuinely want historical backfill — otherwise enabling a DAG fires every missed interval at once and overloads the cluster.
- **Keep tasks small, single-purpose, and idempotent.** Independent tasks run in parallel (express with `[a, b] >> c` or TaskFlow data dependencies); fine-grained tasks are easier to retry and debug.
- **Never process big data inside an Airflow task.** A `pandas`/`PythonOperator` that loads a large dataset will OOM the worker. Push compute to Spark/`KubernetesPodOperator`; Airflow only orchestrates and waits.
- Use **TaskGroups** to organize related steps (e.g. a Spark submit + its sensor) visually and logically.
- Set per-task `container_resources` / pod resources so KubernetesExecutor worker pods are right-sized.
- Pass small handoffs via **XCom**; pass large data via object storage (write in one task, read the path in the next).
