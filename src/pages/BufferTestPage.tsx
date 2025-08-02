import BufferManager from '../components/BufferManager'

export default function BufferTestPage() {
  return (
    <div className="h-screen w-screen bg-gray-50 dark:bg-gray-900">
      <BufferManager initialContent="// Type your code here" language="javascript" />
    </div>
  )
}