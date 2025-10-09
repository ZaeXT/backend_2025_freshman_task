import React, { useState } from "react";

function App() {
  const [input, setInput] = useState("");       // ✅ 定义输入框内容
  const [response, setResponse] = useState(""); // ✅ 定义回复内容

  // ✅ 这里替换成你自己在控制台里拿到的真实 token

  const sendMessage = async () => {
    try {
      const res = await fetch("http://localhost:8080/api/v1/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ question: input }), // ✅ 用 input 发消息
      });

      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }

      const data = await res.json();
      setResponse(data.reply || "无响应");
    } catch (error) {
      console.error("发送失败：", error);
      setResponse("出错了");
    }
  };

  return (
    <div style={{ padding: "20px" }}>
      <h2>Chat Demo</h2>
      <input
        type="text"
        value={input}
        onChange={(e) => setInput(e.target.value)} // ✅ 更新 input 状态
        placeholder="请输入内容"
        style={{ width: "300px", marginRight: "10px" }}
      />
      <button onClick={sendMessage}>发送</button>
      <p>回复：{response}</p>
    </div>
  );
}

export default App;
