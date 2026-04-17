"use client";

import { Plus, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

import { ClusterSelector } from "@/components/ClusterSelector";
import { CreatePolicyDialog } from "@/components/CreatePolicyDialog";
import { NamespaceSelector } from "@/components/NamespaceSelector";
import { PolicyTable } from "@/components/PolicyTable";
import { ThemeToggle } from "@/components/ThemeToggle";
import { fetchPolicies } from "@/lib/api";
import type { PolicyView } from "@/types";

const REFRESH_INTERVAL_MS = 30000;

export default function DashboardPage(): React.ReactElement {
  const [cluster, setCluster] = useState<string | null>(null);
  const [namespace, setNamespace] = useState<string | null>(null);
  const [policies, setPolicies] = useState<readonly PolicyView[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const loadPolicies = useCallback(async (): Promise<void> => {
    if (!cluster || !namespace) return;

    try {
      setLoading(true);
      setError(null);
      const result = await fetchPolicies({ cluster, namespace });
      setPolicies(result);
      setLastUpdated(new Date());
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to load policies";
      setError(message);
    } finally {
      setLoading(false);
    }
  }, [cluster, namespace]);

  useEffect(() => {
    void loadPolicies();
  }, [loadPolicies]);

  useEffect(() => {
    if (!cluster || !namespace) return;

    intervalRef.current = setInterval(() => {
      void loadPolicies();
    }, REFRESH_INTERVAL_MS);

    return (): void => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [cluster, namespace, loadPolicies]);

  const handleClusterChange = (newCluster: string): void => {
    setCluster(newCluster);
    setNamespace(null);
    setPolicies([]);
  };

  const handleNamespaceChange = (newNamespace: string): void => {
    setNamespace(newNamespace);
  };

  const handleRefresh = (): void => {
    void loadPolicies();
  };

  const handlePolicyUpdated = (): void => {
    void loadPolicies();
  };

  const handlePolicyCreated = (): void => {
    void loadPolicies();
  };

  const isProductionNamespace = namespace ? namespace.includes("prod") : false;

  const formatLastUpdated = (date: Date | null): string => {
    if (!date) return "Never";
    return date.toLocaleTimeString();
  };

  return (
    <div style={{ minHeight: "100vh", backgroundColor: "var(--bg-secondary)" }}>
      <header style={{
        backgroundColor: "var(--bg-color)",
        borderBottom: "1px solid var(--border-color)",
        padding: "1rem",
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
      }}>
        <div>
          <div style={{ fontWeight: "bold", fontSize: "1.2rem", color: "var(--text-color)" }}>
            flux-policyctl
          </div>
          <div style={{ fontSize: "0.8rem", color: "var(--text-muted)" }}>
            Flux CD Image Policy Control
          </div>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: "0.75rem" }}>
          <ClusterSelector
            selectedCluster={cluster}
            onClusterChange={handleClusterChange}
          />
          <NamespaceSelector
            cluster={cluster}
            selectedNamespace={namespace}
            onNamespaceChange={handleNamespaceChange}
          />
          <ThemeToggle />
        </div>
      </header>

      <main style={{ padding: "2rem", maxWidth: "1280px", margin: "0 auto" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "1.5rem" }}>
          <span style={{ fontSize: "12px", color: "var(--text-muted)" }}>
            Last updated: {formatLastUpdated(lastUpdated)}
          </span>
          <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: "12px" }}>
            <button
              type="button"
              onClick={handleRefresh}
              disabled={loading}
              style={{
                padding: "8px",
                border: "1px solid var(--border-color)",
                borderRadius: "4px",
                backgroundColor: "var(--bg-color)",
                color: "var(--text-muted)",
                cursor: loading ? "not-allowed" : "pointer",
                opacity: loading ? 0.5 : 1,
                display: "flex",
                alignItems: "center",
              }}
              title="Refresh policies"
            >
              <RefreshCw size={16} />
            </button>
            <button
              type="button"
              onClick={(): void => setShowCreateDialog(true)}
              disabled={!cluster || !namespace}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "6px",
                padding: "8px 16px",
                border: "none",
                borderRadius: "4px",
                backgroundColor: "var(--button-bg)",
                color: "var(--button-text)",
                fontSize: "14px",
                fontWeight: 600,
                cursor: !cluster || !namespace ? "not-allowed" : "pointer",
                opacity: !cluster || !namespace ? 0.5 : 1,
              }}
            >
              <Plus size={16} />
              Create Policy
            </button>
          </div>
        </div>

        {isProductionNamespace && (
          <div style={{
            backgroundColor: "var(--warning-bg)",
            color: "white",
            fontWeight: "bold",
            textAlign: "center",
            borderRadius: "4px",
            padding: "0.5rem",
            marginBottom: "1rem",
          }}>
            You are viewing a production namespace. Changes here will affect live workloads.
          </div>
        )}

        {error && (
          <div style={{
            marginBottom: "1rem",
            padding: "8px 12px",
            borderRadius: "4px",
            backgroundColor: "var(--error-color)",
            color: "white",
            fontSize: "14px",
          }}>
            {error}
          </div>
        )}

        {loading && policies.length === 0 ? (
          <div style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            padding: "3rem",
            color: "var(--text-muted)",
          }}>
            <RefreshCw size={20} />
            <span style={{ marginLeft: "8px", fontSize: "14px" }}>Loading policies...</span>
          </div>
        ) : cluster && namespace ? (
          <PolicyTable
            policies={policies}
            cluster={cluster}
            onPolicyUpdated={handlePolicyUpdated}
          />
        ) : (
          <div style={{
            padding: "32px",
            textAlign: "center",
            fontSize: "14px",
            color: "var(--text-muted)",
            backgroundColor: "var(--bg-color)",
            borderRadius: "8px",
            border: "1px solid var(--border-color)",
          }}>
            Select a cluster and namespace to view image policies.
          </div>
        )}
      </main>

      {cluster && namespace && (
        <CreatePolicyDialog
          cluster={cluster}
          namespace={namespace}
          isOpen={showCreateDialog}
          onClose={(): void => setShowCreateDialog(false)}
          onCreated={handlePolicyCreated}
        />
      )}
    </div>
  );
}
