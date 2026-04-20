"use client";

import { Trash2 } from "lucide-react";
import { useState } from "react";

import { deletePolicy, updatePolicy } from "@/lib/api";
import type { PolicyView } from "@/types";

import { RangeComboBox } from "./RangeComboBox";
import { StatusBadge } from "./StatusBadge";

interface PolicyTableProps {
  readonly policies: readonly PolicyView[];
  readonly cluster: string;
  readonly onPolicyUpdated: () => void;
}

const MAX_VISIBLE_VERSIONS = 5;

const thStyle: React.CSSProperties = {
  padding: "12px",
  textAlign: "left",
  fontWeight: 600,
  fontSize: "12px",
  textTransform: "uppercase",
  letterSpacing: "0.05em",
  color: "var(--text-muted)",
  backgroundColor: "var(--table-header-bg)",
  borderBottom: "1px solid var(--border-color)",
};

const tdStyle: React.CSSProperties = {
  padding: "12px",
  borderBottom: "1px solid var(--border-color)",
  fontSize: "14px",
  color: "var(--text-color)",
};

export function PolicyTable({ policies, cluster, onPolicyUpdated }: PolicyTableProps): React.ReactElement {
  const [confirmingDelete, setConfirmingDelete] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const handleRangeChange = async ({ policy, newRange }: { readonly policy: PolicyView; readonly newRange: string }): Promise<void> => {
    try {
      setActionError(null);
      await updatePolicy({
        cluster,
        namespace: policy.namespace,
        name: policy.name,
        semverRange: newRange,
      });
      onPolicyUpdated();
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to update policy";
      setActionError(message);
    }
  };

  const handleDelete = async ({ policy }: { readonly policy: PolicyView }): Promise<void> => {
    try {
      setActionError(null);
      await deletePolicy({
        cluster,
        namespace: policy.namespace,
        name: policy.name,
      });
      setConfirmingDelete(null);
      onPolicyUpdated();
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to delete policy";
      setActionError(message);
    }
  };

  const policyKey = (policy: PolicyView): string => {
    return `${policy.namespace}/${policy.name}`;
  };

  if (policies.length === 0) {
    return (
      <div style={{
        padding: "32px",
        textAlign: "center",
        fontSize: "14px",
        color: "var(--text-muted)",
        backgroundColor: "var(--bg-color)",
        borderRadius: "8px",
        border: "1px solid var(--border-color)",
      }}>
        No image policies found in this namespace.
      </div>
    );
  }

  return (
    <div>
      {actionError && (
        <div style={{
          marginBottom: "16px",
          padding: "8px 12px",
          borderRadius: "4px",
          backgroundColor: "var(--error-color)",
          color: "white",
          fontSize: "14px",
        }}>
          {actionError}
        </div>
      )}
      <div style={{
        borderRadius: "8px",
        boxShadow: "var(--shadow)",
        backgroundColor: "var(--bg-color)",
      }}>
        <table style={{ width: "100%", borderCollapse: "collapse", textAlign: "left" }}>
          <thead>
            <tr>
              <th style={thStyle}>Name</th>
              <th style={thStyle}>Image</th>
              <th style={thStyle}>Semver Range</th>
              <th style={thStyle}>Selected Version</th>
              <th style={thStyle}>Available Versions</th>
              <th style={thStyle}>Status</th>
              <th style={thStyle}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {policies.map((policy) => {
              const key = policyKey(policy);
              const versions = policy.availableVersions ?? [];
              const visibleVersions = versions.slice(0, MAX_VISIBLE_VERSIONS);
              const remainingCount = versions.length - MAX_VISIBLE_VERSIONS;
              const isConfirming = confirmingDelete === key;

              return (
                <tr key={key}>
                  <td style={{ ...tdStyle, fontWeight: 500, whiteSpace: "nowrap" }}>
                    {policy.name}
                  </td>
                  <td style={tdStyle}>
                    <span style={{
                      fontFamily: "monospace",
                      fontSize: "12px",
                      backgroundColor: "var(--bg-secondary)",
                      padding: "2px 6px",
                      borderRadius: "3px",
                    }}>
                      {policy.imageUrl}
                    </span>
                  </td>
                  <td style={tdStyle}>
                    <RangeComboBox
                      currentRange={policy.semverRange}
                      availableVersions={versions}
                      onRangeChange={(newRange): void => {
                        void handleRangeChange({ policy, newRange });
                      }}
                    />
                  </td>
                  <td style={{ ...tdStyle, whiteSpace: "nowrap" }}>
                    <span style={{
                      fontFamily: "monospace",
                      fontSize: "12px",
                      backgroundColor: "var(--border-color)",
                      padding: "2px 6px",
                      borderRadius: "3px",
                    }}>
                      {policy.latestVersion || "N/A"}
                    </span>
                  </td>
                  <td style={tdStyle}>
                    <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
                      {visibleVersions.map((version) => (
                        <span
                          key={version}
                          style={{
                            fontFamily: "monospace",
                            fontSize: "12px",
                            backgroundColor: "var(--border-color)",
                            padding: "2px 6px",
                            borderRadius: "3px",
                          }}
                        >
                          {version}
                        </span>
                      ))}
                      {remainingCount > 0 && (
                        <span style={{
                          fontSize: "12px",
                          color: "var(--text-muted)",
                          padding: "2px 6px",
                        }}>
                          +{remainingCount} more
                        </span>
                      )}
                    </div>
                  </td>
                  <td style={tdStyle}>
                    <StatusBadge ready={policy.ready} message={policy.message} />
                  </td>
                  <td style={tdStyle}>
                    {isConfirming ? (
                      <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
                        <button
                          type="button"
                          onClick={(): void => {
                            void handleDelete({ policy });
                          }}
                          style={{
                            padding: "4px 10px",
                            border: "none",
                            borderRadius: "4px",
                            backgroundColor: "var(--error-color)",
                            color: "white",
                            fontSize: "12px",
                            fontWeight: 600,
                            cursor: "pointer",
                          }}
                        >
                          Confirm
                        </button>
                        <button
                          type="button"
                          onClick={(): void => setConfirmingDelete(null)}
                          style={{
                            padding: "4px 10px",
                            border: "1px solid var(--border-color)",
                            borderRadius: "4px",
                            backgroundColor: "var(--bg-color)",
                            color: "var(--text-color)",
                            fontSize: "12px",
                            cursor: "pointer",
                          }}
                        >
                          Cancel
                        </button>
                      </div>
                    ) : (
                      <button
                        type="button"
                        onClick={(): void => setConfirmingDelete(key)}
                        style={{
                          padding: "6px",
                          border: "none",
                          borderRadius: "4px",
                          backgroundColor: "transparent",
                          color: "var(--text-muted)",
                          cursor: "pointer",
                        }}
                        title="Delete policy"
                      >
                        <Trash2 size={16} />
                      </button>
                    )}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
