'use client';

import * as React from 'react';
import { Database, Shield, AlertTriangle, Zap, LayoutDashboard, Brain, BookOpen, Book, Network, GitBranch, FileCode, CheckCircle, Bookmark, Activity, RefreshCw, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useViewLearnMode } from '@/contexts/ViewLearnContext';

interface NavItem {
  id: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  badge?: string;
}

interface NavGroup {
  title: string;
  items: NavItem[];
}

const navigation: NavGroup[] = [
  {
    title: 'Database',
    items: [
      { id: 'cve', label: 'CVE', icon: Database, badge: 'count' },
      { id: 'cwe', label: 'CWE', icon: Shield },
      { id: 'capec', label: 'CAPEC', icon: AlertTriangle },
      { id: 'attack', label: 'ATT&CK', icon: Zap },
      { id: 'cce', label: 'CCE', icon: CheckCircle },
    ],
  },
  {
    title: 'Learning',
    items: [
      { id: 'notes-dashboard', label: 'Dashboard', icon: LayoutDashboard },
      { id: 'study-cards', label: 'Study Cards', icon: Brain },
      { id: 'learning-cve', label: 'Learn CVE', icon: BookOpen },
      { id: 'learning-cwe', label: 'Learn CWE', icon: Book },
    ],
  },
  {
    title: 'Analysis',
    items: [
      { id: 'graph', label: 'Graph Analysis', icon: Network },
      { id: 'cweviews', label: 'CWE Views', icon: GitBranch },
      { id: 'ssg', label: 'SSG Guides', icon: FileCode },
      { id: 'asvs', label: 'ASVS', icon: CheckCircle },
    ],
  },
  {
    title: 'System',
    items: [
      { id: 'bookmarks', label: 'Bookmarks', icon: Bookmark },
      { id: 'sysmon', label: 'System Monitor', icon: Activity },
      { id: 'etl', label: 'ETL Status', icon: RefreshCw },
    ],
  },
];

interface SidebarProps {
  activeItem: string;
  onItemClick: (itemId: string) => void;
  className?: string;
}

export function Sidebar({ activeItem, onItemClick, className }: SidebarProps) {
  const { mode, setMode } = useViewLearnMode();
  const [searchQuery, setSearchQuery] = React.useState('');
  const [collapsed, setCollapsed] = React.useState(false);

  const filteredNavigation = navigation.map(group => ({
    ...group,
    items: group.items.filter(item => 
      item.label.toLowerCase().includes(searchQuery.toLowerCase())
    )
  })).filter(group => group.items.length > 0);

  const toggleMode = () => {
    setMode(mode === 'view' ? 'learn' : 'view');
  };

  return (
    <div className={`flex flex-col h-full bg-sidebar border-r border-sidebar-border ${className}`}>
      {/* Header */}
      <div className="p-4 border-b border-sidebar-border">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-sidebar-foreground">
            {collapsed ? 'v2e' : 'v2e Security'}
          </h2>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={() => setCollapsed(!collapsed)}
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            {collapsed ? '»' : '«'}
          </Button>
        </div>
        
        {/* Search */}
        {!collapsed && (
          <div className="mb-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                type="search"
                placeholder="Search..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 pr-4 py-2 w-full"
              />
            </div>
          </div>
        )}

        {/* View/Learn Toggle */}
        {!collapsed && (
          <div className="flex gap-2">
            <Button
              variant={mode === 'view' ? 'default' : 'outline'}
              size="sm"
              className="flex-1"
              onClick={() => setMode('view')}
              aria-pressed={mode === 'view'}
            >
              View
            </Button>
            <Button
              variant={mode === 'learn' ? 'default' : 'outline'}
              size="sm"
              className="flex-1"
              onClick={() => setMode('learn')}
              aria-pressed={mode === 'learn'}
            >
              Learn
            </Button>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto py-2">
        {filteredNavigation.map((group) => (
          !collapsed ? (
            <div key={group.title} className="mb-6">
              <h3 className="px-4 text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                {group.title}
              </h3>
              <ul className="space-y-1">
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const isActive = activeItem === item.id;
                  return (
                    <li key={item.id}>
                      <Button
                        variant={isActive ? 'secondary' : 'ghost'}
                        className={`w-full justify-start px-4 py-2 h-auto ${
                          isActive 
                            ? 'bg-sidebar-accent text-sidebar-accent-foreground border-l-2 border-sidebar-primary' 
                            : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
                        }`}
                        onClick={() => onItemClick(item.id)}
                        aria-current={isActive ? 'page' : undefined}
                      >
                        <Icon className="h-5 w-5 mr-3" />
                        <span className="flex-1 text-left">{item.label}</span>
                        {item.badge && (
                          <span className="bg-sidebar-primary/10 text-sidebar-primary text-xs px-2 py-0.5 rounded-full">
                            0
                          </span>
                        )}
                      </Button>
                    </li>
                  );
                })}
              </ul>
            </div>
          ) : (
            <div key={group.title} className="mb-4">
              <ul className="space-y-1">
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const isActive = activeItem === item.id;
                  return (
                    <li key={item.id}>
                      <Button
                        variant={isActive ? 'secondary' : 'ghost'}
                        size="icon-lg"
                        className={`w-full ${
                          isActive 
                            ? 'bg-sidebar-accent text-sidebar-accent-foreground' 
                            : 'text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground'
                        }`}
                        onClick={() => onItemClick(item.id)}
                        aria-label={item.label}
                        aria-current={isActive ? 'page' : undefined}
                      >
                        <Icon className="h-5 w-5" />
                      </Button>
                    </li>
                  );
                })}
              </ul>
            </div>
          )
        ))}
      </nav>

      {/* Footer */}
      {!collapsed && (
        <div className="p-4 border-t border-sidebar-border">
          <div className="text-xs text-muted-foreground text-center">
            v2e Security Platform
          </div>
        </div>
      )}
    </div>
  );
}