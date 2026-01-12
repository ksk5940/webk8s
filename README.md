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

#Using helm we can deploy

helm upgrade webk8s . -n kube-system -f values.yaml

kubectl -n kube-system rollout restart deployment webk8s

kubectl -n kube-system port-forward svc/webk8s 8080:80


helm list -n kube-system
helm uninstall webk8s -n kube-system