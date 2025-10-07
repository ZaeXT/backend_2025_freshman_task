API Documentation

Base URL: http://localhost:8080

Auth
- POST /api/auth/register
  Request JSON: { "email": "user@example.com", "password": "pass123", "role": "free|pro|admin" }
  Response: { "user": {id, email, role, created_at}, "token": "JWT" }

- POST /api/auth/login
  Request JSON: { "email": "user@example.com", "password": "pass123" }
  Response: { "user": {...}, "token": "JWT" }

Use Authorization: Bearer <token> for all protected endpoints below.

User
- GET /api/me
  Response: { "user_id": 1, "user_email": "user@example.com", "user_role": "free" }

Conversations
- GET /api/conversations
  Response: { "conversations": [ {id, title, model, created_at, updated_at}, ... ] }

- GET /api/conversations/:id/messages
  Response: { "messages": [ {id, role, content, created_at}, ... ] }

Chat
- POST /api/chat
  Request JSON: { "conversation_id": 0, "model": "mock-mini", "message": "hello", "stream": false }
  Response (non-stream): { "conversation_id": 1, "reply": "..." }

  Streaming: set stream=true or ?stream=1 and use text/event-stream (SSE).
  Example curl:
    curl -N -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
      -d '{"model":"mock-mini","message":"你好","stream":true}' \
      http://localhost:8080/api/chat

Models and Permissions
- Roles and allowed models:
  - free: [mock-mini]
  - pro: [mock-mini, mock-pro]
  - admin: [mock-mini, mock-pro, mock-admin]

Notes
- Default provider is Mock (no external key). If OPENAI_API_KEY is set, the backend switches to OpenAI-compatible Chat Completions API.
  - Env:
    - OPENAI_API_KEY=your-key
    - OPENAI_API_BASE=custom endpoint (optional; default https://api.openai.com)
- To integrate another provider (e.g., 火山引擎/Volcengine), implement provider.LLMProvider and update provider.NewProviderFromEnv().
