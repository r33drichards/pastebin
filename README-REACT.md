# React Conversion

This pastebin has been converted to use React with the following features:

## Tech Stack
- **React 18** with TypeScript
- **React Router** for client-side routing
- **React Query** for data fetching and caching
- **Vite** for fast development and building
- **Tailwind CSS** for styling
- **Monaco Editor** for code editing
- **Mermaid** for diagram rendering

## Development

### Prerequisites
- Node.js 20+ (included in Nix shell)
- pnpm (included in Nix shell)

### Setup
```bash
# Enter Nix shell
nix-shell

# Install dependencies
just install
# or
pnpm install
```

### Running in Development
```bash
# Run both frontend and backend
just dev

# Or run separately:
# Frontend only (with proxy to backend)
just dev-frontend

# Backend only
just dev-backend
```

### Building
```bash
# Build frontend
just build-frontend

# Build with Nix (includes both frontend and backend)
nix build
```

## Architecture

### Frontend Structure
```
src/
├── components/     # Reusable components
├── pages/         # Page components (routes)
├── services/      # API services
├── hooks/         # Custom React hooks
├── types/         # TypeScript types
└── main.tsx       # Entry point
```

### API Integration
The React app communicates with the Go backend through these endpoints:
- `POST /paste` - Create a new paste
- `GET /paste?id=...` - Get paste data
- `POST /diff` - Create a new diff
- `GET /diff?id=...` - Get diff data
- `POST /complete` - Get code completions
- `GET /html?id=...` - Get rendered HTML for markdown

### Routing
React Router handles client-side routing:
- `/` - Home page (code editor)
- `/paste?id=...` - View a paste
- `/diff` - Diff editor
- `/diff?id=...` - View a diff
- `/html?id=...` - View rendered HTML

## Production Build
The production build is handled by Nix, which:
1. Builds the React app with Vite
2. Embeds the static files into the Go binary
3. Serves the React app for all non-API routes

The Go server serves:
- API endpoints at their original paths
- Static assets from the embedded filesystem
- The React app's index.html for all other routes (SPA routing)