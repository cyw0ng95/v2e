import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "@/lib/providers";
import { Toaster } from "@/components/ui/sonner";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "v2e - CVE Management",
  description: "CVE (Common Vulnerabilities and Exposures) data management system",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased transition-colors duration-300 ease-in-out`}
      >
        <Providers>
          <div className="min-h-screen min-w-screen flex flex-col bg-background">
            <main className="w-full h-full overflow-auto">{children}</main>
            <Toaster />
          </div>
        </Providers>
      </body>
    </html>
  );
}
