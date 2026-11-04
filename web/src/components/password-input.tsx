import { Eye, EyeOff } from 'lucide-react'
import { forwardRef, useState } from 'react'

import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

export const PasswordInput = forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(function PasswordInput({ className, ...props }, ref) {
  const [visible, setVisible] = useState(false)
  return (
    <div className='relative'>
      <Input
        ref={ref}
        type={visible ? 'text' : 'password'}
        className={cn('pr-10', className)}
        {...props}
      />
      <button
        type='button'
        onClick={() => setVisible((value) => !value)}
        className='absolute right-1 top-1/2 flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded text-muted-foreground hover:text-foreground'
        aria-label={visible ? 'Hide password' : 'Show password'}
        tabIndex={-1}
      >
        {visible ? <EyeOff className='h-4 w-4' /> : <Eye className='h-4 w-4' />}
      </button>
    </div>
  )
})
