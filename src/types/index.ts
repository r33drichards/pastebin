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