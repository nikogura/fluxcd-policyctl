export interface PolicyView {
  readonly name: string;
  readonly namespace: string;
  readonly imageRepository: string;
  readonly imageUrl: string;
  readonly semverRange: string;
  readonly latestVersion: string;
  readonly availableVersions: readonly string[];
  readonly lastUpdated: string;
  readonly ready: boolean;
  readonly message: string;
}

export interface CreatePolicyRequest {
  readonly name: string;
  readonly namespace: string;
  readonly imageRepository: string;
  readonly semverRange: string;
}

export interface UpdatePolicyRequest {
  readonly semverRange: string;
}

export interface ClusterInfo {
  readonly name: string;
  readonly current: boolean;
}

export interface AppConfig {
  readonly accessMode: "local" | "cluster" | "namespaces" | "namespace";
  readonly inCluster: boolean;
  readonly allowedNamespaces: readonly string[] | null;
  readonly fixedNamespace: string | null;
  readonly refreshIntervalSec: number;
}
