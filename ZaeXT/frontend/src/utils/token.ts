const TOKEN_KEY = 'ai-chat-token'

export const getToken = (): string | null =>
  typeof window !== 'undefined' ? window.localStorage.getItem(TOKEN_KEY) : null

export const setToken = (token: string) => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(TOKEN_KEY, token)
}

export const clearToken = () => {
  if (typeof window === 'undefined') return
  window.localStorage.removeItem(TOKEN_KEY)
}

export const getAuthHeader = () => {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export { TOKEN_KEY }
