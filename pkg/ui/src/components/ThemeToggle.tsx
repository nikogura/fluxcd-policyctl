"use client";

import { useEffect, useState } from "react";

export function ThemeToggle(): React.ReactElement {
  const [mounted, setMounted] = useState(false);
  const [isDark, setIsDark] = useState(true);

  useEffect((): void => {
    const stored = localStorage.getItem("theme");
    const dark = stored ? stored === "dark" : true;
    setIsDark(dark);
    document.documentElement.setAttribute("data-theme", dark ? "dark" : "light");
    setMounted(true);
  }, []);

  function handleToggle(): void {
    const next = !isDark;
    setIsDark(next);
    const themeName = next ? "dark" : "light";
    localStorage.setItem("theme", themeName);
    document.documentElement.setAttribute("data-theme", themeName);
  }

  // Don't render until mounted to avoid hydration mismatch
  if (!mounted) {
    return (
      <button
        type="button"
        style={{
          padding: "6px 10px",
          border: "1px solid var(--border-color)",
          borderRadius: "4px",
          backgroundColor: "transparent",
          color: "var(--text-color)",
          cursor: "pointer",
          fontSize: "16px",
          minWidth: "40px",
          lineHeight: 1,
        }}
        aria-label="Toggle theme"
      >
        {" "}
      </button>
    );
  }

  return (
    <button
      onClick={handleToggle}
      type="button"
      style={{
        padding: "6px 10px",
        border: "1px solid var(--border-color)",
        borderRadius: "4px",
        backgroundColor: "transparent",
        color: "var(--text-color)",
        cursor: "pointer",
        fontSize: "16px",
        minWidth: "40px",
        lineHeight: 1,
      }}
      title={isDark ? "Switch to light mode" : "Switch to dark mode"}
      aria-label="Toggle theme"
    >
      {isDark ? "\u2600\uFE0F" : "\uD83C\uDF19"}
    </button>
  );
}
