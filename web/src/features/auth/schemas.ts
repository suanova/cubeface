import { z } from 'zod'

export const loginSchema = z.object({
  username: z.string().min(1, 'Please enter your username or email'),
  password: z.string().min(1, 'Please enter your password'),
})

export type LoginValues = z.infer<typeof loginSchema>

export const registerSchema = z
  .object({
    username: z
      .string()
      .min(1, 'Please enter your username')
      .regex(
        /^[A-Za-z0-9][A-Za-z0-9_.-]{2,31}$/,
        'Username must be 3-32 characters and may contain letters, digits, dots, hyphens and underscores'
      ),
    email: z
      .string()
      .min(1, 'Please enter your email')
      .email('Please enter a valid email address'),
    verificationCode: z.string().optional(),
    password: z
      .string()
      .min(1, 'Please enter your password')
      .min(8, 'Password must be between 8 and 20 characters')
      .max(20, 'Password must be at most 20 characters long')
      .regex(
        /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).*$/,
        'Password must contain at least one uppercase letter, one lowercase letter and one digit'
      ),
    confirmPassword: z.string().min(1, 'Please confirm your password'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match.",
    path: ['confirmPassword'],
  })

export type RegisterValues = z.infer<typeof registerSchema>

export const VERIFICATION_COUNTDOWN_SECONDS = 60
