#!/usr/bin/env bash
# Deploy the opa-hook backend (OPA + the real mcp_auth.rego oe_decision rule)
# to a running minikube cluster. Refuses to run against any other kubectl
# context on purpose — this is a local-dev tool scoped to minikube, not a path
# to any other cluster.
set -euo pipefail
cd "$(dirname "$0")"

CTX="$(kubectl config current-context 2>/dev/null || true)"
if [[ "$CTX" != "minikube" ]]; then
  echo "Refusing to deploy: current kubectl context is '${CTX:-<none>}', not 'minikube'." >&2
  echo "Run: minikube start   then: kubectl config use-context minikube" >&2
  exit 1
fi

POLICY_FILE="${AGENTIC_SENTRY_POLICIES:-../../../../Agentic-Sentry/mcp-policies/policies/mcp_auth.rego}"
if [[ ! -f "$POLICY_FILE" ]]; then
  echo "mcp_auth.rego not found at $POLICY_FILE" >&2
  echo "Clone Agentic-Sentry as a sibling of agentic-harness, or set" >&2
  echo "AGENTIC_SENTRY_POLICIES to the mcp_auth.rego file path." >&2
  exit 1
fi

kubectl create configmap opa-mcp-policies \
  --from-file="mcp_auth.rego=$POLICY_FILE" \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl apply -f opa.yaml
kubectl rollout status deployment/opa-hook-backend --timeout=60s

echo
echo "Deployed. In a separate terminal, keep this running while wiring hooks:"
echo "  kubectl port-forward svc/opa-hook-backend 8181:8181"
echo
echo "Then verify oe_decision is queryable:"
echo "  curl -s http://127.0.0.1:8181/v1/data/mcp/auth/oe_decision \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"input\":{\"server_name\":\"native\",\"tool_name\":\"Bash\",\"arguments\":{},\"groups\":[]}}'"
