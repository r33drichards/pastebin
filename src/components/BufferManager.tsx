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

interface TileConfig {
  bufferId: string
  width?: number // percentage of width
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
  const [tiles, setTiles] = useState<TileConfig[]>([{ bufferId: 'text-1' }])
  const [activeTileIndex, setActiveTileIndex] = useState(0)


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

  const splitHorizontal = useCallback((existingBufferId?: string) => {
    const bufferId = existingBufferId || activeBufferId
    const currentTileIndex = tiles.findIndex(tile => tile.bufferId === bufferId)
    
    if (currentTileIndex === -1) return

    // Create a new buffer or prompt to select one
    const newBuffer: Buffer = {
      id: `text-${Date.now()}`,
      name: 'text',
      content: '',
      type: 'text',
      language: language
    }
    setBuffers(prev => [...prev, newBuffer])

    // Insert new tile after current tile
    const newTiles = [...tiles]
    newTiles.splice(currentTileIndex + 1, 0, { bufferId: newBuffer.id })
    setTiles(newTiles)
    setActiveTileIndex(currentTileIndex + 1)
    setActiveBufferId(newBuffer.id)
  }, [tiles, activeBufferId, language])

  const splitWithExisting = useCallback((targetBufferId: string) => {
    const currentTileIndex = activeTileIndex
    
    // Insert tile with existing buffer
    const newTiles = [...tiles]
    newTiles.splice(currentTileIndex + 1, 0, { bufferId: targetBufferId })
    setTiles(newTiles)
    setActiveTileIndex(currentTileIndex + 1)
    setActiveBufferId(targetBufferId)
  }, [tiles, activeTileIndex])

  const closeTile = useCallback((tileIndex: number) => {
    if (tiles.length <= 1) return // Keep at least one tile
    
    const newTiles = tiles.filter((_, index) => index !== tileIndex)
    setTiles(newTiles)
    
    // Adjust active tile index
    if (tileIndex === activeTileIndex) {
      const newIndex = Math.min(tileIndex, newTiles.length - 1)
      setActiveTileIndex(newIndex)
      setActiveBufferId(newTiles[newIndex].bufferId)
    } else if (tileIndex < activeTileIndex) {
      setActiveTileIndex(activeTileIndex - 1)
    }
  }, [tiles, activeTileIndex])

  const switchToTile = useCallback((tileIndex: number) => {
    if (tileIndex >= 0 && tileIndex < tiles.length) {
      setActiveTileIndex(tileIndex)
      setActiveBufferId(tiles[tileIndex].bufferId)
    }
  }, [tiles])

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
    {
      id: 'split-horizontal',
      title: 'Split horizontal (new buffer)',
      handler: () => splitHorizontal()
    },
    ...buffers
      .filter(b => !tiles.some(tile => tile.bufferId === b.id))
      .map((buffer) => ({
        id: `split-with-${buffer.id}`,
        title: `Split with ${buffer.name}`,
        handler: () => splitWithExisting(buffer.id)
      })),
    ...(tiles.length > 1 ? [{
      id: 'close-tile',
      title: 'Close current tile',
      handler: () => closeTile(activeTileIndex)
    }] : []),
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
        {/* Header showing all tiles */}
        <div className="flex bg-gray-100 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          {tiles.map((tile, index) => {
            const buffer = buffers.find(b => b.id === tile.bufferId)
            const isActive = index === activeTileIndex
            return (
              <div
                key={`${tile.bufferId}-${index}`}
                className={`flex-1 flex items-center justify-between p-2 border-r border-gray-200 dark:border-gray-700 cursor-pointer ${
                  isActive ? 'bg-white dark:bg-gray-700' : 'hover:bg-gray-50 dark:hover:bg-gray-750'
                }`}
                onClick={() => switchToTile(index)}
              >
                <span className="font-mono text-sm text-gray-800 dark:text-gray-200">
                  {buffer?.name || 'Unknown'}
                </span>
                <div className="flex items-center space-x-2">
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {index + 1}/{tiles.length}
                  </span>
                  {tiles.length > 1 && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        closeTile(index)
                      }}
                      className="text-xs text-gray-400 hover:text-red-500 dark:text-gray-500 dark:hover:text-red-400"
                    >
                      ×
                    </button>
                  )}
                </div>
              </div>
            )
          })}
        </div>

        {/* Tiled editor area */}
        <div className="flex-1 min-h-0 flex">
          {tiles.map((tile, index) => {
            const buffer = buffers.find(b => b.id === tile.bufferId)
            const isActive = index === activeTileIndex
            
            if (!buffer) return null

            return (
              <div
                key={`${tile.bufferId}-${index}`}
                className={`flex-1 min-w-0 ${index > 0 ? 'border-l border-gray-200 dark:border-gray-700' : ''} ${
                  isActive ? 'ring-2 ring-blue-500 ring-inset' : ''
                }`}
                onClick={() => switchToTile(index)}
              >
                {buffer.type === 'text' && (
                  <MonacoEditor
                    value={buffer.content}
                    onChange={(value) => handleBufferChange(buffer.id, value)}
                    language={buffer.language || 'javascript'}
                  />
                )}
                {buffer.type === 'output' && (
                  <MonacoEditor
                    value={buffer.content}
                    onChange={(value) => handleBufferChange(buffer.id, value)}
                    language="text"
                    readOnly={true}
                  />
                )}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}