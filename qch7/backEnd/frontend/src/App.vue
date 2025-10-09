<template>
  <div class="container">
    <h2>AI 问答 Demo</h2>
    <section class="card">
      <h3>登录</h3>
      <input v-model.trim="username" placeholder="用户名(注册用)" minlength="2" maxlength="32" />
      <input v-model.trim="email" placeholder="邮箱" type="email" />
      <input v-model="password" placeholder="密码 (≥6 位)" type="password" minlength="6" />
      <button @click="register">注册</button>
      <button @click="login">登录</button>
      <span v-if="token">已登录</span>
    </section>
    <section class="card">
      <h3>聊天</h3>
      <label><input type="checkbox" v-model="stream" />流式</label>
      <textarea v-model="content" placeholder="输入你的问题"></textarea>
      <button @click="ask">发送</button>
      <div class="chat">
        <div v-for="(m,i) in messages" :key="i" :class="m.role">
          <strong>{{ m.role }}:</strong> {{ m.content }}
        </div>
      </div>
    </section>
  </div>
  
</template>

<script setup lang="ts">
import { ref } from 'vue'

const username = ref('')
const email = ref('')
const password = ref('')
const token = ref<string>('')
const content = ref('')
const stream = ref(true)
const conversationId = ref<string>('')
const messages = ref<{role:'user'|'assistant'|'system', content:string}[]>([])

async function login() {
  const res = await fetch('/api/v1/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: email.value, password: password.value })
  })
  const data = await res.json()
  token.value = data.token || ''
}

async function register() {
  const uname = username.value.trim()
  const mail = email.value.trim()
  const pwd = password.value
  if (uname.length < 2) { alert('用户名至少 2 个字符'); return }
  if (!/^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(mail)) { alert('邮箱格式不正确'); return }
  if (pwd.length < 6) { alert('密码需至少 6 位'); return }
  const res = await fetch('/api/v1/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email: mail, username: uname, password: pwd })
  })
  const data = await res.json()
  if (res.ok) {
    alert('注册成功，请登录')
  } else {
    alert('注册失败：' + (data.error || res.statusText))
  }
}

async function ask() {
  if (!token.value) { alert('请先登录'); return }
  const userMsg = { role: 'user' as const, content: content.value }
  messages.value.push(userMsg)
  content.value = ''
  if (!stream.value) {
    const res = await fetch('/api/v1/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token.value}`
      },
      body: JSON.stringify({
        conversationId: conversationId.value || undefined,
        messages: messages.value.slice(-10),
        stream: false
      })
    })
    const data = await res.json()
    conversationId.value = data.conversationId
    messages.value.push({ role: 'assistant', content: data.content || '' })
    return
  }
  // SSE
  const res = await fetch('/api/v1/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token.value}`
    },
    body: JSON.stringify({
      conversationId: conversationId.value || undefined,
      messages: messages.value.slice(-10),
      stream: true
    })
  })
  const reader = res.body?.getReader()
  if (!reader) return
  let assistant = { role: 'assistant' as const, content: '' }
  messages.value.push(assistant)
  const decoder = new TextDecoder()
  let buf = ''
  while (true) {
    const {done, value} = await reader.read()
    if (done) break
    buf += decoder.decode(value, { stream: true })
    const lines = buf.split('\n')
    // keep last partial line in buffer
    buf = lines.pop() || ''
    for (const line of lines) {
      const l = line.trim()
      if (!l) continue
      if (l.startsWith('data:')) {
        let payload = l.slice(5).trim()
        if (payload.startsWith(':')) payload = payload.slice(1).trim()
        if (payload === '[DONE]') { buf=''; break }
        // gin.SSEvent 会把字符串做 JSON 编码，尝试解析
        try {
          const parsed = JSON.parse(payload)
          if (typeof parsed === 'string') {
            assistant.content += parsed
          } else if (parsed && typeof parsed.content === 'string') {
            assistant.content += parsed.content
          } else {
            assistant.content += String(parsed)
          }
        } catch {
          assistant.content += payload
        }
      }
    }
  }
}
</script>

<style scoped>
.container { max-width: 860px; margin: 24px auto; padding: 0 16px; font-family: system-ui, -apple-system, Segoe UI, Roboto, Ubuntu, Cantarell, Noto Sans, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji"; }
.card { padding: 12px; border: 1px solid #e5e5e5; border-radius: 8px; margin-bottom: 16px; }
.chat { background:#fafafa; border:1px solid #eee; padding:8px; min-height:160px; white-space:pre-wrap; }
.user { color:#2c7be5; }
.assistant { color:#2f9e44; }
textarea { width:100%; min-height:80px; }
input { margin-right:8px; }
button { margin-left:8px; }
</style>


