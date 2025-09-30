import random
import string
import secrets
import hashlib
import hmac

import websocket
import json
import time
import threading
import sys


class WebSocketClient:
    def __init__(self, token=None, secret_key="your_shared_secret_key"):
        self.ws_url = "ws://localhost:8080/ws/ai"
        self.token = token
        self.secret_key = secret_key
        self.ws = None
        self.is_authenticated = False

    def _generate_nonce(self):
        """生成加密安全的随机nonce"""
        # 使用secrets模块生成32位随机字符串，确保加密安全性
        return secrets.token_hex(16)  # 生成32位十六进制随机字符串

    def _build_sign_string(self, method, path, timestamp, nonce, query_params=None, body=None):
        """构建签名字符串，与服务端保持一致"""
        # 1. 添加HTTP方法
        sign_parts = [method.upper()]

        # 2. 添加路径
        sign_parts.append(path)

        # 3. 添加时间戳
        sign_parts.append(str(timestamp))

        # 4. 添加nonce
        sign_parts.append(nonce)

        # 5. 添加查询参数（按键排序）
        if query_params:
            query_parts = []
            for key in sorted(query_params.keys()):
                value = query_params[key]
                # 处理不同类型的参数值
                if isinstance(value, list):
                    for v in sorted(value):
                        query_parts.append(f"{key}={v}")
                else:
                    query_parts.append(f"{key}={value}")
            if query_parts:
                sign_parts.append("&".join(query_parts))

        # 6. 添加请求体（如果有）
        if body:
            sign_parts.append(body)

        # 使用&连接所有部分
        return "&".join(sign_parts)

    def _get_headers(self):
        """获取包含认证、时间戳、nonce和签名的请求头"""
        headers = {
            "Content-Type": "application/json"
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
            timestamp = str(int(time.time()))
            nonce = self._generate_nonce()

            headers["X-Timestamp"] = timestamp
            headers["X-Nonce"] = nonce

            # 计算签名
            sign_string = self._build_sign_string("GET", "/ws/ai", timestamp, nonce)
            # 创建HMAC签名
            signature = hmac.new(
                self.secret_key.encode('utf-8'),
                sign_string.encode('utf-8'),
                hashlib.sha256
            ).hexdigest()
            headers["X-Signature"] = signature

        return headers

    def on_message(self, ws, message):
        """处理接收到的WebSocket消息"""
        try:
            # 解析JSON响应
            data = json.loads(message)

            # 处理认证响应
            if 'success' in data:
                if data['success']:
                    print(f"\n认证成功: {data.get('message', '')}")
                    self.is_authenticated = True
                    threading.Thread(target=self.handle_user_input, args=(ws,), daemon=True).start()
                else:
                    print(f"\n认证失败: {data.get('error', '')}")
                    ws.close()
                    return

            # 打印AI的响应内容
            if data.get('text'):
                print(data['text'], end='', flush=True)
            # 当收到完成标记时
            if data.get('done'):
                if data.get('text'):  # 如果有文本内容，打印换行
                    print()
                elif data.get('error'):  # 如果有错误，打印错误信息
                    print(f"\n错误: {data['error']}")
                print("\n--- 响应完成 ---")
                # 重新显示命令提示符
                print("\n请输入命令 (q:退出, c:清空上下文): ", end='', flush=True)
        except json.JSONDecodeError:
            print(f"\n接收到无效的JSON: {message}")

    def on_error(self, ws, error):
        """处理WebSocket错误"""
        print(f"\n错误: {error}")

    def on_close(self, ws, close_status_code, close_msg):
        """处理WebSocket关闭"""
        print("\n连接已关闭")

    def on_open(self, ws):
        """处理WebSocket连接建立"""
        print("已连接到WebSocket服务器")
        print("等待认证结果...")
        # 认证成功后在新线程中处理用户输入
        # 认证结果会在on_message中处理

    def handle_user_input(self, ws):
        """处理用户输入"""
        while True:
            print("\n请输入命令 (q:退出, c:清空上下文): ", end='', flush=True)
            user_input = input().strip()

            if user_input.lower() == 'q':
                ws.close()
                sys.exit(0)
            elif user_input.lower() == 'c':
                # 发送清空上下文请求
                clear_context_message = json.dumps({"clearContext": True})
                ws.send(clear_context_message)
                print("正在清空上下文...")
            elif user_input:
                # 发送带上下文的请求
                request_message = json.dumps({
                    "prompt": user_input,
                    "model": "qwen:7b",  # 可根据需要更改模型
                    "useContext": True,
                    "conversationId": 0
                })
                print(f"\n发送请求: {user_input}")
                print("AI响应:", end='', flush=True)
                ws.send(request_message)
        # while True:
        #     request_message = json.dumps({
        #                 "prompt": "你好",
        #                 "model": "qwen:7b",  # 可根据需要更改模型
        #                 "useContext": True,
        #                 "conversationId": 0
        #             })
        #     ws.send(request_message)
        #     time.sleep(1)

    def connect(self):
        """连接到WebSocket服务器"""
        headers = self._get_headers()

        # 创建WebSocket连接
        self.ws = websocket.WebSocketApp(
            self.ws_url,
            header=headers,
            on_message=self.on_message,
            on_error=self.on_error,
            on_close=self.on_close
        )
        self.ws.on_open = self.on_open

        # 运行WebSocket客户端
        try:
            print(f"正在连接到 {self.ws_url}...")
            self.ws.run_forever()
        except KeyboardInterrupt:
            print("\n程序被用户中断")
            if self.ws:
                self.ws.close()


def main():
    """主函数"""
    # 从命令行参数获取token（如果有）
    token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo4LCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJ1c2VybmFtZSI6InRlc3RfdXNlciIsImV4cCI6MTc1OTI4MTgxMSwiaWF0IjoxNzU5MTk1NDExfQ.210XrDiZcc1kqzbhX0V5XIQGQKjKC4pnlVpHTXKeusg"
    secret_key = "your_shared_secret_key"

    client = WebSocketClient(token, secret_key)
    client.connect()


if __name__ == "__main__":
    print("注意: 运行此脚本前请确保已安装websocket-client库")
    print("安装命令: pip install websocket-client")
    print("\n使用方法:")
    print("1. 交互式运行: python websocket_client.py")
    print("2. 带token运行: python websocket_client.py YOUR_JWT_TOKEN")
    print()
    main()
