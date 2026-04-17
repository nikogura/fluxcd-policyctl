import type { Metadata } from "next";

import "./globals.css";

export const metadata: Metadata = {
  title: "flux-policyctl",
  description: "Flux CD Image Policy Control",
};

export default function RootLayout({
  children,
}: {
  readonly children: React.ReactNode;
}): React.ReactElement {
  return (
    <html lang="en" data-theme="dark" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{
          __html: `
            (function() {
              var theme = localStorage.getItem('theme') || 'dark';
              document.documentElement.setAttribute('data-theme', theme);
            })();
          `
        }} />
      </head>
      <body>{children}</body>
    </html>
  );
}
