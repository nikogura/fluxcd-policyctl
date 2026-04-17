"use client";

import { useEffect, useState } from "react";

import { fetchClusters } from "@/lib/api";
import type { ClusterInfo } from "@/types";

interface ClusterSelectorProps {
  readonly selectedCluster: string | null;
  readonly onClusterChange: (cluster: string) => void;
}

const selectStyle: React.CSSProperties = {
  padding: "8px 12px",
  border: "1px solid var(--border-color)",
  borderRadius: "4px",
  backgroundColor: "var(--bg-color)",
  color: "var(--text-color)",
  fontSize: "14px",
  cursor: "pointer",
  outline: "none",
};

const disabledSelectStyle: React.CSSProperties = {
  ...selectStyle,
  cursor: "default",
  opacity: 0.6,
};

const errorSelectStyle: React.CSSProperties = {
  ...selectStyle,
  borderColor: "var(--error-color)",
  color: "var(--error-color)",
  cursor: "default",
  opacity: 0.6,
};

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
      <select disabled style={disabledSelectStyle}>
        <option>Loading clusters...</option>
      </select>
    );
  }

  if (error) {
    return (
      <select disabled style={errorSelectStyle}>
        <option>Error loading clusters</option>
      </select>
    );
  }

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>): void => {
    onClusterChange(e.target.value);
  };

  return (
    <label style={{ display: "flex", alignItems: "center", gap: "8px", fontSize: "14px", fontWeight: 500, color: "var(--text-color)" }}>
      Cluster
      <select
        value={selectedCluster ?? undefined}
        onChange={handleChange}
        style={selectStyle}
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
