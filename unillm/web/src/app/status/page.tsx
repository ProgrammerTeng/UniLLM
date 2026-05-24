"use client";

import { useState, useEffect } from "react";

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
  const [data, setData] = useState<StatusData | null>(null);
  const [error, setError] = useState("");
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchStatus = async () => {
    try {
      const res = await fetch("/backend/status");
      if (!res.ok) throw new Error("Failed to fetch status");
      const d = await res.json();
      setData(d);
      setLastUpdated(new Date());
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to fetch status");
    }
  };

  useEffect(() => {
    fetchStatus();
    const interval = setInterval(fetchStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  const statusColor = (status: string) => {
    switch (status) {
      case "up":
      case "operational":
        return "text-green-400";
      case "degraded":
        return "text-yellow-400";
      case "down":
        return "text-red-400";
      default:
        return "text-gray-400";
    }
  };

  const statusBg = (status: string) => {
    switch (status) {
      case "up":
      case "operational":
        return "bg-green-500/20 border-green-500/30";
      case "degraded":
        return "bg-yellow-500/20 border-yellow-500/30";
      case "down":
        return "bg-red-500/20 border-red-500/30";
      default:
        return "bg-gray-500/20 border-gray-500/30";
    }
  };

  const circuitBadge = (circuit: string) => {
    switch (circuit) {
      case "closed":
        return "bg-green-500/20 text-green-400";
      case "half-open":
        return "bg-yellow-500/20 text-yellow-400";
      case "open":
        return "bg-red-500/20 text-red-400";
      default:
        return "bg-gray-500/20 text-gray-400";
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <div className="max-w-4xl mx-auto px-6 py-12">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold mb-2">UniLLM Status</h1>
          <p className="text-gray-400">Real-time provider health monitoring</p>
        </div>

        {error && (
          <div className="bg-red-500/10 border border-red-500/30 rounded-lg p-4 mb-8 text-red-400 text-center">
            {error}
          </div>
        )}

        {data && (
          <>
            <div
              className={`rounded-xl border p-6 mb-8 text-center ${statusBg(data.status)}`}
            >
              <div className={`text-3xl font-bold ${statusColor(data.status)}`}>
                {data.status === "operational"
                  ? "All Systems Operational"
                  : "Some Systems Degraded"}
              </div>
              {lastUpdated && (
                <div className="text-gray-400 text-sm mt-2">
                  Last checked: {lastUpdated.toLocaleTimeString()}
                </div>
              )}
            </div>

            <div className="space-y-4">
              <h2 className="text-xl font-semibold mb-4">Providers</h2>
              {data.providers
                .sort((a, b) => a.name.localeCompare(b.name))
                .map((p) => (
                  <div
                    key={p.name}
                    className="bg-gray-900 border border-gray-800 rounded-lg p-4 flex items-center justify-between"
                  >
                    <div className="flex items-center gap-4">
                      <div
                        className={`w-3 h-3 rounded-full ${
                          p.status === "up"
                            ? "bg-green-400"
                            : p.status === "degraded"
                              ? "bg-yellow-400"
                              : "bg-red-400"
                        }`}
                      />
                      <div>
                        <div className="font-medium capitalize">{p.name}</div>
                        <div className="text-gray-500 text-sm">
                          {p.checked_at
                            ? new Date(p.checked_at).toLocaleTimeString()
                            : "-"}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <span
                        className={`px-2 py-1 rounded text-xs font-medium ${circuitBadge(p.circuit)}`}
                      >
                        circuit: {p.circuit}
                      </span>
                      <span className={`font-medium ${statusColor(p.status)}`}>
                        {p.status.toUpperCase()}
                      </span>
                    </div>
                  </div>
                ))}
            </div>

            <div className="mt-8 text-center text-gray-500 text-sm">
              Auto-refreshes every 30 seconds
            </div>
          </>
        )}

        {!data && !error && (
          <div className="text-center text-gray-400">Loading...</div>
        )}
      </div>
    </div>
  );
}
