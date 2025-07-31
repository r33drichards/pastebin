import { useEffect, useState } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import Header from '../components/Header'
import MonacoEditor from '../components/MonacoEditor'

export default function PastePage() {
  const [searchParams] = useSearchParams()
  const id = searchParams.get('id')
  const [pasteData, setPasteData] = useState<{ text: string; language: string; title?: string } | null>(null)

  // Since the API returns HTML, we need to parse it
  const { data: htmlContent } = useQuery({
    queryKey: ['paste', id],
    queryFn: async () => {
      if (!id) throw new Error('No paste ID provided')
      const response = await fetch(`/paste?id=${id}`)
      return response.text()
    },
    enabled: !!id,
  })

  useEffect(() => {
    if (htmlContent) {
      // Parse the HTML to extract the paste data
      const parser = new DOMParser()
      const doc = parser.parseFromString(htmlContent, 'text/html')
      
      // Extract text from the script that sets up monaco
      const scripts = doc.querySelectorAll('script')
      let text = ''
      let language = ''
      let title = ''
      
      scripts.forEach(script => {
        const content = script.textContent || ''
        if (content.includes('monaco.editor.create')) {
          // Extract value
          const valueMatch = content.match(/value:\s*(.+?),\s*language:/s)
          if (valueMatch) {
            try {
              // The value is a Go template expression, we need to evaluate it carefully
              text = JSON.parse(valueMatch[1])
            } catch {
              // If JSON parse fails, try to extract the raw string
              text = valueMatch[1].replace(/^"|"$/g, '')
            }
          }
          
          // Extract language
          const langMatch = content.match(/language:\s*"?(.+?)"?,/)
          if (langMatch) {
            language = langMatch[1].replace(/"/g, '')
          }
        }
      })
      
      // Extract title
      const titleElement = doc.querySelector('title')
      if (titleElement && titleElement.textContent !== 'PBIN pastebin with Monaco Editor') {
        title = titleElement.textContent || ''
      }
      
      setPasteData({ text, language, title })
    }
  }, [htmlContent])

  const copyText = () => {
    if (pasteData?.text) {
      navigator.clipboard.writeText(pasteData.text)
    }
  }

  const shareLink = () => {
    navigator.clipboard.writeText(window.location.href)
  }

  if (!id) {
    return (
      <div className="container-xl h-screen overflow-y-hidden">
        <Header />
        <div className="p-4">
          <p>No paste ID provided. <Link to="/" className="text-blue-500">Go back to home</Link></p>
        </div>
      </div>
    )
  }

  if (!pasteData) {
    return (
      <div className="container-xl h-screen overflow-y-hidden">
        <Header />
        <div className="p-4">Loading...</div>
      </div>
    )
  }

  return (
    <div className="container-xl h-screen overflow-y-hidden">
      <Header>
        {pasteData.title && (
          <span className="py-2 px-4 font-semibold text-gray-700">{pasteData.title}</span>
        )}
        <button
          className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-green-500 hover:bg-green-700 ml-2"
          onClick={copyText}
        >
          <i className="far fa-copy"></i> Copy Text
        </button>
        <button
          className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-green-500 hover:bg-grey-700 ml-2"
          onClick={shareLink}
        >
          <i className="far fa-share-square"></i> Share Link to Text
        </button>
        {pasteData.language === 'markdown' && (
          <Link
            className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-green-500 hover:bg-grey-700 ml-2 inline-block"
            to={`/html?id=${id}`}
          >
            <i className="far fa-eye"></i> View HTML
          </Link>
        )}
      </Header>
      <div id="monacoContainer" className="w-full h-4/5">
        <MonacoEditor
          value={pasteData.text}
          language={pasteData.language}
          readOnly={true}
        />
      </div>
    </div>
  )
}