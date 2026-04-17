"use client";

import { AlertTriangle, Plus, RefreshCw } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

import { ClusterSelector } from "@/components/ClusterSelector";
import { CreatePolicyDialog } from "@/components/CreatePolicyDialog";
import { NamespaceSelector } from "@/components/NamespaceSelector";
import { PolicyTable } from "@/components/PolicyTable";
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
    <div className="min-h-screen bg-gray-50">
      <header className="border-b border-gray-200 bg-white px-6 py-4">
        <h1 className="text-xl font-bold text-gray-900">flux-policyctl</h1>
        <p className="text-sm text-gray-500">Flux CD Image Policy Control</p>
      </header>

      <main className="mx-auto max-w-7xl px-6 py-6">
        <div className="mb-6 flex flex-wrap items-center gap-4">
          <ClusterSelector
            selectedCluster={cluster}
            onClusterChange={handleClusterChange}
          />
          <NamespaceSelector
            cluster={cluster}
            selectedNamespace={namespace}
            onNamespaceChange={handleNamespaceChange}
          />

          <div className="ml-auto flex items-center gap-3">
            <span className="text-xs text-gray-400">
              Last updated: {formatLastUpdated(lastUpdated)}
            </span>
            <button
              type="button"
              onClick={handleRefresh}
              disabled={loading}
              className="rounded border border-gray-300 bg-white p-2 text-gray-600 hover:bg-gray-50 disabled:opacity-50"
              title="Refresh policies"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : null}`} />
            </button>
            <button
              type="button"
              onClick={(): void => setShowCreateDialog(true)}
              disabled={!cluster || !namespace}
              className="flex items-center gap-1.5 rounded bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:bg-blue-400"
            >
              <Plus className="h-4 w-4" />
              Create Policy
            </button>
          </div>
        </div>

        {isProductionNamespace && (
          <div className="mb-6 flex items-center gap-2 rounded border border-amber-400 bg-amber-50 px-4 py-3 text-sm text-amber-800">
            <AlertTriangle className="h-4 w-4 shrink-0" />
            <span>
              You are viewing a production namespace. Changes here will affect live workloads.
            </span>
          </div>
        )}

        {error && (
          <div className="mb-6 rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {loading && policies.length === 0 ? (
          <div className="flex items-center justify-center py-12">
            <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
            <span className="ml-2 text-sm text-gray-500">Loading policies...</span>
          </div>
        ) : cluster && namespace ? (
          <PolicyTable
            policies={policies}
            cluster={cluster}
            onPolicyUpdated={handlePolicyUpdated}
          />
        ) : (
          <div className="rounded border border-gray-200 bg-gray-50 p-8 text-center text-sm text-gray-500">
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
