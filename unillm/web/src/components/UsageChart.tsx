"use client";

import { useI18n } from "@/lib/i18n";

interface DailyData {
  date: string;
  requests: number;
  tokens: number;
  cost: number;
}

export default function UsageChart({ data }: { data: DailyData[] }) {
  const { t } = useI18n();

  if (data.length === 0) {
    return (
      <p className="text-sm text-[var(--muted)]">{t.dashboard.noUsage}</p>
    );
  }

  const maxReqs = Math.max(...data.map((d) => d.requests), 1);
  const maxCost = Math.max(...data.map((d) => d.cost), 0.001);

  return (
    <div className="space-y-4">
      <div>
        <div className="text-xs text-[var(--muted)] mb-2">
          {t.dashboard.chartDailyRequests}
        </div>
        <div className="flex items-end gap-[2px] h-24">
          {data.map((d) => {
            const h = Math.max((d.requests / maxReqs) * 100, 2);
            return (
              <div
                key={d.date}
                className="flex-1 rounded-t transition-all group relative"
                style={{
                  height: `${h}%`,
                  background: "var(--primary)",
                  minWidth: 4,
                }}
              >
                <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block z-10">
                  <div
                    className="px-2 py-1 rounded text-xs whitespace-nowrap"
                    style={{
                      background: "var(--card)",
                      border: "1px solid var(--border)",
                    }}
                  >
                    {d.date}: {d.requests} reqs
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      <div>
        <div className="text-xs text-[var(--muted)] mb-2">
          {t.dashboard.chartDailyCost}
        </div>
        <div className="flex items-end gap-[2px] h-24">
          {data.map((d) => {
            const h = Math.max((d.cost / maxCost) * 100, 2);
            return (
              <div
                key={d.date}
                className="flex-1 rounded-t transition-all group relative"
                style={{
                  height: `${h}%`,
                  background: "var(--success)",
                  minWidth: 4,
                }}
              >
                <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 hidden group-hover:block z-10">
                  <div
                    className="px-2 py-1 rounded text-xs whitespace-nowrap"
                    style={{
                      background: "var(--card)",
                      border: "1px solid var(--border)",
                    }}
                  >
                    {d.date}: ${d.cost.toFixed(4)}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
