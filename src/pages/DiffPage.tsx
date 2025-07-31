import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { DiffEditor } from '@monaco-editor/react'
import Header from '../components/Header'
import { diffService } from '../services/api'
import { Diff } from '../types'

export default function DiffPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const id = searchParams.get('id')
  
  const [original, setOriginal] = useState('')
  const [modified, setModified] = useState('')

  // Fetch diff if ID is provided
  const { data: diffData, isLoading, error } = useQuery({
    queryKey: ['diff', id],
    queryFn: async (): Promise<Diff> => {
      if (!id) throw new Error('No diff ID provided')
      const response = await fetch(`/api/diff?id=${id}`)
      if (!response.ok) {
        throw new Error(`Failed to fetch diff: ${response.status}`)
      }
      return response.json()
    },
    enabled: !!id,
  })

  // Update state when diff data is loaded
  useState(() => {
    if (diffData) {
      setOriginal(diffData.oldText)
      setModified(diffData.newText)
    }
  })

  const createDiffMutation = useMutation({
    mutationFn: () => diffService.create(original, modified),
    onSuccess: (newId) => {
      navigate(`/diff?id=${newId}`)
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (original || modified) {
      createDiffMutation.mutate()
    }
  }

  const theme = window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'vs-dark'
    : 'vs-light'

  if (id && diffData) {
    return (
      <div className="container-xl h-screen overflow-y-hidden flex flex-col">
        <Header />
        <div className="flex-grow">
          <DiffEditor
            original={diffData.oldText}
            modified={diffData.newText}
            language="text"
            theme={theme}
            options={{
              readOnly: true,
              automaticLayout: true,
            }}
          />
        </div>
      </div>
    )
  }

  if (id && isLoading) {
    return (
      <div className="container-xl h-screen overflow-y-hidden">
        <Header />
        <div className="p-4">Loading...</div>
      </div>
    )
  }

  if (id && error) {
    return (
      <div className="container-xl h-screen overflow-y-hidden">
        <Header />
        <div className="p-4">
          <p>Error loading diff: {error.message}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="container-xl h-screen overflow-y-hidden flex flex-col">
      <Header>
        <button
          type="button"
          className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-green-500 hover:bg-green-700 ml-2"
          onClick={handleSubmit}
          disabled={createDiffMutation.isPending || (!original && !modified)}
        >
          <i className="fas fa-code-branch"></i> Create Diff
        </button>
      </Header>
      <div className="flex-grow">
        <DiffEditor
          original={original}
          modified={modified}
          language="text"
          theme={theme}
          onMount={(editor) => {
            const modifiedEditor = editor.getModifiedEditor()
            const originalEditor = editor.getOriginalEditor()
            
            modifiedEditor.onDidChangeModelContent(() => {
              setModified(modifiedEditor.getValue())
            })
            
            originalEditor.onDidChangeModelContent(() => {
              setOriginal(originalEditor.getValue())
            })
          }}
          options={{
            automaticLayout: true,
          }}
        />
      </div>
    </div>
  )
}