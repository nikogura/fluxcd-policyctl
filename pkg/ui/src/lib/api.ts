import type { ClusterInfo, CreatePolicyRequest, PolicyView } from "@/types";

const API_BASE = "/api";

interface ApiOptions {
  readonly method?: string;
  readonly body?: unknown;
}

async function apiCall<T>({ path, options }: { readonly path: string; readonly options?: ApiOptions }): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };

  const token = typeof window !== "undefined" ? localStorage.getItem("oidc_token") : null;
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  // eslint-disable-next-line local-rules/disallow-fetch
  const response = await fetch(`${API_BASE}${path}`, {
    method: options?.method ?? "GET",
    headers,
    body: options?.body ? JSON.stringify(options.body) : undefined,
  });

  if (response.status === 401) {
    window.location.href = "/auth/login";
    throw new Error("Unauthorized");
  }

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(errorText || `API error: ${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const data: T = await response.json();
  return data;
}

export async function fetchClusters(): Promise<readonly ClusterInfo[]> {
  return apiCall<readonly ClusterInfo[]>({ path: "/clusters" });
}

export async function fetchNamespaces({ cluster }: { readonly cluster: string }): Promise<readonly string[]> {
  return apiCall<readonly string[]>({ path: `/namespaces?cluster=${encodeURIComponent(cluster)}` });
}

export async function fetchPolicies({ cluster, namespace }: { readonly cluster: string; readonly namespace: string }): Promise<readonly PolicyView[]> {
  const result = await apiCall<{ readonly policies: readonly PolicyView[] }>({
    path: `/policies?cluster=${encodeURIComponent(cluster)}&namespace=${encodeURIComponent(namespace)}`,
  });
  return result.policies;
}

export async function createPolicy({ cluster, request }: { readonly cluster: string; readonly request: CreatePolicyRequest }): Promise<void> {
  await apiCall<void>({
    path: `/policies?cluster=${encodeURIComponent(cluster)}`,
    options: { method: "POST", body: request },
  });
}

export async function updatePolicy({ cluster, namespace, name, semverRange }: {
  readonly cluster: string;
  readonly namespace: string;
  readonly name: string;
  readonly semverRange: string;
}): Promise<void> {
  await apiCall<void>({
    path: `/policies/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}?cluster=${encodeURIComponent(cluster)}`,
    options: { method: "PUT", body: { semverRange } },
  });
}

export async function deletePolicy({ cluster, namespace, name }: {
  readonly cluster: string;
  readonly namespace: string;
  readonly name: string;
}): Promise<void> {
  await apiCall<void>({
    path: `/policies/${encodeURIComponent(namespace)}/${encodeURIComponent(name)}?cluster=${encodeURIComponent(cluster)}`,
    options: { method: "DELETE" },
  });
}
