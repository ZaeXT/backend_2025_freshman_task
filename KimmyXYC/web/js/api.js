import { getToken } from './state.js';

async function api(path, { method = 'GET', headers = {}, body = undefined } = {}) {
  const h = { 'Accept': 'application/json', ...headers };
  const token = getToken();
  if (token) h['Authorization'] = `Bearer ${token}`;
  const res = await fetch(path, { method, headers: h, body });
  if (!res.ok) {
    let errText = await res.text().catch(() => '');
    try { const j = JSON.parse(errText); errText = j.error || errText; } catch {}
    throw new Error(errText || `${res.status} ${res.statusText}`);
  }
  const ct = res.headers.get('Content-Type') || '';
  if (ct.includes('application/json')) return res.json();
  return res.text();
}

export async function register(email, password, role = 'free') {
  return api('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, role }),
  });
}

export async function login(email, password) {
  return api('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
}

export async function me() {
  return api('/api/me');
}

export async function listConversations() {
  return api('/api/conversations');
}

export async function getMessages(convId) {
  return api(`/api/conversations/${convId}/messages`);
}

export async function sendChat({ conversation_id = 0, model = 'mock-mini', message = '' }) {
  return api('/api/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ conversation_id, model, message, stream: false }),
  });
}

// Streaming via fetch ReadableStream. Parses text/event-stream (SSE-like) lines.
export async function chatStream({ conversation_id = 0, model = 'mock-mini', message = '' }, { onChunk, onDone } = {}) {
  const res = await fetch('/api/chat?stream=1', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${getToken()}`,
    },
    body: JSON.stringify({ conversation_id, model, message, stream: true }),
  });
  if (!res.ok || !res.body) {
    let t = await res.text().catch(() => '');
    try { const j = JSON.parse(t); t = j.error || t; } catch {}
    throw new Error(t || `${res.status} ${res.statusText}`);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let convId = conversation_id;

  const flushEvents = () => {
    let idx;
    while ((idx = buffer.indexOf('\n\n')) >= 0) {
      const evt = buffer.slice(0, idx);
      buffer = buffer.slice(idx + 2);
      const lines = evt.split('\n');
      let eventName = 'message';
      const dataLines = [];
      for (const line of lines) {
        if (line.startsWith('event:')) eventName = line.slice(6).trim();
        else if (line.startsWith('data:')) dataLines.push(line.slice(5).replace(/^\s*/, ''));
      }
      const data = dataLines.join('\n');
      if (eventName === 'done') {
        try {
          const obj = JSON.parse(data);
          if (obj.conversation_id) convId = obj.conversation_id;
        } catch {}
        if (onDone) onDone({ conversation_id: convId });
      } else {
        if (onChunk) onChunk(data);
      }
    }
  };

  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    flushEvents();
  }
  // flush any remainder
  flushEvents();
  return { conversation_id: convId };
}

export const AllowedModelsByRole = {
  free: ['mock-mini', 'gpt-4o-mini'],
  pro: ['mock-mini', 'mock-pro', 'gpt-4o-mini', 'gpt-4o'],
  admin: ['mock-mini', 'mock-pro', 'mock-admin', 'gpt-4o-mini', 'gpt-4o', 'gpt-4.1'],
};

export function roleAllowsModel(role, model) {
  if (!model) return true;
  const list = AllowedModelsByRole[role] || [];
  return list.includes(model);
}
