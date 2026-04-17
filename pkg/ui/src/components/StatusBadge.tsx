interface StatusBadgeProps {
  readonly ready: boolean;
  readonly message: string;
}

export function StatusBadge({ ready, message }: StatusBadgeProps): React.ReactElement {
  return (
    <span
      title={message}
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
        ready
          ? "bg-green-100 text-green-800"
          : "bg-red-100 text-red-800"
      }`}
    >
      {ready ? "Ready" : "Not Ready"}
    </span>
  );
}
