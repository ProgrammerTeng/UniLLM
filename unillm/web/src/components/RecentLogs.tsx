"use client";

import { useI18n } from "@/lib/i18n";

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
  const { t } = useI18n();

  if (logs.length === 0) {
    return (
      <p className="text-sm text-[var(--muted)]">{t.logs.noRequests}</p>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="text-left text-[var(--muted)]">
            <th className="pb-3">{t.logs.colTime}</th>
            <th className="pb-3">{t.logs.colModel}</th>
            <th className="pb-3 text-right">{t.logs.colTokens}</th>
            <th className="pb-3 text-right">{t.logs.colCost}</th>
            <th className="pb-3 text-right">{t.logs.colLatency}</th>
            <th className="pb-3 text-right">{t.logs.colStatus}</th>
          </tr>
        </thead>
        <tbody>
          {logs.map((log) => (
            <tr
              key={log.id}
              className="border-t"
              style={{ borderColor: "var(--border)" }}
            >
              <td className="py-2.5 text-[var(--muted)] text-xs whitespace-nowrap">
                {new Date(log.created_at).toLocaleString()}
              </td>
              <td className="py-2.5 font-mono text-xs">{log.model_name}</td>
              <td className="py-2.5 text-right">{log.total_tokens}</td>
              <td className="py-2.5 text-right font-mono">
                ${log.cost.toFixed(6)}
              </td>
              <td className="py-2.5 text-right">{log.latency_ms}ms</td>
              <td className="py-2.5 text-right">
                <span
                  className="text-xs px-1.5 py-0.5 rounded"
                  style={{
                    color:
                      log.status === "ok" ? "var(--success)" : "var(--danger)",
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
