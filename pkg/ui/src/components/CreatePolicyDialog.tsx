"use client";

import { X } from "lucide-react";
import { useEffect, useState } from "react";

import { createPolicy } from "@/lib/api";

interface CreatePolicyDialogProps {
  readonly cluster: string;
  readonly namespace: string;
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCreated: () => void;
}

const DEFAULT_SEMVER_RANGE = ">=0.0.0";

export function CreatePolicyDialog({ cluster, namespace, isOpen, onClose, onCreated }: CreatePolicyDialogProps): React.ReactElement | null {
  const [policyName, setPolicyName] = useState<string | null>(null);
  const [imageRepository, setImageRepository] = useState<string | null>(null);
  const [semverRange, setSemverRange] = useState(DEFAULT_SEMVER_RANGE);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setPolicyName(null);
      setImageRepository(null);
      setSemverRange(DEFAULT_SEMVER_RANGE);
      setError(null);
    }
  }, [isOpen]);

  if (!isOpen) {
    return null;
  }

  const handleSubmit = async (e: React.FormEvent): Promise<void> => {
    e.preventDefault();

    if (!policyName || !imageRepository) {
      setError("Policy name and image repository are required.");
      return;
    }

    try {
      setSubmitting(true);
      setError(null);
      await createPolicy({
        cluster,
        request: {
          name: policyName,
          namespace,
          imageRepository,
          semverRange,
        },
      });
      onCreated();
      onClose();
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to create policy";
      setError(message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleBackdropClick = (e: React.MouseEvent): void => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      onClick={handleBackdropClick}
      role="presentation"
    >
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-gray-900">Create Image Policy</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {error && (
          <div className="mb-4 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
            {error}
          </div>
        )}

        <form
          onSubmit={(e): void => {
            void handleSubmit(e);
          }}
          className="space-y-4"
        >
          <div>
            <label htmlFor="policyName" className="mb-1 block text-sm font-medium text-gray-700">
              Policy Name
            </label>
            <input
              id="policyName"
              type="text"
              value={policyName ?? undefined}
              onChange={(e): void => setPolicyName(e.target.value || null)}
              placeholder="my-app-policy"
              required
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div>
            <label htmlFor="policyNamespace" className="mb-1 block text-sm font-medium text-gray-700">
              Namespace
            </label>
            <input
              id="policyNamespace"
              type="text"
              value={namespace}
              disabled
              className="w-full rounded border border-gray-200 bg-gray-50 px-3 py-2 text-sm text-gray-500"
            />
          </div>

          <div>
            <label htmlFor="imageRepo" className="mb-1 block text-sm font-medium text-gray-700">
              Image Repository Name
            </label>
            <input
              id="imageRepo"
              type="text"
              value={imageRepository ?? undefined}
              onChange={(e): void => setImageRepository(e.target.value || null)}
              placeholder="my-app"
              required
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
            <p className="mt-1 text-xs text-gray-500">
              Name of an existing Flux ImageRepository resource
            </p>
          </div>

          <div>
            <label htmlFor="semverRange" className="mb-1 block text-sm font-medium text-gray-700">
              Semver Range
            </label>
            <input
              id="semverRange"
              type="text"
              value={semverRange}
              onChange={(e): void => setSemverRange(e.target.value)}
              placeholder=">=1.0.0"
              required
              className="w-full rounded border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="rounded border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:bg-blue-400"
            >
              {submitting ? "Creating..." : "Create Policy"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
