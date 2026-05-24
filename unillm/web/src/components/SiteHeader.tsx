"use client";

import { useEffect, useState, type ReactNode } from "react";
import { isLoggedIn, logout } from "@/lib/api";
import { useI18n } from "@/lib/i18n";
import { PreferencesBar } from "./PreferencesBar";

export type NavKey = "models" | "playground" | "docs" | "calculator";

type SiteHeaderProps = {
  activeNav?: NavKey;
  sticky?: boolean;
  balance?: number | null;
  adminBadge?: boolean;
  trailing?: ReactNode;
};

const navLinkClass =
  "text-sm text-[var(--muted)] hover:text-[var(--foreground)] transition-colors";

export function SiteHeader({
  activeNav,
  sticky = false,
  balance = null,
  adminBadge = false,
  trailing,
}: SiteHeaderProps) {
  const { t } = useI18n();
  const [loggedIn, setLoggedIn] = useState(false);

  useEffect(() => {
    setLoggedIn(isLoggedIn());
  }, []);

  function linkClass(key: NavKey) {
    return activeNav === key
      ? "text-sm text-[var(--foreground)] font-medium"
      : navLinkClass;
  }

  return (
    <header
      className={`border-b px-6 py-3 flex items-center justify-between ${sticky ? "sticky top-0 z-50" : ""}`}
      style={{
        borderColor: "var(--border)",
        background: "var(--background)",
      }}
    >
      <div className="flex items-center gap-4">
        <a
          href="/"
          className="text-lg font-bold hover:opacity-80 transition-opacity"
        >
          UniLLM
        </a>
        {adminBadge && (
          <span
            className="text-xs px-2 py-0.5 rounded font-medium"
            style={{ background: "var(--danger)", color: "#fff" }}
          >
            {t.admin.badge}
          </span>
        )}
        <a href="/models" className={linkClass("models")}>
          {t.nav.models}
        </a>
        <a href="/playground" className={linkClass("playground")}>
          {t.nav.playground}
        </a>
        <a href="/docs" className={linkClass("docs")}>
          {t.nav.docs}
        </a>
        <a href="/calculator" className={linkClass("calculator")}>
          {t.nav.calculator}
        </a>
      </div>

      <div className="flex items-center gap-3">
        {trailing}
        <PreferencesBar />
        {balance !== null && (
          <span className="text-sm text-[var(--muted)]">
            {t.dashboard.balance}: ${balance.toFixed(2)}
          </span>
        )}
        {loggedIn ? (
          <>
            <a href="/dashboard" className={navLinkClass}>
              {t.nav.dashboard}
            </a>
            <button type="button" onClick={logout} className={navLinkClass}>
              {t.nav.logout}
            </button>
          </>
        ) : (
          <a
            href="/login"
            className="text-sm px-4 py-1.5 rounded-lg font-medium"
            style={{ background: "var(--primary)", color: "#fff" }}
          >
            {t.nav.signIn}
          </a>
        )}
      </div>
    </header>
  );
}
