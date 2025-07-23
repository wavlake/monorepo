# Wavlake Monorepo

[![codecov](https://codecov.io/gh/wavlake/monorepo/branch/main/graph/badge.svg)](https://codecov.io/gh/wavlake/monorepo)
[![Test Coverage](https://github.com/wavlake/monorepo/actions/workflows/test-coverage.yml/badge.svg)](https://github.com/wavlake/monorepo/actions/workflows/test-coverage.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wavlake/monorepo)](https://goreportcard.com/report/github.com/wavlake/monorepo)

A comprehensive monorepo for Wavlake's music platform, featuring a React TypeScript frontend, Go backend, and Nostr relay integration with full Test-Driven Development (TDD) support.

## 🏗️ Architecture Overview

```
monorepo/
├── apps/
│   ├── frontend/          # React + TypeScript + Vite
│   └── backend/           # Go API + Firebase + GCP
├── packages/
│   ├── shared-types/      # TypeScript interfaces
│   └── dev-relay/         # Local nostr relay
└── tools/                 # Development tools
```

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- Go 1.21+  
- Docker & Docker Compose
- [Task](https://taskfile.dev/) (task runner)
- Firebase CLI
- GCP CLI (for deployment)

### Setup

```bash
# Clone and setup the monorepo
git clone <repository-url>
cd monorepo

# Initialize everything (installs dependencies, generates types, sets up testing)
task setup

# Start TDD development environment
task dev:tdd
```

## 🧪 Test-Driven Development (TDD)

This monorepo is designed for TDD workflows with comprehensive testing at all levels.

### TDD Workflow Commands

```bash
# Start TDD watch mode for both frontend and backend
task tdd

# Red-Green-Refactor cycle helpers
task red      # 🔴 Create failing test
task green    # 🟢 Run tests & implement
task refactor # ♻️ Improve code while keeping tests green

# Fast test feedback
task test:unit:fast     # Quick unit tests (no coverage)
task test:unit         # Full unit tests with coverage
task test:integration  # Integration tests
task test:e2e         # End-to-end tests
```

### Testing Architecture

**Backend Testing (Go + Ginkgo)**
- Unit tests with mocks for external dependencies
- Integration tests with test database and Firebase emulators  
- Contract tests for API endpoints
- Performance benchmarks

**Frontend Testing (React + Jest + Testing Library)**
- Component unit tests with React Testing Library
- Hook tests for custom React hooks
- Integration tests with MSW (Mock Service Worker)
- E2E tests with Playwright

**Test Services**
- Firebase emulators (Auth, Firestore, Storage)
- Local Nostr relay for testing
- Test database with Docker
- Mock external APIs

## 🔧 Development Commands

### Core Workflows

```bash
# Development
task dev              # Start all services
task dev:tdd         # Start dev + test watchers
task dev:frontend    # Frontend only
task dev:backend     # Backend only
task dev:relay       # Nostr relay only

# Testing  
task test            # Run all tests
task coverage        # Generate coverage reports
task quality:check   # Comprehensive quality check

# Type Generation
task types:generate  # Generate TS types from Go
task types:watch     # Watch & regenerate types

# Building
task build           # Build all for production
task build:frontend  # Build React app
task build:backend   # Build Go binary

# Deployment
task deploy          # Deploy all
task deploy:frontend # Deploy to Vercel
task deploy:backend  # Deploy to GCP Cloud Run
```

## 📁 Project Structure

### Frontend (`apps/frontend/`)
- **React 18** with TypeScript
- **Vite** for fast development and building
- **Firebase Auth** for authentication
- **Generated API types** from Go backend
- **Nostr client** for decentralized features
- Deployed to **Vercel**

### Backend (`apps/backend/`)
- **Go 1.21** with Gin framework
- **Firebase Admin SDK** for auth and Firestore
- **GCP Cloud Storage** for file uploads
- **TypeScript generation** from Go structs
- Deployed to **GCP Cloud Run**

### Shared Types (`packages/shared-types/`)
- **Generated API types** from Go structs
- **Nostr event definitions** (manually maintained)
- **Common utility types** shared across frontend/backend

### Development Relay (`packages/dev-relay/`)
- Local Nostr relay using `nak serve`
- Configuration for development/testing
- Event storage for local development

## 🧬 Type Generation System

The monorepo features automatic TypeScript interface generation from Go structs:

```bash
# Manual generation
task types:generate

# Watch mode (regenerates on Go file changes)  
task types:watch
```

**How it works:**
1. Go tool analyzes structs in `internal/models`, `internal/handlers`, `internal/types`
2. Generates TypeScript interfaces in `packages/shared-types/api/`
3. Frontend imports types for full type safety
4. Build process ensures types are always up-to-date

## 🎯 TDD Best Practices

### Red-Green-Refactor Cycle

1. **🔴 RED**: Write a failing test
   ```bash
   task red  # Shows commands to create tests
   ```

2. **🟢 GREEN**: Make test pass with minimal code
   ```bash
   task green  # Runs fast tests
   ```

3. **♻️ REFACTOR**: Improve code while keeping tests green
   ```bash
   task refactor  # Runs tests + linting
   ```

### Test Organization

**Backend Tests:**
```go
//go:generate mockgen -source=interface.go -destination=mocks/mock.go

func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockDB := mocks.NewMockDatabase(ctrl)
    service := NewUserService(mockDB)
    
    // Act
    result, err := service.CreateUser(ctx, userData)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser, result)
}
```

**Frontend Tests:**
```typescript
describe('TrackPlayer', () => {
  it('should play track when play button clicked', async () => {
    // Arrange
    const track = mockTrack();
    render(<TrackPlayer track={track} />);
    
    // Act
    await user.click(screen.getByRole('button', { name: /play/i }));
    
    // Assert
    expect(screen.getByRole('button', { name: /pause/i })).toBeInTheDocument();
  });
});
```

## 🔐 Environment Setup

### Environment Variables

Create `.env.local` in the root:

```bash
# Firebase
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_API_KEY=your-api-key

# GCP  
GOOGLE_CLOUD_PROJECT=your-gcp-project
GOOGLE_APPLICATION_CREDENTIALS=path/to/service-account.json

# Nostr
DEFAULT_RELAY_URLS=ws://localhost:7777,wss://relay.wavlake.com

# Development
NODE_ENV=development
GO_ENV=development
```

### Firebase Setup

```bash
# Install Firebase CLI
npm install -g firebase-tools

# Login and initialize
firebase login
firebase init

# Start emulators for development
task firebase:emulators
```

### GCP Setup

```bash
# Install gcloud CLI
# https://cloud.google.com/sdk/docs/install

# Authenticate
gcloud auth login
gcloud config set project your-project-id

# Enable required APIs
gcloud services enable run.googleapis.com
gcloud services enable storage.googleapis.com
```

## 📊 Quality & Testing Metrics

### Current Coverage Status
- **Backend Handlers**: 95.1% coverage (172 comprehensive tests)
- **API Routes**: 100% production endpoint coverage (10/10 routes tested)
- **Handler Methods**: 100% coverage (AuthHandlers, TracksHandler, LegacyHandler)
- **Service Interfaces**: 113 comprehensive interface tests

### Coverage Targets
- **Backend**: 80%+ unit test coverage ✅ **EXCEEDED**
- **Frontend**: 75%+ component coverage  
- **Integration**: 60%+ critical path coverage
- **E2E**: 90%+ user journey coverage

### Quality Gates
All builds require:
- ✅ All tests passing
- ✅ Linting passed
- ✅ Type checking passed  
- ✅ Security scans passed
- ✅ Coverage thresholds met

## 🚀 Deployment

### Frontend (Vercel)
```bash
task deploy:frontend
```

### Backend (GCP Cloud Run)
```bash  
task deploy:backend
```

### Full Deployment
```bash
task deploy  # Deploys both frontend and backend
```

## 🤝 Contributing

1. **Clone & Setup**:
   ```bash
   git clone <repo>
   task setup
   ```

2. **TDD Workflow**:
   ```bash
   task tdd        # Start TDD environment
   task red        # Write failing test
   task green      # Implement minimal code
   task refactor   # Improve code structure
   ```

3. **Quality Checks**:
   ```bash
   task quality:check  # Comprehensive quality check
   ```

4. **Commit**:
   ```bash
   git add .
   git commit -m "feat: add new feature"  # Pre-commit hooks run automatically
   ```

### Git Hooks

The monorepo includes pre-commit hooks that:
- Regenerate types if Go structs changed
- Run linting and formatting
- Execute fast tests  
- Check builds
- Scan for sensitive data

Install hooks: `task hooks:install`

## 🔍 Debugging & Monitoring

### Logs
```bash
task logs:backend   # View backend logs
task logs:relay     # View relay logs
```

### Health Checks
```bash
task health         # Check all services
```

### Performance Monitoring
- Frontend: Vercel Analytics
- Backend: GCP Cloud Monitoring  
- Database: Firebase Console
- Relay: Built-in metrics

## 📚 Documentation

- **API Documentation**: Generated with Swagger (`task docs`)
- **Type Documentation**: Auto-generated from Go structs
- **Component Documentation**: Storybook (coming soon)
- **Architecture Docs**: `/docs` directory (coming soon)

## 🎵 Nostr Integration

Wavlake uses Nostr for decentralized features:

- **User profiles** (kind 0)
- **Track metadata** (kind 31337)
- **Album metadata** (kind 31338)  
- **Playlists** (kind 31340)
- **Lightning payments** (kind 40001-40004)

See `packages/shared-types/nostr/` for complete type definitions.

## 📞 Support

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions  
- **Development**: Join our Discord
- **Documentation**: Check `/docs` directory

---

Built with ❤️ for the music community by the Wavlake team.