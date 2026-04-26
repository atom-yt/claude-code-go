'use client';

import React, { useEffect, useState, useMemo } from 'react';
import { Search, LayoutGrid, List, Package } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Artifact } from '@/types';
import { useArtifactsStore } from '@/stores/artifactsStore';
import { ArtifactCard } from './ArtifactCard';
import { ArtifactViewer } from './ArtifactViewer';

function groupByDate(artifacts: Artifact[]): { recent: Artifact[]; older: Artifact[] } {
  const now = new Date();
  const sevenDaysAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const recent: Artifact[] = [];
  const older: Artifact[] = [];

  for (const artifact of artifacts) {
    const created = new Date(artifact.createdAt);
    if (created >= sevenDaysAgo) {
      recent.push(artifact);
    } else {
      older.push(artifact);
    }
  }

  return { recent, older };
}

export function ArtifactList() {
  const {
    artifacts,
    total,
    search,
    viewMode,
    isLoading,
    fetchArtifacts,
    setSearch,
    setViewMode,
  } = useArtifactsStore();

  const [selectedArtifact, setSelectedArtifact] = useState<Artifact | null>(null);

  useEffect(() => {
    fetchArtifacts();
  }, [fetchArtifacts]);

  useEffect(() => {
    const timer = setTimeout(() => {
      fetchArtifacts(search || undefined);
    }, 300);
    return () => clearTimeout(timer);
  }, [search, fetchArtifacts]);

  const { recent, older } = useMemo(() => groupByDate(artifacts), [artifacts]);

  const gridClass = viewMode === 'grid'
    ? 'grid grid-cols-1 gap-3 sm:grid-cols-2'
    : 'grid grid-cols-1 gap-3';

  return (
    <div className="mx-auto w-full max-w-4xl px-6 py-8">
      {/* Header */}
      <div className="mb-1">
        <div className="flex items-center gap-3">
          <h1 className="text-2xl font-bold text-foreground">产物</h1>
          <span className="inline-flex items-center rounded-full bg-atom-glow/30 px-2.5 py-0.5 text-xs font-medium text-atom-deep">
            {total}
          </span>
        </div>
        <p className="mt-1 text-sm text-muted-foreground">
          atom 造物，遇见更好的日常
        </p>
      </div>

      {/* Toolbar */}
      <div className="mt-5 flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="搜索产物..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className={cn(
              'h-9 w-full rounded-md border border-input bg-background pl-9 pr-3 text-sm',
              'placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-atom-spark/40'
            )}
          />
        </div>
        <div className="flex items-center rounded-md border border-input">
          <button
            onClick={() => setViewMode('grid')}
            className={cn(
              'flex h-9 w-9 items-center justify-center rounded-l-md transition-colors',
              viewMode === 'grid'
                ? 'bg-atom-mist text-atom-core'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <LayoutGrid className="h-4 w-4" />
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={cn(
              'flex h-9 w-9 items-center justify-center rounded-r-md transition-colors',
              viewMode === 'list'
                ? 'bg-atom-mist text-atom-core'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            <List className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Content */}
      {isLoading && artifacts.length === 0 ? (
        <div className="mt-16 flex flex-col items-center justify-center text-muted-foreground">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-atom-spark border-t-transparent" />
          <p className="mt-3 text-sm">加载中...</p>
        </div>
      ) : artifacts.length === 0 ? (
        <div className="mt-16 flex flex-col items-center justify-center text-muted-foreground">
          <Package className="h-12 w-12 text-muted-foreground/40" />
          <p className="mt-3 text-sm">暂无产物</p>
          <p className="mt-1 text-xs text-muted-foreground/60">
            与 atom 对话即可生成产物
          </p>
        </div>
      ) : (
        <div className="mt-6 space-y-6">
          {recent.length > 0 && (
            <section>
              <h2 className="mb-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                近 7 天
              </h2>
              <div className={gridClass}>
                {recent.map((artifact) => (
                  <ArtifactCard
                    key={artifact.id}
                    artifact={artifact}
                    onClick={() => setSelectedArtifact(artifact)}
                  />
                ))}
              </div>
            </section>
          )}

          {older.length > 0 && (
            <section>
              <h2 className="mb-3 text-xs font-medium uppercase tracking-wider text-muted-foreground">
                更早
              </h2>
              <div className={gridClass}>
                {older.map((artifact) => (
                  <ArtifactCard
                    key={artifact.id}
                    artifact={artifact}
                    onClick={() => setSelectedArtifact(artifact)}
                  />
                ))}
              </div>
            </section>
          )}
        </div>
      )}

      {/* Viewer Modal */}
      <ArtifactViewer
        artifact={selectedArtifact}
        onClose={() => setSelectedArtifact(null)}
      />
    </div>
  );
}
