export interface JwtPayload {
  exp?: number
  iat?: number
  sub?: string
  [key: string]: unknown
}

export const decodeToken = (token: string): JwtPayload | null => {
  try {
    const [, payload] = token.split('.')
    if (!payload) {
      return null
    }
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const decoded =
      typeof window !== 'undefined'
        ? window.atob(normalized)
        : Buffer.from(normalized, 'base64').toString('utf-8')
    return JSON.parse(decoded) as JwtPayload
  } catch (error) {
    console.warn('[jwt] Failed to decode token', error)
    return null
  }
}

export const isTokenExpired = (token: string | null) => {
  if (!token) return true
  const payload = decodeToken(token)
  if (!payload?.exp) return false
  const expiry = payload.exp * 1000
  return Date.now() >= expiry
}
