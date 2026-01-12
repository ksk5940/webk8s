# webk8s

Lightweight Kubernetes Web UI (single container) deployed as system pod in kube-system.

Features:
- Namespace dropdown (default selected)
- Resource dropdown (pods, deployments, rs, sts, ds, svc, jobs, cronjobs, cm, secrets)
- List resources in table
- Click Pod -> Details drawer:
  - Overview (node, ip, phase, containers, restarts)
  - Logs (stream via SSE)
  - Events
  - Metrics (CPU/Memory via metrics-server)

No pod exec feature (safe).

## Build
```bash
docker build -t webk8s:0.1.0 .
