"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { isLoggedIn, getMe } from "@/lib/api";
import {
  getGlobalStats,
  getUsers,
  updateBalance,
  getProviders,
  toggleProvider,
  getModels,
  updateModel,
  getProviderKeys,
  addProviderKey,
} from "@/lib/admin-api";

interface Stats {
  total_users: number;
  total_requests: number;
  total_cost: number;
  total_tokens: number;
  active_keys: number;
}

interface User {
  id: number;
  email: string;
  name: string;
  role: string;
  balance: number;
  created_at: string;
}

interface ProviderInfo {
  id: number;
  name: string;
  base_url: string;
  is_active: boolean;
}

interface ModelInfo {
  id: number;
  public_name: string;
  provider_id: number;
  upstream_model: string;
  input_price_per_1m: number;
  output_price_per_1m: number;
  is_active: boolean;
  max_tokens: number;
}

interface ProviderKeyInfo {
  id: number;
  provider_id: number;
  key_prefix: string;
  rpm: number;
  is_active: boolean;
}

type Tab = "overview" | "users" | "providers" | "models" | "keys";

export default function AdminPage() {
  const router = useRouter();
  const [tab, setTab] = useState<Tab>("overview");
  const [stats, setStats] = useState<Stats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [providerKeys, setProviderKeys] = useState<ProviderKeyInfo[]>([]);
  const [error, setError] = useState("");

  // Balance form
  const [balUserId, setBalUserId] = useState("");
  const [balDelta, setBalDelta] = useState("");
  const [balReason, setBalReason] = useState("");

  // Add key form
  const [newKeyProviderId, setNewKeyProviderId] = useState("");
  const [newKeyValue, setNewKeyValue] = useState("");
  const [newKeyRpm, setNewKeyRpm] = useState("60");

  useEffect(() => {
    if (!isLoggedIn()) {
      router.push("/login");
      return;
    }
    getMe()
      .then((u) => {
        if (u.role !== "admin") {
          router.push("/dashboard");
          return;
        }
        loadAll();
      })
      .catch(() => router.push("/login"));
  }, [router]);

  async function loadAll() {
    try {
      const [s, u, p, m, k] = await Promise.all([
        getGlobalStats(),
        getUsers(),
        getProviders(),
        getModels(),
        getProviderKeys(),
      ]);
      setStats(s);
      setUsers(u.users || []);
      setProviders(p.providers || []);
      setModels(m.models || []);
      setProviderKeys(k.keys || []);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    }
  }

  async function handleUpdateBalance() {
    if (!balUserId || !balDelta) return;
    try {
      await updateBalance(Number(balUserId), Number(balDelta), balReason);
      setBalUserId("");
      setBalDelta("");
      setBalReason("");
      loadAll();
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed");
    }
  }

  async function handleToggleProvider(id: number, current: boolean) {
    await toggleProvider(id, !current);
    loadAll();
  }

  async function handleToggleModel(id: number, current: boolean) {
    await updateModel({ id, is_active: !current });
    loadAll();
  }

  async function handleAddKey() {
    if (!newKeyProviderId || !newKeyValue) return;
    try {
      await addProviderKey(
        Number(newKeyProviderId),
        newKeyValue,
        Number(newKeyRpm)
      );
      setNewKeyValue("");
      loadAll();
    } catch (e) {
      alert(e instanceof Error ? e.message : "Failed");
    }
  }

  if (!stats) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-[var(--muted)]">
          {error || "Loading admin panel..."}
        </div>
      </div>
    );
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: "overview", label: "Overview" },
    { key: "users", label: "Users" },
    { key: "providers", label: "Providers" },
    { key: "models", label: "Models" },
    { key: "keys", label: "Provider Keys" },
  ];

  return (
    <div className="min-h-screen">
      <header
        className="border-b px-6 py-3 flex items-center justify-between"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="flex items-center gap-4">
          <a href="/dashboard" className="text-lg font-bold hover:opacity-80">
            UniLLM
          </a>
          <span
            className="text-xs px-2 py-0.5 rounded"
            style={{ background: "var(--danger)", color: "white" }}
          >
            Admin
          </span>
        </div>
      </header>

      <div className="max-w-6xl mx-auto p-6">
        {/* Tabs */}
        <div className="flex gap-2 mb-6 flex-wrap">
          {tabs.map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className="px-4 py-2 rounded-lg text-sm font-medium"
              style={{
                background:
                  tab === t.key ? "var(--primary)" : "var(--card)",
                border: `1px solid ${tab === t.key ? "var(--primary)" : "var(--border)"}`,
              }}
            >
              {t.label}
            </button>
          ))}
        </div>

        {/* Overview */}
        {tab === "overview" && (
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
            <Card label="Total Users" value={stats.total_users.toString()} />
            <Card
              label="Total Requests"
              value={stats.total_requests.toLocaleString()}
            />
            <Card label="Total Cost" value={`$${stats.total_cost.toFixed(4)}`} />
            <Card
              label="Total Tokens"
              value={stats.total_tokens.toLocaleString()}
            />
            <Card label="Active Keys" value={stats.active_keys.toString()} />
          </div>
        )}

        {/* Users */}
        {tab === "users" && (
          <Panel title="USER MANAGEMENT">
            <div className="flex gap-2 mb-4 flex-wrap items-end">
              <Input
                label="User ID"
                value={balUserId}
                onChange={setBalUserId}
                placeholder="ID"
                width="w-20"
              />
              <Input
                label="Amount ($)"
                value={balDelta}
                onChange={setBalDelta}
                placeholder="10"
                width="w-24"
              />
              <Input
                label="Reason"
                value={balReason}
                onChange={setBalReason}
                placeholder="top-up"
                width="flex-1"
              />
              <button
                onClick={handleUpdateBalance}
                className="px-4 py-2 rounded-lg text-sm font-medium text-white"
                style={{ background: "var(--primary)" }}
              >
                Add Balance
              </button>
            </div>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-[var(--muted)]">
                  <th className="pb-2">ID</th>
                  <th className="pb-2">Email</th>
                  <th className="pb-2">Name</th>
                  <th className="pb-2">Role</th>
                  <th className="pb-2 text-right">Balance</th>
                  <th className="pb-2">Joined</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr
                    key={u.id}
                    className="border-t"
                    style={{ borderColor: "var(--border)" }}
                  >
                    <td className="py-2">{u.id}</td>
                    <td className="py-2 font-mono text-xs">{u.email}</td>
                    <td className="py-2">{u.name}</td>
                    <td className="py-2">
                      <span
                        className="text-xs px-1.5 py-0.5 rounded"
                        style={{
                          background:
                            u.role === "admin"
                              ? "rgba(239,68,68,0.15)"
                              : "rgba(59,130,246,0.15)",
                          color:
                            u.role === "admin"
                              ? "var(--danger)"
                              : "var(--primary)",
                        }}
                      >
                        {u.role}
                      </span>
                    </td>
                    <td className="py-2 text-right font-mono">
                      ${u.balance.toFixed(2)}
                    </td>
                    <td className="py-2 text-[var(--muted)] text-xs">
                      {new Date(u.created_at).toLocaleDateString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Panel>
        )}

        {/* Providers */}
        {tab === "providers" && (
          <Panel title="PROVIDERS">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-[var(--muted)]">
                  <th className="pb-2">ID</th>
                  <th className="pb-2">Name</th>
                  <th className="pb-2">Base URL</th>
                  <th className="pb-2">Status</th>
                  <th className="pb-2 text-right">Action</th>
                </tr>
              </thead>
              <tbody>
                {providers.map((p) => (
                  <tr
                    key={p.id}
                    className="border-t"
                    style={{ borderColor: "var(--border)" }}
                  >
                    <td className="py-2">{p.id}</td>
                    <td className="py-2 font-medium">{p.name}</td>
                    <td className="py-2 font-mono text-xs text-[var(--muted)]">
                      {p.base_url}
                    </td>
                    <td className="py-2">
                      <StatusBadge active={p.is_active} />
                    </td>
                    <td className="py-2 text-right">
                      <button
                        onClick={() => handleToggleProvider(p.id, p.is_active)}
                        className="text-xs px-2 py-1 rounded"
                        style={{
                          color: p.is_active
                            ? "var(--danger)"
                            : "var(--success)",
                        }}
                      >
                        {p.is_active ? "Disable" : "Enable"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Panel>
        )}

        {/* Models */}
        {tab === "models" && (
          <Panel title="MODEL CONFIGURATIONS">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-[var(--muted)]">
                  <th className="pb-2">Public Name</th>
                  <th className="pb-2">Upstream Model</th>
                  <th className="pb-2 text-right">Input $/1M</th>
                  <th className="pb-2 text-right">Output $/1M</th>
                  <th className="pb-2 text-right">Max Tokens</th>
                  <th className="pb-2">Status</th>
                  <th className="pb-2 text-right">Action</th>
                </tr>
              </thead>
              <tbody>
                {models.map((m) => (
                  <tr
                    key={m.id}
                    className="border-t"
                    style={{ borderColor: "var(--border)" }}
                  >
                    <td className="py-2 font-mono">{m.public_name}</td>
                    <td className="py-2 font-mono text-xs text-[var(--muted)]">
                      {m.upstream_model}
                    </td>
                    <td className="py-2 text-right">
                      ${m.input_price_per_1m.toFixed(2)}
                    </td>
                    <td className="py-2 text-right">
                      ${m.output_price_per_1m.toFixed(2)}
                    </td>
                    <td className="py-2 text-right">
                      {m.max_tokens.toLocaleString()}
                    </td>
                    <td className="py-2">
                      <StatusBadge active={m.is_active} />
                    </td>
                    <td className="py-2 text-right">
                      <button
                        onClick={() => handleToggleModel(m.id, m.is_active)}
                        className="text-xs px-2 py-1 rounded"
                        style={{
                          color: m.is_active
                            ? "var(--danger)"
                            : "var(--success)",
                        }}
                      >
                        {m.is_active ? "Disable" : "Enable"}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Panel>
        )}

        {/* Provider Keys */}
        {tab === "keys" && (
          <Panel title="UPSTREAM API KEYS">
            <div className="flex gap-2 mb-4 flex-wrap items-end">
              <div className="w-32">
                <label className="block text-xs text-[var(--muted)] mb-1">
                  Provider ID
                </label>
                <select
                  value={newKeyProviderId}
                  onChange={(e) => setNewKeyProviderId(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                  style={{
                    background: "var(--background)",
                    border: "1px solid var(--border)",
                    color: "var(--foreground)",
                  }}
                >
                  <option value="">Select...</option>
                  {providers.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name} ({p.id})
                    </option>
                  ))}
                </select>
              </div>
              <Input
                label="API Key"
                value={newKeyValue}
                onChange={setNewKeyValue}
                placeholder="sk-..."
                width="flex-1"
              />
              <Input
                label="RPM"
                value={newKeyRpm}
                onChange={setNewKeyRpm}
                placeholder="60"
                width="w-20"
              />
              <button
                onClick={handleAddKey}
                className="px-4 py-2 rounded-lg text-sm font-medium text-white"
                style={{ background: "var(--primary)" }}
              >
                Add Key
              </button>
            </div>
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-[var(--muted)]">
                  <th className="pb-2">ID</th>
                  <th className="pb-2">Provider</th>
                  <th className="pb-2">Key Prefix</th>
                  <th className="pb-2 text-right">RPM</th>
                  <th className="pb-2">Status</th>
                </tr>
              </thead>
              <tbody>
                {providerKeys.map((k) => (
                  <tr
                    key={k.id}
                    className="border-t"
                    style={{ borderColor: "var(--border)" }}
                  >
                    <td className="py-2">{k.id}</td>
                    <td className="py-2">
                      {providers.find((p) => p.id === k.provider_id)?.name ||
                        k.provider_id}
                    </td>
                    <td className="py-2 font-mono text-xs text-[var(--muted)]">
                      {k.key_prefix}
                    </td>
                    <td className="py-2 text-right">{k.rpm}</td>
                    <td className="py-2">
                      <StatusBadge active={k.is_active} />
                    </td>
                  </tr>
                ))}
                {providerKeys.length === 0 && (
                  <tr>
                    <td
                      colSpan={5}
                      className="py-4 text-center text-[var(--muted)] text-sm"
                    >
                      No provider keys configured
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </Panel>
        )}
      </div>
    </div>
  );
}

function Card({ label, value }: { label: string; value: string }) {
  return (
    <div
      className="rounded-xl p-4"
      style={{
        background: "var(--card)",
        border: "1px solid var(--border)",
      }}
    >
      <div className="text-xs text-[var(--muted)] mb-1">{label}</div>
      <div className="text-xl font-bold">{value}</div>
    </div>
  );
}

function Panel({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div
      className="rounded-xl p-5"
      style={{
        background: "var(--card)",
        border: "1px solid var(--border)",
      }}
    >
      <h3 className="text-sm font-semibold mb-4 text-[var(--muted)]">
        {title}
      </h3>
      {children}
    </div>
  );
}

function StatusBadge({ active }: { active: boolean }) {
  return (
    <span
      className="text-xs px-1.5 py-0.5 rounded"
      style={{
        background: active ? "rgba(34,197,94,0.15)" : "rgba(239,68,68,0.15)",
        color: active ? "var(--success)" : "var(--danger)",
      }}
    >
      {active ? "active" : "disabled"}
    </span>
  );
}

function Input({
  label,
  value,
  onChange,
  placeholder,
  width,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  placeholder: string;
  width: string;
}) {
  return (
    <div className={width}>
      <label className="block text-xs text-[var(--muted)] mb-1">{label}</label>
      <input
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full px-3 py-2 rounded-lg text-sm outline-none"
        style={{
          background: "var(--background)",
          border: "1px solid var(--border)",
        }}
      />
    </div>
  );
}
