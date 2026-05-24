"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { isLoggedIn, logout, getModelCatalog } from "@/lib/api";

interface ModelInfo {
  id: string;
  vendor: string;
  input_price_per_1m: number;
  output_price_per_1m: number;
  max_tokens: number;
  supports_stream: boolean;
  supports_tools: boolean;
  supports_vision: boolean;
}

const VENDOR_COLORS: Record<string, string> = {
  Anthropic: "#d97706",
  Google: "#2563eb",
  OpenAI: "#10b981",
  DeepSeek: "#8b5cf6",
  Alibaba: "#ff6a00",
  Meta: "#1877f2",
  Mistral: "#f43f5e",
};

function getVendor(model: ModelInfo): string {
  return model.vendor || "Other";
}

function getVendorColor(vendor: string): string {
  return VENDOR_COLORS[vendor] || "#737373";
}

export default function ModelsPage() {
  const router = useRouter();
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [vendorFilter, setVendorFilter] = useState<string>("all");
  const [capFilter, setCapFilter] = useState<string>("all");

  useEffect(() => {
    loadModels();
  }, []);

  async function loadModels() {
    try {
      const data = await getModelCatalog();
      setModels(data.models || []);
    } catch {
      // non-critical, just show empty
    } finally {
      setLoading(false);
    }
  }

  const vendors = Array.from(new Set(models.map(getVendor))).sort();

  const filtered = models.filter((m) => {
    if (search && !m.id.toLowerCase().includes(search.toLowerCase())) return false;
    if (vendorFilter !== "all" && getVendor(m) !== vendorFilter) return false;
    if (capFilter === "vision" && !m.supports_vision) return false;
    if (capFilter === "tools" && !m.supports_tools) return false;
    if (capFilter === "stream" && !m.supports_stream) return false;
    return true;
  });

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header
        className="border-b px-6 py-3 flex items-center justify-between"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="flex items-center gap-4">
          <a href="/dashboard" className="text-lg font-bold hover:opacity-80 transition-opacity">
            UniLLM
          </a>
          <a
            href="/models"
            className="text-sm text-white font-medium"
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
          {isLoggedIn() ? (
            <>
              <a
                href="/dashboard"
                className="text-sm text-[var(--muted)] hover:text-white transition-colors"
              >
                Dashboard
              </a>
              <button
                onClick={logout}
                className="text-sm text-[var(--muted)] hover:text-white"
              >
                Logout
              </button>
            </>
          ) : (
            <a
              href="/login"
              className="text-sm px-4 py-1.5 rounded-lg font-medium"
              style={{ background: "var(--primary)", color: "#fff" }}
            >
              Sign In
            </a>
          )}
        </div>
      </header>

      <div className="max-w-6xl mx-auto p-6">
        {/* Title + Stats */}
        <div className="mb-6">
          <h2 className="text-2xl font-bold mb-1">Model Marketplace</h2>
          <p className="text-[var(--muted)] text-sm">
            {models.length} models available across {vendors.length} vendors.
            All accessible via a single OpenAI-compatible API.
          </p>
        </div>

        {/* Filters */}
        <div className="flex flex-wrap gap-3 mb-6">
          <input
            type="text"
            placeholder="Search models..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="px-3 py-2 rounded-lg text-sm flex-1 min-w-[200px] outline-none"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
              color: "var(--foreground)",
            }}
          />
          <select
            value={vendorFilter}
            onChange={(e) => setVendorFilter(e.target.value)}
            className="px-3 py-2 rounded-lg text-sm outline-none cursor-pointer"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
              color: "var(--foreground)",
            }}
          >
            <option value="all">All Vendors</option>
            {vendors.map((v) => (
              <option key={v} value={v}>
                {v}
              </option>
            ))}
          </select>
          <select
            value={capFilter}
            onChange={(e) => setCapFilter(e.target.value)}
            className="px-3 py-2 rounded-lg text-sm outline-none cursor-pointer"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
              color: "var(--foreground)",
            }}
          >
            <option value="all">All Capabilities</option>
            <option value="vision">Vision</option>
            <option value="tools">Function Calling</option>
            <option value="stream">Streaming</option>
          </select>
        </div>

        {/* Model Cards Grid */}
        {loading ? (
          <div className="text-center py-20 text-[var(--muted)]">Loading models...</div>
        ) : filtered.length === 0 ? (
          <div className="text-center py-20 text-[var(--muted)]">
            No models match your filters.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filtered.map((m) => {
              const vendor = getVendor(m);
              const color = getVendorColor(vendor);
              return (
                <div
                  key={m.id}
                  className="rounded-xl p-5 flex flex-col gap-3 hover:border-[var(--primary)] transition-colors"
                  style={{
                    background: "var(--card)",
                    border: "1px solid var(--border)",
                  }}
                >
                  {/* Model name + vendor badge */}
                  <div className="flex items-start justify-between gap-2">
                    <h3 className="text-base font-semibold leading-tight break-all">
                      {m.id}
                    </h3>
                    <span
                      className="text-xs px-2 py-0.5 rounded-full font-medium shrink-0"
                      style={{
                        background: `${color}20`,
                        color: color,
                        border: `1px solid ${color}40`,
                      }}
                    >
                      {vendor}
                    </span>
                  </div>

                  {/* Pricing */}
                  <div className="flex gap-4 text-sm">
                    <div>
                      <div className="text-[var(--muted)] text-xs mb-0.5">Input</div>
                      <div className="font-mono">
                        ${m.input_price_per_1m.toFixed(2)}
                        <span className="text-[var(--muted)] text-xs"> /1M</span>
                      </div>
                    </div>
                    <div>
                      <div className="text-[var(--muted)] text-xs mb-0.5">Output</div>
                      <div className="font-mono">
                        ${m.output_price_per_1m.toFixed(2)}
                        <span className="text-[var(--muted)] text-xs"> /1M</span>
                      </div>
                    </div>
                    <div>
                      <div className="text-[var(--muted)] text-xs mb-0.5">Max Tokens</div>
                      <div className="font-mono">
                        {m.max_tokens >= 1000
                          ? `${(m.max_tokens / 1000).toFixed(0)}K`
                          : m.max_tokens}
                      </div>
                    </div>
                  </div>

                  {/* Capability badges */}
                  <div className="flex flex-wrap gap-1.5 mt-auto">
                    {m.supports_stream && (
                      <span
                        className="text-xs px-2 py-0.5 rounded"
                        style={{
                          background: "var(--success)15",
                          color: "var(--success)",
                          border: "1px solid var(--success)30",
                        }}
                      >
                        Streaming
                      </span>
                    )}
                    {m.supports_tools && (
                      <span
                        className="text-xs px-2 py-0.5 rounded"
                        style={{
                          background: "var(--primary)15",
                          color: "var(--primary)",
                          border: "1px solid var(--primary)30",
                        }}
                      >
                        Function Calling
                      </span>
                    )}
                    {m.supports_vision && (
                      <span
                        className="text-xs px-2 py-0.5 rounded"
                        style={{
                          background: "var(--warning)15",
                          color: "var(--warning)",
                          border: "1px solid var(--warning)30",
                        }}
                      >
                        Vision
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {/* API Example */}
        <div
          className="mt-8 rounded-xl p-5"
          style={{
            background: "var(--card)",
            border: "1px solid var(--border)",
          }}
        >
          <h3 className="text-sm font-semibold mb-3">Quick Start</h3>
          <pre
            className="text-xs overflow-x-auto p-4 rounded-lg"
            style={{ background: "var(--background)" }}
          >
{`curl https://your-domain.com/v1/chat/completions \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "${filtered[0]?.id || "claude-sonnet-4-6"}",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`}
          </pre>
        </div>
      </div>
    </div>
  );
}
