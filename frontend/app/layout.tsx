import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "FinOps AI - Cloud Cost Optimization",
  description: "AI-powered cloud cost optimization and FinOps intelligence platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
