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
    const constraint = `>=${version}`;
    setInputValue(constraint);
    onRangeChange(constraint);
    setIsOpen(false);
  };

  const handleToggle = (): void => {
    setIsOpen(!isOpen);
  };

  return (
    <div ref={containerRef} className="relative">
      <div className="flex">
        <input
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          onFocus={handleToggle}
          className="w-32 rounded-l border border-gray-300 px-2 py-1 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        <button
          type="button"
          onClick={handleToggle}
          className="rounded-r border border-l-0 border-gray-300 bg-gray-50 px-2 py-1 hover:bg-gray-100"
        >
          <ChevronDown className="h-3 w-3" />
        </button>
      </div>
      {isOpen && sortedVersions.length > 0 && (
        <ul className="absolute z-10 mt-1 max-h-48 w-48 overflow-auto rounded border border-gray-200 bg-white shadow-lg">
          {sortedVersions.map((version) => {
            const isSelected = currentRange === `>=${version}`;
            return (
              <li key={version}>
                <button
                  type="button"
                  onClick={(): void => handleVersionClick(version)}
                  className="flex w-full items-center gap-2 px-3 py-1.5 text-left text-sm hover:bg-blue-50"
                >
                  {isSelected ? (
                    <Check className="h-3 w-3 text-blue-600" />
                  ) : (
                    <span className="h-3 w-3" />
                  )}
                  <span className={isPreRelease(version) ? "text-gray-400" : "text-gray-900"}>
                    {version}
                  </span>
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
