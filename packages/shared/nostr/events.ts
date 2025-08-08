// Nostr Event Type Definitions
// Based on NIPs (Nostr Implementation Possibilities)
// https://github.com/nostr-protocol/nips

/**
 * Base Nostr Event structure (NIP-01)
 */
export interface NostrEvent {
  id: string;           // 32-bytes lowercase hex-encoded sha256 of the serialized event data
  pubkey: string;       // 32-bytes lowercase hex-encoded public key of the event creator
  created_at: number;   // unix timestamp in seconds
  kind: number;         // integer between 0 and 65535
  tags: string[][];     // array of arrays of arbitrary strings
  content: string;      // arbitrary string
  sig: string;          // 64-bytes lowercase hex signature of the sha256 hash
}

/**
 * Unsigned event (for signing)
 */
export interface UnsignedNostrEvent {
  pubkey: string;
  created_at: number;
  kind: number;
  tags: string[][];
  content: string;
}

/**
 * Event kinds used in Wavlake
 */
export enum NostrEventKind {
  // Standard kinds (NIP-01)
  SET_METADATA = 0,           // User metadata
  TEXT_NOTE = 1,              // Short text note
  RECOMMEND_SERVER = 2,       // Recommend server
  CONTACTS = 3,               // Contact list
  ENCRYPTED_DIRECT_MESSAGE = 4, // Encrypted direct message
  DELETE = 5,                 // Event deletion

  // Music-specific kinds (NIP-XX - custom or proposed)
  TRACK_METADATA = 31337,     // Track metadata and info
  ALBUM_METADATA = 31338,     // Album metadata and info
  ARTIST_METADATA = 31339,    // Artist metadata and info
  PLAYLIST = 31340,           // Playlist
  MUSIC_REVIEW = 31341,       // Music review/rating
  MUSIC_SHARE = 31342,        // Music share/recommendation
  LIVE_PERFORMANCE = 31343,   // Live performance event
  
  // Wavlake-specific kinds
  PAYMENT_REQUEST = 40001,    // Lightning payment request
  PAYMENT_RECEIPT = 40002,    // Payment confirmation
  TIP = 40003,               // Lightning tip
  BOOST = 40004,             // Boost/amplify content
  
  // Social features
  LIKE = 7,                  // Like/reaction (NIP-25)
  REPOST = 6,               // Repost (NIP-18)
  COMMENT = 1,              // Comment on music (uses regular text note)
}

/**
 * User metadata event (kind 0)
 */
export interface UserMetadataEvent extends NostrEvent {
  kind: NostrEventKind.SET_METADATA;
  content: string; // JSON stringified UserMetadata
}

export interface UserMetadata {
  name?: string;
  about?: string;
  picture?: string;
  banner?: string;
  display_name?: string;
  website?: string;
  nip05?: string;        // NIP-05 identifier
  lud16?: string;        // Lightning address (LNURL)
  lud06?: string;        // LNURL-pay
}

/**
 * Text note event (kind 1)
 */
export interface TextNoteEvent extends NostrEvent {
  kind: NostrEventKind.TEXT_NOTE;
  content: string;       // Plain text content
}

/**
 * Contact list event (kind 3)
 */
export interface ContactListEvent extends NostrEvent {
  kind: NostrEventKind.CONTACTS;
  content: string;       // Usually empty or relay list JSON
  tags: [string, string, string?, string?][]; // ["p", pubkey, relay?, petname?]
}

/**
 * Track metadata event (custom kind)
 */
export interface TrackMetadataEvent extends NostrEvent {
  kind: NostrEventKind.TRACK_METADATA;
  content: string;       // JSON stringified TrackMetadata
}

export interface TrackMetadata {
  title: string;
  artist: string;
  album?: string;
  genre?: string;
  duration?: number;     // Duration in seconds
  artwork_url?: string;
  audio_url?: string;
  isrc?: string;         // International Standard Recording Code
  bpm?: number;
  key?: string;          // Musical key
  year?: number;
  description?: string;
  tags?: string[];       // Genre tags, mood tags, etc.
  license?: string;      // Creative Commons, etc.
  price_msat?: number;   // Price in millisatoshi
}

/**
 * Album metadata event (custom kind)
 */
export interface AlbumMetadataEvent extends NostrEvent {
  kind: NostrEventKind.ALBUM_METADATA;
  content: string;       // JSON stringified AlbumMetadata
}

export interface AlbumMetadata {
  title: string;
  artist: string;
  artwork_url?: string;
  year?: number;
  genre?: string;
  description?: string;
  total_tracks?: number;
  tracks?: string[];     // Array of track event IDs
  upc?: string;          // Universal Product Code
  label?: string;        // Record label
  price_msat?: number;   // Album price in millisatoshi
}

/**
 * Artist metadata event (custom kind)
 */
export interface ArtistMetadataEvent extends NostrEvent {
  kind: NostrEventKind.ARTIST_METADATA;
  content: string;       // JSON stringified ArtistMetadata
}

export interface ArtistMetadata {
  name: string;
  bio?: string;
  image_url?: string;
  banner_url?: string;
  website?: string;
  location?: string;
  genres?: string[];
  social_links?: {
    twitter?: string;
    instagram?: string;
    bandcamp?: string;
    spotify?: string;
    soundcloud?: string;
  };
  lud16?: string;        // Lightning address for tips
}

/**
 * Playlist event (custom kind)
 */
export interface PlaylistEvent extends NostrEvent {
  kind: NostrEventKind.PLAYLIST;
  content: string;       // JSON stringified PlaylistMetadata
  tags: [string, string][]; // ["e", track_event_id] for each track
}

export interface PlaylistMetadata {
  name: string;
  description?: string;
  image_url?: string;
  is_public: boolean;
  created_at: number;
  updated_at: number;
}

/**
 * Music review event (custom kind)
 */
export interface MusicReviewEvent extends NostrEvent {
  kind: NostrEventKind.MUSIC_REVIEW;
  content: string;       // Review text
  tags: [string, string][]; // ["e", reviewed_event_id], ["rating", "1-5"]
}

/**
 * Payment request event (custom kind)
 */
export interface PaymentRequestEvent extends NostrEvent {
  kind: NostrEventKind.PAYMENT_REQUEST;
  content: string;       // JSON stringified PaymentRequest
}

export interface PaymentRequest {
  bolt11: string;        // Lightning invoice
  amount_msat: number;   // Amount in millisatoshi
  description?: string;
  expires_at?: number;   // Unix timestamp
  for_event?: string;    // Event ID this payment is for (track, album, tip)
}

/**
 * Payment receipt event (custom kind)
 */
export interface PaymentReceiptEvent extends NostrEvent {
  kind: NostrEventKind.PAYMENT_RECEIPT;
  content: string;       // JSON stringified PaymentReceipt
}

export interface PaymentReceipt {
  payment_hash: string;
  preimage: string;
  amount_msat: number;
  paid_at: number;       // Unix timestamp
  for_event?: string;    // Event ID this payment was for
}

/**
 * Like/reaction event (NIP-25)
 */
export interface LikeEvent extends NostrEvent {
  kind: NostrEventKind.LIKE;
  content: string;       // Usually "+" or emoji
  tags: [string, string, string?][]; // ["e", event_id, relay?], ["p", pubkey, relay?]
}

/**
 * Repost event (NIP-18)
 */
export interface RepostEvent extends NostrEvent {
  kind: NostrEventKind.REPOST;
  content: string;       // JSON of reposted event or empty
  tags: [string, string, string?][]; // ["e", event_id, relay?], ["p", pubkey, relay?]
}

/**
 * Delete event (NIP-09)
 */
export interface DeleteEvent extends NostrEvent {
  kind: NostrEventKind.DELETE;
  content: string;       // Reason for deletion (optional)
  tags: [string, string][]; // ["e", deleted_event_id]
}

/**
 * Event filter for querying (NIP-01)
 */
export interface NostrFilter {
  ids?: string[];        // Event IDs
  authors?: string[];    // Pubkeys
  kinds?: number[];      // Event kinds
  since?: number;        // Unix timestamp
  until?: number;        // Unix timestamp
  limit?: number;        // Maximum number of events
  '#e'?: string[];       // Events referenced in 'e' tags
  '#p'?: string[];       // Pubkeys referenced in 'p' tags
  '#t'?: string[];       // Hashtags referenced in 't' tags
  search?: string;       // Full-text search (NIP-50)
}

/**
 * Relay message types (NIP-01)
 */
export type RelayMessage = 
  | ['EVENT', string, NostrEvent]
  | ['REQ', string, ...NostrFilter[]]
  | ['CLOSE', string]
  | ['EOSE', string]
  | ['NOTICE', string]
  | ['OK', string, boolean, string];

/**
 * Client message types
 */
export type ClientMessage =
  | ['EVENT', NostrEvent]
  | ['REQ', string, ...NostrFilter[]]
  | ['CLOSE', string];

/**
 * Subscription state for client
 */
export interface NostrSubscription {
  id: string;
  filters: NostrFilter[];
  active: boolean;
  created_at: number;
  events: NostrEvent[];
  eose_received: boolean; // End of stored events
}

/**
 * Relay connection info
 */
export interface RelayInfo {
  url: string;
  name?: string;
  description?: string;
  pubkey?: string;
  contact?: string;
  supported_nips?: number[];
  software?: string;
  version?: string;
  limitation?: {
    max_message_length?: number;
    max_subscriptions?: number;
    max_filters?: number;
    max_limit?: number;
    max_subid_length?: number;
    min_prefix?: number;
    max_event_tags?: number;
    max_content_length?: number;
    payment_required?: boolean;
    restricted_writes?: boolean;
  };
  payments_url?: string;  // LNURL for paid relays
  fees?: {
    admission?: { amount: number; unit: string };
    subscription?: { amount: number; unit: string; period: number };
    publication?: { amount: number; unit: string };
  };
}

/**
 * Utility types for the frontend
 */
export interface NostrEventWithMetadata extends NostrEvent {
  relay_url?: string;    // Which relay this event came from
  verified?: boolean;    // Signature verified
  seen_at?: number;      // When we first saw this event
}

/**
 * Parsed content types for different event kinds
 */
export type ParsedContent<T extends NostrEvent> = 
  T extends UserMetadataEvent ? UserMetadata :
  T extends TrackMetadataEvent ? TrackMetadata :
  T extends AlbumMetadataEvent ? AlbumMetadata :
  T extends ArtistMetadataEvent ? ArtistMetadata :
  T extends PlaylistEvent ? PlaylistMetadata :
  T extends PaymentRequestEvent ? PaymentRequest :
  T extends PaymentReceiptEvent ? PaymentReceipt :
  string;

/**
 * Event validation result
 */
export interface EventValidationResult {
  valid: boolean;
  error?: string;
  warnings?: string[];
}

/**
 * Relay connection status
 */
export enum RelayConnectionStatus {
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  ERROR = 'error',
  RECONNECTING = 'reconnecting'
}

export interface RelayConnection {
  url: string;
  status: RelayConnectionStatus;
  last_connected?: number;
  error?: string;
  stats: {
    events_sent: number;
    events_received: number;
    subscriptions_active: number;
  };
}