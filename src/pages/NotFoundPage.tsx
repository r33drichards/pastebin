import { Link } from 'react-router-dom'

export default function NotFoundPage() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-gray-900 dark:text-gray-100 mb-4">404</h1>
        <h2 className="text-2xl font-semibold text-gray-700 dark:text-gray-300 mb-4">Page Not Found</h2>
        <p className="text-gray-600 dark:text-gray-400 mb-8">
          The page you're looking for doesn't exist.
        </p>
        <div className="space-x-4">
          <Link
            to="/"
            className="inline-block py-2 px-4 bg-blue-500 hover:bg-blue-700 text-white font-semibold rounded-lg shadow-md transition-colors"
          >
            Go Home
          </Link>
          <Link
            to="/buffers"
            className="inline-block py-2 px-4 bg-green-500 hover:bg-green-700 text-white font-semibold rounded-lg shadow-md transition-colors"
          >
            Try Buffers
          </Link>
        </div>
      </div>
    </div>
  )
}