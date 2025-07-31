declare global {
  interface Window {
    GuessLang: any
    QJS: any
    getQuickJS?: () => Promise<any>
  }
}

declare module 'ninja-keys' {
  const component: any
  export default component
}

export {}