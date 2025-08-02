import { DefaultService, OpenAPI } from '../generated'
import { Paste, Diff, CompletionResponse } from '../types'

// Configure the OpenAPI client
OpenAPI.BASE = 'http://localhost:8000'

export const pasteService = {
  create: async (text: string, lang: string): Promise<string> => {
    // The response is a redirect, we need to extract the ID from the Location header
    // For now, we'll use the old implementation until we fix the redirect handling
    const formData = new URLSearchParams()
    formData.append('text', text)
    formData.append('lang', lang)
    
    const fetchResponse = await fetch('/api/paste', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: formData,
    })
    
    const redirectUrl = fetchResponse.headers.get('location') || fetchResponse.url
    const url = new URL(redirectUrl)
    const id = url.searchParams.get('id')
    
    if (!id) throw new Error('Failed to create paste')
    return id
  },

  get: async (id: string): Promise<Paste> => {
    return DefaultService.getPaste(id)
  },

  getCompletion: async (text: string): Promise<CompletionResponse> => {
    return DefaultService.getCompletion({ text })
  },
}

export const diffService = {
  create: async (original: string, modified: string): Promise<string> => {
    // The response is a redirect, we need to extract the ID from the Location header
    // For now, we'll use the old implementation until we fix the redirect handling
    const formData = new URLSearchParams()
    formData.append('original', original)
    formData.append('modified', modified)
    
    const fetchResponse = await fetch('/api/diff', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: formData,
    })
    
    const redirectUrl = fetchResponse.headers.get('location') || fetchResponse.url
    const url = new URL(redirectUrl)
    const id = url.searchParams.get('id')
    
    if (!id) throw new Error('Failed to create diff')
    return id
  },

  update: async (id: string, original: string, modified: string): Promise<Diff> => {
    const formData = new URLSearchParams()
    formData.append('original', original)
    formData.append('modified', modified)
    
    const response = await fetch(`/api/diff?id=${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: formData,
    })
    
    if (!response.ok) {
      throw new Error(`Failed to update diff: ${response.statusText}`)
    }
    
    return response.json()
  },

  get: async (id: string): Promise<Diff> => {
    return DefaultService.getDiff(id)
  },
}