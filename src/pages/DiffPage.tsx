import { useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { DiffEditor } from '@monaco-editor/react'
import Header from '../components/Header'
import { diffService } from '../services/api'

export default function DiffPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const id = searchParams.get('id')
  
  const [original, setOriginal] = useState('')
  const [modified, setModified] = useState('')

  // Fetch diff if ID is provided
  const { data: diffData } = useQuery({
    queryKey: ['diff', id],
    queryFn: async () => {
      if (!id) return null
      const response = await fetch(`/diff?id=${id}`)
      const html = await response.text()
      
      // Parse the HTML to extract diff data
      const parser = new DOMParser()
      const doc = parser.parseFromString(html, 'text/html')
      
      // Find the script that contains the diff data
      const scripts = doc.querySelectorAll('script')
      let oldText = ''
      let newText = ''
      
      scripts.forEach(script => {
        const content = script.textContent || ''
        if (content.includes('oldText:') && content.includes('newText:')) {
          // Extract oldText
          const oldMatch = content.match(/oldText:\s*`([^`]*)`/)
          if (oldMatch) {
            oldText = oldMatch[1]
          }
          
          // Extract newText
          const newMatch = content.match(/newText:\s*`([^`]*)`/)
          if (newMatch) {
            newText = newMatch[1]
          }
        }
      })
      
      return { oldText, newText }
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

  return (
    <div className="container-xl h-screen overflow-y-hidden flex flex-col">
      <Header>
        <button
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