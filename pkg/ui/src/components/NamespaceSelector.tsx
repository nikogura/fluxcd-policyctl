"use client";

import { useEffect, useState } from "react";

import { fetchNamespaces } from "@/lib/api";

interface NamespaceSelectorProps {
  readonly cluster: string | null;
  readonly selectedNamespace: string | null;
  readonly onNamespaceChange: (namespace: string) => void;
}

const DEFAULT_NAMESPACE = "flux-system";

export function NamespaceSelector({ cluster, selectedNamespace, onNamespaceChange }: NamespaceSelectorProps): React.ReactElement {
  const [namespaces, setNamespaces] = useState<readonly string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!cluster) return;

    let cancelled = false;

    const loadNamespaces = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);
        const result = await fetchNamespaces({ cluster });
        if (cancelled) return;
        setNamespaces(result);

        if (!selectedNamespace) {
          const hasDefault = result.includes(DEFAULT_NAMESPACE);
          onNamespaceChange(hasDefault ? DEFAULT_NAMESPACE : result[0]);
        }
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : "Failed to load namespaces";
        setError(message);
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    void loadNamespaces();

    return (): void => {
      cancelled = true;
    };
  }, [cluster]); // eslint-disable-line react-hooks/exhaustive-deps

  if (!cluster) {
    return (
      <select disabled className="rounded border border-gray-300 bg-gray-50 px-3 py-2 text-sm">
        <option>Select a cluster first</option>
      </select>
    );
  }

  if (loading) {
    return (
      <select disabled className="rounded border border-gray-300 bg-gray-50 px-3 py-2 text-sm">
        <option>Loading namespaces...</option>
      </select>
    );
  }

  if (error) {
    return (
      <select disabled className="rounded border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-600">
        <option>Error loading namespaces</option>
      </select>
    );
  }

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>): void => {
    onNamespaceChange(e.target.value);
  };

  return (
    <label className="flex items-center gap-2 text-sm font-medium text-gray-700">
      Namespace
      <select
        value={selectedNamespace ?? undefined}
        onChange={handleChange}
        className="rounded border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
      >
        {namespaces.map((ns) => (
          <option key={ns} value={ns}>
            {ns}
          </option>
        ))}
      </select>
    </label>
  );
}
