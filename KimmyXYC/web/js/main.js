import { isLoggedIn, clearToken, clearUser, setUser } from './state.js';
import { me } from './api.js';
import { initAuthUI } from './auth.js';
import { initChatUI } from './chat.js';

function show(id) {
  document.querySelectorAll('.view').forEach(v => v.classList.add('hidden'));
  document.getElementById(id).classList.remove('hidden');
}

async function enterApp() {
  show('app-view');
  initChatUI();
}

async function enterAuth() {
  show('auth-view');
  initAuthUI({ onAuthenticated: enterApp });
}

async function bootstrap() {
  const logoutBtn = document.getElementById('logout-btn');
  if (logoutBtn) {
    logoutBtn.addEventListener('click', () => {
      clearToken();
      clearUser();
      location.reload();
    });
  }

  if (!isLoggedIn()) {
    await enterAuth();
    return;
  }

  // Validate token and fetch profile
  try {
    const profile = await me();
    setUser({ email: profile.user_email, role: profile.user_role, id: profile.user_id });
    await enterApp();
  } catch (err) {
    console.warn('Token invalid, returning to auth', err);
    clearToken();
    clearUser();
    await enterAuth();
  }
}

window.addEventListener('DOMContentLoaded', bootstrap);
