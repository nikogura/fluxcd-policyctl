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

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: "8px 12px",
  border: "1px solid var(--border-color)",
  borderRadius: "4px",
  backgroundColor: "var(--bg-color)",
  color: "var(--text-color)",
  fontSize: "14px",
  outline: "none",
};

const disabledInputStyle: React.CSSProperties = {
  ...inputStyle,
  backgroundColor: "var(--bg-secondary)",
  color: "var(--text-muted)",
  cursor: "not-allowed",
};

const labelStyle: React.CSSProperties = {
  display: "block",
  marginBottom: "4px",
  fontSize: "14px",
  fontWeight: 500,
  color: "var(--text-color)",
};

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
      onClick={handleBackdropClick}
      role="presentation"
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 50,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "rgba(0, 0, 0, 0.5)",
      }}
    >
      <div style={{
        width: "100%",
        maxWidth: "450px",
        borderRadius: "8px",
        backgroundColor: "var(--bg-color)",
        padding: "24px",
        boxShadow: "var(--shadow-lg)",
      }}>
        <div style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          marginBottom: "16px",
        }}>
          <h2 style={{ fontSize: "18px", fontWeight: 600, color: "var(--text-color)" }}>
            Create Image Policy
          </h2>
          <button
            type="button"
            onClick={onClose}
            style={{
              padding: "4px",
              border: "none",
              borderRadius: "4px",
              backgroundColor: "transparent",
              color: "var(--text-muted)",
              cursor: "pointer",
            }}
          >
            <X size={20} />
          </button>
        </div>

        {error && (
          <div style={{
            marginBottom: "16px",
            padding: "8px 12px",
            borderRadius: "4px",
            backgroundColor: "var(--error-color)",
            color: "white",
            fontSize: "14px",
          }}>
            {error}
          </div>
        )}

        <form
          onSubmit={(e): void => {
            void handleSubmit(e);
          }}
          style={{ display: "flex", flexDirection: "column", gap: "16px" }}
        >
          <div>
            <label htmlFor="policyName" style={labelStyle}>
              Policy Name
            </label>
            <input
              id="policyName"
              type="text"
              value={policyName ?? undefined}
              onChange={(e): void => setPolicyName(e.target.value || null)}
              placeholder="my-app-policy"
              required
              style={inputStyle}
            />
          </div>

          <div>
            <label htmlFor="policyNamespace" style={labelStyle}>
              Namespace
            </label>
            <input
              id="policyNamespace"
              type="text"
              value={namespace}
              disabled
              style={disabledInputStyle}
            />
          </div>

          <div>
            <label htmlFor="imageRepo" style={labelStyle}>
              Image Repository Name
            </label>
            <input
              id="imageRepo"
              type="text"
              value={imageRepository ?? undefined}
              onChange={(e): void => setImageRepository(e.target.value || null)}
              placeholder="my-app"
              required
              style={inputStyle}
            />
            <p style={{ marginTop: "4px", fontSize: "12px", color: "var(--text-muted)" }}>
              Name of an existing Flux ImageRepository resource
            </p>
          </div>

          <div>
            <label htmlFor="semverRange" style={labelStyle}>
              Semver Range
            </label>
            <input
              id="semverRange"
              type="text"
              value={semverRange}
              onChange={(e): void => setSemverRange(e.target.value)}
              placeholder=">=1.0.0"
              required
              style={{ ...inputStyle, fontFamily: "monospace" }}
            />
          </div>

          <div style={{ display: "flex", justifyContent: "flex-end", gap: "12px", paddingTop: "8px" }}>
            <button
              type="button"
              onClick={onClose}
              style={{
                padding: "8px 16px",
                border: "1px solid var(--border-color)",
                borderRadius: "4px",
                backgroundColor: "var(--bg-color)",
                color: "var(--text-color)",
                fontSize: "14px",
                fontWeight: 600,
                cursor: "pointer",
              }}
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={submitting}
              style={{
                padding: "8px 16px",
                border: "none",
                borderRadius: "4px",
                backgroundColor: "var(--button-bg)",
                color: "var(--button-text)",
                fontSize: "14px",
                fontWeight: 600,
                cursor: submitting ? "not-allowed" : "pointer",
                opacity: submitting ? 0.6 : 1,
              }}
            >
              {submitting ? "Creating..." : "Create Policy"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
