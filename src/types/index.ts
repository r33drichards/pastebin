// Core types for the Pastebin service
// These align with the protobuf definitions in proto/pastebin.proto

export interface Paste {
  id: string
  text: string
  language: string
  title?: string
}

export interface Diff {
  id: string
  oldText: string
  newText: string
}

export interface CompletionResponse {
  completions: string[]
}

// Request/Response types for API calls
export interface CreatePasteRequest {
  text: string
  language: string
}

export interface CreatePasteResponse {
  id: string
}

export interface GetPasteRequest {
  id: string
}

export interface GetPasteResponse {
  id: string
  text: string
  language: string
  title: string
}

export interface CreateDiffRequest {
  original: string
  modified: string
}

export interface CreateDiffResponse {
  id: string
}

export interface GetDiffRequest {
  id: string
}

export interface GetDiffResponse {
  id: string
  old_text: string
  new_text: string
}

export interface GetCompletionRequest {
  text: string
}

export interface GetCompletionResponse {
  completions: string[]
}