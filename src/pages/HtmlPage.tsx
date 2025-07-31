import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'

export default function HtmlPage() {
  const [searchParams] = useSearchParams()
  const id = searchParams.get('id')

  const { data: htmlContent, isLoading } = useQuery({
    queryKey: ['html', id],
    queryFn: async () => {
      if (!id) throw new Error('No paste ID provided')
      const response = await fetch(`/html?id=${id}`)
      return response.text()
    },
    enabled: !!id,
  })

  if (!id) {
    return <div>No paste ID provided</div>
  }

  if (isLoading) {
    return <div>Loading...</div>
  }

  return (
    <div
      className="w-full h-screen"
      dangerouslySetInnerHTML={{ __html: htmlContent || '' }}
    />
  )
}