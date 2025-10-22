import requests
import json
import random
import string

# --- 配置 ---
BASE_URL = "http://localhost:8080/api/v1"

# --- 辅助函数 (与之前的脚本相同) ---
def random_string(length=8):
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

def get_auth_headers(token):
    return {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}

def register_and_login(username, password):
    """注册并登录一个用户，返回其Token"""
    # 注册
    reg_res = requests.post(f"{BASE_URL}/register", json={"username": username, "password": password})
    if reg_res.status_code != 200:
        print_fail(f"为权限测试准备用户 '{username}' 时注册失败", reg_res)
    
    # 登录
    login_res = requests.post(f"{BASE_URL}/login", json={"username": username, "password": password})
    if login_res.status_code != 200 or "data" not in login_res.json():
        print_fail(f"为权限测试准备用户 '{username}' 时登录失败", login_res)
    
    return login_res.json()["data"]["token"]

# --- 测试流程 ---

def test_model_permission_inheritance():
    """测试模型权限的正确性和继承性"""
    print_test_header("模型权限继承测试")

    password = "password123"
    
    # 1. 创建一个 free 用户和一个 premium 用户
    free_user_username = f"free_user_{random_string()}"
    premium_user_username = f"premium_user_{random_string()}" # 注意前缀
    
    free_user_token = register_and_login(free_user_username, password)
    premium_user_token = register_and_login(premium_user_username, password)
    
    print(f"  -> 已成功创建并登录 'free' 用户 ({free_user_username}) 和 'premium' 用户 ({premium_user_username})")
    
    # 2. 获取两个用户的可用模型列表并进行验证
    print("  -> 验证 'free' 用户的模型列表")
    res_free_models = requests.get(f"{BASE_URL}/models", headers=get_auth_headers(free_user_token))
    if res_free_models.status_code != 200:
        print_fail("'free' 用户获取模型列表失败", res_free_models)
    
    free_models = res_free_models.json().get("data", [])
    free_model_names = [m['name'] for m in free_models]
    # 假设 'free' 等级只能看到 '4k' 模型
    if "豆包 Pro 4k" in free_model_names and "豆包 Pro 32k" not in free_model_names:
        print_success("'free' 用户只能看到 free 等级的模型")
    else:
        print_fail(f"'free' 用户的模型列表不符合预期: {free_model_names}")
    
    free_model_id = free_models[0]['id']
    
    print("  -> 验证 'premium' 用户的模型列表 (继承性)")
    res_premium_models = requests.get(f"{BASE_URL}/models", headers=get_auth_headers(premium_user_token))
    if res_premium_models.status_code != 200:
        print_fail("'premium' 用户获取模型列表失败", res_premium_models)
        
    premium_models = res_premium_models.json().get("data", [])
    premium_model_names = [m['name'] for m in premium_models]
    # 'premium' 用户应该能看到 '4k' 和 '32k' 两个模型
    if "豆包 Pro 4k" in premium_model_names and "豆包 Pro 32k" in premium_model_names:
        print_success("'premium' 用户能看到 free 和 premium 两个等级的模型")
    else:
        print_fail(f"'premium' 用户的模型列表不符合预期 (未体现继承性): {premium_model_names}")
        
    premium_model_id = next(m['id'] for m in premium_models if m['name'] == "豆包 Pro 32k")

    # 3. 【核心测试】验证 'free' 用户无法使用 'premium' 模型
    print("  -> 【安全测试】尝试让 'free' 用户使用 'premium' 模型")
    
    # 'free' 用户先创建一个对话
    res_conv = requests.post(f"{BASE_URL}/conversations", headers=get_auth_headers(free_user_token), json={})
    conv_id = res_conv.json()["data"]["id"]
    
    # 然后在该对话中，故意指定一个 premium 模型ID
    chat_payload = {
        "message": "test",
        "model_id": premium_model_id # 这是 'free' 用户不应能使用的模型
    }
    
    res_permission_denied = requests.post(f"{BASE_URL}/conversations/{conv_id}/messages", 
                                          headers=get_auth_headers(free_user_token),
                                          json=chat_payload)
    
    # 预期服务器返回权限错误 (403 Forbidden)
    if res_permission_denied.status_code == 403:
        print_success("'free' 用户使用 'premium' 模型被按预期拒绝 (403 Forbidden)")
    else:
        print_fail("'free' 用户使用 'premium' 模型的请求未被正确拒绝", res_permission_denied)
        
    # 4. 验证 'premium' 用户可以使用 'premium' 模型
    print("  -> 验证 'premium' 用户可以使用 'premium' 模型")
    
    # 'premium' 用户创建对话
    res_conv_prem = requests.post(f"{BASE_URL}/conversations", headers=get_auth_headers(premium_user_token), json={})
    conv_id_prem = res_conv_prem.json()["data"]["id"]
    
    chat_payload_prem = {
        "message": "test",
        "model_id": premium_model_id
    }
    
    # 这里我们只关心请求是否成功建立，不关心AI的流式返回
    res_permission_ok = requests.post(f"{BASE_URL}/conversations/{conv_id_prem}/messages",
                                      headers=get_auth_headers(premium_user_token),
                                      json=chat_payload_prem,
                                      stream=True)
                                      
    if res_permission_ok.status_code == 200:
        print_success("'premium' 用户使用 'premium' 模型请求成功 (200 OK)")
    else:
        print_fail("'premium' 用户使用 'premium' 模型失败", res_permission_ok)


if __name__ == "__main__":
    print("🚀  开始对AI对话系统后端进行【模型权限继承】自动化测试...")
    test_model_permission_inheritance()
    print("\n" + "="*50)
    print("  🎉  \033[92m模型权限继承测试已成功通过！\033[0m")
    print("="*50)