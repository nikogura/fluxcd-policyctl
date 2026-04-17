"use client";

import { useEffect, useState } from "react";

import { fetchClusters } from "@/lib/api";
import type { ClusterInfo } from "@/types";

interface ClusterSelectorProps {
  readonly selectedCluster: string | null;
  readonly onClusterChange: (cluster: string) => void;
}

export function ClusterSelector({ selectedCluster, onClusterChange }: ClusterSelectorProps): React.ReactElement {
  const [clusters, setClusters] = useState<readonly ClusterInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    const loadClusters = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);
        const result = await fetchClusters();
        if (cancelled) return;
        setClusters(result);

        if (!selectedCluster) {
          const currentCluster = result.find((c) => c.current);
          if (currentCluster) {
            onClusterChange(currentCluster.name);
          } else if (result.length > 0) {
            onClusterChange(result[0].name);
          }
        }
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : "Failed to load clusters";
        setError(message);
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void loadClusters();

    return (): void => {
      cancelled = true;
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  if (loading) {
    return (
      <select disabled className="rounded border border-gray-300 bg-gray-50 px-3 py-2 text-sm">
        <option>Loading clusters...</option>
      </select>
    );
  }

  if (error) {
    return (
      <select disabled className="rounded border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-600">
        <option>Error loading clusters</option>
      </select>
    );
  }

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>): void => {
    onClusterChange(e.target.value);
  };

  return (
    <label className="flex items-center gap-2 text-sm font-medium text-gray-700">
      Cluster
      <select
        value={selectedCluster ?? undefined}
        onChange={handleChange}
        className="rounded border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      >
        {clusters.map((cluster) => (
          <option key={cluster.name} value={cluster.name}>
            {cluster.name}{cluster.current ? " (current)" : null}
          </option>
        ))}
      </select>
    </label>
  );
}
