'use client';

import React from 'react';
import { FileText } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Artifact } from '@/types';

interface ArtifactCardProps {
  artifact: Artifact;
  onClick: () => void;
}

export function ArtifactCard({ artifact, onClick }: ArtifactCardProps) {
  const formattedDate = new Date(artifact.createdAt).toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div
      onClick={onClick}
      className={cn(
        'group cursor-pointer rounded-lg border border-border bg-card p-4',
        'transition-all duration-200 hover:shadow-md hover:border-atom-spark/40',
        'hover:bg-atom-mist/30'
      )}
    >
      <div className="flex items-start gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-atom-mist text-atom-core">
          <FileText className="h-4 w-4" />
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center rounded-sm bg-atom-glow/30 px-1.5 py-0.5 text-[10px] font-medium text-atom-deep">
              {artifact.fileType?.toUpperCase() || 'MD'}
            </span>
          </div>
          <h3 className="mt-1.5 truncate text-sm font-medium text-foreground">
            {artifact.title}
          </h3>
          <p className="mt-1 line-clamp-3 text-xs text-muted-foreground">
            {artifact.content}
          </p>
          <p className="mt-2 text-[11px] text-muted-foreground/60">
            {formattedDate}
          </p>
        </div>
      </div>
    </div>
  );
}
