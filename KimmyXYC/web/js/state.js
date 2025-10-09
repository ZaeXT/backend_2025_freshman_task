const TOKEN_KEY = 'aib_token';
const USER_KEY = 'aib_user';

export function getToken() {
  return localStorage.getItem(TOKEN_KEY) || '';
}

export function setToken(t) {
  if (t) localStorage.setItem(TOKEN_KEY, t);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

export function getUser() {
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try { return JSON.parse(raw); } catch { return null; }
}

export function setUser(u) {
  if (u) localStorage.setItem(USER_KEY, JSON.stringify(u));
}

export function clearUser() {
  localStorage.removeItem(USER_KEY);
}

export function isLoggedIn() {
  return !!getToken();
}
