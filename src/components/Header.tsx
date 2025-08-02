import { Link, useLocation } from 'react-router-dom'

interface HeaderProps {
  children?: React.ReactNode
}

export default function Header({ children }: HeaderProps) {
  const location = useLocation()
  
  return (
    <div className="max-h-1/5 p-4">
      <Link 
        className={`py-2 px-4 font-bold ${location.pathname === '/' ? 'text-blue-600' : 'text-black'}`} 
        to="/"
      >
        PBIN
      </Link>
      <Link 
        className={`py-2 px-4 font-bold ${location.pathname === '/diff' ? 'text-blue-600' : 'text-black'}`} 
        to="/diff"
      >
        DIFF
      </Link>
      <Link 
        className={`py-2 px-4 font-bold ${location.pathname === '/buffers' ? 'text-blue-600' : 'text-black'}`} 
        to="/buffers"
      >
        BUFFERS
      </Link>
      {children}
    </div>
  )
}