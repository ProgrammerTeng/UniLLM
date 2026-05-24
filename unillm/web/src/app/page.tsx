"use client";

import { useEffect, useState } from "react";
import { getModelCatalog } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { SiteHeader } from "@/components/SiteHeader";

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

export default function Home() {
  const { t } = useI18n();
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [modelsLoaded, setModelsLoaded] = useState(false);

  useEffect(() => {
    loadModels();
  }, []);

  async function loadModels() {
    try {
      const data = await getModelCatalog();
      setModels(data.models || []);
    } catch {
      // non-critical
    } finally {
      setModelsLoaded(true);
    }
  }

  const previewModels = models.slice(0, 6);

  const features = [
    { key: "openai" as const, icon: "{}" },
    { key: "pricing" as const, icon: "$" },
    { key: "multiVendor" as const, icon: "\u2194" },
    { key: "failover" as const, icon: "\u21bb" },
    { key: "dashboard" as const, icon: "\u25ce" },
    { key: "devFirst" as const, icon: ">_" },
  ];

  const stats = [
    { label: t.home.statsModels, value: "7+" },
    { label: t.home.statsVendors, value: "3" },
    { label: t.home.statsUptime, value: "99.9%" },
    { label: t.home.statsLatency, value: "<1s" },
  ];

  return (
    <div className="min-h-screen">
      <SiteHeader sticky />

      <section className="max-w-5xl mx-auto px-6 pt-20 pb-16 text-center">
        <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6 leading-tight">
          {t.home.heroTitle}{" "}
          <span style={{ color: "var(--primary)" }}>
            {t.home.heroTitleHighlight}
          </span>
        </h1>
        <p className="text-lg md:text-xl text-[var(--muted)] max-w-2xl mx-auto mb-10 leading-relaxed">
          {t.home.heroSubtitle}
        </p>
        <div className="flex items-center justify-center gap-4 flex-wrap mb-14">
          <a
            href="/login"
            className="px-6 py-3 rounded-lg font-medium text-white text-sm transition-colors"
            style={{ background: "var(--primary)" }}
            onMouseEnter={(e) =>
              (e.currentTarget.style.background = "var(--primary-hover)")
            }
            onMouseLeave={(e) =>
              (e.currentTarget.style.background = "var(--primary)")
            }
          >
            {t.home.getStarted}
          </a>
          <a
            href="/models"
            className="px-6 py-3 rounded-lg font-medium text-sm transition-colors"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
              color: "var(--foreground)",
            }}
          >
            {t.home.viewModels}
          </a>
        </div>

        <div className="max-w-2xl mx-auto text-left">
          <div
            className="rounded-xl overflow-hidden"
            style={{ border: "1px solid var(--border)" }}
          >
            <div
              className="px-4 py-2 text-xs font-mono text-[var(--muted)] flex items-center gap-2"
              style={{
                background: "var(--card)",
                borderBottom: "1px solid var(--border)",
              }}
            >
              <span
                className="w-2.5 h-2.5 rounded-full inline-block"
                style={{ background: "#ef4444" }}
              />
              <span
                className="w-2.5 h-2.5 rounded-full inline-block"
                style={{ background: "#f59e0b" }}
              />
              <span
                className="w-2.5 h-2.5 rounded-full inline-block"
                style={{ background: "#22c55e" }}
              />
              <span className="ml-2">{t.home.terminal}</span>
            </div>
            <pre
              className="text-xs md:text-sm p-5 overflow-x-auto font-mono leading-relaxed"
              style={{ background: "var(--background)" }}
            >
              <span style={{ color: "var(--success)" }}>curl</span>
              {" https://api.unillm.dev/v1/chat/completions \\\n"}
              {"  -H "}
              <span style={{ color: "#d97706" }}>
                {'"Authorization: Bearer YOUR_API_KEY"'}
              </span>
              {" \\\n"}
              {"  -H "}
              <span style={{ color: "#d97706" }}>
                {'"Content-Type: application/json"'}
              </span>
              {" \\\n"}
              {"  -d "}
              <span style={{ color: "#d97706" }}>{"'"}</span>
              <span style={{ color: "var(--foreground)" }}>
                {'{\n    "model": "claude-sonnet-4-20250514",\n    "messages": [{"role": "user", "content": "Hello!"}]\n  }'}
              </span>
              <span style={{ color: "#d97706" }}>{"'"}</span>
            </pre>
          </div>
        </div>
      </section>

      <section
        className="border-y"
        style={{ borderColor: "var(--border)", background: "var(--card)" }}
      >
        <div className="max-w-5xl mx-auto px-6 py-8 grid grid-cols-2 md:grid-cols-4 gap-6 text-center">
          {stats.map((s) => (
            <div key={s.label}>
              <div
                className="text-3xl font-bold mb-1"
                style={{ color: "var(--primary)" }}
              >
                {s.value}
              </div>
              <div className="text-sm text-[var(--muted)]">{s.label}</div>
            </div>
          ))}
        </div>
      </section>

      <section className="max-w-5xl mx-auto px-6 py-20">
        <h2 className="text-2xl md:text-3xl font-bold text-center mb-3">
          {t.home.featuresTitle}
        </h2>
        <p className="text-[var(--muted)] text-center mb-12 max-w-xl mx-auto">
          {t.home.featuresSubtitle}
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {features.map((f) => {
            const feat = t.home.features[f.key];
            return (
              <div
                key={f.key}
                className="rounded-xl p-6 transition-colors"
                style={{
                  background: "var(--card)",
                  border: "1px solid var(--border)",
                }}
              >
                <div
                  className="w-10 h-10 rounded-lg flex items-center justify-center text-sm font-bold mb-4"
                  style={{
                    background: "var(--primary)15",
                    color: "var(--primary)",
                    border: "1px solid var(--primary)30",
                  }}
                >
                  {f.icon}
                </div>
                <h3 className="text-base font-semibold mb-2">{feat.title}</h3>
                <p className="text-sm text-[var(--muted)] leading-relaxed">
                  {feat.description}
                </p>
              </div>
            );
          })}
        </div>
      </section>

      <section className="border-y" style={{ borderColor: "var(--border)" }}>
        <div className="max-w-5xl mx-auto px-6 py-20">
          <h2 className="text-2xl md:text-3xl font-bold text-center mb-3">
            {t.home.modelsTitle}
          </h2>
          <p className="text-[var(--muted)] text-center mb-12 max-w-xl mx-auto">
            {t.home.modelsSubtitle}
          </p>

          {!modelsLoaded ? (
            <div className="text-center py-12 text-[var(--muted)]">
              {t.home.loadingModels}
            </div>
          ) : previewModels.length === 0 ? (
            <div className="text-center py-12 text-[var(--muted)]">
              {t.home.noModels}
            </div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
                {previewModels.map((m) => {
                  const vendor = m.vendor || "Other";
                  const color = getVendorColor(vendor);
                  return (
                    <div
                      key={m.id}
                      className="rounded-xl p-5 flex flex-col gap-3"
                      style={{
                        background: "var(--card)",
                        border: "1px solid var(--border)",
                      }}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <h3 className="text-sm font-semibold leading-tight break-all font-mono">
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
                      <div className="flex gap-4 text-sm">
                        <div>
                          <div className="text-[var(--muted)] text-xs mb-0.5">
                            {t.home.input}
                          </div>
                          <div className="font-mono">
                            ${m.input_price_per_1m.toFixed(2)}
                            <span className="text-[var(--muted)] text-xs">
                              {" "}
                              {t.home.perMillion}
                            </span>
                          </div>
                        </div>
                        <div>
                          <div className="text-[var(--muted)] text-xs mb-0.5">
                            {t.home.output}
                          </div>
                          <div className="font-mono">
                            ${m.output_price_per_1m.toFixed(2)}
                            <span className="text-[var(--muted)] text-xs">
                              {" "}
                              {t.home.perMillion}
                            </span>
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
              <div className="text-center">
                <a
                  href="/models"
                  className="text-sm font-medium transition-colors"
                  style={{ color: "var(--primary)" }}
                >
                  {t.home.viewAllModels} &rarr;
                </a>
              </div>
            </>
          )}
        </div>
      </section>

      <section className="max-w-5xl mx-auto px-6 py-20">
        <h2 className="text-2xl md:text-3xl font-bold text-center mb-3">
          {t.home.pricingTitle}
        </h2>
        <p className="text-[var(--muted)] text-center mb-12 max-w-xl mx-auto">
          {t.home.pricingSubtitle}
        </p>

        <div
          className="rounded-xl overflow-hidden mb-10"
          style={{
            background: "var(--card)",
            border: "1px solid var(--border)",
          }}
        >
          <table className="w-full text-sm">
            <thead>
              <tr
                className="text-left text-[var(--muted)]"
                style={{ borderBottom: "1px solid var(--border)" }}
              >
                <th className="px-5 py-4 font-medium">{t.home.pricingModel}</th>
                <th className="px-5 py-4 font-medium text-right">
                  {t.home.pricingInput}
                </th>
                <th className="px-5 py-4 font-medium text-right">
                  {t.home.pricingOutput}
                </th>
              </tr>
            </thead>
            <tbody>
              {(models.length > 0 ? models.slice(0, 6) : []).map((m, i) => (
                <tr
                  key={m.id}
                  style={{
                    borderTop: i > 0 ? "1px solid var(--border)" : undefined,
                  }}
                >
                  <td className="px-5 py-3.5">
                    <span className="font-mono text-sm">{m.id}</span>
                    <span
                      className="ml-2 text-xs px-1.5 py-0.5 rounded"
                      style={{
                        color: getVendorColor(m.vendor || "Other"),
                        opacity: 0.8,
                      }}
                    >
                      {m.vendor || "Other"}
                    </span>
                  </td>
                  <td className="px-5 py-3.5 text-right font-mono">
                    ${m.input_price_per_1m.toFixed(2)}
                  </td>
                  <td className="px-5 py-3.5 text-right font-mono">
                    ${m.output_price_per_1m.toFixed(2)}
                  </td>
                </tr>
              ))}
              {models.length === 0 && (
                <tr>
                  <td
                    colSpan={3}
                    className="px-5 py-8 text-center text-[var(--muted)]"
                  >
                    {t.home.pricingLoading}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="text-center">
          <a
            href="/login"
            className="inline-block px-6 py-3 rounded-lg font-medium text-white text-sm transition-colors"
            style={{ background: "var(--primary)" }}
            onMouseEnter={(e) =>
              (e.currentTarget.style.background = "var(--primary-hover)")
            }
            onMouseLeave={(e) =>
              (e.currentTarget.style.background = "var(--primary)")
            }
          >
            {t.home.startFreeCredit}
          </a>
        </div>
      </section>

      <footer
        className="border-t px-6 py-8 text-center"
        style={{ borderColor: "var(--border)" }}
      >
        <p className="text-sm text-[var(--muted)]">
          &copy; 2026 UniLLM. {t.home.footer}
        </p>
      </footer>
    </div>
  );
}
