"use client";

import { useState, useEffect } from "react";
import { SiteHeader } from "@/components/SiteHeader";
import { useI18n } from "@/lib/i18n";

interface ProviderHealth {
  name: string;
  status: string;
  latency_ms: number;
  circuit: string;
  checked_at: string;
}

interface StatusData {
  status: string;
  providers: ProviderHealth[];
}

export default function StatusPage() {
  const { t } = useI18n();
  const [data, setData] = useState<StatusData | null>(null);
  const [error, setError] = useState("");
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchStatus = async () => {
    try {
      const res = await fetch("/backend/status");
      if (!res.ok) throw new Error(t.status.fetchError);
      const d = await res.json();
      setData(d);
      setLastUpdated(new Date());
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : t.status.fetchError);
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 30000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const statusColor = (status: string) => {
    switch (status) {
      case "up":
      case "operational":
        return "var(--success)";
      case "degraded":
        return "#eab308";
      case "down":
        return "var(--danger)";
      default:
        return "var(--muted)";
    }
  };

  const statusBg = (status: string) => {
    switch (status) {
      case "up":
      case "operational":
        return "rgba(34, 197, 94, 0.12)";
      case "degraded":
        return "rgba(234, 179, 8, 0.12)";
      case "down":
        return "rgba(239, 68, 68, 0.12)";
      default:
        return "var(--card)";
    }
  };

  const circuitBadge = (circuit: string) => {
    switch (circuit) {
      case "closed":
        return { bg: "rgba(34, 197, 94, 0.15)", color: "var(--success)" };
      case "half-open":
        return { bg: "rgba(234, 179, 8, 0.15)", color: "#eab308" };
      case "open":
        return { bg: "rgba(239, 68, 68, 0.15)", color: "var(--danger)" };
      default:
        return { bg: "var(--card)", color: "var(--muted)" };
    }
  };

  return (
    <div className="min-h-screen" style={{ background: "var(--background)" }}>
      <SiteHeader />
      <div className="max-w-4xl mx-auto px-6 py-12">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold mb-2">{t.status.title}</h1>
          <p className="text-[var(--muted)]">{t.status.subtitle}</p>
        </div>

        {error && (
          <div
            className="rounded-lg p-4 mb-8 text-center text-sm"
            style={{
              background: "rgba(239, 68, 68, 0.1)",
              border: "1px solid var(--danger)",
              color: "var(--danger)",
            }}
          >
            {error}
          </div>
        )}

        {data && (
          <>
            <div
              className="rounded-xl p-6 mb-8 text-center"
              style={{
                background: statusBg(data.status),
                border: "1px solid var(--border)",
              }}
            >
              <div
                className="text-3xl font-bold"
                style={{ color: statusColor(data.status) }}
              >
                {data.status === "operational"
                  ? t.status.operational
                  : t.status.degraded}
              </div>
              {lastUpdated && (
                <div className="text-[var(--muted)] text-sm mt-2">
                  {t.status.lastUpdated}: {lastUpdated.toLocaleTimeString()}
                </div>
              )}
            </div>

            <div className="space-y-4">
              <h2 className="text-xl font-semibold mb-4">{t.status.providers}</h2>
              {data.providers.length === 0 ? (
                <p className="text-sm text-[var(--muted)]">{t.status.noData}</p>
              ) : (
                data.providers
                  .sort((a, b) => a.name.localeCompare(b.name))
                  .map((p) => {
                    const circuit = circuitBadge(p.circuit);
                    return (
                      <div
                        key={p.name}
                        className="rounded-lg p-4 flex items-center justify-between"
                        style={{
                          background: "var(--card)",
                          border: "1px solid var(--border)",
                        }}
                      >
                        <div className="flex items-center gap-4">
                          <div
                            className="w-3 h-3 rounded-full"
                            style={{ background: statusColor(p.status) }}
                          />
                          <div>
                            <div className="font-medium capitalize">
                              {p.name}
                            </div>
                            <div className="text-[var(--muted)] text-sm">
                              {p.checked_at
                                ? new Date(p.checked_at).toLocaleTimeString()
                                : "-"}
                            </div>
                          </div>
                        </div>
                        <div className="flex items-center gap-4">
                          <span
                            className="px-2 py-1 rounded text-xs font-medium"
                            style={{
                              background: circuit.bg,
                              color: circuit.color,
                            }}
                          >
                            {t.status.circuitLabel}: {p.circuit}
                          </span>
                          <span
                            className="font-medium text-sm"
                            style={{ color: statusColor(p.status) }}
                          >
                            {p.status.toUpperCase()}
                          </span>
                        </div>
                      </div>
                    );
                  })
              )}
            </div>

            <div className="mt-8 text-center text-[var(--muted)] text-sm">
              {t.status.autoRefresh}
            </div>
          </>
        )}

        {!data && !error && (
          <div className="text-center text-[var(--muted)]">
            {t.common.loading}
          </div>
        )}
      </div>
    </div>
  );
}
