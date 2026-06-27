# Fractional GPU & Device Sharing with Volcano

Standard Kubernetes restricts accelerator allocation to integer boundaries (1 container gets 1 physical GPU). This is highly inefficient for smaller inference tasks, dev environments, and data preparation steps. Volcano's `deviceshare` plugin enables multiple containers to share the same physical GPU safely.

---

## 1. How Volcano Device Sharing Works

The `deviceshare` plugin intercepts resource scheduling decisions, looking for fractional GPU requests. It operates along two main parameters:
- **GPU Core Shares**: The percentage of GPU processing time requested (expressed from 1 to 100).
- **GPU Memory**: The amount of VRAM requested.

Instead of registering nodes with native `nvidia.com/gpu` (integers), Volcano maps fractional assets in its internal cache and registers nodes with virtual core/memory shares.

---

## 2. Scheduler Config Configuration

Ensure the `deviceshare` plugin is loaded under your actions tier in `volcano-scheduler.conf`.

```yaml
actions: "enqueue, allocate, backfill"
tiers:
  - plugins:
    - name: priority
  - plugins:
    - name: deviceshare        # Activates fractional scheduling
      arguments:
        # Define fractional resource identifiers
        vgpu.core: "volcano.sh/vgpu-cores"
        vgpu.memory: "volcano.sh/vgpu-memory"
```

---

## 3. Deploying a Fractional GPU Job

Rather than requesting standard `nvidia.com/gpu`, request fractional vGPU resource units:

```yaml
apiVersion: batch.volcano.sh/v1alpha1
kind: Job
metadata:
  name: shared-inference-job
spec:
  minAvailable: 2
  schedulerName: volcano
  tasks:
    - name: predictor
      replicas: 2
      template:
        spec:
          containers:
            - name: model-server
              image: triton-inference:latest
              resources:
                requests:
                  # Request 30% of a GPU core
                  volcano.sh/vgpu-cores: "30"
                  # Request 4GiB of GPU memory (VRAM)
                  volcano.sh/vgpu-memory: "4Gi"
                  cpu: "2"
                  memory: "8Gi"
                limits:
                  volcano.sh/vgpu-cores: "30"
                  volcano.sh/vgpu-memory: "4Gi"
                  cpu: "2"
                  memory: "8Gi"
```

- `volcano.sh/vgpu-cores: "30"`: The scheduler will binpack up to three pods with this request onto a single physical GPU (utilizing 90% capacity).
- `volcano.sh/vgpu-memory: "4Gi"`: Restricts the pod from consuming more than 4GiB of VRAM. If the node has a 16GiB GPU, the scheduler can pack 4 such pods safely.

---

## 4. Integration with NVIDIA MPS / MIG

- **CUDA Multi-Process Service (MPS)**: Highly recommended when using `deviceshare` on shared nodes. MPS enables multiple CUDA contexts to run concurrently on a single GPU with physical hardware-enforced memory and compute partitioning, reducing context-switch overhead to near-zero.
- **Multi-Instance GPU (MIG)**: Hardware-level slicing. While MIG physically slices a GPU into static chunks (e.g., 1g.5gb), Volcano's `deviceshare` operates at the *software level*, offering higher density and dynamic scaling without re-configuring the GPU hardware profile.

---

## 5. Troubleshooting GPU Sharing

### A. CUDA Out of Memory (OOM) Errors:
If containers crash with CUDA OOM, check if they are requesting vgpu-memory limits.
- **Anti-pattern**: Specifying vgpu-cores without setting `vgpu-memory`. The model will start up and consume all physical VRAM, crashing sister containers. Always specify both.

### B. View Shared Allocations on Nodes:
```bash
# Query node capacity to verify virtual device registration
kubectl describe node gpu-worker-01
# Look for:
# Allocatable:
#   volcano.sh/vgpu-cores: 100
#   volcano.sh/vgpu-memory: 16Gi
```
