import requests
import json
import random
import string

# --- 配置 ---
BASE_URL = "http://localhost:8080/api/v1"
# 定义预期的HTTP错误代码
# 更多信息: https://developer.mozilla.org/en-US/docs/Web/HTTP/Status
STATUS_BAD_REQUEST = 400
STATUS_UNAUTHORIZED = 401
STATUS_FORBIDDEN = 403
STATUS_NOT_FOUND = 404

# --- 辅助函数 ---
def random_string(length=8):
    """生成一个随机字符串，用于创建唯一的用户名"""
    return ''.join(random.choices(string.ascii_lowercase + string.digits, k=length))

def print_test_header(title):
    print("\n" + "="*50)
    print(f"  🧪  开始测试: {title}")
    print("="*50)

def print_success(message):
    print(f"  ✅  \033[92m成功:\033[0m {message}")

def print_fail(message, response=None):
    print(f"  ❌  \033[91m失败:\033[0m {message}")
    if response is not None:
        try:
            print(f"      响应状态码: {response.status_code}, 内容: {response.json()}")
        except json.JSONDecodeError:
            print(f"      响应状态码: {response.status_code}, 内容: {response.text}")
    exit(1)

def assert_status_code(response, expected_code, test_name):
    """断言响应的状态码是否符合预期"""
    if response.status_code == expected_code:
        print_success(f"{test_name} - 服务器返回了预期的状态码 {expected_code}")
    else:
        print_fail(f"{test_name} - 预期状态码 {expected_code}，但收到了 {response.status_code}", response)

# --- 测试流程 ---

def test_registration_failures():
    """测试用户注册时的失败场景"""
    print_test_header("注册失败路径")
    
    # 1. 注册一个已存在的用户
    username = f"failure_user_{random_string()}"
    password = "password123"
    payload = {"username": username, "password": password}
    
    print("  -> 第一次注册 (应该成功)")
    res_first = requests.post(f"{BASE_URL}/register", json=payload)
    if res_first.status_code != 200:
        print_fail("为测试重复注册做准备时，首次注册失败", res_first)
    
    print("  -> 尝试用相同的用户名再次注册")
    res_second = requests.post(f"{BASE_URL}/register", json=payload)
    # 预期服务器会返回一个通用错误 (500) 或一个更具体的冲突错误 (409)
    # 根据我们的 handler 实现，返回的是 500 + 错误信息
    if res_second.status_code == 500 and "already exists" in res_second.json().get("msg", ""):
        print_success("重复注册 - 服务器按预期拒绝了请求")
    else:
        print_fail("重复注册 - 服务器未按预期响应", res_second)

    # 2. 注册时缺少字段
    print("  -> 尝试注册时缺少密码字段")
    res_missing = requests.post(f"{BASE_URL}/register", json={"username": username})
    # Gin的binding:"required"会返回400 Bad Request
    assert_status_code(res_missing, STATUS_BAD_REQUEST, "缺少字段")

def test_login_and_auth_failures():
    """测试登录失败和认证中间件的保护"""
    print_test_header("登录与认证失败路径")

    username = f"failure_user_{random_string()}"
    password = "password123"
    requests.post(f"{BASE_URL}/register", json={"username": username, "password": password})

    # 1. 使用错误密码登录
    print("  -> 尝试使用错误的密码登录")
    res_wrong_pass = requests.post(f"{BASE_URL}/login", json={"username": username, "password": "wrongpassword"})
    assert_status_code(res_wrong_pass, STATUS_UNAUTHORIZED, "错误密码登录")

    # 2. 登录一个不存在的用户
    print("  -> 尝试登录一个不存在的用户")
    res_no_user = requests.post(f"{BASE_URL}/login", json={"username": "nonexistentuser", "password": "password"})
    assert_status_code(res_no_user, STATUS_UNAUTHORIZED, "不存在的用户登录")
    
    # 3. 不带Token访问受保护的路由
    print("  -> 尝试不带Token访问 /profile")
    res_no_token = requests.get(f"{BASE_URL}/profile")
    assert_status_code(res_no_token, STATUS_UNAUTHORIZED, "不带Token访问")
    
    # 4. 使用无效或伪造的Token访问
    print("  -> 尝试使用无效Token访问 /profile")
    invalid_headers = {"Authorization": "Bearer thisisafaketoken"}
    res_invalid_token = requests.get(f"{BASE_URL}/profile", headers=invalid_headers)
    assert_status_code(res_invalid_token, STATUS_UNAUTHORIZED, "无效Token访问")

def test_permission_isolation():
    """测试不同用户之间的权限隔离"""
    print_test_header("权限隔离 (安全测试)")

    # 1. 创建两个用户：Alice 和 Bob
    alice_username = f"alice_{random_string()}"
    bob_username = f"bob_{random_string()}"
    password = "password123"
    
    # --- 新增：为关键的准备步骤添加断言 ---
    res_alice_reg = requests.post(f"{BASE_URL}/register", json={"username": alice_username, "password": password})
    if res_alice_reg.status_code != 200:
        print_fail(f"为权限测试创建用户Alice失败", res_alice_reg)
        
    res_bob_reg = requests.post(f"{BASE_URL}/register", json={"username": bob_username, "password": password})
    if res_bob_reg.status_code != 200:
        print_fail(f"为权限测试创建用户Bob失败", res_bob_reg)


    # 2. Alice 和 Bob 分别登录，获取各自的Token
    res_alice_login = requests.post(f"{BASE_URL}/login", json={"username": alice_username, "password": password})
    # --- 核心修改：增加健壮的检查 ---
    if res_alice_login.status_code != 200 or "data" not in res_alice_login.json():
        print_fail("Alice 登录失败，无法继续进行权限测试", res_alice_login)
    alice_token = res_alice_login.json()["data"]["token"]
    alice_headers = {"Authorization": f"Bearer {alice_token}", "Content-Type": "application/json"}
    
    res_bob_login = requests.post(f"{BASE_URL}/login", json={"username": bob_username, "password": password})
    if res_bob_login.status_code != 200 or "data" not in res_bob_login.json():
        print_fail("Bob 登录失败，无法继续进行权限测试", res_bob_login)
    bob_token = res_bob_login.json()["data"]["token"]
    bob_headers = {"Authorization": f"Bearer {bob_token}", "Content-Type": "application/json"}
    
    print("  -> 已成功创建并登录用户 Alice 和 Bob")

    # 3. Alice 创建一个对话
    res_alice_conv = requests.post(f"{BASE_URL}/conversations", headers=alice_headers, json={})
    if res_alice_conv.status_code != 200 or "data" not in res_alice_conv.json():
        print_fail("Alice 创建对话失败，无法继续进行权限测试", res_alice_conv)
    alice_conv_id = res_alice_conv.json()["data"]["id"]
    print(f"  -> Alice 创建了对话 (ID: {alice_conv_id})")

    # --- 后续的核心测试逻辑保持不变 ---
    # 4. 【核心测试】Bob 尝试删除 Alice 的对话
    print(f"  -> 尝试用 Bob 的身份删除 Alice 的对话 (ID: {alice_conv_id})")
    res_bob_delete = requests.delete(f"{BASE_URL}/conversations/{alice_conv_id}", headers=bob_headers)
    if res_bob_delete.status_code == 500 and "permission denied" in res_bob_delete.json().get("msg", ""):
        print_success("权限隔离 - Bob 删除 Alice 对话被按预期拒绝")
    else:
        print_fail("权限隔离 - Bob 删除 Alice 对话的请求未被正确拒绝", res_bob_delete)
        
    # 5. 【核心测试】Bob 尝试向 Alice 的对话中发送消息
    print(f"  -> 尝试用 Bob 的身份向 Alice 的对话发送消息 (ID: {alice_conv_id})")
    chat_payload = {"message": "这是一个恶意消息"}
    res_bob_chat = requests.post(f"{BASE_URL}/conversations/{alice_conv_id}/messages", headers=bob_headers, json=chat_payload)
    if res_bob_chat.status_code != 200:
         print_success("权限隔离 - Bob 向 Alice 对话发消息被按预期拒绝")
    else:
         print_fail("权限隔离 - Bob 向 Alice 对话发消息的请求未被正确拒绝", res_bob_chat)

    # 6. 验证 Alice 的对话仍然存在
    print("  -> 验证 Alice 的对话仍然存在")
    res_verify = requests.get(f"{BASE_URL}/conversations", headers=alice_headers)
    found = any(conv['id'] == alice_conv_id for conv in res_verify.json().get("data", []))
    if found:
        print_success("权限隔离验证 - Alice 的对话安然无恙")
    else:
        print_fail("权限隔离验证 - Alice 的对话被错误地删除了")

if __name__ == "__main__":
    print("🚀  开始对AI对话系统后端进行【失败路径与安全】自动化测试...")
    try:
        test_registration_failures()
        test_login_and_auth_failures()
        test_permission_isolation()
        print("\n" + "="*50)
        print("  🎉  \033[92m所有失败路径与安全测试均已成功通过！\033[0m")
        print("="*50)
    except Exception as e:
        print(f"\n❌  测试过程中出现意外错误: {e}")