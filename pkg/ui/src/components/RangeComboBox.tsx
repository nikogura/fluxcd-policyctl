"use client";

import { Check, ChevronDown } from "lucide-react";
import { useEffect, useRef, useState } from "react";

interface RangeComboBoxProps {
  readonly currentRange: string;
  readonly availableVersions: readonly string[];
  readonly onRangeChange: (range: string) => void;
}

function isPreRelease(version: string): boolean {
  return version.includes("-");
}

function sortVersionsNewestFirst(versions: readonly string[]): readonly string[] {
  return [...versions].sort((a, b) => {
    const aIsPreRelease = isPreRelease(a);
    const bIsPreRelease = isPreRelease(b);

    if (aIsPreRelease !== bIsPreRelease) {
      return aIsPreRelease ? 1 : -1;
    }

    // eslint-disable-next-line local-rules/disallow-empty-string
    const aParts = a.replace(/^v/, "").split(/[.-]/).map(Number);
    // eslint-disable-next-line local-rules/disallow-empty-string
    const bParts = b.replace(/^v/, "").split(/[.-]/).map(Number);

    for (let i = 0; i < Math.max(aParts.length, bParts.length); i++) {
      const aPart = aParts[i] ?? 0;
      const bPart = bParts[i] ?? 0;
      if (aPart !== bPart) return bPart - aPart;
    }

    return 0;
  });
}

export function RangeComboBox({ currentRange, availableVersions, onRangeChange }: RangeComboBoxProps): React.ReactElement {
  const [inputValue, setInputValue] = useState(currentRange);
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setInputValue(currentRange);
  }, [currentRange]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent): void => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return (): void => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  const sortedVersions = sortVersionsNewestFirst(availableVersions);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    setInputValue(e.target.value);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>): void => {
    if (e.key === "Enter") {
      onRangeChange(inputValue);
      setIsOpen(false);
    }
    if (e.key === "Escape") {
      setIsOpen(false);
      setInputValue(currentRange);
    }
  };

  const handleVersionClick = (version: string): void => {
    setInputValue(version);
    onRangeChange(version);
    setIsOpen(false);
  };

  const handleToggle = (): void => {
    setIsOpen(!isOpen);
  };

  return (
    <div ref={containerRef} style={{ position: "relative" }}>
      <div style={{ display: "flex" }}>
        <input
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          onFocus={handleToggle}
          style={{
            width: "130px",
            padding: "6px 8px",
            border: "1px solid var(--border-color)",
            borderRight: "none",
            borderRadius: "4px 0 0 4px",
            backgroundColor: "var(--bg-color)",
            color: "var(--text-color)",
            fontSize: "13px",
            fontFamily: "monospace",
            outline: "none",
          }}
        />
        <button
          type="button"
          onClick={handleToggle}
          style={{
            padding: "6px 8px",
            border: "1px solid var(--border-color)",
            borderRadius: "0 4px 4px 0",
            backgroundColor: "var(--bg-secondary)",
            color: "var(--text-muted)",
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
          }}
        >
          <ChevronDown size={12} />
        </button>
      </div>
      {isOpen && sortedVersions.length > 0 && (
        <ul style={{
          position: "absolute",
          zIndex: 10,
          marginTop: "4px",
          maxHeight: "200px",
          width: "200px",
          overflowY: "auto",
          borderRadius: "4px",
          border: "1px solid var(--border-color)",
          backgroundColor: "var(--bg-color)",
          boxShadow: "var(--shadow-lg)",
          listStyle: "none",
          padding: 0,
        }}>
          {sortedVersions.map((version) => {
            const isSelected = currentRange === version || currentRange === `>=${version}`;
            return (
              <li key={version}>
                <button
                  type="button"
                  onClick={(): void => handleVersionClick(version)}
                  style={{
                    display: "flex",
                    width: "100%",
                    alignItems: "center",
                    gap: "8px",
                    padding: "6px 12px",
                    textAlign: "left",
                    fontSize: "13px",
                    fontFamily: "monospace",
                    backgroundColor: "transparent",
                    border: "none",
                    color: isPreRelease(version) ? "var(--text-muted)" : "var(--text-color)",
                    cursor: "pointer",
                  }}
                >
                  {isSelected ? (
                    <Check size={12} style={{ color: "var(--button-bg)" }} />
                  ) : (
                    <span style={{ width: "12px" }} />
                  )}
                  {version}
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
