import { Monitor, Moon, Sun } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'
import { type Theme, useTheme } from '@/context/theme-provider'

const NEXT_THEME: Record<Theme, Theme> = {
  light: 'dark',
  dark: 'system',
  system: 'light',
}

export function ThemeToggle() {
  const { t } = useTranslation()
  const { theme, setTheme } = useTheme()
  const Icon = theme === 'light' ? Sun : theme === 'dark' ? Moon : Monitor
  const next = NEXT_THEME[theme]
  const label = `${t('Toggle theme')}: ${t(next === 'light' ? 'Light' : next === 'dark' ? 'Dark' : 'System')}`

  return (
    <Button
      variant='ghost'
      size='icon'
      className='h-9 w-9 rounded-full'
      onClick={() => setTheme(next)}
      aria-label={label}
      title={label}
    >
      <Icon className='h-[18px] w-[18px]' />
    </Button>
  )
}
