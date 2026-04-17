"use client";

import { useEffect, useState } from "react";

import { fetchNamespaces } from "@/lib/api";

interface NamespaceSelectorProps {
  readonly cluster: string | null;
  readonly selectedNamespace: string | null;
  readonly onNamespaceChange: (namespace: string) => void;
  readonly fixedNamespace?: string | null;
  readonly allowedNamespaces?: readonly string[] | null;
}

const DEFAULT_NAMESPACE = "flux-system";

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

const staticLabelStyle: React.CSSProperties = {
  padding: "8px 12px",
  border: "1px solid var(--border-color)",
  borderRadius: "4px",
  backgroundColor: "var(--bg-color)",
  color: "var(--text-color)",
  fontSize: "14px",
};

export function NamespaceSelector({ cluster, selectedNamespace, onNamespaceChange, fixedNamespace, allowedNamespaces }: NamespaceSelectorProps): React.ReactElement {
  const [namespaces, setNamespaces] = useState<readonly string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (fixedNamespace) {
      onNamespaceChange(fixedNamespace);
      return;
    }

    if (!cluster) return;

    let cancelled = false;

    const loadNamespaces = async (): Promise<void> => {
      try {
        setLoading(true);
        setError(null);
        const result = await fetchNamespaces({ cluster });
        if (cancelled) return;

        const filtered = allowedNamespaces
          ? result.filter((ns) => allowedNamespaces.includes(ns))
          : result;

        setNamespaces(filtered);

        if (!selectedNamespace) {
          const hasDefault = filtered.includes(DEFAULT_NAMESPACE);
          onNamespaceChange(hasDefault ? DEFAULT_NAMESPACE : filtered[0]);
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
  }, [cluster, fixedNamespace, allowedNamespaces]); // eslint-disable-line react-hooks/exhaustive-deps

  if (fixedNamespace) {
    return (
      <label style={{ display: "flex", alignItems: "center", gap: "8px", fontSize: "14px", fontWeight: 500, color: "var(--text-color)" }}>
        Namespace
        <span style={staticLabelStyle}>{fixedNamespace}</span>
      </label>
    );
  }

  if (!cluster) {
    return (
      <select disabled style={disabledSelectStyle}>
        <option>Select a cluster first</option>
      </select>
    );
  }

  if (loading) {
    return (
      <select disabled style={disabledSelectStyle}>
        <option>Loading namespaces...</option>
      </select>
    );
  }

  if (error) {
    return (
      <select disabled style={errorSelectStyle}>
        <option>Error loading namespaces</option>
      </select>
    );
  }

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>): void => {
    onNamespaceChange(e.target.value);
  };

  return (
    <label style={{ display: "flex", alignItems: "center", gap: "8px", fontSize: "14px", fontWeight: 500, color: "var(--text-color)" }}>
      Namespace
      <select
        value={selectedNamespace ?? undefined}
        onChange={handleChange}
        style={selectStyle}
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
