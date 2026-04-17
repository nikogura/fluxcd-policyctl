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
      <div className="rounded border border-gray-200 bg-gray-50 p-8 text-center text-sm text-gray-500">
        No image policies found in this namespace.
      </div>
    );
  }

  return (
    <div>
      {actionError && (
        <div className="mb-4 rounded border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {actionError}
        </div>
      )}
      <div className="overflow-x-auto rounded border border-gray-200">
        <table className="w-full text-left text-sm">
          <thead className="bg-gray-50 text-xs uppercase text-gray-500">
            <tr>
              <th className="px-4 py-3">Name</th>
              <th className="px-4 py-3">Image</th>
              <th className="px-4 py-3">Semver Range</th>
              <th className="px-4 py-3">Latest Version</th>
              <th className="px-4 py-3">Available Versions</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {policies.map((policy) => {
              const key = policyKey(policy);
              const visibleVersions = policy.availableVersions.slice(0, MAX_VISIBLE_VERSIONS);
              const remainingCount = policy.availableVersions.length - MAX_VISIBLE_VERSIONS;
              const isConfirming = confirmingDelete === key;

              return (
                <tr key={key} className="hover:bg-gray-50">
                  <td className="whitespace-nowrap px-4 py-3 font-medium text-gray-900">
                    {policy.name}
                  </td>
                  <td className="px-4 py-3 text-gray-600">
                    <code className="rounded bg-gray-100 px-1.5 py-0.5 text-xs">
                      {policy.imageUrl}
                    </code>
                  </td>
                  <td className="px-4 py-3">
                    <RangeComboBox
                      currentRange={policy.semverRange}
                      availableVersions={policy.availableVersions}
                      onRangeChange={(newRange): void => {
                        void handleRangeChange({ policy, newRange });
                      }}
                    />
                  </td>
                  <td className="whitespace-nowrap px-4 py-3">
                    <span className="rounded bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-800">
                      {policy.latestVersion || "N/A"}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-1">
                      {visibleVersions.map((version) => (
                        <span
                          key={version}
                          className="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600"
                        >
                          {version}
                        </span>
                      ))}
                      {remainingCount > 0 && (
                        <span className="rounded bg-gray-200 px-1.5 py-0.5 text-xs text-gray-500">
                          +{remainingCount} more
                        </span>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <StatusBadge ready={policy.ready} message={policy.message} />
                  </td>
                  <td className="px-4 py-3">
                    {isConfirming ? (
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={(): void => {
                            void handleDelete({ policy });
                          }}
                          className="rounded bg-red-600 px-2 py-1 text-xs text-white hover:bg-red-700"
                        >
                          Confirm
                        </button>
                        <button
                          type="button"
                          onClick={(): void => setConfirmingDelete(null)}
                          className="rounded bg-gray-200 px-2 py-1 text-xs text-gray-700 hover:bg-gray-300"
                        >
                          Cancel
                        </button>
                      </div>
                    ) : (
                      <button
                        type="button"
                        onClick={(): void => setConfirmingDelete(key)}
                        className="rounded p-1.5 text-gray-400 hover:bg-red-50 hover:text-red-600"
                        title="Delete policy"
                      >
                        <Trash2 className="h-4 w-4" />
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
