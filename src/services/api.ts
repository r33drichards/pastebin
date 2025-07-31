import axios from 'axios'
import { Paste, Diff, CompletionResponse } from '../types'

const api = axios.create({
  headers: {
    'Content-Type': 'application/json',
  },
  maxRedirects: 0,
  validateStatus: (status) => status < 400,
})

export const pasteService = {
  create: async (text: string, lang: string): Promise<string> => {
    const formData = new URLSearchParams()
    formData.append('text', text)
    formData.append('lang', lang)
    
    const response = await api.post('/paste', formData, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })
    
    const redirectUrl = response.request.responseURL || response.headers.location
    const url = new URL(redirectUrl)
    const id = url.searchParams.get('id')
    
    if (!id) throw new Error('Failed to create paste')
    return id
  },

  get: async (id: string): Promise<Paste> => {
    const response = await api.get(`/api/paste?id=${id}`)
    return response.data
  },

  getCompletion: async (text: string): Promise<CompletionResponse> => {
    const response = await api.post(`/api/complete?text=${encodeURIComponent(text)}`)
    return response.data
  },
}

export const diffService = {
  create: async (original: string, modified: string): Promise<string> => {
    const formData = new URLSearchParams()
    formData.append('original', original)
    formData.append('modified', modified)
    
    const response = await api.post('/diff', formData, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })
    
    const redirectUrl = response.request.responseURL || response.headers.location
    const url = new URL(redirectUrl)
    const id = url.searchParams.get('id')
    
    if (!id) throw new Error('Failed to create diff')
    return id
  },

  get: async (id: string): Promise<Diff> => {
    const response = await api.get(`/api/diff?id=${id}`)
    return response.data
  },
}