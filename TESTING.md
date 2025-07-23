# Testing Strategy & Implementation Plan

This document outlines the comprehensive testing strategy for the Wavlake monorepo, with focus on achieving 80%+ backend unit test coverage.

## Current Testing Status

### Backend Coverage Analysis
- **Current Coverage**: 0.0% unit test coverage
- **Target Coverage**: 80%+ (per project requirements)
- **Existing Tests**: 7 integration test suites (all passing)
- **Framework**: Ginkgo + Gomega (configured)
- **Critical Gap**: No unit tests for core business logic (26 Go source files)

### Test Infrastructure
- **Integration Tests**: Comprehensive coverage in `tests/integration/`
- **Mock Generation**: `mockgen` configured but unused
- **Test Services**: Docker Compose setup for Firebase emulators, PostgreSQL, Nostr relay
- **TDD Workflow**: Taskfile commands available (`task tdd`, `task red`, `task green`, `task refactor`)

## Backend Unit Testing Implementation Plan

### 🎯 Strategic Overview

**Approach**: Test-driven development with mock-based isolation  
**Timeline**: 4-week phased implementation  
**Framework**: Ginkgo + Gomega with generated mocks  

### 📋 Phase 1: Foundation & Critical Path (Week 1-2)

#### Priority 1: Core Services Layer
**Target Coverage**: 85%+ (highest business logic concentration)

##### UserService (`internal/services/user_service.go`)
**File**: `internal/services/user_service_test.go`

```go
// Test structure to implement:
├── LinkPubkeyToUser tests
│   ├── Success: new user creation
│   ├── Success: existing user pubkey addition  
│   ├── Error: pubkey already linked to different user
│   ├── Error: Firestore transaction failure
│   └── Edge cases: empty pubkey, invalid firebaseUID
├── UnlinkPubkeyFromUser tests
│   ├── Success: pubkey unlinked successfully
│   ├── Error: pubkey not found
│   ├── Error: pubkey belongs to different user
│   ├── Error: pubkey already unlinked
│   └── Error: Firestore transaction failure
├── GetLinkedPubkeys tests
│   ├── Success: ordered query with results
│   ├── Success: fallback to simple query when OrderBy fails
│   ├── Success: empty results
│   └── Error: query failure
├── GetFirebaseUIDByPubkey tests
│   ├── Success: active pubkey returns UID
│   ├── Error: pubkey not found
│   └── Error: pubkey not active
└── GetUserEmail tests
    ├── Success: returns user email
    └── Error: user not found in Firebase Auth
```

**Mock Requirements**:
- `firestore.Client` (transactions, collections, documents)
- `auth.Client` (GetUser method)

##### NostrTrackService (`internal/services/nostr_track.go`)
**File**: `internal/services/nostr_track_test.go`

```go
├── CreateTrack tests
│   ├── Success: track created with valid data
│   ├── Error: invalid extension
│   └── Error: service failure
├── GetTrack tests
│   ├── Success: track found
│   └── Error: track not found
├── GetTracksByPubkey tests
│   ├── Success: returns user tracks
│   └── Success: empty results for user with no tracks
└── DeleteTrack tests
    ├── Success: track deleted (soft delete)
    └── Error: track not found
```

#### Priority 2: Handler Layer
**Target Coverage**: 80%+ (HTTP interface validation)

##### AuthHandlers (`internal/handlers/auth.go`)
**File**: `internal/handlers/auth_test.go`

```go
├── LinkPubkey endpoint tests
│   ├── Success: dual auth present, pubkey linked
│   ├── Error: missing Firebase auth
│   ├── Error: missing Nostr auth
│   ├── Error: request pubkey mismatch
│   └── Error: service failure
├── UnlinkPubkey endpoint tests
│   ├── Success: pubkey unlinked
│   ├── Error: missing Firebase auth
│   ├── Error: invalid request body
│   └── Error: service failure
├── GetLinkedPubkeys endpoint tests
│   ├── Success: returns linked pubkeys list
│   ├── Success: returns empty array when no pubkeys
│   ├── Error: missing Firebase auth
│   └── Error: service failure
└── CheckPubkeyLink endpoint tests
    ├── Success: pubkey is linked, returns Firebase UID
    ├── Success: pubkey not linked
    ├── Error: missing Nostr auth
    ├── Error: pubkey mismatch (user can only check own)
    └── Error: invalid request body
```

**Mock Requirements**:
- `services.UserServiceInterface`
- `gin.Context` (test context creation helpers)

##### TracksHandler (`internal/handlers/tracks.go`)
**File**: `internal/handlers/tracks_test.go`

```go
├── CreateTrackNostr tests
│   ├── Success: track created with valid extension
│   ├── Error: missing extension field
│   ├── Error: unsupported audio format
│   ├── Error: missing authentication
│   └── Error: service failure
├── GetMyTracks tests
│   ├── Success: returns user tracks
│   ├── Error: missing authentication
│   └── Error: service failure
├── GetTrack tests
│   ├── Success: track found and returned
│   ├── Error: missing track ID
│   └── Error: track not found
└── DeleteTrack tests
    ├── Success: track deleted by owner
    ├── Error: missing track ID
    ├── Error: missing authentication
    ├── Error: track not found
    ├── Error: user not owner of track
    └── Error: service failure
```

### 📋 Phase 2: Authentication & Middleware (Week 3)

#### Authentication Modules
**Target Coverage**: 90%+ (security-critical)

##### Firebase Authentication (`internal/auth/firebase.go`)
**File**: `internal/auth/firebase_test.go`

```go
├── FirebaseMiddleware tests
│   ├── Success: valid token, context set with UID and email
│   ├── Error: missing Authorization header
│   ├── Error: invalid Bearer token format
│   ├── Error: Firebase token verification failure
│   └── Edge cases: malformed headers, empty tokens
├── extractBearerToken tests
│   ├── Success: extracts token from valid Bearer header
│   ├── Returns empty: missing header
│   ├── Returns empty: invalid format (not Bearer)
│   └── Returns empty: missing token part
```

##### NIP-98 Authentication (`internal/auth/nip98.go`)
**File**: `internal/auth/nip98_test.go`

```go
├── NIP-98 signature validation tests
│   ├── Success: valid signature and event
│   ├── Error: invalid signature
│   ├── Error: malformed event
│   ├── Error: expired event
│   └── Error: invalid pubkey format
├── Event verification tests
│   ├── Success: valid NIP-98 event structure
│   ├── Error: missing required fields
│   └── Error: invalid timestamp
└── Pubkey extraction tests
    ├── Success: extracts pubkey from valid event
    └── Error: malformed event data
```

##### Dual Authentication (`internal/auth/dual.go`)
**File**: `internal/auth/dual_test.go`

```go
├── Dual authentication flow tests
│   ├── Success: both Firebase and NIP-98 auth present
│   ├── Error: missing Firebase auth
│   ├── Error: missing NIP-98 auth
│   └── Error: authentication mismatch
├── Missing auth scenarios tests
│   ├── Partial auth: only Firebase present
│   ├── Partial auth: only NIP-98 present
│   └── No auth: both missing
└── Auth combination validation tests
    ├── Success: valid combination sets context
    └── Error: invalid combination blocks request
```

**Mock Requirements**:
- `auth.Client.VerifyIDToken`
- HTTP request/response contexts
- NIP-98 event validation functions

### 📋 Phase 3: Utilities & Models (Week 4)

#### Utility Functions

##### Audio Processing (`internal/utils/audio.go`)
**File**: `internal/utils/audio_test.go`

```go
├── AudioProcessor tests
│   ├── Initialization: processor created successfully
│   ├── Configuration: options set correctly
│   └── Error handling: invalid configuration
├── Format validation tests
│   ├── Success: supported formats (mp3, wav, flac, etc.)
│   ├── Error: unsupported formats
│   └── Edge cases: empty extension, case sensitivity
├── Compression tests (with ffmpeg mocks)
│   ├── Success: audio compressed with default options
│   ├── Success: audio compressed with custom options  
│   ├── Error: ffmpeg not available
│   ├── Error: invalid input file
│   └── Error: compression failure
└── Metadata extraction tests
    ├── Success: extracts metadata from valid audio
    ├── Error: ffprobe not available
    ├── Error: invalid audio file
    └── Error: metadata extraction failure
```

##### Storage Paths (`internal/utils/storage_paths.go`)
**File**: `internal/utils/storage_paths_test.go`

```go
├── Path generation tests
│   ├── Success: generates valid storage paths
│   ├── Success: path uniqueness across calls
│   └── Success: proper file extension handling
├── Validation tests
│   ├── Success: validates proper path formats
│   ├── Error: invalid characters in paths
│   └── Error: path length limits
└── Security tests (path traversal)
    ├── Blocks: ../ sequences
    ├── Blocks: absolute paths
    └── Blocks: symbolic link patterns
```

#### Models & Configuration

##### User Models (`internal/models/user.go`)
**File**: `internal/models/user_test.go`

```go
├── Model validation tests
│   ├── APIUser: valid struct creation and validation
│   ├── NostrAuth: proper field validation
│   ├── LinkedPubkeyInfo: data marshaling
│   └── LegacyUser: backward compatibility
├── JSON marshaling tests
│   ├── Success: proper JSON serialization
│   ├── Success: JSON deserialization
│   ├── Success: omitempty fields handled correctly
│   └── Error: invalid JSON format
└── Database mapping tests
    ├── Firestore: document to struct mapping
    ├── PostgreSQL: row to struct mapping
    └── Field tag validation: json, firestore, db tags
```

##### Service Configuration (`internal/config/service_config.go`)
**File**: `internal/config/service_config_test.go`

```go
├── Configuration loading tests
│   ├── Success: loads from environment variables
│   ├── Success: uses default values when env vars missing
│   ├── Success: development vs production configs
│   └── Error: invalid configuration values
├── Environment variable tests
│   ├── Required vars: FIREBASE_PROJECT_ID, etc.
│   ├── Optional vars: proper defaults applied
│   └── Type conversion: string to int, bool conversion
└── Validation tests
    ├── Firebase config: project ID, credentials validation
    ├── Database config: connection string validation
    └── Service URLs: proper URL format validation
```

## 🛠️ Mock Generation Strategy

### Interface-Based Mocking

Add mock generation directives to key files:

```go
// In internal/services/interfaces.go
//go:generate mockgen -source=interfaces.go -destination=../../tests/mocks/service_mocks.go -package=mocks

// In internal/services/user_service.go  
//go:generate mockgen -package=mocks cloud.google.com/go/firestore Client,Transaction,DocumentRef
//go:generate mockgen -package=mocks firebase.google.com/go/v4/auth Client

// In internal/utils/audio.go
//go:generate mockgen -source=audio.go -destination=../../tests/mocks/audio_mocks.go -package=mocks
```

### Mock Directory Structure

```
tests/
├── mocks/
│   ├── service_mocks.go      # Generated service interface mocks
│   ├── firebase_mocks.go     # Firebase SDK mocks  
│   ├── gin_mocks.go          # Gin framework mocks
│   ├── audio_mocks.go        # Audio processing mocks
│   └── external_mocks.go     # External dependency mocks
├── testutil/
│   ├── fixtures.go           # Test data fixtures
│   ├── helpers.go            # Test helper functions
│   ├── context.go            # Test context creation
│   └── assertions.go         # Custom Gomega matchers
└── integration/              # Existing integration tests
    ├── api_integration_test.go
    ├── auth_flows_test.go
    └── ...
```

### Test Data Fixtures

Create reusable test data in `tests/testutil/fixtures.go`:

```go
package testutil

import (
    "time"
    "github.com/wavlake/monorepo/internal/models"
)

// User fixtures
func ValidAPIUser() models.APIUser {
    return models.APIUser{
        FirebaseUID:   "test-firebase-uid",
        ActivePubkeys: []string{"test-pubkey-1", "test-pubkey-2"},
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }
}

func ValidNostrAuth() models.NostrAuth {
    return models.NostrAuth{
        Pubkey:      "test-pubkey",
        FirebaseUID: "test-firebase-uid", 
        Active:      true,
        LinkedAt:    time.Now(),
        LastUsedAt:  time.Now(),
    }
}

// HTTP request fixtures
func ValidLinkPubkeyRequest() map[string]interface{} {
    return map[string]interface{}{
        "pubkey": "test-pubkey",
    }
}

// Add more fixtures as needed...
```

## 📊 Implementation Priorities

### Critical Path (Must Have - 80% of value)
1. **UserService** - Core business logic (40% impact)
2. **AuthHandlers** - API endpoints (25% impact)  
3. **Firebase Auth** - Security layer (15% impact)

### High Value (Should Have - 15% of value)  
4. **TracksHandler** - Track management
5. **NostrTrackService** - Nostr integration
6. **NIP-98 Auth** - Alternative authentication

### Nice to Have (Could Have - 5% of value)
7. **Audio Utils** - Audio processing helpers
8. **Models** - Data structures
9. **Configuration** - Setup logic

## 🎯 Success Metrics

### Coverage Targets by Layer
- **Services**: 85%+ (complex business logic)
- **Handlers**: 80%+ (HTTP interface)  
- **Auth**: 90%+ (security critical)
- **Utils**: 70%+ (utility functions)
- **Models**: 60%+ (data structures)

### Quality Gates
- All tests pass before merge
- No regression in integration tests  
- Coverage reports generated automatically (`task coverage:backend`)
- Mock coverage ≥95% for external dependencies
- Test execution time <30 seconds for unit tests

## 🔧 Development Workflow

### TDD Implementation Process

```bash
# 1. Generate mocks for new interfaces
task mocks:generate

# 2. Create failing test first (RED)
task red

# 3. Implement minimal code (GREEN)
task green  

# 4. Refactor with passing tests (REFACTOR)
task refactor

# 5. Verify coverage
task coverage:backend

# 6. Run full test suite
task test:unit:backend
```

### File Naming Convention
- **Test files**: `*_test.go` alongside source files
- **Mock files**: `tests/mocks/*_mocks.go`  
- **Test utilities**: `tests/testutil/*.go`
- **Integration tests**: `tests/integration/*_test.go` (existing)

### Test Organization Pattern

Each test file should follow this structure:

```go
package services_test

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/golang/mock/gomock"
    
    "github.com/wavlake/monorepo/internal/services"
    "github.com/wavlake/monorepo/tests/mocks"
    "github.com/wavlake/monorepo/tests/testutil"
)

var _ = Describe("UserService", func() {
    var (
        ctrl            *gomock.Controller
        mockFirestore   *mocks.MockClient
        mockAuth        *mocks.MockClient  
        userService     *services.UserService
    )

    BeforeEach(func() {
        ctrl = gomock.NewController(GinkgoT())
        mockFirestore = mocks.NewMockClient(ctrl)
        mockAuth = mocks.NewMockClient(ctrl)
        userService = services.NewUserService(mockFirestore, mockAuth)
    })

    AfterEach(func() {
        ctrl.Finish()
    })

    Describe("LinkPubkeyToUser", func() {
        Context("when linking a new pubkey", func() {
            It("should create user and link pubkey successfully", func() {
                // Test implementation
            })
        })
        
        Context("when pubkey already exists", func() {
            It("should return error for different user", func() {
                // Test implementation  
            })
        })
    })
})
```

## 🚀 Implementation Timeline

### Week 1: Foundation Setup
- [ ] Set up mock generation infrastructure
- [ ] Create test utilities and fixtures
- [ ] Implement UserService unit tests
- [ ] Target: 40% overall coverage

### Week 2: Core Services & Handlers  
- [ ] Complete AuthHandlers unit tests
- [ ] Implement NostrTrackService unit tests
- [ ] Add TracksHandler unit tests
- [ ] Target: 65% overall coverage

### Week 3: Authentication Layer
- [ ] Firebase authentication tests
- [ ] NIP-98 authentication tests  
- [ ] Dual authentication tests
- [ ] Target: 75% overall coverage

### Week 4: Utilities & Polish
- [ ] Audio processing tests
- [ ] Storage utilities tests
- [ ] Models and configuration tests
- [ ] Final coverage optimization
- [ ] Target: 80%+ overall coverage

## 📝 Additional Considerations

### External Dependencies
- **FFmpeg**: Mock for audio processing tests (currently missing)
- **Firebase Emulators**: Use for integration tests, mock for unit tests
- **PostgreSQL**: Mock for unit tests, real DB for integration tests

### Performance Considerations
- Unit tests should run in <30 seconds total
- Use table-driven tests for multiple scenarios
- Parallel test execution where possible
- Optimize mock setup/teardown

### Continuous Integration
- Add coverage reporting to CI/CD pipeline
- Fail builds on coverage regression
- Generate and store coverage badges
- Run tests on multiple Go versions

This comprehensive plan provides a clear roadmap from 0% to 80%+ backend unit test coverage while maintaining code quality and following TDD best practices.