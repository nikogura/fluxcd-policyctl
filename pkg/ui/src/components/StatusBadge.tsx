interface StatusBadgeProps {
  readonly ready: boolean;
  readonly message: string;
}

export function StatusBadge({ ready, message }: StatusBadgeProps): React.ReactElement {
  const color = ready ? "#28a745" : "#dc3545";

  return (
    <span
      title={message}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: "6px",
        fontSize: "13px",
        color: "var(--text-color)",
      }}
    >
      <span style={{ color, fontSize: "10px" }}>{"\u25CF"}</span>
      {ready ? "Ready" : "Not Ready"}
    </span>
  );
}
