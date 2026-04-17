# fluxcd-policyctl

**Flux CD Image Policy Control**

A web-based tool for managing Flux CD ImagePolicies. Non-technical users select container image versions through a friendly UI; Flux handles the actual deployment through its GitOps reconciliation loop.

ImageRepositories (the image sources) are managed via GitOps. This tool only controls ImagePolicies (the version selection).

## Features

- **Full ImagePolicy CRUD** — Create, read, update, and delete Flux CD ImagePolicy resources from the browser
- **Four access modes** — local (kubeconfig), cluster-wide, namespace list, or single namespace
- **Multi-cluster support** — Manage policies across multiple Kubernetes clusters in local mode
- **In-cluster deployment** — Run as a pod with Kustomize overlays or Helm chart
- **Optional OIDC authentication** — Integrate with any OpenID Connect provider
- **Dark/light mode** — Theme toggle with dark as default
- **Embedded SPA** — Next.js frontend compiled into a single Go binary
- **No runtime dependencies** — One statically-linked executable

## Quick Start

### Local (uses your kubeconfig)

```bash
# Install script
curl -sSL https://raw.githubusercontent.com/nikogura/fluxcd-policyctl/main/install.sh | sh

# Or download from releases
# https://github.com/nikogura/fluxcd-policyctl/releases

# Or go install
go install github.com/nikogura/fluxcd-policyctl@latest

# Run
fluxcd-policyctl
```

Then open http://localhost:9999.

### Kubernetes (Kustomize)

```bash
# Single namespace (own namespace only)
kubectl apply -k kubernetes/overlays/namespace

# Full cluster access
kubectl apply -k kubernetes/overlays/cluster

# Specific namespaces
# Edit kubernetes/overlays/namespaces/config-patch.yaml first
kubectl apply -k kubernetes/overlays/namespaces
```

### Kubernetes (Helm)

```bash
# Single namespace (default)
helm install policyctl charts/fluxcd-policyctl -n flux-system

# Full cluster access
helm install policyctl charts/fluxcd-policyctl -n flux-system --set accessMode=cluster

# Specific namespaces
helm install policyctl charts/fluxcd-policyctl -n flux-system \
  --set accessMode=namespaces \
  --set allowedNamespaces="dev-01\,stage-01"
```

### Docker

```bash
docker run --rm -p 9999:9999 \
  -v ~/.kube/config:/home/policyctl/.kube/config:ro \
  ghcr.io/nikogura/fluxcd-policyctl:latest
```

## Access Modes

| Mode | Flag / Env | Description | Cluster Selector | Namespace Selector |
|------|-----------|-------------|-----------------|-------------------|
| **local** | `--access-mode=local` / `POLICYCTL_ACCESS_MODE=local` | Uses kubeconfig. Multi-cluster. | Dropdown | Dropdown (all) |
| **cluster** | `--access-mode=cluster` / `POLICYCTL_ACCESS_MODE=cluster` | In-cluster. All namespaces. | Hidden | Dropdown (all) |
| **namespaces** | `--access-mode=namespaces` / `POLICYCTL_ACCESS_MODE=namespaces` | In-cluster. Allowed namespaces only. | Hidden | Dropdown (filtered) |
| **namespace** | `--access-mode=namespace` / `POLICYCTL_ACCESS_MODE=namespace` | In-cluster. Own namespace only. | Hidden | Static label |

When running locally (`--access-mode=local`, the default), the tool uses your kubeconfig and can access any cluster and namespace you have access to.

When running in a Kubernetes cluster, the access mode controls what the tool can see and modify. RBAC is enforced at the Kubernetes level — the tool's ServiceAccount only has access to what the Role/ClusterRole grants.

For the `namespaces` mode, set the allowed namespace list:

```bash
# Via flag
fluxcd-policyctl --access-mode=namespaces --allowed-namespaces=dev-01,stage-01

# Via env var
POLICYCTL_ACCESS_MODE=namespaces POLICYCTL_NAMESPACES=dev-01,stage-01 fluxcd-policyctl
```

## Configuration

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `--bind-address`, `-b` | — | `0.0.0.0:9999` | Listen address |
| `--namespace`, `-n` | — | `flux-system` | Default namespace for Flux resources |
| `--access-mode`, `-m` | `POLICYCTL_ACCESS_MODE` | `local` | Access mode (local/cluster/namespaces/namespace) |
| `--allowed-namespaces` | `POLICYCTL_NAMESPACES` | — | Comma-separated namespace list (namespaces mode) |
| `--oidc-issuer` | — | — | OIDC issuer URL (enables auth when set) |
| `--oidc-audience` | — | — | OIDC audience for token validation |
| `--oidc-groups` | — | — | Comma-separated allowed OIDC groups |
| `--verbose`, `-v` | — | `false` | Verbose logging |
| `--log-level`, `-l` | — | `info` | Log level (debug/info/warn/error) |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/config` | Access mode and constraints (for frontend) |
| `GET` | `/api/user` | Current user (OIDC) or 204 |
| `GET` | `/api/clusters` | List kubeconfig contexts (local mode) |
| `GET` | `/api/namespaces?cluster={ctx}` | List namespaces |
| `GET` | `/api/policies?cluster={ctx}&namespace={ns}` | List ImagePolicies |
| `POST` | `/api/policies?cluster={ctx}` | Create ImagePolicy |
| `GET` | `/api/policies/{namespace}/{name}?cluster={ctx}` | Get single policy |
| `PUT` | `/api/policies/{namespace}/{name}?cluster={ctx}` | Update semver range |
| `DELETE` | `/api/policies/{namespace}/{name}?cluster={ctx}` | Delete policy |

## Kubernetes Deployment

### RBAC Requirements

| Access Mode | RBAC | Resources |
|-------------|------|-----------|
| **cluster** | ClusterRole | imagepolicies (CRUD), imagerepositories (read), namespaces (list) |
| **namespaces** | Role per namespace | imagepolicies (CRUD), imagerepositories (read) |
| **namespace** | Role in own namespace | imagepolicies (CRUD), imagerepositories (read) |

### Kustomize

```
kubernetes/
  base/                   # Deployment, Service, ServiceAccount
  overlays/
    cluster/              # ClusterRole for full access
    namespaces/           # Role templates for specific namespaces
    namespace/            # Role for own namespace only
```

### Helm Chart

```bash
helm install policyctl charts/fluxcd-policyctl \
  -n flux-system \
  --set accessMode=cluster \
  --set oidc.enabled=true \
  --set oidc.issuerUrl=https://dex.example.com \
  --set oidc.audience=fluxcd-policyctl
```

See `charts/fluxcd-policyctl/values.yaml` for all configurable values.

## Architecture

```
fluxcd-policyctl/
├── cmd/                    # Cobra CLI
├── pkg/
│   ├── policyctl/          # HTTP server, handlers, K8s client, OIDC auth
│   └── ui/                 # Next.js app + go:embed
├── test/                   # Go tests
├── kubernetes/             # Kustomize manifests
├── charts/                 # Helm chart
├── main.go
├── Makefile
└── Dockerfile
```

The Next.js frontend is compiled to static files at build time and embedded into the binary via `//go:embed`. The Go HTTP server serves the API on `/api/` and the SPA on all other routes. No Gin, no gorilla/mux — pure `net/http` stdlib.

## Development

### Prerequisites

- Go 1.25+
- Node.js 18+
- `namedreturns` (`go install github.com/nikogura/namedreturns@latest`)
- `golangci-lint` v2 (`go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`)

### Workflow

```bash
# Lint first
make lint

# Then test
make test

# Build (UI + binary)
make build

# Run locally
./fluxcd-policyctl
```

| Target | Description |
|--------|-------------|
| `make build` | Build UI and Go binary |
| `make build-ui` | Build Next.js frontend only |
| `make build-go` | Build Go binary only |
| `make lint` | namedreturns + golangci-lint + ESLint |
| `make test` | Run Go tests |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build Docker image |

## Security

When deployed beyond localhost, enable OIDC:

```bash
fluxcd-policyctl \
  --oidc-issuer=https://dex.example.com \
  --oidc-audience=fluxcd-policyctl \
  --oidc-groups=engineering,ops
```

Without OIDC, all endpoints are open — suitable for local development or environments with network-level access control.

In-cluster deployments should use the most restrictive access mode appropriate:
- **namespace** for team-specific instances (one per team/environment)
- **namespaces** for shared instances with scoped access
- **cluster** only when broad access is required and OIDC is enabled

## License

Apache 2.0. See [LICENSE](LICENSE).

---

*Built by [Nik Ogura](https://nikogura.com) at [KATN Solutions](https://katn-solutions.io).*
