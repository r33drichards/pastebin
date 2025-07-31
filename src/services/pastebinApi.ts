import {
  CreatePasteRequest,
  CreatePasteResponse,
  GetPasteRequest,
  GetPasteResponse,
  CreateDiffRequest,
  CreateDiffResponse,
  GetDiffRequest,
  GetDiffResponse,
  GetCompletionRequest,
  GetCompletionResponse
} from '../types';

// Simple HTTP-based API client for the Pastebin service
export class PastebinApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8000') {
    this.baseUrl = baseUrl;
  }

  async createPaste(request: CreatePasteRequest): Promise<CreatePasteResponse> {
    const response = await fetch(`${this.baseUrl}/api/paste`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Failed to create paste: ${response.statusText}`);
    }

    return response.json();
  }

  async getPaste(request: GetPasteRequest): Promise<GetPasteResponse> {
    const response = await fetch(`${this.baseUrl}/api/paste/${request.id}`);

    if (!response.ok) {
      throw new Error(`Failed to get paste: ${response.statusText}`);
    }

    return response.json();
  }

  async createDiff(request: CreateDiffRequest): Promise<CreateDiffResponse> {
    const response = await fetch(`${this.baseUrl}/api/diff`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Failed to create diff: ${response.statusText}`);
    }

    return response.json();
  }

  async getDiff(request: GetDiffRequest): Promise<GetDiffResponse> {
    const response = await fetch(`${this.baseUrl}/api/diff/${request.id}`);

    if (!response.ok) {
      throw new Error(`Failed to get diff: ${response.statusText}`);
    }

    return response.json();
  }

  async getCompletion(request: GetCompletionRequest): Promise<GetCompletionResponse> {
    const response = await fetch(`${this.baseUrl}/api/completion`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error(`Failed to get completion: ${response.statusText}`);
    }

    return response.json();
  }
}

// Export a default instance
export const pastebinApi = new PastebinApiClient(); 