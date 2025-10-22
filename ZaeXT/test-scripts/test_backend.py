import requests
import json
import time
import random
import string
from sseclient import SSEClient

# --- 配置 ---
BASE_URL = "http://localhost:8080/api/v1"

# --- 全局变量，用于在不同测试函数间共享状态 ---
auth_token = None
user_credentials = {}
available_models = []
parent_category_id = None
conversation_id = None

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
            print(f"      响应内容: {response.json()}")
        except json.JSONDecodeError:
            print(f"      响应内容: {response.text}")
    # 强制退出，因为后续测试可能依赖于此
    exit(1)

def get_auth_headers():
    """获取包含JWT的请求头"""
    if not auth_token:
        raise ValueError("用户未登录，无法获取 auth_token")
    return {"Authorization": f"Bearer {auth_token}", "Content-Type": "application/json"}

# --- 测试流程 ---

def test_user_flow():
    """测试用户注册、登录、获取信息和更新记忆"""
    global auth_token, user_credentials
    print_test_header("用户账户流程")

    # 1. 注册
    username = f"testuser_{random_string()}"
    password = "password123"
    user_credentials = {"username": username, "password": password}
    
    print(f"  -> 注册新用户: {username}")
    res = requests.post(f"{BASE_URL}/register", json=user_credentials)
    if res.status_code == 200:
        print_success("用户注册成功")
    else:
        print_fail(f"用户注册失败 (状态码: {res.status_code})", res)

    # 2. 登录
    print(f"  -> 登录用户: {username}")
    res = requests.post(f"{BASE_URL}/login", json=user_credentials)
    if res.status_code == 200 and "token" in res.json().get("data", {}):
        auth_token = res.json()["data"]["token"]
        print_success("用户登录成功，获取到Token")
    else:
        print_fail("用户登录失败", res)

    # 3. 获取用户信息
    print("  -> 获取用户信息")
    res = requests.get(f"{BASE_URL}/profile", headers=get_auth_headers())
    if res.status_code == 200 and res.json().get("data", {}).get("username") == username:
        print_success(f"成功获取到用户 '{username}' 的信息")
        print(f"      默认分类已创建 (下一步验证)")
    else:
        print_fail("获取用户信息失败", res)
        
    # 4. 更新用户记忆
    print("  -> 更新用户记忆")
    memory_info = "我是一名Python开发者，对AI技术非常感兴趣。"
    res = requests.put(f"{BASE_URL}/profile/memory", headers=get_auth_headers(), json={"memory_info": memory_info})
    if res.status_code == 200:
        print_success("更新用户记忆成功")
    else:
        print_fail("更新用户记忆失败", res)
        
    # 5. 验证记忆更新
    print("  -> 验证用户记忆更新")
    res = requests.get(f"{BASE_URL}/profile", headers=get_auth_headers())
    if res.status_code == 200 and res.json().get("data", {}).get("memory_info") == memory_info:
        print_success("用户记忆已正确保存")
    else:
        print_fail("用户记忆验证失败", res)

def test_models_and_categories_flow():
    """测试模型列表和分类管理"""
    global available_models, parent_category_id
    print_test_header("模型与分类流程")
    
    # 1. 获取可用模型列表
    print("  -> 获取可用模型列表")
    res = requests.get(f"{BASE_URL}/models", headers=get_auth_headers())
    if res.status_code == 200 and isinstance(res.json().get("data"), list):
        available_models = res.json()["data"]
        print_success(f"成功获取到 {len(available_models)} 个可用模型")
        for model in available_models:
            print(f"      - {model['name']} (ID: {model['id']})")
    else:
        print_fail("获取模型列表失败", res)

    # 2. 获取默认分类
    print("  -> 验证默认分类")
    res = requests.get(f"{BASE_URL}/categories", headers=get_auth_headers())
    if res.status_code == 200 and len(res.json().get("data", [])) > 0:
        print_success("注册时创建的默认分类已存在")
        parent_category_id = res.json()["data"][0]["id"]
        print(f"      将使用分类 '{res.json()['data'][0]['name']}' (ID: {parent_category_id}) 进行后续测试")
    else:
        print_fail("默认分类不存在或获取失败", res)

def test_conversation_and_chat_flow():
    """测试对话创建、流式聊天、标题和自动分类"""
    global conversation_id
    print_test_header("核心对话与聊天流程")
    
    # 1. 创建新对话
    print("  -> 创建新对话")
    res = requests.post(f"{BASE_URL}/conversations", headers=get_auth_headers(), json={"is_temporary": False})
    if res.status_code == 200 and "id" in res.json().get("data", {}):
        conversation_id = res.json()["data"]["id"]
        print_success(f"新对话创建成功 (ID: {conversation_id})")
    else:
        print_fail("创建对话失败", res)
        
    # 2. 进行流式聊天 (SSE)
    print("  -> 进行流式聊天 (等待AI响应...)")
    chat_payload = {
        "message": "你好，请用Python写一个简单的Hello World程序。（仅输出“HelloWorld!”）",
        "model_id": available_models[0]["id"] # 使用第一个可用的模型
    }
    
    try:
        # requests本身可以处理stream，但sseclient更优雅
        response = requests.post(f"{BASE_URL}/conversations/{conversation_id}/messages", 
                                 headers=get_auth_headers(), 
                                 json=chat_payload, 
                                 stream=True)
        response.raise_for_status()
        client = SSEClient(response)
        
        print("      AI回复: ", end="", flush=True)
        full_response = ""
        for event in client.events():
            if event.event == 'message':
                print(event.data, end="", flush=True)
                full_response += event.data
        print("\n") # 换行
        
        if "print(\"HelloWorld!\")" in full_response:
             print_success("流式聊天测试成功，并收到预期内容")
        else:
             print_fail("流式聊天未收到预期内容")

    except Exception as e:
        print_fail(f"流式聊天请求失败: {e}")

    # 3. 验证AI自动生成标题
    print("  -> 验证AI自动生成标题 (等待10秒)")
    time.sleep(10)
    res = requests.get(f"{BASE_URL}/conversations", headers=get_auth_headers())
    conv_found = False
    for conv in res.json().get("data", []):
        if conv["id"] == conversation_id:
            conv_found = True
            if "New Chat" not in conv["title"]:
                print_success(f"AI自动生成标题成功: '{conv['title']}'")
            else:
                print_fail("AI未能自动生成标题")
            break
    if not conv_found:
        print_fail("验证标题时找不到对话")

    # 4. AI自动分类
    print("  -> 测试AI自动分类")
    res = requests.post(f"{BASE_URL}/conversations/{conversation_id}/auto-classify", headers=get_auth_headers())
    if res.status_code == 200:
        print_success("AI自动分类请求成功")
    else:
        print_fail("AI自动分类请求失败", res)
        
    # 5. 验证分类结果
    print("  -> 验证分类结果")
    res = requests.get(f"{BASE_URL}/conversations", headers=get_auth_headers())
    cat_id_found = False
    for conv in res.json().get("data", []):
        if conv["id"] == conversation_id:
            if conv.get("category_id") == parent_category_id:
                cat_id_found = True
                print_success(f"对话已成功自动分类到ID为 {parent_category_id} 的分类中")
            break
    if not cat_id_found:
        print_fail("自动分类后，对话的category_id不正确")

    # 6. 删除对话 (移入回收站)
    print(f"  -> 删除对话 (ID: {conversation_id})")
    res = requests.delete(f"{BASE_URL}/conversations/{conversation_id}", headers=get_auth_headers())
    if res.status_code == 200:
        print_success("对话已成功移入回收站")
    else:
        print_fail("删除对话失败", res)

def test_recycle_bin_flow():
    """测试回收站的列表、恢复和永久删除"""
    print_test_header("回收站流程")
    
    # 1. 查看回收站
    print("  -> 查看回收站")
    res = requests.get(f"{BASE_URL}/recycle-bin", headers=get_auth_headers())
    if res.status_code == 200:
        items = res.json().get("data", [])
        found = any(item['id'] == conversation_id for item in items)
        if found:
            print_success(f"在回收站中找到对话 (ID: {conversation_id})")
        else:
            print_fail("在回收站中未找到目标对话")
    else:
        print_fail("查看回收站失败", res)

    # 2. 恢复对话
    print(f"  -> 从回收站恢复对话 (ID: {conversation_id})")
    res = requests.post(f"{BASE_URL}/recycle-bin/restore/{conversation_id}", headers=get_auth_headers())
    if res.status_code == 200:
        print_success("恢复对话成功")
    else:
        print_fail("恢复对话失败", res)
        
    # 3. 验证恢复
    print("  -> 验证对话已恢复")
    res = requests.get(f"{BASE_URL}/conversations", headers=get_auth_headers())
    found = any(conv['id'] == conversation_id for conv in res.json().get("data", []))
    if found:
        print_success("已在对话列表中找到恢复的对话")
    else:
        print_fail("验证恢复失败，对话未出现在列表中")
        
    # 4. 再次删除，为永久删除做准备
    print(f"  -> 再次删除对话 (ID: {conversation_id})")
    requests.delete(f"{BASE_URL}/conversations/{conversation_id}", headers=get_auth_headers())
    
    # 5. 永久删除
    print(f"  -> 永久删除对话 (ID: {conversation_id})")
    res = requests.delete(f"{BASE_URL}/recycle-bin/permanent/{conversation_id}", headers=get_auth_headers())
    if res.status_code == 200:
        print_success("永久删除请求成功")
    else:
        print_fail("永久删除请求失败", res)
        
    # 6. 验证永久删除
    print("  -> 验证对话已被永久删除")
    res = requests.get(f"{BASE_URL}/recycle-bin", headers=get_auth_headers())
    found = any(item['id'] == conversation_id for item in res.json().get("data", []))
    if not found:
        print_success("已确认对话不在回收站中，永久删除成功")
    else:
        print_fail("永久删除失败，对话仍在回收站中")


if __name__ == "__main__":
    print("🚀  开始对AI对话系统后端进行全功能自动化测试...")
    try:
        test_user_flow()
        test_models_and_categories_flow()
        test_conversation_and_chat_flow()
        test_recycle_bin_flow()
        print("\n" + "="*50)
        print("  🎉  \033[92m所有测试均已成功通过！\033[0m")
        print("="*50)
    except Exception as e:
        print(f"\n❌  测试过程中出现意外错误: {e}")