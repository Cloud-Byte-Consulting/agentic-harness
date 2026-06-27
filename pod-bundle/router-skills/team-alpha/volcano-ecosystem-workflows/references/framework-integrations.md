# Framework & Operator Integrations with Volcano

Volcano acts as the foundational scheduling substrate for major AI/ML and big data operators. This reference guide covers how to integrate Volcano with Kubeflow, MPI, Apache Spark, and Ray.

---

## 1. Kubeflow Operators (PyTorch / TensorFlow)

Kubeflow operators (like `pytorch-operator` and `tf-operator`) reconcile their own CRDs (`PyTorchJob`, `TFJob`). However, they do not schedule pods — they delegate scheduling to Kubernetes. By telling them to use Volcano, you enforce gang scheduling and queue management on their distributed pods.

### Integrating PyTorchJob:
Set the `schedulerName` to `volcano` in the `PyTorchJob` specification:

```yaml
apiVersion: "kubeflow.org/v1"
kind: "PyTorchJob"
metadata:
  name: "pytorch-dist-mnist"
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      restartPolicy: OnFailure
      template:
        spec:
          schedulerName: volcano      # Delegate to Volcano
          containers:
            - name: pytorch
              image: pytorch/pytorch:latest
              command: ["python", "mnist.py"]
    Worker:
      replicas: 3
      restartPolicy: OnFailure
      template:
        spec:
          schedulerName: volcano      # Delegate to Volcano
          containers:
            - name: pytorch
              image: pytorch/pytorch:latest
              command: ["python", "mnist.py"]
```

Volcano's controller automatically observes these pods, detects they belong to the same distributed group, generates a matching `PodGroup` behind the scenes, and applies gang scheduling (requiring 4 pods to be ready).

---

## 2. MPI (Message Passing Interface) with Job Plugins

MPI jobs (like those deployed by Kubeflow's `mpi-operator`) require complex networking setups. Volcano provides native Job Plugins to handle this securely and automatically.

### Example: MPI Job with `ssh` and `svc` Plugins:
```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: mpi-hpc-job
spec:
  minAvailable: 3             # Master + 2 workers
  schedulerName: volcano
  plugins:
    ssh: []                   # Generates passwordless SSH keypairs and mounts them
    svc: []                   # Generates headless services & hosts file discovery
  tasks:
    - name: mpimaster
      replicas: 1
      policies:
        - event: TaskCompleted
          action: CompleteJob # Complete worker tasks once master finishes
      template:
        spec:
          containers:
            - name: mpimaster
              image: volcanosh/example-mpi:0.0.3
              command:
                - /bin/sh
                - -c
                - |
                  # Volcano's svc plugin writes worker hostnames here
                  MPI_HOST=`cat /etc/volcano/mpiworker.host | tr "\n" ","`;
                  /usr/sbin/sshd
                  mpiexec --allow-run-as-root --host ${MPI_HOST} -np 2 mpi_hello;
    - name: mpiworker
      replicas: 2
      template:
        spec:
          containers:
            - name: mpiworker
              image: volcanosh/example-mpi:0.0.3
              command:
                - /bin/sh
                - -c
                - "/usr/sbin/sshd -D"
```

#### How the Plugins Work:
1. **`ssh`**: Generates a temporary, cryptographically secure SSH key pair. It mounts the public key to `/root/.ssh/authorized_keys` and the private key to `/root/.ssh/id_rsa` on all master and worker containers, enabling passwordless `mpiexec` execution.
2. **`svc`**: Sets up a headless cluster DNS service. It writes the list of active worker pod IP/hostname allocations to `/etc/volcano/mpiworker.host` inside the master container.

---

## 3. Apache Spark on Volcano

Apache Spark (v3.3.0+) native Kubernetes scheduler supports Volcano natively.

### Submit Spark Job with Volcano:
Submit jobs using the Spark CLI, referencing Volcano:

```bash
spark-submit \
  --master k8s://https://<kubernetes-api-server> \
  --deploy-mode cluster \
  --conf spark.kubernetes.scheduler=volcano \  # Opt-in to Volcano
  --conf spark.kubernetes.driver.podTemplateFile=driver-template.yaml \
  --conf spark.kubernetes.executor.podTemplateFile=executor-template.yaml \
  --class org.apache.spark.examples.SparkPi \
  local:///opt/spark/examples/jars/spark-examples_2.12-3.4.0.jar
```

Alternatively, if using the `spark-on-k8s-operator`, specify the scheduler in the YAML spec:
```yaml
spec:
  type: Scala
  mode: cluster
  image: "gcr.io/spark-operator/spark:v3.1.1"
  batchScheduler: Volcano     # Enforces gang-scheduling of executor pods
```

---

## 4. Ray on Volcano

Ray manages clusters of worker nodes for AI. Ray pod scaling can cause race conditions where too many worker pods compete for resources.

By setting the scheduler to Volcano, Ray's autoscaling driver can bundle worker pod claims into a single Volcano Queue.

### `RayCluster` Snippet:
```yaml
apiVersion: ray.io/v1
kind: RayCluster
metadata:
  name: ray-cluster
spec:
  headGroupSpec:
    rayStartParams:
      dashboard-host: '0.0.0.0'
    template:
      spec:
        schedulerName: volcano  # Head node uses Volcano
  workerGroupSpecs:
    - groupName: gpu-group
      replicas: 4
      template:
        spec:
          schedulerName: volcano  # Workers use Volcano
```
This guarantees that Ray GPU worker pods scale in clean, scheduling-controlled batches.
