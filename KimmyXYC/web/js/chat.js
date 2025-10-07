import { chatStream, sendChat, listConversations, getMessages, roleAllowsModel } from './api.js';
import { getUser } from './state.js';

export function initChatUI() {
  const userInfoEmail = document.getElementById('user-email');
  const userInfoRole = document.getElementById('user-role');
  const modelSelect = document.getElementById('model-select');
  const streamToggle = document.getElementById('stream-toggle');
  const convList = document.getElementById('conv-list');
  const newChatBtn = document.getElementById('new-chat-btn');
  const messagesEl = document.getElementById('messages');
  const chatForm = document.getElementById('chat-form');
  const chatInput = document.getElementById('chat-input');
  const sendHint = document.getElementById('send-hint');

  const user = getUser();
  userInfoEmail.textContent = user?.email || '';
  userInfoRole.textContent = user?.role || 'free';

  let currentConv = 0;
  let sending = false;

  const updateModelHint = () => {
    const model = modelSelect.value;
    const allowed = roleAllowsModel(user?.role || 'free', model);
    if (!allowed) {
      sendHint.textContent = `当前角色无权使用 ${model}，尝试发送将被后端拒绝`;
    } else {
      sendHint.textContent = '';
    }
  };
  modelSelect.addEventListener('change', updateModelHint);
  updateModelHint();

  const scrollToBottom = () => {
    messagesEl.scrollTop = messagesEl.scrollHeight;
  };

  const fmtTime = (iso) => {
    try { return new Date(iso).toLocaleString(); } catch { return ''; }
  };

  const renderMessage = (m) => {
    const div = document.createElement('div');
    div.className = `message ${m.role}`;
    const meta = document.createElement('div');
    meta.className = 'meta';
    meta.textContent = `${m.role}`;
    const content = document.createElement('div');
    content.className = 'content';
    content.textContent = m.content || '';
    div.appendChild(meta);
    div.appendChild(content);
    messagesEl.appendChild(div);
    scrollToBottom();
    return content; // return content node for streaming update
  };

  const clearMessages = () => { messagesEl.innerHTML = ''; };

  async function loadConversations() {
    convList.innerHTML = '';
    try {
      const resp = await listConversations();
      const convs = resp.conversations || [];
      for (const c of convs) {
        const li = document.createElement('li');
        li.dataset.id = c.id;
        li.className = (c.id === currentConv) ? 'active' : '';
        const title = c.title || `对话 #${c.id}`;
        li.innerHTML = `<div>${title}</div><small>${c.model || ''}</small>`;
        li.addEventListener('click', async () => {
          currentConv = c.id;
          document.querySelectorAll('#conv-list li').forEach(x => x.classList.remove('active'));
          li.classList.add('active');
          await loadMessages(c.id);
        });
        convList.appendChild(li);
      }
    } catch (err) {
      console.error('加载会话失败', err);
    }
  }

  async function loadMessages(convId) {
    clearMessages();
    if (!convId) return;
    try {
      const resp = await getMessages(convId);
      const msgs = resp.messages || [];
      for (const m of msgs) renderMessage(m);
    } catch (err) {
      console.error('加载消息失败', err);
    }
  }

  newChatBtn.addEventListener('click', () => {
    currentConv = 0; // backend will create on first send
    clearMessages();
    chatInput.focus();
  });

  chatForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    if (sending) return;
    const text = chatInput.value.trim();
    if (!text) return;
    sending = true;
    try {
      const model = modelSelect.value || 'mock-mini';
      // render user message immediately
      renderMessage({ role: 'user', content: text });
      chatInput.value = '';

      const doStream = streamToggle.checked;
      if (!doStream) {
        const r = await sendChat({ conversation_id: currentConv, model, message: text });
        currentConv = r.conversation_id || currentConv;
        renderMessage({ role: 'assistant', content: r.reply || '' });
      } else {
        let assistantNode = renderMessage({ role: 'assistant', content: '' });
        await chatStream(
          { conversation_id: currentConv, model, message: text },
          {
            onChunk: (chunk) => { assistantNode.textContent += chunk; scrollToBottom(); },
            onDone: ({ conversation_id }) => { if (conversation_id) currentConv = conversation_id; },
          }
        );
      }
      await loadConversations();
    } catch (err) {
      renderMessage({ role: 'assistant', content: `错误：${err.message || err}` });
    } finally {
      sending = false;
    }
  });

  // Load initial
  loadConversations();
}
