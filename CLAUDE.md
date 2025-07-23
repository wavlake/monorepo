# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a TDD-focused monorepo for Wavlake's decentralized music platform, featuring:
- **React TypeScript frontend** (Vite + Tailwind CSS) deployed to Vercel
- **Go backend API** with Firebase integration deployed to GCP Cloud Run  
- **Nostr relay integration** for decentralized features
- **Automatic TypeScript type generation** from Go structs
- **Comprehensive testing setup** across all layers

## Essential Commands

### Quick Start
```bash
task setup           # Initialize entire monorepo (dependencies + types + tests)
task dev:tdd         # Start development with test watchers
```

### TDD Workflow (Core Development Pattern)
```bash
task tdd             # Start test watchers for frontend + backend
task red             # Helper for creating failing tests
task green           # Run fast tests to verify implementation
task refactor        # Run tests + linting after code improvements

# Individual test suites
task test:unit:fast  # Quick feedback loop (no coverage)
task test:unit       # Full unit tests with coverage
task test:integration # Integration tests with Docker services
task test:e2e        # End-to-end Playwright tests
```

### Type Generation (Critical for Development)
```bash
task types:generate  # Generate TypeScript from Go structs
task types:watch     # Auto-regenerate on Go file changes
```

### Service Management
```bash
task dev:frontend    # React dev server only
task dev:backend     # Go API server only  
task dev:relay       # Local Nostr relay only
task dev:services    # All services without test watchers
```

### Quality Gates
```bash
task quality:check   # Comprehensive: lint + test + coverage + build
task coverage        # Generate coverage reports for both apps
task build           # Production builds (requires tests to pass)
```

## Architecture & Key Systems

### Monorepo Structure
- `apps/frontend/` - React + TypeScript + Vite + Tailwind CSS
- `apps/backend/` - Go API with Firebase Admin SDK
- `packages/shared-types/` - Generated TypeScript interfaces + Nostr types
- `packages/dev-relay/` - Local Nostr relay configuration
- `tools/` - Development utilities and scripts

### Type Generation System
**Critical for maintaining type safety between Go and TypeScript:**

1. **Source**: Go structs in `apps/backend/internal/{models,handlers,types}/`
2. **Tool**: Custom Go tool at `apps/backend/tools/typegen/main.go`
3. **Output**: TypeScript interfaces in `packages/shared-types/api/`
4. **Usage**: Frontend imports from `@shared` alias

**Process**:
- Analyzes Go structs and JSON tags
- Generates categorized TypeScript files (models, requests, responses)
- Handles custom type mappings (time.Time → string, etc.)
- Build process ensures types are current before deployment

### Testing Architecture
**Backend (Go + Ginkgo)**:
- Mock generation with `//go:generate mockgen` directives
- Test services: Firebase emulators, test database, local relay
- Integration tests tagged with `// +build integration`
- **IMPORTANT**: Run backend test suite after any backend changes: `task test:unit:backend`

**Frontend (Vitest + React Testing Library)**:
- Component tests with user event simulation
- MSW for API mocking
- Playwright for E2E testing
- **IMPORTANT**: Run frontend test suite after any frontend changes: `task test:unit:frontend` (suite to be implemented)

**Test Services (Docker Compose)**:
- Firebase emulators (Auth, Firestore, Storage) on ports 9099, 8080, 9199
- Local Nostr relay on port 7777
- Test PostgreSQL database on port 5433

### Nostr Integration
Custom event kinds for music platform:
- Standard: User profiles (0), Text notes (1), Contacts (3)
- Music: Track metadata (31337), Albums (31338), Artists (31339), Playlists (31340)
- Payments: Lightning invoices (40001-40004)

Type definitions in `packages/shared-types/nostr/events.ts` include full NIP compliance.

### Pre-commit Workflow
Automated via `tools/scripts/pre-commit.sh`:
1. Regenerates types if Go structs changed
2. Runs linting and formatting
3. Executes fast tests
4. Validates builds
5. Scans for sensitive data

Install with: `task hooks:install`

## Development Patterns

### TDD Cycle Implementation
1. **Red Phase**: Use `task red` for test creation guidance
   - Backend: `ginkgo generate [package]` 
   - Frontend: Create `.test.tsx` files
2. **Green Phase**: Implement minimal code, verify with `task green`
3. **Refactor Phase**: Improve code, validate with `task refactor`

### Adding New API Endpoints
1. Define Go structs in `internal/models/` with JSON tags
2. Run `task types:generate` to update TypeScript interfaces
3. Frontend automatically gets type safety for new endpoints
4. Write tests first, then implement handlers
5. **Always run tests after changes**: `task test:unit:backend` for backend, `task test:unit:frontend` for frontend

### Working with Shared Types
- **Import**: `import { SomeType } from '@shared'` in frontend
- **Categories**: `api/` (generated), `nostr/` (manual), `common/` (utilities)
- **Regeneration**: Automatic on backend changes, manual with `task types:generate`

### Testing Requirements
**Critical**: Always run appropriate test suites after making changes:

**Backend Changes**:
```bash
task test:unit:backend        # Unit tests with coverage
task test:integration         # Integration tests (if applicable)
task deploy:staging:test      # Test against staging environment
```

**Frontend Changes**:
```bash
task test:unit:frontend       # Frontend test suite (to be implemented)
task test:e2e                 # End-to-end Playwright tests
```

**Full Validation**:
```bash
task quality:check            # Comprehensive: lint + test + coverage + build
```

## Environment Requirements

### Required Tools
- Node.js 18+ and Go 1.21+
- Task runner: https://taskfile.dev/
- Docker & Docker Compose
- Firebase CLI (for emulators)
- GCP CLI (for deployment)

### Environment Variables (.env.local)
```bash
FIREBASE_PROJECT_ID=your-project-id
GOOGLE_CLOUD_PROJECT=your-gcp-project
DEFAULT_RELAY_URLS=ws://localhost:7777,wss://relay.wavlake.com
```

### Coverage Targets
- Backend: 80%+ unit test coverage
- Frontend: 75%+ component coverage
- Integration: 60%+ critical path coverage
- E2E: 90%+ user journey coverage

## Deployment Pipeline
- Frontend: `task deploy:frontend` → Vercel
- Backend: `task deploy:backend` → GCP Cloud Run
- Full: `task deploy` (both applications)
- Requires: Tests passing, types current, linting clean