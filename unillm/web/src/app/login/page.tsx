"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login, register } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { PreferencesBar } from "@/components/PreferencesBar";

export default function LoginPage() {
  const router = useRouter();
  const { t } = useI18n();
  const [isRegister, setIsRegister] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (isRegister) {
        await register(email, password, name);
      } else {
        await login(email, password);
      }
      router.push("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : t.login.failed);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex flex-col">
      <div className="flex justify-end p-4">
        <PreferencesBar />
      </div>

      <div className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <h1 className="text-3xl font-bold mb-2">UniLLM</h1>
            <p className="text-[var(--muted)]">{t.login.tagline}</p>
          </div>

          <div
            className="rounded-xl p-6"
            style={{
              background: "var(--card)",
              border: "1px solid var(--border)",
            }}
          >
            <h2 className="text-xl font-semibold mb-6">
              {isRegister ? t.login.createAccount : t.login.signIn}
            </h2>

            {error && (
              <div
                className="mb-4 p-3 rounded-lg text-sm"
                style={{
                  background: "rgba(239,68,68,0.1)",
                  color: "var(--danger)",
                }}
              >
                {error}
              </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              {isRegister && (
                <div>
                  <label className="block text-sm mb-1.5 text-[var(--muted)]">
                    {t.login.name}
                  </label>
                  <input
                    type="text"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required={isRegister}
                    className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                    style={{
                      background: "var(--background)",
                      border: "1px solid var(--border)",
                      color: "var(--foreground)",
                    }}
                    placeholder={t.login.namePlaceholder}
                  />
                </div>
              )}

              <div>
                <label className="block text-sm mb-1.5 text-[var(--muted)]">
                  {t.login.email}
                </label>
                <input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                  style={{
                    background: "var(--background)",
                    border: "1px solid var(--border)",
                    color: "var(--foreground)",
                  }}
                  placeholder={t.login.emailPlaceholder}
                />
              </div>

              <div>
                <label className="block text-sm mb-1.5 text-[var(--muted)]">
                  {t.login.password}
                </label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={8}
                  className="w-full px-3 py-2 rounded-lg text-sm outline-none"
                  style={{
                    background: "var(--background)",
                    border: "1px solid var(--border)",
                    color: "var(--foreground)",
                  }}
                  placeholder={t.login.passwordPlaceholder}
                />
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full py-2.5 rounded-lg font-medium text-sm text-white transition-colors"
                style={{
                  background: loading ? "var(--muted)" : "var(--primary)",
                }}
              >
                {loading
                  ? t.login.loading
                  : isRegister
                    ? t.login.createAccount
                    : t.login.signIn}
              </button>
            </form>

            <div className="mt-4 text-center text-sm text-[var(--muted)]">
              {isRegister ? t.login.hasAccount : t.login.noAccount}{" "}
              <button
                type="button"
                onClick={() => {
                  setIsRegister(!isRegister);
                  setError("");
                }}
                className="underline"
                style={{ color: "var(--primary)" }}
              >
                {isRegister ? t.login.signIn : t.login.register}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
