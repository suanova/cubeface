import { Languages } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { Button } from '@/components/ui/button'

export function LanguageToggle() {
  const { i18n, t } = useTranslation()
  const isChinese = i18n.resolvedLanguage?.startsWith('zh')

  return (
    <Button
      variant='ghost'
      size='sm'
      className='h-9 gap-1.5 rounded-full px-3 text-xs font-medium'
      onClick={() => void i18n.changeLanguage(isChinese ? 'en' : 'zh')}
      aria-label={t('Switch language')}
      title={t('Switch language')}
    >
      <Languages className='h-4 w-4' />
      {isChinese ? 'EN' : '中文'}
    </Button>
  )
}
