"use client";

import { useTheme } from "next-themes";
import { useEffect, useState } from "react";
import { useI18n } from "@/lib/i18n";

export function ThemeToggle() {
  const { resolvedTheme, setTheme } = useTheme();
  const { t } = useI18n();
  const [mounted, setMounted] = useState(false);

  useEffect(() => setMounted(true), []);

  if (!mounted) {
    return (
      <button
        type="button"
        className="w-9 h-9 rounded-lg"
        style={{ border: "1px solid var(--border)" }}
        aria-hidden
      />
    );
  }

  const isDark = resolvedTheme !== "light";

  return (
    <button
      type="button"
      onClick={() => setTheme(isDark ? "light" : "dark")}
      className="w-9 h-9 rounded-lg flex items-center justify-center text-base transition-colors hover:opacity-80"
      style={{
        border: "1px solid var(--border)",
        background: "var(--card)",
        color: "var(--foreground)",
      }}
      aria-label={isDark ? t.theme.switchToLight : t.theme.switchToDark}
      title={isDark ? t.theme.switchToLight : t.theme.switchToDark}
    >
      {isDark ? "☀" : "☾"}
    </button>
  );
}
