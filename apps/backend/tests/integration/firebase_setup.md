# Firebase Emulator Integration Test Setup

## Prerequisites

### 1. Java Installation
Firebase emulators require Java. Install via:

```bash
# macOS
brew install openjdk@11
echo 'export PATH="/opt/homebrew/opt/openjdk@11/bin:$PATH"' >> ~/.zshrc

# Verify installation
java -version
```

### 2. Firebase Configuration
Create `firebase.json` in project root:

```json
{
  "emulators": {
    "auth": {
      "port": 9099
    },
    "firestore": {
      "port": 8080
    },
    "storage": {
      "port": 9199
    },
    "ui": {
      "enabled": true,
      "port": 4000
    }
  }
}
```

### 3. Environment Variables
The integration tests automatically set these when emulators are running:
- `FIRESTORE_EMULATOR_HOST=localhost:8080`
- `FIREBASE_AUTH_EMULATOR_HOST=localhost:9099`
- `FIREBASE_STORAGE_EMULATOR_HOST=localhost:9199`

## Running Integration Tests

### Start Emulators
```bash
task firebase:emulators
```

### Run Tests
```bash
# All integration tests
task test:integration

# UserService integration tests only
cd apps/backend && go test -tags=integration ./tests/integration -run TestUserServiceIntegration -v
```

## What the Tests Cover

### UserService Implementation Testing
- **Real Firestore Transactions**: Tests actual transaction logic, not mocks
- **Data Persistence**: Verifies data is correctly written to/read from Firestore
- **Concurrent Operations**: Tests transaction safety under concurrent access
- **Error Scenarios**: Real Firebase error handling and validation
- **Data Consistency**: Ensures users and nostr_auth collections stay synchronized

### Benefits Over Interface Testing
- Tests actual Firebase client behavior
- Validates transaction isolation and consistency
- Catches Firestore-specific edge cases
- Tests real error conditions and recovery
- Validates data serialization/deserialization

## Current Coverage Impact

Running these integration tests will significantly improve the **4.2% services coverage** gap by testing the actual implementation code paths that interface tests cannot reach.

**Expected Coverage Improvement**: 4.2% → ~35% for services layer