# fluxcd-policyctl

**Flux CD Image Policy Control**

A web-based tool for managing Flux CD ImagePolicies. Non-technical users select
container image versions through a friendly UI; Flux handles the actual
deployment through its GitOps reconciliation loop.

## Features

- **Full ImagePolicy CRUD** -- Create, read, update, and delete Flux CD
  ImagePolicy resources directly from the browser.
- **Multi-cluster support** -- Manage policies across multiple Kubernetes
  clusters from a single instance.
- **Optional OIDC authentication** -- Integrate with any OpenID Connect
  provider to control who can modify policies.
- **Embedded SPA** -- A Next.js frontend is compiled into the Go binary at
  build time. No separate web server required.
- **Single binary** -- One statically-linked executable. No runtime
  dependencies beyond a kubeconfig.

## Quick Start

### Install script (Linux / macOS)

```bash
curl -sSL https://raw.githubusercontent.com/nikogura/fluxcd-policyctl/main/install.sh | sh
```

### Download a release binary

Grab the latest release from the
[Releases](https://github.com/nikogura/fluxcd-policyctl/releases) page, make it
executable, and run it.

### Docker

```bash
docker run --rm -p 9999:9999 \
  -v ~/.kube/config:/home/policyctl/.kube/config:ro \
  ghcr.io/nikogura/fluxcd-policyctl:latest
```

### Go install

```bash
go install github.com/nikogura/fluxcd-policyctl@latest
```

Then open [http://localhost:9999](http://localhost:9999) in your browser.

## Configuration

All configuration is through CLI flags.

| Flag | Default | Description |
|------|---------|-------------|
| `--bind-address` | `:9999` | Address and port to listen on |
| `--kubeconfig` | `~/.kube/config` | Path to kubeconfig file |
| `--context` | (current context) | Kubernetes context to use |
| `--oidc-issuer` | (none) | OIDC issuer URL (enables authentication) |
| `--oidc-client-id` | (none) | OIDC client ID |
| `--oidc-client-secret` | (none) | OIDC client secret |
| `--oidc-redirect-url` | (none) | OIDC redirect URL |
| `--log-level` | `info` | Log level (debug, info, warn, error) |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/policies` | List all ImagePolicies |
| `GET` | `/api/v1/policies/:namespace/:name` | Get a single ImagePolicy |
| `POST` | `/api/v1/policies` | Create an ImagePolicy |
| `PUT` | `/api/v1/policies/:namespace/:name` | Update an ImagePolicy |
| `DELETE` | `/api/v1/policies/:namespace/:name` | Delete an ImagePolicy |
| `GET` | `/api/v1/clusters` | List configured clusters |
| `GET` | `/api/v1/health` | Health check |
| `GET` | `/` | Serve the embedded UI |

## Architecture

fluxcd-policyctl is a single Go binary with an embedded Next.js frontend.

```
fluxcd-policyctl
├── cmd/            # Cobra CLI commands
├── pkg/
│   ├── policyctl/  # Core business logic, Kubernetes client
│   └── ui/         # Next.js app + Go embed directive
├── test/           # Integration / E2E tests
├── main.go         # Entry point
└── Makefile
```

At build time the Next.js app is compiled to static files under `pkg/ui/dist`,
which are embedded into the binary via `//go:embed`. The Go HTTP server serves
the API on `/api/` and the SPA on all other routes.

## Development

### Prerequisites

- Go 1.24+
- Node.js 18+
- `golangci-lint`
- `namedreturns` (`go install github.com/nikogura/namedreturns@latest`)

### Make targets

| Target | Description |
|--------|-------------|
| `make build` | Build the UI and Go binary |
| `make build-ui` | Build the Next.js frontend only |
| `make build-go` | Build the Go binary only |
| `make test` | Run all Go tests |
| `make lint` | Run namedreturns, golangci-lint, and ESLint |
| `make clean` | Remove build artifacts |
| `make docker-build` | Build the Docker image |

### Typical workflow

```bash
# Install tools
go install github.com/nikogura/namedreturns@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linters
make lint

# Run tests
make test

# Build
make build
```

## Security

### When to enable OIDC

If fluxcd-policyctl is exposed beyond localhost -- for example behind an ingress
or shared on a team network -- enable OIDC authentication so that only
authorized users can modify ImagePolicies.

```bash
fluxcd-policyctl \
  --oidc-issuer=https://accounts.example.com \
  --oidc-client-id=fluxcd-policyctl \
  --oidc-client-secret=<secret> \
  --oidc-redirect-url=https://policyctl.example.com/callback
```

When OIDC is not configured the server runs without authentication, suitable
for local development or environments where network-level access control is
sufficient.

### Kubernetes RBAC

fluxcd-policyctl requires a ServiceAccount (or kubeconfig identity) with
permissions to list, get, create, update, and delete ImagePolicy resources in
the target namespaces.

## License

Apache 2.0. See [LICENSE](LICENSE) for the full text.
