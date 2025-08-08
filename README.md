# Wavlake Monorepo

[![codecov](https://codecov.io/gh/wavlake/monorepo/branch/main/graph/badge.svg)](https://codecov.io/gh/wavlake/monorepo)
[![Test Coverage](https://github.com/wavlake/monorepo/actions/workflows/test-coverage.yml/badge.svg)](https://github.com/wavlake/monorepo/actions/workflows/test-coverage.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wavlake/monorepo)](https://goreportcard.com/report/github.com/wavlake/monorepo)

A comprehensive monorepo for Wavlake's music platform, featuring a React TypeScript frontend, Go backend, and Nostr relay integration with full Test-Driven Development (TDD) support.

## 🏗️ Architecture Overview

```
monorepo/
├── apps/
│   ├── web/               # React + TypeScript + Vite (web client)
│   └── api/               # Go API + Firebase + GCP
├── packages/
│   ├── shared/            # TypeScript interfaces
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

# Configure your environment:
cp .env.example .env.local
# Edit .env.local and uncomment the setup section you want (see file for details)

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

### Frontend (`apps/web/`)
- **React 18** with TypeScript
- **Vite** for fast development and building
- **Firebase Auth** for authentication
- **Generated API types** from Go backend
- **Nostr client** for decentralized features
- Deployed to **Vercel**

### API (`apps/api/`)
- **Go 1.21** with Gin framework
- **Firebase Admin SDK** for auth and Firestore
- **GCP Cloud Storage** for file uploads
- **TypeScript generation** from Go structs
- Deployed to **GCP Cloud Run**

### Shared Types (`packages/shared/`)
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
2. Generates TypeScript interfaces in `packages/shared/api/`
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

## 🔐 Environment Configuration

### Quick Start for New Developers

The monorepo supports three environment configurations to match your development needs. Copy the template and choose your setup:

```bash
# Copy the template
cp .env.example .env.local

# Edit .env.local and uncomment ONE of the three setup sections
```

### Configuration Options

#### 1. Minimal Setup (Recommended for New Developers)
**Best for**: First-time setup, quick development, no external dependencies

```bash
DEVELOPMENT=true
SKIP_AUTH=true
MOCK_STORAGE=true
MOCK_STORAGE_PATH=./dev-storage
FILE_SERVER_URL=http://localhost:8081
DEFAULT_RELAY_URLS=ws://localhost:10547
LOG_REQUESTS=true
LOG_RESPONSES=true
```

**What you get**:
- ✅ No Firebase/GCP setup required
- ✅ Authentication bypassed for development
- ✅ Local file storage in `./dev-storage`
- ✅ Local Nostr relay only
- ✅ Detailed request/response logging

**Start developing**:
```bash
task setup
task dev:tdd
```

#### 2. Firebase Emulator Setup
**Best for**: Full local development with realistic auth/database behavior

```bash
DEVELOPMENT=true
FIREBASE_PROJECT_ID=demo-project
GOOGLE_CLOUD_PROJECT=demo-project
FIRESTORE_EMULATOR_HOST=localhost:8080
FIREBASE_AUTH_EMULATOR_HOST=localhost:9099
FIREBASE_STORAGE_EMULATOR_HOST=localhost:9199
DEFAULT_RELAY_URLS=ws://localhost:10547
LOG_REQUESTS=true
LOG_RESPONSES=true
```

**What you get**:
- ✅ Real Firebase Auth behavior (with emulator)
- ✅ Real Firestore database (with emulator)  
- ✅ Real Firebase Storage (with emulator)
- ✅ No internet required after setup
- ✅ Realistic development environment

**Setup steps**:
```bash
# Install Firebase CLI
npm install -g firebase-tools

# Start emulators (in separate terminal)
task firebase:emulators

# Start development
task dev:tdd
```

#### 3. Production Services Setup (Advanced)
**Best for**: Testing integration with real GCP services, advanced development

```bash
DEVELOPMENT=true
FIREBASE_PROJECT_ID=your-firebase-project-id
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
GCS_BUCKET_NAME=your-storage-bucket
PROD_POSTGRES_CONNECTION_STRING_RO=postgresql://user:pass@host:5432/db
DEFAULT_RELAY_URLS=ws://localhost:10547,wss://relay.wavlake.com
LOG_REQUESTS=true
```

**What you get**:
- ✅ Real Firebase Auth and Firestore
- ✅ Real GCP Cloud Storage
- ✅ Production database (read-only)
- ✅ Remote and local Nostr relays
- ⚠️ Requires GCP project and credentials

**Setup steps**:
```bash
# Authenticate with GCP
gcloud auth login
gcloud config set project your-project-id

# Set up service account (if needed)
export GOOGLE_APPLICATION_CREDENTIALS=path/to/service-account-key.json

# Start development
task dev:tdd
```

### Environment Variable Reference

All available environment variables:

| Variable | Purpose | Default | Required |
|----------|---------|---------|----------|
| `DEVELOPMENT` | Enable development mode | `false` | ✅ |
| `BACKEND_PORT` | Backend server port | `3000` | ❌ |
| `VITE_PORT` | Frontend dev server port (Vite standard) | `8080` | ❌ |
| `SKIP_AUTH` | Bypass authentication | `false` | ❌ |
| `MOCK_STORAGE` | Use local file storage | `false` | ❌ |
| `MOCK_STORAGE_PATH` | Local storage directory | `./dev-storage` | ❌ |
| `FILE_SERVER_URL` | File server endpoint | `http://localhost:8081` | ❌ |
| `FIREBASE_PROJECT_ID` | Firebase project ID | - | Conditional |
| `GOOGLE_CLOUD_PROJECT` | GCP project ID | - | Conditional |
| `GCS_BUCKET_NAME` | Cloud Storage bucket | - | Conditional |
| `DEFAULT_RELAY_URLS` | Nostr relay URLs | `ws://localhost:10547` | ❌ |
| `LOG_REQUESTS` | Log HTTP requests | `false` | ❌ |
| `LOG_RESPONSES` | Log HTTP responses | `false` | ❌ |
| `LOG_HEADERS` | Log HTTP headers | `false` | ❌ |
| `LOG_REQUEST_BODY` | Log request bodies | `false` | ❌ |
| `LOG_RESPONSE_BODY` | Log response bodies | `false` | ❌ |

### Switching Between Configurations

You can easily switch between development configurations:

```bash
# Switch to minimal setup
cp .env.example .env.local
# Edit and uncomment MINIMAL SETUP section

# Switch to Firebase emulators
cp .env.example .env.local  
# Edit and uncomment FIREBASE EMULATOR SETUP section

# Switch to production services
cp .env.example .env.local
# Edit and uncomment PRODUCTION SERVICES SETUP section
```

### Service Substitution

The monorepo allows flexible service substitution:

| Service | Local Option | Remote Option | Notes |
|---------|--------------|---------------|-------|
| **Authentication** | `SKIP_AUTH=true` | Firebase Auth | Skip for development |
| **Database** | Firebase emulator | Real Firestore | Emulator recommended |
| **Storage** | `MOCK_STORAGE=true` | GCS bucket | Mock for development |
| **Nostr Relay** | `ws://localhost:10547` | Remote relays | Local for development |
| **Frontend** | Always local | - | Must run locally |
| **Backend** | Always local | - | Must run locally |

### Advanced Setup Instructions

#### Firebase CLI Setup
Required for Firebase emulator configuration:

```bash
# Install Firebase CLI
npm install -g firebase-tools

# Login and initialize (one-time setup)
firebase login
firebase init

# Start emulators for development
task firebase:emulators
```

#### GCP CLI Setup  
Required for production services configuration:

```bash
# Install gcloud CLI
# https://cloud.google.com/sdk/docs/install

# Authenticate
gcloud auth login
gcloud config set project your-project-id

# Enable required APIs
gcloud services enable run.googleapis.com
gcloud services enable storage.googleapis.com

# Set up service account (if needed)
gcloud iam service-accounts create wavlake-dev
gcloud iam service-accounts keys create key.json --iam-account=wavlake-dev@your-project.iam.gserviceaccount.com
export GOOGLE_APPLICATION_CREDENTIALS=./key.json
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

See `packages/shared/nostr/` for complete type definitions.

## 📞 Support

- **Issues**: GitHub Issues
- **Discussions**: GitHub Discussions  
- **Development**: Join our Discord
- **Documentation**: Check `/docs` directory

---

Built with ❤️ for the music community by the Wavlake team.