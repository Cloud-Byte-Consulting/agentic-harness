# ConfigMaps and Secrets

Decoupling configuration from images and injecting it into Pods.

## Contents
- Why decouple config
- ConfigMaps: create and inspect
- Secrets: create, types, base64 reality
- Consuming as environment variables (env / envFrom)
- Consuming as volume mounts (+ subPath)
- Projected volumes
- Immutability
- Update behavior & gotchas
- Security guidance

## Why decouple config

Container images are immutable; baking environment-specific values (DB endpoints, API
keys, feature flags) into the image forces a rebuild per environment and couples code to
config. ConfigMaps (non-sensitive) and Secrets (sensitive) externalize configuration so
one image runs unchanged across dev/staging/prod. Always create the ConfigMap/Secret
**before** the Pod that references it — a Pod referencing a missing ConfigMap/Secret won't
start (it sticks in `ContainerCreating` with a `FailedMount` event) unless the reference is
marked `optional: true`.

## ConfigMaps

Store key/value pairs — short literals or whole config-file contents.

Declarative:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  COLOR: "blue"
  LOG_LEVEL: "info"
  app.properties: |          # a whole file as one value
    server.port=8080
    cache.ttl=300
```

Imperative:
```bash
kubectl create configmap app-config --from-literal=COLOR=blue --from-literal=LOG_LEVEL=info
kubectl create configmap app-config --from-file=./app.properties      # key = filename
kubectl create configmap app-config --from-file=./conf.d/             # one key per file in dir
kubectl create configmap app-config --from-env-file=./app.env         # key=value lines
```

Inspect (data is plaintext — never put secrets here):
```bash
kubectl get cm                       # DATA column = number of keys
kubectl describe cm app-config
kubectl get cm app-config -o yaml
```

## Secrets

Like ConfigMaps but for sensitive data; values are stored base64-encoded. The default
type is `Opaque`.

Imperative (Kubernetes base64-encodes for you):
```bash
kubectl create secret generic db-secret \
  --from-literal=DB_USER=appadmin \
  --from-literal=DB_PASSWORD='s3cr3t'
kubectl create secret generic tls-key --from-file=./password.txt   # key = filename
```

Declarative with `stringData` (plaintext in the manifest; Kubernetes encodes it — avoids
hand-encoding and is the cleaner option):
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
stringData:
  DB_USER: appadmin
  DB_PASSWORD: s3cr3t
```

Declarative with pre-encoded `data` (you must base64 yourself):
```bash
echo -n 's3cr3t' | base64        # -> czNjcjN0
```
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  DB_PASSWORD: czNjcjN0
```

Common Secret types beyond `Opaque`: `kubernetes.io/dockerconfigjson` (registry pull
secrets, referenced via `imagePullSecrets`), `kubernetes.io/tls` (`tls.crt`/`tls.key`),
`kubernetes.io/basic-auth`, `kubernetes.io/ssh-auth`.

Reading: `kubectl describe secret` shows only key names and byte sizes (not values), but
`kubectl get secret <s> -o yaml` exposes the base64 `data`, and anyone with that can
decode it:
```bash
echo 'czNjcjN0' | base64 --decode
```
**base64 is encoding, not encryption.** Treat Secret access as equivalent to plaintext
access.

## Consuming as environment variables

One key → one env var:
```yaml
spec:
  containers:
    - name: app
      image: my-app:1.0
      env:
        - name: COLOR                    # env var name (your choice)
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: COLOR
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: DB_PASSWORD
```

All keys at once (env var names inherited from the keys):
```yaml
      envFrom:
        - configMapRef:
            name: app-config
        - secretRef:
            name: db-secret
        # - prefix: APP_           # optional: prefix all imported var names
        #   configMapRef: { name: app-config }
```

Notes: keys that aren't valid env var names are skipped. Env vars are captured at Pod
start — updating the ConfigMap/Secret does **not** change them in a running Pod (restart
needed). Avoid env vars for secrets where possible (they can leak via logs, `describe`,
child processes, crash dumps).

## Consuming as volume mounts

Each key becomes a file in the mount directory (filename = key, contents = value). Ideal
for whole config files.

```yaml
spec:
  containers:
    - name: app
      image: my-app:1.0
      volumeMounts:
        - name: config-vol
          mountPath: /etc/conf            # /etc/conf/COLOR, /etc/conf/app.properties, ...
        - name: secret-vol
          mountPath: /etc/secrets
          readOnly: true
  volumes:
    - name: config-vol
      configMap:
        name: app-config
    - name: secret-vol
      secret:
        secretName: db-secret
```

Mount a **single** key to a specific path with `items`:
```yaml
  volumes:
    - name: config-vol
      configMap:
        name: app-config
        items:
          - key: app.properties
            path: application.properties   # mounts only this key, renamed
```

**subPath** — mount one key as a single file into a directory that has other files,
without masking the whole directory:
```yaml
      volumeMounts:
        - name: config-vol
          mountPath: /etc/nginx/nginx.conf   # a file, not a dir
          subPath: nginx.conf                # the key inside the ConfigMap
```
Caveat: `subPath` mounts do **not** receive live updates when the ConfigMap/Secret changes
(whole-directory mounts do, eventually, via kubelet sync).

## Projected volumes

Combine multiple sources (ConfigMaps, Secrets, downward API, service-account tokens) into
one directory:
```yaml
  volumes:
    - name: combined
      projected:
        sources:
          - configMap:
              name: app-config
          - secret:
              name: db-secret
          - downwardAPI:
              items:
                - path: pod-name
                  fieldRef: { fieldPath: metadata.name }
```

## Immutability

Mark a ConfigMap or Secret immutable to prevent accidental edits and reduce API-server
load (kubelet stops watching it):
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
immutable: true
data:
  COLOR: "blue"
```
Once immutable, you cannot change `data` — delete and recreate (and recreate any Pods that
must pick up new values).

## Update behavior & gotchas

- **Env vars / envFrom**: never updated live; the Pod must be recreated.
- **Volume mounts (whole dir)**: updated automatically after a sync delay (typically up to
  ~1 min); the app must re-read the file. `subPath` mounts are **not** updated.
- Deleting a ConfigMap/Secret in use doesn't disturb running Pods, but a subsequent Pod
  (re)create will fail to mount until it's recreated.
- A referenced ConfigMap/Secret must exist at Pod start unless marked `optional: true` on
  the `configMapKeyRef`/`secretKeyRef`/volume source.

## Security guidance

- Use RBAC to restrict who can read Secrets (read = plaintext access).
- Prefer volume-mounted Secrets over env vars for sensitive values.
- Enable encryption at rest (KMS provider) for Secrets (and optionally ConfigMaps) in
  etcd; base64 alone is not protection.
- Rotate secrets; never hardcode them in images or commit them to source control.
- For advanced needs, integrate an external secret manager (HashiCorp Vault, cloud secret
  managers) via operators/CSI drivers.
