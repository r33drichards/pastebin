import WindowManager from '../components/WindowManager'

export default function BufferTestPage() {
  return (
    <div className="h-screen w-screen bg-gray-50 dark:bg-gray-900">
      <WindowManager initialContent="// Type your code here" language="javascript" />
    </div>
  )
}