# Namespaces Overlay

This overlay configures fluxcd-policyctl to operate on a specific list of namespaces.

## Setup

### 1. Edit the namespace list

In `config-patch.yaml`, set `POLICYCTL_NAMESPACES` to a comma-separated list of namespaces you want policyctl to manage:

```yaml
- name: POLICYCTL_NAMESPACES
  value: "dev-01,stage-01,prod-01"
```

### 2. Create RBAC in each target namespace

The `role.yaml` and `rolebinding.yaml` files are templates. You must create a Role and RoleBinding in **each** namespace that policyctl needs access to.

For each namespace, copy the templates and replace `CHANGEME` with the namespace name:

```bash
# Example: grant access to the dev-01 namespace
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: fluxcd-policyctl
  namespace: dev-01
rules:
  - apiGroups: ["image.toolkit.fluxcd.io"]
    resources: ["imagepolicies"]
    verbs: ["get", "list", "create", "update", "patch", "delete"]
  - apiGroups: ["image.toolkit.fluxcd.io"]
    resources: ["imagerepositories"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: fluxcd-policyctl
  namespace: dev-01
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: fluxcd-policyctl
subjects:
  - kind: ServiceAccount
    name: fluxcd-policyctl
    namespace: flux-system
EOF
```

Repeat for each namespace in your `POLICYCTL_NAMESPACES` list.

### 3. Deploy

```bash
kubectl apply -k kubernetes/overlays/namespaces/
```

If managing RBAC via GitOps, commit separate Role/RoleBinding manifests for each target namespace alongside this overlay.
