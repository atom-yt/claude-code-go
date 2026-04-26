'use client';

import React from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Artifact } from '@/types';

interface ArtifactViewerProps {
  artifact: Artifact | null;
  onClose: () => void;
}

export function ArtifactViewer({ artifact, onClose }: ArtifactViewerProps) {
  if (!artifact) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className={cn(
          'relative mx-4 flex h-[85vh] w-full max-w-3xl flex-col',
          'rounded-xl border border-border bg-background shadow-2xl'
        )}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border px-6 py-4">
          <div className="min-w-0 flex-1">
            <h2 className="truncate text-lg font-semibold text-foreground">
              {artifact.title}
            </h2>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {new Date(artifact.createdAt).toLocaleString('zh-CN')}
              {artifact.tags && artifact.tags.length > 0 && (
                <span className="ml-2">
                  {artifact.tags.map((tag) => (
                    <span
                      key={tag}
                      className="ml-1 inline-flex items-center rounded-sm bg-atom-glow/30 px-1.5 py-0.5 text-[10px] font-medium text-atom-deep"
                    >
                      {tag}
                    </span>
                  ))}
                </span>
              )}
            </p>
          </div>
          <button
            onClick={onClose}
            className={cn(
              'ml-4 flex h-8 w-8 shrink-0 items-center justify-center rounded-md',
              'text-muted-foreground transition-colors hover:bg-muted hover:text-foreground'
            )}
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto px-6 py-4">
          <pre className="whitespace-pre-wrap break-words font-mono text-sm leading-relaxed text-foreground/90">
            {artifact.content}
          </pre>
        </div>
      </div>
    </div>
  );
}
