// Re-export auto-generated types from OpenAPI
export type { Paste, Diff, CompletionResponse } from '../generated'

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