import { useState, useCallback } from 'react'
import MonacoEditor from './MonacoEditor'
import { useNinjaKeys } from '../hooks/useNinjaKeys'

interface Buffer {
  id: string
  name: string
  content: string
  type: 'text' | 'output'
  language?: string
}

interface BufferManagerProps {
  initialContent?: string
  language?: string
  onBufferSwitch?: (bufferId: string) => void
}

export default function BufferManager({ initialContent = '', language = 'javascript' }: BufferManagerProps) {
  const [buffers, setBuffers] = useState<Buffer[]>([
    { id: 'text-1', name: 'text', content: initialContent, type: 'text', language },
    { id: 'output-1', name: 'output', content: 'Output buffer ready...', type: 'output', language: 'text' }
  ])
  const [activeBufferId, setActiveBufferId] = useState('text-1')

  const activeBuffer = buffers.find(b => b.id === activeBufferId)

  const handleBufferChange = useCallback((bufferId: string, content: string) => {
    setBuffers(prev => prev.map(buffer => 
      buffer.id === bufferId ? { ...buffer, content } : buffer
    ))
  }, [])

  const switchBuffer = useCallback((bufferId: string) => {
    setActiveBufferId(bufferId)
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
  }, [language])

  const deleteBuffer = useCallback((bufferId: string) => {
    if (buffers.length <= 1) return // Keep at least one buffer
    
    const bufferIndex = buffers.findIndex(b => b.id === bufferId)
    const newBuffers = buffers.filter(b => b.id !== bufferId)
    setBuffers(newBuffers)
    
    // Switch to another buffer if we're deleting the active one
    if (bufferId === activeBufferId) {
      const newIndex = Math.min(bufferIndex, newBuffers.length - 1)
      setActiveBufferId(newBuffers[newIndex].id)
    }
  }, [buffers, activeBufferId])

  const renameBuffer = useCallback((bufferId: string) => {
    const buffer = buffers.find(b => b.id === bufferId)
    if (!buffer) return
    
    const newName = prompt('Enter new buffer name:', buffer.name)
    if (newName && newName.trim()) {
      setBuffers(prev => prev.map(b => 
        b.id === bufferId ? { ...b, name: newName.trim() } : b
      ))
    }
  }, [buffers])

  // Create ninja-keys actions
  const ninjaActions = [
    ...buffers.map((buffer) => ({
      id: `switch-${buffer.id}`,
      title: `${buffer.name} (${buffer.type})`,
      handler: () => switchBuffer(buffer.id)
    })),
    {
      id: 'create-text-buffer',
      title: 'New text buffer',
      handler: () => createBuffer('text')
    },
    {
      id: 'create-output-buffer',
      title: 'New output buffer',
      handler: () => createBuffer('output')
    },
    ...buffers.map((buffer) => ({
      id: `rename-${buffer.id}`,
      title: `Rename ${buffer.name}`,
      handler: () => renameBuffer(buffer.id)
    })),
    ...(buffers.length > 1 ? buffers.map((buffer) => ({
      id: `delete-${buffer.id}`,
      title: `Delete ${buffer.name}`,
      handler: () => deleteBuffer(buffer.id)
    })) : [])
  ]

  // Use the ninja-keys hook
  useNinjaKeys(ninjaActions)

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
    </div>
  )
}