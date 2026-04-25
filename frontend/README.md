# Frontend

Next.js 14+ application for the Atom AI Platform.

## To be initialized by the Frontend Agent

This directory will contain:
- React components for chat interface
- Session management UI
- Agent configuration dashboard
- Real-time WebSocket connections
- User authentication flows

## Planned Structure

```
frontend/
├── app/                  # Next.js App Router
│   ├── (auth)/          # Authentication routes
│   ├── (dashboard)/     # Dashboard routes
│   ├── chat/            # Chat interface
│   └── api/             # API routes (if needed)
├── components/          # React components
│   ├── chat/           # Chat-related components
│   ├── session/        # Session components
│   └── ui/             # UI primitives
├── lib/                 # Utility functions
├── hooks/              # React hooks
├── types/              # TypeScript types
└── public/             # Static assets
```

## Getting Started (when initialized)

```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to view the application.