"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  isLoggedIn,
  logout,
  getUsageSummary,
  getUsageByModel,
  getUsageDaily,
  getUsageRecent,
  getAPIKeys,
  createAPIKey,
  deleteAPIKey,
  changePassword,
} from "@/lib/api";
import UsageChart from "@/components/UsageChart";
import RecentLogs from "@/components/RecentLogs";

interface Summary {
  total_requests: number;
  total_tokens: number;
  total_cost: number;
  avg_latency: number;
  success_rate: number;
  today_cost: number;
  balance: number;
}

interface ModelStat {
  model_name: string;
  requests: number;
  total_tokens: number;
  total_cost: number;
  avg_latency: number;
}

interface APIKey {
  id: number;
  name: string;
  key_prefix: string;
  scope: string;
  is_active: boolean;
  last_used: string;
  created_at: string;
}

export default function DashboardPage() {
  const router = useRouter();
  const [summary, setSummary] = useState<Summary | null>(null);
  const [models, setModels] = useState<ModelStat[]>([]);
  const [keys, setKeys] = useState<APIKey[]>([]);
  const [newKeyName, setNewKeyName] = useState("");
  const [newKeyResult, setNewKeyResult] = useState("");
  const [daily, setDaily] = useState<Array<{ date: string; requests: number; tokens: number; cost: number }>>([]);
  const [recentLogs, setRecentLogs] = useState<Array<{
    id: number; model_name: string; provider: string;
    prompt_tokens: number; completion_tokens: number; total_tokens: number;
    cost: number; latency_ms: number; status: string; created_at: string;
  }>>([]);
  const [tab, setTab] = useState<"overview" | "keys" | "logs" | "settings">("overview");
  const [oldPwd, setOldPwd] = useState("");
  const [newPwd, setNewPwd] = useState("");
  const [pwdMsg, setPwdMsg] = useState("");

  useEffect(() => {
    if (!isLoggedIn()) {
      router.push("/login");
      return;
    }
    loadData();
  }, [router]);

  async function loadData() {
    try {
      const [s, m, k, d, r] = await Promise.all([
        getUsageSummary(),
        getUsageByModel(),
        getAPIKeys(),
        getUsageDaily().catch(() => ({ days: [] })),
        getUsageRecent().catch(() => ({ logs: [] })),
      ]);
      setSummary(s);
      setModels(m.models || []);
      setKeys(k.keys || []);
      setDaily(d.days || []);
      setRecentLogs(r.logs || []);
    } catch {
      router.push("/login");
    }
  }

  async function handleCreateKey() {
    if (!newKeyName.trim()) return;
    try {
      const result = await createAPIKey(newKeyName);
      setNewKeyResult(result.key);
      setNewKeyName("");
      loadData();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Failed");
    }
  }

  async function handleDeleteKey(id: number) {
    if (!confirm("Delete this API key?")) return;
    await deleteAPIKey(id);
    loadData();
  }

  if (!summary) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-[var(--muted)]">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header
        className="border-b px-6 py-3 flex items-center justify-between"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="flex items-center gap-4">
          <h1 className="text-lg font-bold">UniLLM</h1>
          <a
            href="/models"
            className="text-sm text-[var(--muted)] hover:text-white transition-colors"
          >
            Models
          </a>
          <a
            href="/playground"
            className="text-sm text-[var(--muted)] hover:text-white transition-colors"
          >
            Playground
          </a>
          <a
            href="/docs"
            className="text-sm text-[var(--muted)] hover:text-white transition-colors"
          >
            Docs
          </a>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-[var(--muted)]">
            Balance: ${summary.balance.toFixed(2)}
          </span>
          <button
            onClick={logout}
            className="text-sm text-[var(--muted)] hover:text-white"
          >
            Logout
          </button>
        </div>
      </header>

      <div className="max-w-6xl mx-auto p-6">
        {/* Tabs */}
        <div className="flex gap-4 mb-6">
          {(["overview", "logs", "keys", "settings"] as const).map((t) => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className="px-4 py-2 rounded-lg text-sm font-medium transition-colors"
              style={{
                background: tab === t ? "var(--primary)" : "var(--card)",
                border: `1px solid ${tab === t ? "var(--primary)" : "var(--border)"}`,
              }}
            >
              {t === "overview" ? "Overview" : t === "logs" ? "Logs" : "API Keys"}
            </button>
          ))}
        </div>

        {tab === "overview" && (
          <>
            {/* Stats Cards */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
              <StatCard
                label="Total Requests"
                value={summary.total_requests.toLocaleString()}
              />
              <StatCard
                label="Total Tokens"
                value={summary.total_tokens.toLocaleString()}
              />
              <StatCard
                label="Total Cost"
                value={`$${summary.total_cost.toFixed(4)}`}
              />
              <StatCard
                label="Today Cost"
                value={`$${summary.today_cost.toFixed(4)}`}
              />
              <StatCard
                label="Avg Latency"
                value={`${summary.avg_latency.toFixed(1)}s`}
              />
              <StatCard
                label="Success Rate"
                value={`${summary.success_rate.toFixed(1)}%`}
                color={
                  summary.success_rate >= 99
                    ? "var(--success)"
                    : "var(--warning)"
                }
              />
            </div>

            {/* Usage Chart */}
            <div
              className="rounded-xl p-5 mb-6"
              style={{
                background: "var(--card)",
                border: "1px solid var(--border)",
              }}
            >
              <h3 className="text-sm font-semibold mb-4 text-[var(--muted)]">
                USAGE TREND
              </h3>
              <UsageChart data={daily} />
            </div>

            {/* Per-Model Breakdown */}
            <div
              className="rounded-xl p-5"
              style={{
                background: "var(--card)",
                border: "1px solid var(--border)",
              }}
            >
              <h3 className="text-sm font-semibold mb-4 text-[var(--muted)]">
                USAGE BY MODEL
              </h3>
              {models.length === 0 ? (
                <p className="text-sm text-[var(--muted)]">No usage yet</p>
              ) : (
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-[var(--muted)]">
                      <th className="pb-3">Model</th>
                      <th className="pb-3 text-right">Requests</th>
                      <th className="pb-3 text-right">Tokens</th>
                      <th className="pb-3 text-right">Cost</th>
                      <th className="pb-3 text-right">Avg Latency</th>
                    </tr>
                  </thead>
                  <tbody>
                    {models.map((m) => (
                      <tr
                        key={m.model_name}
                        className="border-t"
                        style={{ borderColor: "var(--border)" }}
                      >
                        <td className="py-3 font-mono">{m.model_name}</td>
                        <td className="py-3 text-right">{m.requests}</td>
                        <td className="py-3 text-right">
                          {m.total_tokens.toLocaleString()}
                        </td>
                        <td className="py-3 text-right">
                          ${m.total_cost.toFixed(6)}
                        </td>
                        <td className="py-3 text-right">
                          {m.avg_latency.toFixed(1)}s
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}
            </div>
          </>
        )}

        {tab === "logs" && (
          <div
            className="rounded-xl p-5"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
            }}
          >
            <h3 className="text-sm font-semibold mb-4 text-[var(--muted)]">
              RECENT REQUESTS
            </h3>
            <RecentLogs logs={recentLogs} />
          </div>
        )}

        {tab === "keys" && (
          <div
            className="rounded-xl p-5"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
            }}
          >
            <h3 className="text-sm font-semibold mb-4 text-[var(--muted)]">
              API KEYS
            </h3>

            {/* Create Key */}
            <div className="flex gap-2 mb-4">
              <input
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                placeholder="Key name (e.g. production)"
                className="flex-1 px-3 py-2 rounded-lg text-sm outline-none"
                style={{
                  background: "var(--background)",
                  border: "1px solid var(--border)",
                }}
                onKeyDown={(e) => e.key === "Enter" && handleCreateKey()}
              />
              <button
                onClick={handleCreateKey}
                className="px-4 py-2 rounded-lg text-sm font-medium text-white"
                style={{ background: "var(--primary)" }}
              >
                Create
              </button>
            </div>

            {newKeyResult && (
              <div
                className="mb-4 p-3 rounded-lg text-sm"
                style={{
                  background: "rgba(34,197,94,0.1)",
                  border: "1px solid var(--success)",
                }}
              >
                <p className="mb-1 font-medium" style={{ color: "var(--success)" }}>
                  Key created! Copy it now — it won&apos;t be shown again.
                </p>
                <code className="block font-mono text-xs break-all">
                  {newKeyResult}
                </code>
              </div>
            )}

            {/* Key List */}
            {keys.length === 0 ? (
              <p className="text-sm text-[var(--muted)]">No API keys yet</p>
            ) : (
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-[var(--muted)]">
                    <th className="pb-3">Name</th>
                    <th className="pb-3">Key</th>
                    <th className="pb-3">Scope</th>
                    <th className="pb-3">Created</th>
                    <th className="pb-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody>
                  {keys.map((k) => (
                    <tr
                      key={k.id}
                      className="border-t"
                      style={{ borderColor: "var(--border)" }}
                    >
                      <td className="py-3">{k.name}</td>
                      <td className="py-3 font-mono text-[var(--muted)]">
                        {k.key_prefix}...
                      </td>
                      <td className="py-3">{k.scope}</td>
                      <td className="py-3 text-[var(--muted)]">
                        {new Date(k.created_at).toLocaleDateString()}
                      </td>
                      <td className="py-3 text-right">
                        <button
                          onClick={() => handleDeleteKey(k.id)}
                          className="text-xs px-2 py-1 rounded"
                          style={{ color: "var(--danger)" }}
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {tab === "settings" && (
          <div
            className="rounded-xl p-5 max-w-md"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
            }}
          >
            <h3 className="text-sm font-semibold mb-4 text-[var(--muted)]">
              CHANGE PASSWORD
            </h3>
            <div className="flex flex-col gap-3">
              <input
                type="password"
                value={oldPwd}
                onChange={(e) => setOldPwd(e.target.value)}
                placeholder="Current password"
                className="px-3 py-2 rounded-lg text-sm outline-none"
                style={{
                  background: "var(--background)",
                  border: "1px solid var(--border)",
                }}
              />
              <input
                type="password"
                value={newPwd}
                onChange={(e) => setNewPwd(e.target.value)}
                placeholder="New password (min 8 characters)"
                className="px-3 py-2 rounded-lg text-sm outline-none"
                style={{
                  background: "var(--background)",
                  border: "1px solid var(--border)",
                }}
              />
              <button
                onClick={async () => {
                  setPwdMsg("");
                  try {
                    await changePassword(oldPwd, newPwd);
                    setPwdMsg("Password changed successfully");
                    setOldPwd("");
                    setNewPwd("");
                  } catch (err) {
                    setPwdMsg(err instanceof Error ? err.message : "Failed");
                  }
                }}
                className="px-4 py-2 rounded-lg text-sm font-medium text-white w-fit"
                style={{ background: "var(--primary)" }}
              >
                Update Password
              </button>
              {pwdMsg && (
                <p
                  className="text-sm"
                  style={{
                    color: pwdMsg.includes("success")
                      ? "var(--success)"
                      : "var(--danger)",
                  }}
                >
                  {pwdMsg}
                </p>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  color,
}: {
  label: string;
  value: string;
  color?: string;
}) {
  return (
    <div
      className="rounded-xl p-4"
      style={{
        background: "var(--card)",
        border: "1px solid var(--border)",
      }}
    >
      <div className="text-xs text-[var(--muted)] mb-1">{label}</div>
      <div className="text-xl font-bold" style={{ color }}>
        {value}
      </div>
    </div>
  );
}
