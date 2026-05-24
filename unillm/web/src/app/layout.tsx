import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "UniLLM - Multi-Model AI API Platform",
  description: "One API for all leading AI models",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
