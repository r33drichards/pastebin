import { useState, useRef, useCallback, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import mermaid from 'mermaid'
import Header from '../components/Header'
import LanguageSelector from '../components/LanguageSelector'
import MonacoEditor from '../components/MonacoEditor'
import { pasteService } from '../services/api'
import { useNinjaKeys } from '../hooks/useNinjaKeys'


mermaid.initialize({
  theme: 'base',
  flowchart: {
    curve: 'basis',
  },
})

function debounce<T extends (...args: any[]) => void>(func: T, wait: number): T {
  let timeout: NodeJS.Timeout
  return ((...args) => {
    clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }) as T
}

export default function HomePage() {
  const navigate = useNavigate()
  const [code, setCode] = useState('')
  const [language, setLanguage] = useState('detect')
  const [output, setOutput] = useState('')
  const [showOutput, setShowOutput] = useState(true)
  const [autoRun, setAutoRun] = useState(false)
  const editorValueRef = useRef('')

  const createPasteMutation = useMutation({
    mutationFn: async () => {
      const lang = language === 'detect' ? await detectLanguage(code) : language
      return pasteService.create(code, lang)
    },
    onSuccess: (id) => {
      navigate(`/paste?id=${id}`)
    },
  })

  const detectLanguage = async (text: string): Promise<string> => {
    if (!window.GuessLang) return 'text'
    
    const guesser = new window.GuessLang()
    const result = await guesser.runModel(text)
    const lang = result.reduce(
      (a: any, b: any) => (a.confidence > b.confidence ? a : b),
      { languageId: 'text', confidence: 0 }
    ).languageId
    
    return convertGuessLangToMonacoLang(lang)
  }

  const convertGuessLangToMonacoLang = (lang: string): string => {
    const mapping: Record<string, string> = {
      asm: 'asm',
      bat: 'bat',
      c: 'cpp',
      cpp: 'cpp',
      clj: 'clojure',
      cmake: 'cmake',
      cbl: 'cobol',
      coffee: 'coffee',
      css: 'css',
      csv: 'csv',
      dart: 'dart',
      dm: 'dm',
      dockerfile: 'dockerfile',
      ex: 'elixir',
      erl: 'erlang',
      f90: 'fortran',
      go: 'go',
      groovy: 'groovy',
      hs: 'haskell',
      html: 'html',
      ini: 'ini',
      java: 'java',
      js: 'javascript',
      json: 'json',
      jl: 'julia',
      kt: 'kotlin',
      lisp: 'lisp',
      lua: 'lua',
      makefile: 'makefile',
      md: 'markdown',
      matlab: 'matlab',
      mm: 'objective-c',
      ml: 'ocaml',
      pas: 'pascal',
      pm: 'perl',
      php: 'php',
      ps1: 'powershell',
      prolog: 'prolog',
      py: 'python',
      r: 'r',
      rb: 'ruby',
      rs: 'rust',
      scala: 'scala',
      sh: 'shell',
      sql: 'sql',
      swift: 'swift',
      tex: 'tex',
      toml: 'toml',
      ts: 'typescript',
      v: 'verilog',
      vba: 'vb',
      xml: 'xml',
      yaml: 'yaml',
      mermaid: 'mermaid',
    }
    return mapping[lang] || 'text'
  }

  const runCode = useCallback(async () => {
    if (language === 'mermaid') {
      try {
        const { svg } = await mermaid.render('theGraph', code)
        setOutput(svg)
      } catch (err: any) {
        setOutput(`Mermaid render error: ${err}`)
      }
    } else {
      try {
        if (window.QJS) {
          const result = window.QJS.evalCode(code)
          setOutput(result)
        } else {
          setOutput('QuickJS not loaded')
        }
      } catch (err: any) {
        setOutput(`Error: ${err.message}`)
      }
    }
  }, [code, language])

  const debouncedRunCode = useCallback(debounce(runCode, 500), [runCode])

  useEffect(() => {
    if (autoRun && code) {
      debouncedRunCode()
    }
  }, [code, autoRun, debouncedRunCode])

  const toggleOutput = () => {
    setShowOutput(!showOutput)
  }

  const handlePaste = () => {
    createPasteMutation.mutate()
  }

  useNinjaKeys([
    {
      id: 'run code',
      title: 'run code',
      handler: runCode,
    },
    {
      id: 'paste',
      title: 'paste',
      handler: handlePaste,
    },
    {
      id: 'toggle output',
      title: 'toggle output',
      handler: toggleOutput,
    },
  ])

  return (
    <div className="container-xl h-screen overflow-y-hidden">
      <Header>
        <LanguageSelector value={language} onChange={setLanguage} />
        <button
          className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-green-500 hover:bg-green-700 ml-2"
          onClick={handlePaste}
          disabled={createPasteMutation.isPending}
        >
          <i className="fas fa-paste"></i> Paste
        </button>
        <button
          id="runButton"
          className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-blue-500 hover:bg-blue-700 ml-2"
          onClick={runCode}
          style={{ display: autoRun ? 'none' : '' }}
        >
          <i className="fas fa-play"></i> Run
        </button>
        <button
          id="toggleOutputButton"
          className="py-2 px-4 font-semibold rounded-lg shadow-md text-white bg-gray-500 hover:bg-gray-700 ml-2"
          onClick={toggleOutput}
        >
          <i className={`fas fa-eye${showOutput ? '-slash' : ''}`}></i>{' '}
          {showOutput ? 'Hide' : 'Show'} Output
        </button>
        <label className="ml-2 font-medium text-gray-700">
          <input
            type="checkbox"
            id="autoRunCheckbox"
            className="mr-1 align-middle"
            checked={autoRun}
            onChange={(e) => setAutoRun(e.target.checked)}
          />
          Auto Run
        </label>
      </Header>
      <div className="flex-grow overflow-auto">
        <div className="grid" id="panelContainer">
          <div className="panel" id="editorPanel">
            <div className="panel-header">Code Editor</div>
            <div className="panel-content">
              <div id="monacoContainer" className="w-full h-full">
                <MonacoEditor
                  value={code}
                  onChange={setCode}
                  language={language}
                  onRunCode={runCode}
                  onPaste={handlePaste}
                />
              </div>
            </div>
          </div>
          <div
            className="panel"
            id="outputPanel"
            style={{ display: showOutput ? 'block' : 'none' }}
          >
            <div className="panel-header">Output</div>
            <div className="panel-content">
              <div
                id="output"
                className="w-full h-full overflow-auto p-4 bg-gray-100"
                dangerouslySetInnerHTML={{ __html: output }}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}