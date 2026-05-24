"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { login, register } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
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
      setError(err instanceof Error ? err.message : "Failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold mb-2">UniLLM</h1>
          <p className="text-[var(--muted)]">
            One API for all leading AI models
          </p>
        </div>

        <div
          className="rounded-xl p-6"
          style={{
            background: "var(--card)",
            border: "1px solid var(--border)",
          }}
        >
          <h2 className="text-xl font-semibold mb-6">
            {isRegister ? "Create Account" : "Sign In"}
          </h2>

          {error && (
            <div
              className="mb-4 p-3 rounded-lg text-sm"
              style={{ background: "rgba(239,68,68,0.1)", color: "var(--danger)" }}
            >
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            {isRegister && (
              <div>
                <label className="block text-sm mb-1.5 text-[var(--muted)]">
                  Name
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
                  }}
                  placeholder="Your name"
                />
              </div>
            )}

            <div>
              <label className="block text-sm mb-1.5 text-[var(--muted)]">
                Email
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
                }}
                placeholder="you@example.com"
              />
            </div>

            <div>
              <label className="block text-sm mb-1.5 text-[var(--muted)]">
                Password
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
                }}
                placeholder="Min 8 characters"
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
                ? "..."
                : isRegister
                  ? "Create Account"
                  : "Sign In"}
            </button>
          </form>

          <div className="mt-4 text-center text-sm text-[var(--muted)]">
            {isRegister ? "Already have an account?" : "Don't have an account?"}{" "}
            <button
              onClick={() => {
                setIsRegister(!isRegister);
                setError("");
              }}
              className="underline"
              style={{ color: "var(--primary)" }}
            >
              {isRegister ? "Sign In" : "Register"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
