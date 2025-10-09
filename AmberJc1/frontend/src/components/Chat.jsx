import React, { useState } from 'react'

function Chat({ messages, onSend }) {
  const [input, setInput] = useState('')

  const handleSend = () => {
    onSend(input)
    setInput('')
  }

  return (
    <div className="chat-container">
      <div className="chat-box">
        {messages.map((msg, i) => (
          <div key={i} className={`msg ${msg.role}`}>
            <strong>{msg.role === 'user' ? '你' : 'AI'}：</strong> {msg.content}
          </div>
        ))}
      </div>

      <div className="input-box">
        <input
          type="text"
          value={input}
          placeholder="输入你的消息..."
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
        />
        <button onClick={handleSend}>发送</button>
      </div>
    </div>
  )
}

export default Chat
