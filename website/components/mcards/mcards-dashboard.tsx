'use client';

import { LayoutGrid, Clock, BookOpen, GraduationCap, RefreshCw, Flame, Trophy } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { StatCard } from './stat-card';
import { useMemoryCards } from '@/lib/mcards/hooks';

export function McardsDashboard() {
  const { data, isLoading, refetch } = useMemoryCards({ limit: 1000 });

  const cards = data?.payload?.memory_cards || [];

  // Calculate stats
  const stats = {
    totalCards: cards.length,
    dueToday: cards.filter((c: any) => {
      if (!c.next_review_at) return false;
      const nextReview = new Date(c.next_review_at);
      return nextReview <= new Date();
    }).length,
    learning: cards.filter((c: any) => c.learning_state === 'learning').length,
    mastered: cards.filter((c: any) => c.learning_state === 'mastered').length,
    archived: cards.filter((c: any) => c.learning_state === 'archived').length,
  };

  // Calculate state distribution for chart
  const stateDistribution = [
    { state: 'To Review', count: cards.filter((c: any) => c.learning_state === 'to_review').length, color: 'bg-rose-500' },
    { state: 'Learning', count: stats.learning, color: 'bg-amber-500' },
    { state: 'Mastered', count: stats.mastered, color: 'bg-emerald-500' },
    { state: 'Archived', count: stats.archived, color: 'bg-slate-400' },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold text-slate-900 dark:text-slate-100">Dashboard</h2>
          <p className="text-sm text-slate-600 dark:text-slate-400">
            Memory cards overview and statistics
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={() => refetch()}>
          <RefreshCw className="w-4 h-4 mr-2" />
          Refresh
        </Button>
      </div>

      {/* Stat Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Cards"
          value={stats.totalCards}
          icon={LayoutGrid}
          color="indigo"
        />
        <StatCard
          title="Due Today"
          value={stats.dueToday}
          icon={Clock}
          color="rose"
        />
        <StatCard
          title="Learning"
          value={stats.learning}
          icon={BookOpen}
          color="amber"
        />
        <StatCard
          title="Mastered"
          value={stats.mastered}
          icon={GraduationCap}
          color="emerald"
        />
      </div>

      {/* Charts Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Learning State Distribution */}
        <Card>
          <CardHeader>
            <CardTitle>Learning State Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {stateDistribution.map((item) => (
                <div key={item.state} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className={`w-3 h-3 rounded-full ${item.color}`} />
                    <span className="text-sm text-slate-600 dark:text-slate-400">{item.state}</span>
                  </div>
                  <span className="font-medium text-slate-900 dark:text-slate-100">{item.count}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-slate-600 dark:text-slate-400">Cards studied this week</span>
                <span className="font-medium text-slate-900 dark:text-slate-100">
                  {cards.filter((c: any) => {
                    if (!c.last_reviewed_at) return false;
                    const lastReviewed = new Date(c.last_reviewed_at);
                    const weekAgo = new Date();
                    weekAgo.setDate(weekAgo.getDate() - 7);
                    return lastReviewed >= weekAgo;
                  }).length}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-slate-600 dark:text-slate-400 flex items-center gap-2">
                  <Flame className="w-4 h-4" />
                  Current Streak
                </span>
                <span className="font-medium text-slate-900 dark:text-slate-100">5 days</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-slate-600 dark:text-slate-400 flex items-center gap-2">
                  <Trophy className="w-4 h-4" />
                  Longest Streak
                </span>
                <span className="font-medium text-slate-900 dark:text-slate-100">14 days</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* SM-2 Metrics */}
      <Card>
        <CardHeader>
          <CardTitle>Spaced Repetition Metrics</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-6">
            <div>
              <div className="text-sm text-slate-600 dark:text-slate-400 mb-1">Avg Ease Factor</div>
              <div className="text-3xl font-bold text-slate-900 dark:text-slate-100">
                {cards.length > 0
                  ? (cards.reduce((sum: number, c: any) => sum + (c.ease_factor || 2.5), 0) / cards.length).toFixed(2)
                  : '0.00'}
              </div>
            </div>
            <div>
              <div className="text-sm text-slate-600 dark:text-slate-400 mb-1">Avg Interval</div>
              <div className="text-3xl font-bold text-slate-900 dark:text-slate-100">
                {cards.length > 0
                  ? Math.round(cards.reduce((sum: number, c: any) => sum + (c.interval || 0), 0) / cards.length)
                  : 0}
                </div>
              <div className="text-sm text-slate-400">days</div>
            </div>
            <div>
              <div className="text-sm text-slate-600 dark:text-slate-400 mb-1">Total Repetitions</div>
              <div className="text-3xl font-bold text-slate-900 dark:text-slate-100">
                {cards.reduce((sum: number, c: any) => sum + (c.repetitions || 0), 0)}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
