import { Routes, Route } from 'react-router-dom'
import HomePage from './pages/HomePage'
import PastePage from './pages/PastePage'
import DiffPage from './pages/DiffPage'
import HtmlPage from './pages/HtmlPage'
import BufferTestPage from './pages/BufferTestPage'

function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/paste" element={<PastePage />} />
      <Route path="/diff" element={<DiffPage />} />
      <Route path="/html" element={<HtmlPage />} />
      <Route path="/buffers" element={<BufferTestPage />} />
    </Routes>
  )
}

export default App