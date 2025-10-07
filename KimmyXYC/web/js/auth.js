import { setToken, setUser } from './state.js';
import { login, register, me } from './api.js';

export function initAuthUI({ onAuthenticated } = {}) {
  const loginForm = document.getElementById('login-form');
  const loginEmail = document.getElementById('login-email');
  const loginPassword = document.getElementById('login-password');
  const loginError = document.getElementById('login-error');

  const regForm = document.getElementById('register-form');
  const regEmail = document.getElementById('register-email');
  const regPassword = document.getElementById('register-password');
  const regRole = document.getElementById('register-role');
  const regError = document.getElementById('register-error');

  loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    loginError.textContent = '';
    try {
      const resp = await login(loginEmail.value.trim(), loginPassword.value);
      setToken(resp.token);
      // Get me to store role/email consistently
      const profile = await me();
      setUser({ email: profile.user_email, role: profile.user_role, id: profile.user_id });
      if (onAuthenticated) onAuthenticated();
    } catch (err) {
      loginError.textContent = err.message || '登录失败';
    }
  });

  regForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    regError.textContent = '';
    try {
      const resp = await register(regEmail.value.trim(), regPassword.value, regRole.value);
      setToken(resp.token);
      const profile = await me();
      setUser({ email: profile.user_email, role: profile.user_role, id: profile.user_id });
      if (onAuthenticated) onAuthenticated();
    } catch (err) {
      regError.textContent = err.message || '注册失败';
    }
  });
}
