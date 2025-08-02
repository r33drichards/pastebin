import { useRef, useEffect } from 'react'
import Editor, { OnMount, Monaco } from '@monaco-editor/react'
import { pasteService } from '../services/api'

interface MonacoEditorProps {
  value: string
  onChange?: (value: string) => void
  language: string
  readOnly?: boolean
  onRunCode?: () => void
}

function convertGuessLangToMonacoLang(lang: string): string {
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
  return mapping[lang] || lang
}

export default function MonacoEditor({
  value,
  onChange,
  language,
  readOnly = false,
  onRunCode,
}: MonacoEditorProps) {
  const editorRef = useRef<any>(null)
  const monacoRef = useRef<Monaco | null>(null)

  const handleEditorDidMount: OnMount = (editor, monaco) => {
    editorRef.current = editor
    monacoRef.current = monaco

    // Register Mermaid language if not already present
    if (!monaco.languages.getLanguages().some((lang) => lang.id === 'mermaid')) {
      monaco.languages.register({ id: 'mermaid' })
      monaco.languages.setMonarchTokensProvider('mermaid', {
        tokenizer: {
          root: [
            [
              /(graph|flowchart|sequenceDiagram|classDiagram|stateDiagram|erDiagram|gantt|pie|journey|mindmap|timeline|gitGraph)/,
              'keyword',
            ],
            [/".*?"/, 'string'],
            [/\[.*?\]/, 'string'],
            [/\(.*?\)/, 'string'],
            [/\{.*?\}/, 'string'],
            [/\-\->|==\>|\-\-|==/, 'operator'],
            [/\w+/, 'identifier'],
          ],
        },
      })
    }

    // Register inline completion provider
    const inlineCompletionProvider = {
      freeInlineCompletions: () => {},
      provideInlineCompletions: async function (model: any, position: any) {
        const range = new monaco.Range(1, 1, position.lineNumber, position.column)
        const textUntilPosition = model.getValueInRange(range)
        
        try {
          const data = await pasteService.getCompletion(textUntilPosition)
          
          return {
            items: data.completions.map((completion) => ({
              text: completion,
              range: new monaco.Range(
                position.lineNumber,
                position.column,
                position.lineNumber,
                position.column
              ),
              insertText: completion,
              filterText: textUntilPosition,
            })),
            dispose: () => {},
          }
        } catch (error) {
          console.error('Completion error:', error)
          return { items: [], dispose: () => {} }
        }
      },
    }

    // Register the inline completion provider for all languages
    monaco.languages.getLanguages().forEach((lang) => {
      monaco.languages.registerInlineCompletionsProvider(
        lang.id,
        inlineCompletionProvider
      )
    })

    // Add keyboard shortcuts
    if (onRunCode) {
      editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, onRunCode)
    }

    // Prevent Monaco's default Command+K behavior to allow WindowManager to handle it
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyK, () => {
      // Do nothing - let WindowManager handle Command+K
    })

    editor.focus()
  }

  useEffect(() => {
    if (language === 'detect' && editorRef.current && monacoRef.current && window.GuessLang) {
      const detectLanguage = async () => {
        const guesser = new window.GuessLang()
        const result = await guesser.runModel(value)
        const detectedLang = result.reduce(
          (a: any, b: any) => (a.confidence > b.confidence ? a : b),
          { languageId: 'text', confidence: 0 }
        ).languageId
        
        const monacoLang = convertGuessLangToMonacoLang(detectedLang)
        const model = editorRef.current.getModel()
        if (model && monacoRef.current) {
          monacoRef.current.editor.setModelLanguage(model, monacoLang)
        }
      }
      
      detectLanguage()
    }
  }, [value, language])

  const theme = window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'vs-dark'
    : 'vs-light'

  return (
    <Editor
      height="100%"
      defaultLanguage={language === 'detect' ? 'text' : language}
      language={language === 'detect' ? undefined : language}
      value={value}
      onChange={(val) => onChange?.(val || '')}
      onMount={handleEditorDidMount}
      theme={theme}
      options={{
        readOnly,
        automaticLayout: true,
        wordWrap: 'on',
        inlineSuggest: {
          enabled: true,
          mode: 'prefix',
        },
      }}
    />
  )
}