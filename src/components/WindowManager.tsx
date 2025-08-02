import { useState, useEffect, useCallback } from 'react'
import MonacoEditor from './MonacoEditor'

interface Buffer {
  id: string
  name: string
  content: string
  type: 'text' | 'output'
  language?: string
}

interface WindowManagerProps {
  initialContent?: string
  language?: string
  onBufferSwitch?: (bufferId: string) => void
}

export default function WindowManager({ initialContent = '', language = 'javascript' }: WindowManagerProps) {
  const [buffers, setBuffers] = useState<Buffer[]>([
    { id: 'text-1', name: 'text', content: initialContent, type: 'text', language },
    { id: 'output-1', name: 'output', content: 'Output buffer ready...', type: 'output', language: 'text' }
  ])
  const [activeBufferId, setActiveBufferId] = useState('text-1')
  const [showCommandPalette, setShowCommandPalette] = useState(false)

  const activeBuffer = buffers.find(b => b.id === activeBufferId)

  const handleBufferChange = useCallback((bufferId: string, content: string) => {
    setBuffers(prev => prev.map(buffer => 
      buffer.id === bufferId ? { ...buffer, content } : buffer
    ))
  }, [])

  const switchBuffer = useCallback((bufferId: string) => {
    setActiveBufferId(bufferId)
    setShowCommandPalette(false)
  }, [])

  const createBuffer = useCallback((type: 'text' | 'output') => {
    const newBuffer: Buffer = {
      id: `${type}-${Date.now()}`,
      name: `${type}`,
      content: '',
      type,
      language: type === 'text' ? language : 'text'
    }
    setBuffers(prev => [...prev, newBuffer])
    setActiveBufferId(newBuffer.id)
    setShowCommandPalette(false)
  }, [language])

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setShowCommandPalette(true)
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])

  useEffect(() => {
    if (showCommandPalette) {
      const handleCommandKey = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
          setShowCommandPalette(false)
        } else if (e.key >= '1' && e.key <= '9') {
          const index = parseInt(e.key) - 1
          if (index < buffers.length) {
            switchBuffer(buffers[index].id)
          }
        } else if (e.key === 't') {
          createBuffer('text')
        } else if (e.key === 'o') {
          createBuffer('output')
        }
      }

      window.addEventListener('keydown', handleCommandKey)
      return () => window.removeEventListener('keydown', handleCommandKey)
    }
  }, [showCommandPalette, buffers, switchBuffer, createBuffer])

  return (
    <div className="h-full w-full relative bg-white dark:bg-gray-800">
      <div className="h-full flex flex-col">
        <div className="flex items-center justify-between p-2 bg-gray-100 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          <span className="font-mono text-sm text-gray-800 dark:text-gray-200">{activeBuffer?.name}</span>
          <span className="text-xs text-gray-500 dark:text-gray-400">Buffer {buffers.findIndex(b => b.id === activeBufferId) + 1}/{buffers.length}</span>
        </div>
        <div className="flex-1 min-h-0">
          {activeBuffer && activeBuffer.type === 'text' && (
            <MonacoEditor
              value={activeBuffer.content}
              onChange={(value) => handleBufferChange(activeBuffer.id, value)}
              language={activeBuffer.language || 'javascript'}
            />
          )}
          {activeBuffer && activeBuffer.type === 'output' && (
            <MonacoEditor
              value={activeBuffer.content}
              onChange={(value) => handleBufferChange(activeBuffer.id, value)}
              language="text"
              readOnly={true}
            />
          )}
        </div>
      </div>

      {showCommandPalette && (
        <div className="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 min-w-[400px] max-w-lg">
            <h3 className="text-lg font-semibold mb-4 text-gray-900 dark:text-gray-100">Command Palette (⌘K)</h3>
            <div className="space-y-4">
              <div>
                <div className="font-semibold mb-2 text-sm text-gray-700 dark:text-gray-300">Buffers ({buffers.length}):</div>
                <div className="space-y-1">
                  {buffers.map((buffer, index) => (
                    <div 
                      key={buffer.id} 
                      className={`px-3 py-2 rounded cursor-pointer flex items-center justify-between hover:bg-gray-100 dark:hover:bg-gray-700 ${
                        buffer.id === activeBufferId ? 'bg-gray-100 dark:bg-gray-700' : ''
                      }`}
                      onClick={() => switchBuffer(buffer.id)}
                    >
                      <div className="flex items-center space-x-2">
                        <span className="text-xs font-mono text-gray-500 dark:text-gray-400">{index + 1}</span>
                        <span className="text-sm text-gray-800 dark:text-gray-200">{buffer.name}</span>
                      </div>
                      <span className="text-xs px-2 py-0.5 rounded bg-gray-200 dark:bg-gray-600 text-gray-600 dark:text-gray-300">
                        {buffer.type}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
              
              <div className="border-t dark:border-gray-700 pt-3">
                <div className="font-semibold mb-2 text-sm text-gray-700 dark:text-gray-300">Actions:</div>
                <div className="space-y-1">
                  <div 
                    className="px-3 py-2 rounded cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center space-x-2"
                    onClick={() => createBuffer('text')}
                    onKeyDown={(e) => e.key === 'Enter' && createBuffer('text')}
                    tabIndex={0}
                    role="button"
                  >
                    <span className="text-xs font-mono text-gray-500 dark:text-gray-400">T</span>
                    <span className="text-sm text-gray-800 dark:text-gray-200">New text buffer</span>
                  </div>
                  <div 
                    className="px-3 py-2 rounded cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center space-x-2"
                    onClick={() => createBuffer('output')}
                    onKeyDown={(e) => e.key === 'Enter' && createBuffer('output')}
                    tabIndex={0}
                    role="button"
                  >
                    <span className="text-xs font-mono text-gray-500 dark:text-gray-400">O</span>
                    <span className="text-sm text-gray-800 dark:text-gray-200">New output buffer</span>
                  </div>
                </div>
              </div>
              
              <div className="text-xs text-gray-500 dark:text-gray-400 pt-2">
                Press number key to switch • T/O to create • ESC to close
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}