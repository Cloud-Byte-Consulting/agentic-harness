# Single-Host Networking & Docker Compose

Connecting, isolating, and exposing containers on one host, then orchestrating multi-service stacks locally
with Compose.

## Contents
- [Container Network Model](#container-network-model)
- [Network types](#network-types)
- [Network firewalling (least-privilege)](#network-firewalling-least-privilege)
- [The bridge network](#the-bridge-network)
- [Custom bridge networks and DNS](#custom-bridge-networks-and-dns)
- [host and none](#host-and-none)
- [Sharing a network namespace](#sharing-a-network-namespace)
- [Publishing ports](#publishing-ports)
- [Reverse proxy (Traefik)](#reverse-proxy-traefik)
- [Docker Compose](#docker-compose)
- [Compose vs an orchestrator](#compose-vs-an-orchestrator)

## Container Network Model

Docker's CNM has three elements:
- **Sandbox** — a container's isolated network stack (its network namespace). By default no inbound traffic.
- **Endpoint** — how a sandbox plugs into a network. A sandbox can have zero, one, or many endpoints (so a
  container can join multiple networks).
- **Network** — the pathway carrying packets between endpoints; local (one host) or global (across a cluster).

The CNM is just a model; implementations plug in via drivers.

## Network types

| Driver | Scope | Summary |
|---|---|---|
| **bridge** | Local | Default single-host network (Linux bridge). |
| **host** | Local | Container shares the host's network stack (no isolation). |
| **none** | Local | No networking (only loopback). |
| macvlan / ipvlan | Local | Container gets an L2/L3 identity on the physical network. |
| overlay | Global | Multi-host (VXLAN) for Swarm/clusters. |
| Calico / Weave / Contiv | Global | Third-party cluster networking & policy. |

This skill covers the **local** drivers. Global/overlay and CNI plugins (Calico, Cilium) belong to the
**kubernetes-networking** skill.

## Network firewalling (least-privilege)

By default Docker creates SDNs that are isolated: containers on the **same** network talk freely; containers on
**different** networks can't reach each other unless a container is deliberately attached to both. This is a
built-in firewall — design each network to contain only services that must talk, so a compromise of one has a
small blast radius. Example: put `webAPI` + `productCatalog` on a front network and `productCatalog` + `database`
on a back network; `productCatalog` bridges both, but a breached `webAPI` still can't reach `database` directly.
Avoid the temptation to dump everything on the default bridge.

## The bridge network

When `dockerd` starts it creates the `docker0` Linux bridge and a default network named `bridge` (subnet
`172.17.0.0/16`, gateway `172.17.0.1`). Containers started without `--network` attach here and get the next
free IP (`172.17.0.2`, `…3`, …). Egress is allowed; **ingress is blocked** unless you publish a port.

```bash
docker network ls
docker network inspect bridge          # IPAM subnet/gateway, attached containers
```
To avoid corporate/VPN subnet clashes, change the default in `/etc/docker/daemon.json`:
```json
{ "bip": "192.168.100.1/24" }
```

## Custom bridge networks and DNS

**Always prefer user-defined bridge networks** over the default. Their big advantage: Docker's embedded DNS lets
containers resolve each other **by name** (the default bridge does not).

```bash
docker network create --driver bridge sample-net
docker network create --driver bridge --subnet 10.1.0.0/16 test-net      # custom subnet
docker network create --driver bridge --opt com.docker.network.driver.mtu=1400 mynet-mtu   # MTU for VPN/overlay
docker network create --ipv6 --subnet 10.10.0.0/16 --subnet fd00:dead:beef::/48 mynet-v6   # dual-stack

docker run --name c3 --rm -d --network sample-net alpine:3.22 ping 127.0.0.1
docker run --name c4 --rm -d --network sample-net alpine:3.22 ping 127.0.0.1
docker exec -it c3 ping c4        # resolves by name via Docker DNS
```
Containers on different networks can't ping each other (name resolution fails across networks) — that's the
isolation working. Attach a container to a second network with `docker network connect <net> <c>`.

```bash
docker network rm sample-net          # fails if containers are attached
docker network prune --force          # remove all unused networks
```

## host and none

**host** (`--network host`) removes the network namespace — the container shares the host's IP and interfaces.
Use only for niche cases (ultra-low-latency, host broadcast/multicast discovery, host-network debugging).
Caveats: **Linux only** (on Docker Desktop it attaches to the *VM*, not your physical host, so you still need
`-p`); **no isolation** (a breach can sniff host traffic); **port conflicts** with host services. For business
apps, don't use host networking in production.

**none** (`--network none`) gives only a loopback interface — no external connectivity. For fully isolated
batch/offline/forensics workloads. `ip addr show eth0` inside returns "can't find device".

## Sharing a network namespace

Run a container *inside another container's* network namespace so they share IP/ports and talk over
`localhost`:
```bash
docker run --name web -d --network test-net nginx:1.29-alpine
docker run -it --rm --network container:web alpine:3.22 /bin/sh      # shares web's netns
# inside: wget localhost  -> reaches nginx
```
This is exactly how **Kubernetes Pods** group containers (shared netns + loopback). Great for sidecar debugging
without modifying the target image:
```bash
docker run --rm -it --network container:myapp nicolaka/netshoot      # curl/dig/tcpdump in myapp's netns
```
Security note: sharing a netns shares its network privileges.

## Publishing ports

The container netns and host are independent; to expose a service, **publish** a container port to a host port.

```bash
docker run --name web -P -d nginx:1.29-alpine        # -P: auto-map exposed ports to random host ports
docker container port web                            # 80/tcp -> 0.0.0.0:55000
docker run --name web2 -p 8080:80 -d nginx:1.29-alpine   # -p host:container, explicit
docker run -d -p 127.0.0.1:8080:80 my/dev-app        # bind to loopback only (don't expose on LAN)
docker run -d -p 3000:4321/udp my/app                # UDP (map TCP and UDP separately if both needed)
docker run -d -p 12000-12010:12000-12010 my/worker   # port range
```

Key distinctions and rules:
- **`EXPOSE` (image metadata) does NOT open anything** — it documents intent. Only `-p`/`-P` (or `--publish`)
  actually wires a host port and installs the NAT/forward rules (iptables/nftables).
- `0.0.0.0:port` means any host interface can reach it; prefer `127.0.0.1:` binding for local dev to avoid
  exposing dev services on Wi-Fi/LAN.
- **Two containers cannot share one host port** — give each a unique host port, or front them with one reverse
  proxy and keep app ports unpublished on a private network.
- From a container to the host on Docker Desktop, use `host.docker.internal` rather than guessing the host IP.
- Security checklist: publish only what you must; keep internal services on private user-defined bridges;
  terminate TLS at a proxy and keep app ports private in production; be careful with UDP.

## Reverse proxy (Traefik)

To present one stable public entry point while routing to multiple backends (e.g. extracting microservices from
a monolith without changing client URLs), put a reverse proxy in front. Traefik reads routing rules from
container labels:

```bash
docker run --rm -d --name catalog \
  --label traefik.enable=true \
  --label traefik.port=3000 \
  --label traefik.priority=10 \
  --label traefik.http.routers.catalog.rule='Host(`acme.com`) && PathPrefix(`/catalog`)' \
  acme/catalog:1.0

docker run -d --name traefik -p 8080:8080 -p 80:80 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  traefik:v2.0 --api.insecure=true --providers.docker=true \
  --entrypoints.web.address=:80 --providers.docker.exposedbydefault=false
```
Now `/catalog` requests route to the new service while everything else hits the monolith — TLS termination,
load balancing, and incremental migration all from a single entry point. In Kubernetes this role is played by
Ingress/Gateway API (see **kubernetes-networking**).

## Docker Compose

Compose declaratively defines and runs a multi-service app on one host (dev, CI, demos). It's now part of the
Docker CLI: use `docker compose …` (the standalone `docker-compose` is deprecated). Files are `compose.yaml`
or `docker-compose.yml`; the top-level `version:` field is obsolete (Compose follows the Compose Spec). Builds
use BuildKit by default.

A **service** is a definition of how to run one or more identical containers. Resources are prefixed with the
project name (parent folder by default; override with `--project-name`/`-p`).

Minimal multi-service file (DB + admin UI) with a healthcheck dependency:
```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: dockeruser
      POSTGRES_PASSWORD: dockerpass
      POSTGRES_DB: pets
    volumes:
      - pg-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "dockeruser"]
      interval: 10s
      timeout: 5s
      retries: 5
  pgadmin:
    image: dpage/pgadmin4
    ports:
      - 5050:80
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@acme.com
      PGADMIN_DEFAULT_PASSWORD: admin
    depends_on:
      db:
        condition: service_healthy

volumes:
  pg-data:
```

Building your own image from a service (the `build:` key points at a Dockerfile context):
```yaml
services:
  web:
    build: ./web                 # or: build: { context: ./web, dockerfile: Dockerfile.dev }
    image: acme/web:1.0
    ports:
      - 3000:3000
    depends_on:
      db:
        condition: service_healthy
```

Common commands:
```bash
docker compose up                 # create networks/volumes/containers, run in foreground
docker compose up -d              # detached
docker compose up db --detach     # start a single service
docker compose build [web]        # build images
docker compose ps                 # list this project's services
docker compose logs -f web        # follow logs
docker compose push               # push built images (must be logged in)
docker compose down -v            # stop and remove containers, networks, AND volumes (-v destroys data)
```

**Scaling** a stateless service. If a service hard-maps a host port (`3000:3000`), only one replica can start
(port collision). Map only the container port so Docker assigns ephemeral host ports:
```yaml
ports:
  - 3000          # container port only -> dynamic host ports
```
```bash
docker compose up -d --scale web=3
docker compose ps                 # web-1/2/3 on different host ports
```

**Overrides** layer environment-specific settings on a base file. `compose.yaml` + `compose.override.yaml`
merge automatically on `docker compose up`; otherwise name them explicitly:
```bash
docker compose -f compose.base.yaml -f compose.ci.yaml up -d --build
```

**`include`** (Compose v2.20+) modularizes large stacks — pull in entire Compose files, each keeping its own
relative paths and env. Unlike overrides (which silently merge same-named services), `include` **errors on name
clashes**:
```yaml
include:
  - ./db/compose.yaml
  - ./web/compose.yaml
```

## Compose vs an orchestrator

| Aspect | Docker Compose | Orchestrator (Kubernetes) |
|---|---|---|
| Scope | Single host | Multi-node cluster |
| Use | Local dev, CI, demos | Production, HA, scale |
| Scaling | Manual, host-limited | Declarative, automated |
| State | Starts services; **no self-healing** | Maintains desired state, reschedules on failure |
| Upgrades | Stop/restart | Rolling updates, zero-downtime, rollback |

Use Compose on one host; when you need multi-node HA, self-healing, and rolling upgrades, move to Kubernetes —
covered by the **kubernetes-workloads** and **kubernetes-networking** skills.
