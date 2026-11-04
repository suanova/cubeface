import { describe, expect, it } from 'vitest'

import { loginSchema, registerSchema } from './schemas'

const valid = {
  username: 'alice_01',
  email: 'alice@example.com',
  verificationCode: '123456',
  password: 'Passw0rd',
  confirmPassword: 'Passw0rd',
}

function errorsFor(input: Record<string, unknown>) {
  const result = registerSchema.safeParse({ ...valid, ...input })
  if (result.success) return {}
  const out: Record<string, string> = {}
  for (const issue of result.error.issues) {
    out[issue.path.join('.')] ??= issue.message
  }
  return out
}

describe('loginSchema', () => {
  it('accepts a username identifier', () => {
    expect(loginSchema.safeParse({ username: 'alice', password: 'x' }).success).toBe(true)
  })

  it('requires both fields', () => {
    const result = loginSchema.safeParse({ username: '', password: '' })
    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues.map((issue) => issue.path.join('.'))).toEqual([
        'username',
        'password',
      ])
    }
  })
})

describe('registerSchema', () => {
  it('accepts a valid registration', () => {
    expect(registerSchema.safeParse(valid).success).toBe(true)
  })

  it('accepts an omitted verification code', () => {
    const { verificationCode: _ignored, ...withoutCode } = valid
    expect(registerSchema.safeParse(withoutCode).success).toBe(true)
  })

  it.each([
    ['ab', 'too short'],
    ['_alice', 'starts with an underscore'],
    ['alice smith', 'contains a space'],
    ['a'.repeat(33), 'too long'],
    ['alice@example', 'contains an at sign'],
  ])('rejects username %j (%s)', (username) => {
    expect(errorsFor({ username }).username).toMatch(/3-32 characters/)
  })

  it('rejects malformed emails', () => {
    expect(errorsFor({ email: 'not-an-email' }).email).toBe('Please enter a valid email address')
  })

  it.each([
    ['short1A', /between 8 and 20/],
    ['lowercase1', /uppercase/],
    ['UPPERCASE1', /lowercase/],
    ['NoDigitsHere', /digit/],
    ['LongerThanTwentyPassw0rd', /at most 20/],
  ])('rejects password %j', (password, message) => {
    expect(errorsFor({ password, confirmPassword: password }).password).toMatch(message)
  })

  it("reports mismatched passwords on confirmPassword", () => {
    const errors = errorsFor({ confirmPassword: 'Different1' })
    expect(errors.confirmPassword).toBe("Passwords don't match.")
  })

  it('requires a password confirmation', () => {
    expect(errorsFor({ confirmPassword: '' }).confirmPassword).toBe(
      'Please confirm your password'
    )
  })
})
