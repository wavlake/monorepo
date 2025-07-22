// Shared Types Index - Central export for all shared types

// API types (generated from Go backend)
export * from './api';

// Nostr types (manually maintained)
export * from './nostr';

// Common utility types
export * from './common';

// Re-export frequently used types for convenience
export type {
  // Common types
  ApiResponse,
  PaginatedResponse,
  ErrorResponse,
  
  // Nostr types
  NostrEvent,
  NostrEventKind,
  TrackMetadata,
  AlbumMetadata,
  UserMetadata,
  
  // API types (will be generated)
  // User,
  // Track,
  // Album,
  // Playlist
} from './api';