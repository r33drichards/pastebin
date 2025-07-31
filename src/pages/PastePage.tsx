import { useSearchParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import Header from '../components/Header'
import MonacoEditor from '../components/MonacoEditor'

interface PasteData {
  id: string
  text: string
  language: string
  title: string
}

export default function PastePage() {
  const [searchParams] = useSearchParams()
  const id = searchParams.get('id')

  // Use the JSON API directly
  const { data: pasteData, isLoading, error } = useQuery({
    queryKey: ['paste', id],
    queryFn: async (): Promise<PasteData> => {
      if (!id) throw new Error('No paste ID provided')
      const response = await fetch(`/paste?id=${id}`)
      if (!response.ok) {
        throw new Error(`Failed to fetch paste: ${response.status}`)
      }
      return response.json()
    },
    enabled: !!id,
  })

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

  if (isLoading) {
    return (
      <div className="container-xl h-screen overflow-y-hidden">
        <Header />
        <div className="p-4">Loading...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container-xl h-screen overflow-y-hidden">
        <Header />
        <div className="p-4">
          <p>Error loading paste: {error.message}</p>
          <Link to="/" className="text-blue-500">Go back to home</Link>
        </div>
      </div>
    )
  }

  if (!pasteData) {
    return (
      <div className="container-xl h-screen overflow-y-hidden">
        <Header />
        <div className="p-4">Paste not found.</div>
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
          type="button"
          className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-green-500 hover:bg-green-700 ml-2"
          onClick={copyText}
        >
          <i className="far fa-copy"></i> Copy Text
        </button>
        <button
          type="button"
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