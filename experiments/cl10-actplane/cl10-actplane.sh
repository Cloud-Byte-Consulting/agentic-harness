#!/usr/bin/env bash
# CLO-10 — ActPlane indirect-execution catch (live demo).
# Proves ActPlane's KERNEL enforcement blocks a denied `git` reached DIRECTLY and
# via a wrapper SCRIPT — the indirect path our tool-call hooks (opa_delegate / opa-hook)
# MISS (they see `bash`, not `git`). Needs root (eBPF load) + builds ActPlane once.
#   sudo bash ~/air/workspace/scripts/cl10-actplane.sh
set -e
[ "$(id -u)" = 0 ] || { echo "Run with sudo (eBPF load needs root):  sudo bash $0"; exit 1; }
AP=/home/bittahcriminal/air/workspace/research-repos/ActPlane

echo ">> 1/4 build ActPlane (apt deps + Rust + eBPF) — first run is slow"
make -C "$AP" install
[ -f "$HOME/.cargo/env" ] && . "$HOME/.cargo/env"
command -v cargo >/dev/null 2>&1 || { [ -f /root/.cargo/env ] && . /root/.cargo/env; }
make -C "$AP"
ACTPLANE="$AP/target/release/actplane"
[ -x "$ACTPLANE" ] || { echo "build failed: $ACTPLANE missing"; exit 1; }

echo ">> 2/4 stage a demo policy (block git for the agent subtree) + a demo 'agent'"
DEMO=$(mktemp -d)
cat > "$DEMO/actplane.yaml" <<'YAML'
version: 1
policy: |
  source AGENT = exec "**/demo-agent.sh"
  rule no-git:
    block exec "git" if AGENT
    because "git blocked — the kernel catches DIRECT and SCRIPT-INDIRECT git alike"
YAML
cat > "$DEMO/demo-agent.sh" <<'SH'
#!/bin/bash
echo "[agent] (1) DIRECT   git --version:"
git --version >/dev/null 2>&1 && echo "   NOT blocked  <-- gap" || echo "   BLOCKED  <-- kernel caught it"
echo "[agent] (2) INDIRECT bash -c 'git --version'  (what tool-call hooks MISS):"
bash -c 'git --version' >/dev/null 2>&1 && echo "   NOT blocked  <-- gap" || echo "   BLOCKED  <-- kernel caught it"
SH
chmod +x "$DEMO/demo-agent.sh"

echo ">> 3/4 compile the policy (validate; no privileges needed for this step)"
( cd "$DEMO" && "$ACTPLANE" compile )

echo ">> 4/4 run the demo agent under ActPlane (eBPF enforcement)"
( cd "$DEMO" && "$ACTPLANE" run -- ./demo-agent.sh ) || true

echo
echo ">> EXPECTED: BOTH (1) direct and (2) indirect git report BLOCKED."
echo "   The indirect (script-launched) git is exactly what opa_delegate/opa-hook miss"
echo "   (they see 'bash', not 'git') — ActPlane's syscall-level enforcement closes the gap."
echo "   (cleanup: rm -rf $DEMO)"
