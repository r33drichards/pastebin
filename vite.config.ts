import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'static',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom', 'react-router-dom'],
          query: ['@tanstack/react-query', 'axios'],
          monaco: ['@monaco-editor/react'],
          mermaid: ['mermaid']
        }
      }
    }
  },
  server: {
    proxy: {
      '/paste': 'http://localhost:8000',
      '/diff': 'http://localhost:8000',
      '/html': 'http://localhost:8000',
      '/complete': 'http://localhost:8000',
      '/health': 'http://localhost:8000'
    }
  }
})