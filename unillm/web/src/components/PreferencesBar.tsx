"use client";

import { ThemeToggle } from "./ThemeToggle";
import { LocaleSwitcher } from "./LocaleSwitcher";

/** Theme + locale controls (placed left of Sign In or on login page). */
export function PreferencesBar() {
  return (
    <div className="flex items-center gap-2">
      <ThemeToggle />
      <LocaleSwitcher />
    </div>
  );
}
