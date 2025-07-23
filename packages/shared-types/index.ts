// Shared Types Index - Central export for all shared types

// Common utility types
export * from './common';

// Nostr types (manually maintained)
export * from './nostr';

// API types (generated from Go backend using tygo)
export * from './api';

// Re-export frequently used types for convenience
export type {
  // Common types
  ApiResponse,
  PaginatedResponse,
  ErrorResponse,
} from './common';

// Nostr types will be re-exported when needed
// export type {
//   NostrEvent,
//   NostrEventKind,
//   TrackMetadata,
//   AlbumMetadata,
//   UserMetadata,
// } from './nostr';