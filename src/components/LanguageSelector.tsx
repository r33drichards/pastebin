interface LanguageSelectorProps {
  value: string
  onChange: (value: string) => void
}

const languages = [
  { value: 'detect', label: 'detect' },
  { value: '', label: 'text' },
  { value: 'abap', label: 'abap' },
  { value: 'apex', label: 'apex' },
  { value: 'azcli', label: 'azcli' },
  { value: 'bat', label: 'bat' },
  { value: 'bicep', label: 'bicep' },
  { value: 'cameligo', label: 'cameligo' },
  { value: 'clojure', label: 'clojure' },
  { value: 'coffee', label: 'coffee' },
  { value: 'cpp', label: 'cpp' },
  { value: 'csharp', label: 'csharp' },
  { value: 'csp', label: 'csp' },
  { value: 'css', label: 'css' },
  { value: 'dart', label: 'dart' },
  { value: 'dockerfile', label: 'dockerfile' },
  { value: 'ecl', label: 'ecl' },
  { value: 'elixir', label: 'elixir' },
  { value: 'fsharp', label: 'fsharp' },
  { value: 'go', label: 'go' },
  { value: 'graphql', label: 'graphql' },
  { value: 'handlebars', label: 'handlebars' },
  { value: 'hcl', label: 'hcl' },
  { value: 'html', label: 'html' },
  { value: 'ini', label: 'ini' },
  { value: 'java', label: 'java' },
  { value: 'javascript', label: 'javascript' },
  { value: 'julia', label: 'julia' },
  { value: 'kotlin', label: 'kotlin' },
  { value: 'less', label: 'less' },
  { value: 'lexon', label: 'lexon' },
  { value: 'liquid', label: 'liquid' },
  { value: 'lua', label: 'lua' },
  { value: 'm3', label: 'm3' },
  { value: 'markdown', label: 'markdown' },
  { value: 'mips', label: 'mips' },
  { value: 'msdax', label: 'msdax' },
  { value: 'mysql', label: 'mysql' },
  { value: 'objective-c', label: 'objective-c' },
  { value: 'pascal', label: 'pascal' },
  { value: 'pascaligo', label: 'pascaligo' },
  { value: 'perl', label: 'perl' },
  { value: 'pgsql', label: 'pgsql' },
  { value: 'php', label: 'php' },
  { value: 'postiats', label: 'postiats' },
  { value: 'powerquery', label: 'powerquery' },
  { value: 'powershell', label: 'powershell' },
  { value: 'pug', label: 'pug' },
  { value: 'python', label: 'python' },
  { value: 'qsharp', label: 'qsharp' },
  { value: 'r', label: 'r' },
  { value: 'razor', label: 'razor' },
  { value: 'redis', label: 'redis' },
  { value: 'redshift', label: 'redshift' },
  { value: 'restructuredtext', label: 'restructuredtext' },
  { value: 'ruby', label: 'ruby' },
  { value: 'rust', label: 'rust' },
  { value: 'sb', label: 'sb' },
  { value: 'scala', label: 'scala' },
  { value: 'scheme', label: 'scheme' },
  { value: 'scss', label: 'scss' },
  { value: 'shell', label: 'shell' },
  { value: 'solidity', label: 'solidity' },
  { value: 'sophia', label: 'sophia' },
  { value: 'sparql', label: 'sparql' },
  { value: 'sql', label: 'sql' },
  { value: 'st', label: 'st' },
  { value: 'swift', label: 'swift' },
  { value: 'systemverilog', label: 'systemverilog' },
  { value: 'tcl', label: 'tcl' },
  { value: 'twig', label: 'twig' },
  { value: 'typescript', label: 'typescript' },
  { value: 'vb', label: 'vb' },
  { value: 'xml', label: 'xml' },
  { value: 'yaml', label: 'yaml' },
  { value: 'mermaid', label: 'mermaid' },
]

export default function LanguageSelector({ value, onChange }: LanguageSelectorProps) {
  return (
    <>
      <label className="font-medium text-gray-700" htmlFor="language">
        Select a language:
      </label>
      <select
        className="py-2 px-4 rounded-lg shadow-md"
        name="lang"
        id="language"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      >
        {languages.map((lang) => (
          <option key={lang.value} value={lang.value}>
            {lang.label}
          </option>
        ))}
      </select>
    </>
  )
}