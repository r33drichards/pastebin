import { useEffect } from 'react'

interface HotkeyAction {
  id: string
  title: string
  handler: () => void
}

export function useNinjaKeys(actions: HotkeyAction[]) {
  useEffect(() => {
    const ninja = document.querySelector('ninja-keys') as any
    if (ninja) {
      ninja.data = actions
    }
  }, [actions])
}