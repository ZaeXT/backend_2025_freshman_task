import random
import string
import secrets
import hashlib
import hmac

import requests
import json
import time
import sys


class HTTPClient:
    def __init__(self, base_url="http://localhost:8080", secret_key="your_shared_secret_key"):
        self.base_url = base_url
        # 硬编码的token（请替换为实际的JWT token）
        self.token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo4LCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJ1c2VybmFtZSI6InRlc3RfdXNlciIsImV4cCI6MTc1OTM5MzYwOCwiaWF0IjoxNzU5MzA3MjA4fQ.K8aLw999JPr6FL6HBhkhHpMLPpPqHwaGXgu-0kYis-U"
        self.secret_key = secret_key

        self.session = requests.Session()

    def _generate_nonce(self):
        """生成加密安全的随机nonce"""
        # 使用secrets模块生成32位随机字符串，确保加密安全性
        return secrets.token_hex(16)  # 生成32位十六进制随机字符串

    def get_auth_headers(self, method, path, body=None):
        """生成登录和注册接口所需的请求头（包含时间戳、nonce和签名）"""
        headers = {
            "Content-Type": "application/json"
        }

        # 添加时间戳和nonce
        timestamp = str(int(time.time()))
        nonce = self._generate_nonce()

        headers["X-Timestamp"] = timestamp
        headers["X-Nonce"] = nonce

        # 构建签名字符串
        sign_string = self._build_sign_string(method, path, timestamp, nonce, None, body)

        # 创建HMAC签名
        signature = hmac.new(
            self.secret_key.encode('utf-8'),
            sign_string.encode('utf-8'),
            hashlib.sha256
        ).hexdigest()

        headers["X-Signature"] = signature
        return headers

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

    def _get_headers(self, method=None, path=None, query_params=None, body=None):
        """获取包含认证、时间戳、nonce和签名的请求头"""
        headers = {
            "Content-Type": "application/json"
        }
        if self.token:
            headers["Authorization"] = f"Bearer {self.token}"
            timestamp = str(int(time.time()))  # 生成当前时间戳
            nonce = self._generate_nonce()  # 生成随机nonce

            headers["X-Timestamp"] = timestamp
            headers["X-Nonce"] = nonce

            # 如果提供了请求信息，则计算签名
            if method and path is not None:
                # 1. 构建签名字符串
                sign_string = self._build_sign_string(method, path, timestamp, nonce, query_params, body)

                # 2. 创建HMAC签名
                signature = hmac.new(
                    self.secret_key.encode('utf-8'),  # 共享密钥
                    sign_string.encode('utf-8'),  # 签名字符串
                    hashlib.sha256  # 使用SHA256算法
                ).hexdigest()  # 转换为十六进制字符串

                headers["X-Signature"] = signature  # 添加到请求头

        return headers

    def login(self, email, password):
        """用户登录"""
        url = f"{self.base_url}/api/auth/login"
        payload = {
            "email": email,
            "password": password
        }
        body = json.dumps(payload, separators=(',', ':'))

        try:
            response = self.session.post(url, data=body, headers=self.get_auth_headers("POST", "/api/auth/login", body))
            if response.status_code == 200:
                data = response.json()
                self.token = data.get("token")
                user_info = data.get("user", {})
                print(f"登录成功！用户: {user_info.get('username')}, 剩余token: {user_info.get('tokenCount')}")
                return True
            else:
                print(f"登录失败: {response.json().get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"登录请求失败: {e}")
            return False

    def register(self, username, email, password):
        """用户注册"""
        url = f"{self.base_url}/api/auth/register"
        payload = {
            "username": username,
            "email": email,
            "password": password
        }
        body = json.dumps(payload, separators=(',', ':'))

        try:
            response = self.session.post(url, data=body,
                                         headers=self.get_auth_headers("POST", "/api/auth/register", body))
            if response.status_code == 201:
                data = response.json()
                print(f"注册成功！用户: {data.get('username')}")
                return True
            else:
                print(f"注册失败: {response.json().get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"注册请求失败: {e}")
            return False

    def recharge(self, token_amount):
        """充值token"""
        if not self.token:
            print("请先登录")
            return False

        url = f"{self.base_url}/api/recharge"
        payload = {
            "tokenAmount": token_amount
        }
        body = json.dumps(payload, separators=(',', ':'))

        try:
            response = self.session.post(url, data=body, headers=self._get_headers("POST", "/api/recharge", None, body))
            if response.status_code == 200:
                data = response.json()
                print(f"充值成功！当前token: {data.get('tokenCount')}, 本次增加: {data.get('added')}")
                return True
            else:
                print(f"充值失败: {response.json().get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"充值请求失败: {e}")
            return False

    def get_token_info(self):
        """获取token信息"""
        if not self.token:
            print("请先登录")
            return False

        url = f"{self.base_url}/api/token-info"

        try:
            response = self.session.get(url, headers=self._get_headers("GET", "/api/token-info"))
            if response.status_code == 200:
                data = response.json()
                print(f"Token信息: {data}")
                return True
            else:
                print(f"获取token信息失败: {response.json().get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"获取token信息失败: {e}")
            return False

    def delete_user(self):
        """删除用户及其所有数据"""
        if not self.token:
            print("请先登录")
            return False

        url = f"{self.base_url}/api/user"

        try:
            response = self.session.delete(url, headers=self._get_headers("DELETE", "/api/user"))
            if response.status_code == 200:
                data = response.json()
                print(f"用户数据删除成功！{data.get('message', '')}")
                # 清除本地token
                self.token = None
                return True
            else:
                error_msg = response.json().get('error', '未知错误')
                print(f"删除用户数据失败: {error_msg}")
                return False
        except Exception as e:
            print(f"删除用户数据请求失败: {e}")
            return False

    def get_user_profile(self):
        """获取用户信息"""

        url = f"{self.base_url}/api/user"

        try:
            response = self.session.get(url, headers=self._get_headers("GET", "/api/user"))
            if response.status_code == 200:
                data = response.json()
                print(f"用户信息: {data}")
                return True
            else:
                print(f"获取用户信息失败: {response.json().get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"获取用户信息失败: {e}")
            return False

    def get_available_models(self):
        """获取可用的AI模型列表"""
        if not self.token:
            print("请先登录")
            return False

        url = f"{self.base_url}/api/models"

        try:
            response = self.session.get(url, headers=self._get_headers("GET", "/api/models"))
            if response.status_code == 200:
                data = response.json()
                models = data.get("models", [])
                print(f"可用的AI模型 ({data.get('count', 0)} 个):")
                for i, model in enumerate(models, 1):
                    print(f"  {i}. {model}")
                return models
            else:
                print(f"获取模型列表失败: {response.json().get('error', '未知错误')}")
                return []
        except Exception as e:
            print(f"获取模型列表失败: {e}")
            return []

    def get_conversations(self, page=1, page_size=10):
        """获取对话记录列表"""
        if not self.token:
            print("请先登录")
            return False

        url = f"{self.base_url}/api/conversations"
        params = {
            "page": page,
            "pageSize": page_size
        }

        try:
            response = self.session.get(url, headers=self._get_headers("GET", "/api/conversations", params),
                                        params=params)
            if response.status_code == 200:
                data = response.json()
                conversations = data.get("conversations", [])
                print(f"对话记录列表 (第{page}页，共{data.get('totalPages', 0)}页):")
                for conv in conversations:
                    print(f"  ID: {conv.get('id')}, 标题: {conv.get('title')}, "
                          f"消息数: {conv.get('messageCount')}, Token: {conv.get('tokenUsed')}, "
                          f"创建时间: {conv.get('createdAt')}")
                return conversations
            else:
                print(f"获取对话记录列表失败: {response.json().get('error', '未知错误')}")
                return []
        except Exception as e:
            print(f"获取对话记录列表失败: {e}")
            return []

    def get_conversation(self, conversation_id):
        """获取对话详情"""
        if not self.token:
            print("请先登录")
            return False

        path = f"/api/conversations/{conversation_id}"
        url = f"{self.base_url}{path}"

        try:
            response = self.session.get(url, headers=self._get_headers("GET", path))
            if response.status_code == 200:
                data = response.json()
                print(f"对话详情:")
                print(f"  ID: {data.get('id')}")
                print(f"  标题: {data.get('title')}")
                print(f"  消息数: {data.get('messageCount')}")
                print(f"  Token: {data.get('tokenUsed')}")
                print(f"  创建时间: {data.get('createdAt')}")
                print(f"  消息内容:")
                messages = data.get("messages", [])
                for msg in messages:
                    print(f"    [{msg.get('role')}] {msg.get('content')[:100]}...")
                return data
            else:
                print(f"获取对话详情失败: {response.json().get('error', '未知错误')}")
                return None
        except Exception as e:
            print(f"获取对话详情失败: {e}")
            return None

    def delete_conversation(self, conversation_id):
        """删除对话记录"""
        if not self.token:
            print("请先登录")
            return False

        path = f"/api/conversations/{conversation_id}"
        url = f"{self.base_url}{path}"

        try:
            response = self.session.delete(url, headers=self._get_headers("DELETE", path))
            if response.status_code == 200:
                data = response.json()
                print(f"删除对话记录成功: {data.get('message', '删除成功')}")
                return True
            else:
                print(f"删除对话记录失败: {response.json().get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"删除对话记录失败: {e}")
            return False

    def update_conversation_title(self, conversation_id, new_title):
        """更新对话标题"""
        if not self.token:
            print("请先登录")
            return False

        path = f"/api/conversations/{conversation_id}/title"
        url = f"{self.base_url}{path}"
        payload = {
            "title": new_title
        }
        body = json.dumps(payload, separators=(',', ':'))

        try:
            response = self.session.put(url, data=body, headers=self._get_headers("PUT", path, None, body))
            if response.status_code == 200:
                data = response.json()
                print(f"更新对话标题成功: {data.get('message', '更新成功')}")
                print(f"新标题: {data.get('title', new_title)}")
                return True
            else:
                print(f"更新对话标题失败: {response.json().get('error', '未知错误')}")
                return False
        except Exception as e:
            print(f"更新对话标题失败: {e}")
            return False


def main():
    """主函数"""
    # 使用硬编码的token创建客户端
    client = HTTPClient()

    # 从命令行参数获取token（如果有）
    if len(sys.argv) > 1:
        client.token = sys.argv[1]
        print(f"使用提供的token")
    else:
        print(f"使用硬编码token: {client.token}")

    while True:
        print("\n=== HTTP客户端命令菜单 ===")
        print("1. 登录（使用邮箱密码）")
        print("2. 充值token")
        print("3. 查看token信息")
        print("4. 查看用户信息")
        print("5. 查看可用模型")
        print("6. 获取对话记录列表")
        print("7. 获取对话详情")
        print("8. 删除对话记录")
        print("9. 更新对话标题")
        print("10. 注销账号")
        print("11. 退出")

        choice = input("请选择操作 (1-11): ").strip()

        if choice == "1":
            email = input("请输入邮箱: ")
            password = input("请输入密码: ")
            client.login(email, password)
        elif choice == "2":
            try:
                amount = int(input("请输入充值数量: "))
                client.recharge(amount)
            except ValueError:
                print("请输入有效的数字")
        elif choice == "3":
            client.get_token_info()
        elif choice == "4":
            client.get_user_profile()
        elif choice == "5":
            client.get_available_models()
        elif choice == "6":
            try:
                page = int(input("请输入页码 (默认1): ") or "1")
                page_size = int(input("请输入每页数量 (默认10): ") or "10")
                client.get_conversations(page, page_size)
            except ValueError:
                print("请输入有效的数字")
                client.get_conversations()
        elif choice == "7":
            conversation_id = input("请输入对话ID: ")
            if conversation_id:
                client.get_conversation(conversation_id)
            else:
                print("对话ID不能为空")
        elif choice == "8":
            conversation_id = input("请输入要删除的对话ID: ")
            if conversation_id:
                client.delete_conversation(conversation_id)
            else:
                print("对话ID不能为空")
        elif choice == "9":
            conversation_id = input("请输入对话ID: ")
            if conversation_id:
                new_title = input("请输入新标题: ")
                if new_title:
                    client.update_conversation_title(conversation_id, new_title)
                else:
                    print("标题不能为空")
            else:
                print("对话ID不能为空")
        elif choice == "10":
            client.delete_user()
        elif choice == "11":
            print("再见！")
            break
        else:
            print("无效选择，请重新输入")


def test():
    """主函数"""
    # 使用硬编码的token创建客户端
    client = HTTPClient()
    while True:
        client.get_user_profile()
        time.sleep(0.8)


if __name__ == "__main__":
    print("注意: 运行此脚本前请确保已安装requests库")
    print("安装命令: pip install requests")
    print("\n使用方法:")
    print("1. 使用硬编码token运行: python http_client.py")
    print("2. 使用自定义token运行: python http_client.py YOUR_JWT_TOKEN")
    print()
    main()
