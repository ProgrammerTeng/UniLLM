export type Locale = "en" | "zh-CN";

export const LOCALE_STORAGE_KEY = "unillm-locale";
export const DEFAULT_LOCALE: Locale = "en";

export const LOCALE_LABELS: Record<Locale, string> = {
  en: "EN",
  "zh-CN": "中文",
};
