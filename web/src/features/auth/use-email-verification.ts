import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { sendVerificationCode } from '@/features/auth/api'
import { VERIFICATION_COUNTDOWN_SECONDS } from '@/features/auth/schemas'
import { apiErrorMessage } from '@/lib/api'

export function useEmailVerification() {
  const { t } = useTranslation()
  const [secondsLeft, setSecondsLeft] = useState(0)
  const [isSending, setIsSending] = useState(false)

  useEffect(() => {
    if (secondsLeft <= 0) return
    const timer = window.setTimeout(() => setSecondsLeft((value) => value - 1), 1000)
    return () => window.clearTimeout(timer)
  }, [secondsLeft])

  const sendCode = useCallback(
    async (email: string): Promise<boolean> => {
      const value = email.trim()
      if (!value || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
        toast.error(t('Please enter a valid email address'))
        return false
      }
      setIsSending(true)
      try {
        const response = await sendVerificationCode(value)
        if (!response.success) {
          toast.error(t(response.message || 'Failed to send verification code'))
          return false
        }
        toast.success(t('Verification code sent'))
        setSecondsLeft(VERIFICATION_COUNTDOWN_SECONDS)
        return true
      } catch (error) {
        toast.error(apiErrorMessage(error) ?? t('Failed to send verification code'))
        return false
      } finally {
        setIsSending(false)
      }
    },
    [t]
  )

  return {
    isSending,
    secondsLeft,
    isActive: secondsLeft > 0,
    sendCode,
  }
}
