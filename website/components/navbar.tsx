'use client';

import { ShieldIcon, Search, Settings, Sun, Moon, ExternalLink, Github } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useTheme } from 'next-themes';
import { useEffect, useState } from 'react';

export function Navbar() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  return (
    <nav className="sticky top-0 z-50 w-full border-b border-border/40 bg-background/80 backdrop-blur-md supports-[backdrop-filter]:bg-background/60">
      <div className="flex h-[var(--app-header-height)] items-center justify-between px-4 sm:px-6">
        {/* Logo and brand */}
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-primary/10 to-primary/5 border border-primary/20">
            <ShieldIcon className="h-4.5 w-4.5 text-primary" />
          </div>
          <div className="flex flex-col">
            <span className="text-sm font-bold tracking-tight text-foreground">v2e</span>
            <span className="text-[10px] text-muted-foreground/80 uppercase tracking-wide">CVE Management</span>
          </div>
        </div>

        {/* Search bar - desktop only */}
        <div className="hidden md:flex flex-1 max-w-lg mx-6 lg:mx-8">
          <div className="relative w-full group">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground/70 group-focus-within:text-primary transition-colors duration-200" />
            <input
              type="search"
              placeholder="Search CVEs, CWEs, CAPECs..."
              className="h-9 w-full rounded-lg border border-border/60 bg-muted/30 px-9 pr-4 text-sm placeholder:text-muted-foreground/60 outline-none focus:border-primary/60 focus:bg-background focus:ring-4 focus:ring-primary/8 transition-all duration-200"
              aria-label="Search CVEs, CWEs, CAPECs"
            />
          </div>
        </div>

        {/* Right side actions */}
        <div className="flex items-center gap-1.5">
          {/* External links */}
          <Button
            variant="ghost"
            size="icon-sm"
            className="h-8 w-8 text-muted-foreground hover:text-foreground"
            asChild
          >
            <a
              href="https://github.com/cyw0ng95/v2e"
              target="_blank"
              rel="noopener noreferrer"
              aria-label="View on GitHub"
            >
              <Github className="h-4 w-4" />
              <span className="sr-only">GitHub</span>
            </a>
          </Button>

          <div className="h-5 w-px bg-border/40 mx-0.5" />

          {/* Theme toggle */}
          {mounted && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
              className="h-8 w-8 text-muted-foreground hover:text-foreground"
              aria-label="Toggle theme between light and dark mode"
            >
              {theme === 'dark' ? (
                <Sun className="h-4 w-4" />
              ) : (
                <Moon className="h-4 w-4" />
              )}
              <span className="sr-only">Toggle theme</span>
            </Button>
          )}

          {/* Settings */}
          <Button variant="ghost" size="icon-sm" className="h-8 w-8 text-muted-foreground hover:text-foreground" aria-label="Open settings">
            <Settings className="h-4 w-4" />
            <span className="sr-only">Settings</span>
          </Button>
        </div>
      </div>
    </nav>
  );
}
