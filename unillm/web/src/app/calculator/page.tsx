"use client";

import { useEffect, useState } from "react";
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

function getVendorColor(vendor: string): string {
  return VENDOR_COLORS[vendor] || "#737373";
}

function calcCostPerRequest(
  model: ModelInfo,
  inputTokens: number,
  outputTokens: number
): number {
  return (
    (inputTokens / 1_000_000) * model.input_price_per_1m +
    (outputTokens / 1_000_000) * model.output_price_per_1m
  );
}

function formatCost(cost: number): string {
  if (cost < 0.0001) return `$${cost.toExponential(2)}`;
  if (cost < 0.01) return `$${cost.toFixed(6)}`;
  if (cost < 1) return `$${cost.toFixed(4)}`;
  return `$${cost.toFixed(2)}`;
}

export default function CalculatorPage() {
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedModelId, setSelectedModelId] = useState<string>("");
  const [inputTokens, setInputTokens] = useState(500);
  const [outputTokens, setOutputTokens] = useState(200);
  const [requestsPerDay, setRequestsPerDay] = useState(100);

  useEffect(() => {
    loadModels();
  }, []);

  async function loadModels() {
    try {
      const data = await getModelCatalog();
      const list: ModelInfo[] = data.models || [];
      setModels(list);
      if (list.length > 0) {
        setSelectedModelId(list[0].id);
      }
    } catch {
      // non-critical
    } finally {
      setLoading(false);
    }
  }

  const selectedModel = models.find((m) => m.id === selectedModelId);

  const costPerRequest = selectedModel
    ? calcCostPerRequest(selectedModel, inputTokens, outputTokens)
    : 0;
  const dailyCost = costPerRequest * requestsPerDay;
  const monthlyCost = dailyCost * 30;
  const yearlyCost = dailyCost * 365;

  const sortedModels = [...models]
    .map((m) => ({
      ...m,
      costPerReq: calcCostPerRequest(m, inputTokens, outputTokens),
    }))
    .sort((a, b) => a.costPerReq - b.costPerReq);

  return (
    <div className="min-h-screen">
      {/* Header */}
      <header
        className="border-b px-6 py-3 flex items-center justify-between"
        style={{ borderColor: "var(--border)" }}
      >
        <div className="flex items-center gap-4">
          <a
            href="/dashboard"
            className="text-lg font-bold hover:opacity-80 transition-opacity"
          >
            UniLLM
          </a>
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
          <a href="/calculator" className="text-sm text-white font-medium">
            Calculator
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
        {/* Title */}
        <div className="mb-6">
          <h2 className="text-2xl font-bold mb-1">Cost Calculator</h2>
          <p className="text-[var(--muted)] text-sm">
            Estimate your API costs based on expected usage
          </p>
        </div>

        {loading ? (
          <div className="text-center py-20 text-[var(--muted)]">
            Loading models...
          </div>
        ) : (
          <>
            {/* Calculator Card */}
            <div
              className="rounded-xl p-6 mb-6"
              style={{
                background: "var(--card)",
                border: "1px solid var(--border)",
              }}
            >
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
                {/* Model selector */}
                <div>
                  <label className="text-xs text-[var(--muted)] mb-1.5 block">
                    Model
                  </label>
                  <select
                    value={selectedModelId}
                    onChange={(e) => setSelectedModelId(e.target.value)}
                    className="w-full px-3 py-2 rounded-lg text-sm outline-none cursor-pointer"
                    style={{
                      background: "var(--background)",
                      border: "1px solid var(--border)",
                      color: "var(--foreground)",
                    }}
                  >
                    {models.map((m) => (
                      <option key={m.id} value={m.id}>
                        {m.id} ({m.vendor})
                      </option>
                    ))}
                  </select>
                </div>

                {/* Input tokens */}
                <div>
                  <label className="text-xs text-[var(--muted)] mb-1.5 block">
                    Input tokens per request
                  </label>
                  <input
                    type="number"
                    min={0}
                    value={inputTokens}
                    onChange={(e) =>
                      setInputTokens(Math.max(0, Number(e.target.value)))
                    }
                    className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                    style={{
                      background: "var(--background)",
                      border: "1px solid var(--border)",
                      color: "var(--foreground)",
                    }}
                  />
                </div>

                {/* Output tokens */}
                <div>
                  <label className="text-xs text-[var(--muted)] mb-1.5 block">
                    Output tokens per request
                  </label>
                  <input
                    type="number"
                    min={0}
                    value={outputTokens}
                    onChange={(e) =>
                      setOutputTokens(Math.max(0, Number(e.target.value)))
                    }
                    className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                    style={{
                      background: "var(--background)",
                      border: "1px solid var(--border)",
                      color: "var(--foreground)",
                    }}
                  />
                </div>

                {/* Requests per day */}
                <div>
                  <label className="text-xs text-[var(--muted)] mb-1.5 block">
                    Requests per day
                  </label>
                  <input
                    type="number"
                    min={0}
                    value={requestsPerDay}
                    onChange={(e) =>
                      setRequestsPerDay(Math.max(0, Number(e.target.value)))
                    }
                    className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                    style={{
                      background: "var(--background)",
                      border: "1px solid var(--border)",
                      color: "var(--foreground)",
                    }}
                  />
                </div>
              </div>

              {/* Results */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div
                  className="rounded-lg p-4"
                  style={{ background: "var(--background)" }}
                >
                  <div className="text-xs text-[var(--muted)] mb-1">
                    Cost per Request
                  </div>
                  <div className="text-lg font-mono font-semibold">
                    {formatCost(costPerRequest)}
                  </div>
                </div>
                <div
                  className="rounded-lg p-4"
                  style={{ background: "var(--background)" }}
                >
                  <div className="text-xs text-[var(--muted)] mb-1">
                    Daily Cost
                  </div>
                  <div className="text-lg font-mono font-semibold">
                    {formatCost(dailyCost)}
                  </div>
                </div>
                <div
                  className="rounded-lg p-4"
                  style={{ background: "var(--background)" }}
                >
                  <div className="text-xs text-[var(--muted)] mb-1">
                    Monthly Cost (30d)
                  </div>
                  <div className="text-lg font-mono font-semibold text-[var(--primary)]">
                    {formatCost(monthlyCost)}
                  </div>
                </div>
                <div
                  className="rounded-lg p-4"
                  style={{ background: "var(--background)" }}
                >
                  <div className="text-xs text-[var(--muted)] mb-1">
                    Yearly Cost (365d)
                  </div>
                  <div className="text-lg font-mono font-semibold">
                    {formatCost(yearlyCost)}
                  </div>
                </div>
              </div>
            </div>

            {/* Comparison Table */}
            <div
              className="rounded-xl overflow-hidden"
              style={{
                background: "var(--card)",
                border: "1px solid var(--border)",
              }}
            >
              <div className="px-5 py-4">
                <h3 className="text-sm font-semibold">
                  Model Comparison (sorted by cost)
                </h3>
              </div>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr
                      className="text-xs text-[var(--muted)] text-left"
                      style={{
                        borderTop: "1px solid var(--border)",
                        borderBottom: "1px solid var(--border)",
                      }}
                    >
                      <th className="px-5 py-2.5 font-medium">Model</th>
                      <th className="px-5 py-2.5 font-medium">Vendor</th>
                      <th className="px-5 py-2.5 font-medium text-right">
                        Input $/1M
                      </th>
                      <th className="px-5 py-2.5 font-medium text-right">
                        Output $/1M
                      </th>
                      <th className="px-5 py-2.5 font-medium text-right">
                        Cost/Request
                      </th>
                      <th className="px-5 py-2.5 font-medium text-right">
                        Daily Cost
                      </th>
                      <th className="px-5 py-2.5 font-medium text-right">
                        Monthly Cost
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {sortedModels.map((m) => {
                      const isSelected = m.id === selectedModelId;
                      const daily = m.costPerReq * requestsPerDay;
                      const monthly = daily * 30;
                      const vendorColor = getVendorColor(m.vendor || "Other");
                      return (
                        <tr
                          key={m.id}
                          className="cursor-pointer hover:bg-[var(--background)] transition-colors"
                          style={{
                            borderBottom: "1px solid var(--border)",
                            background: isSelected
                              ? "var(--primary)10"
                              : undefined,
                            borderLeft: isSelected
                              ? "3px solid var(--primary)"
                              : "3px solid transparent",
                          }}
                          onClick={() => setSelectedModelId(m.id)}
                        >
                          <td className="px-5 py-2.5 font-medium">{m.id}</td>
                          <td className="px-5 py-2.5">
                            <span
                              className="text-xs px-2 py-0.5 rounded-full font-medium"
                              style={{
                                background: `${vendorColor}20`,
                                color: vendorColor,
                                border: `1px solid ${vendorColor}40`,
                              }}
                            >
                              {m.vendor || "Other"}
                            </span>
                          </td>
                          <td className="px-5 py-2.5 text-right font-mono">
                            ${m.input_price_per_1m.toFixed(2)}
                          </td>
                          <td className="px-5 py-2.5 text-right font-mono">
                            ${m.output_price_per_1m.toFixed(2)}
                          </td>
                          <td className="px-5 py-2.5 text-right font-mono">
                            {formatCost(m.costPerReq)}
                          </td>
                          <td className="px-5 py-2.5 text-right font-mono">
                            {formatCost(daily)}
                          </td>
                          <td className="px-5 py-2.5 text-right font-mono">
                            {formatCost(monthly)}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
