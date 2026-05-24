"use client";

interface LogEntry {
  id: number;
  model_name: string;
  provider: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  cost: number;
  latency_ms: number;
  status: string;
  created_at: string;
}

export default function RecentLogs({ logs }: { logs: LogEntry[] }) {
  if (logs.length === 0) {
    return (
      <p className="text-sm text-[var(--muted)]">No requests yet</p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="text-left text-[var(--muted)]">
            <th className="pb-3">Time</th>
            <th className="pb-3">Model</th>
            <th className="pb-3 text-right">Tokens</th>
            <th className="pb-3 text-right">Cost</th>
            <th className="pb-3 text-right">Latency</th>
            <th className="pb-3 text-right">Status</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((log) => (
            <tr
              key={log.id}
              className="border-t"
              style={{ borderColor: "var(--border)" }}
            >
              <td className="py-2.5 text-[var(--muted)] text-xs">
                {formatTime(log.created_at)}
              </td>
              <td className="py-2.5 font-mono text-xs">{log.model_name}</td>
              <td className="py-2.5 text-right text-xs">
                {log.total_tokens.toLocaleString()}
              </td>
              <td className="py-2.5 text-right text-xs">
                ${log.cost.toFixed(6)}
              </td>
              <td className="py-2.5 text-right text-xs">
                {(log.latency_ms / 1000).toFixed(2)}s
              </td>
              <td className="py-2.5 text-right">
                <span
                  className="text-xs px-1.5 py-0.5 rounded"
                  style={{
                    background:
                      log.status === "success"
                        ? "rgba(34,197,94,0.15)"
                        : "rgba(239,68,68,0.15)",
                    color:
                      log.status === "success"
                        ? "var(--success)"
                        : "var(--danger)",
                  }}
                >
                  {log.status}
                </span>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formatTime(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
