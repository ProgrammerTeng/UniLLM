"use client";

import { useI18n } from "@/lib/i18n";
import { LOCALE_LABELS, type Locale } from "@/lib/i18n/locales";

const LOCALES: Locale[] = ["en", "zh-CN"];

export function LocaleSwitcher() {
  const { locale, setLocale, t } = useI18n();

  return (
    <div
      className="flex rounded-lg overflow-hidden text-xs font-medium"
      style={{ border: "1px solid var(--border)" }}
      role="group"
      aria-label={t.locale.label}
    >
      {LOCALES.map((loc) => {
        const active = locale === loc;
        return (
          <button
            key={loc}
            type="button"
            onClick={() => setLocale(loc)}
            className="px-2.5 py-1.5 transition-colors"
            style={{
              background: active ? "var(--primary)" : "var(--card)",
              color: active ? "#fff" : "var(--muted)",
            }}
          >
            {LOCALE_LABELS[loc]}
          </button>
        );
      })}
    </div>
  );
}
