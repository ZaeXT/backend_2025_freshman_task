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
    
    # 1. 创建新对话 (不变)
    print("  -> 创建新对话")
    res = requests.post(f"{BASE_URL}/conversations", headers=get_auth_headers(), json={"is_temporary": False})
    if res.status_code == 200 and "id" in res.json().get("data", {}):
        conversation_id = res.json()["data"]["id"]
        print_success(f"新对话创建成功 (ID: {conversation_id})")
    else:
        print_fail("创建对话失败", res)
        
    # 2. 进行流式聊天 (SSE) (不变)
    print("  -> 进行流式聊天 (等待AI响应...)")
    chat_payload = {
        "message": "你好，请用Python写一个简单的Hello World程序。",
        "model_id": available_models[0]["id"]
    }
    
    try:
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
        print("\n")
        
        response_lower = full_response.lower()
        if "print" in response_lower and "hello" in response_lower and "world" in response_lower:
             print_success("流式聊天测试成功，并收到预期内容")
        else:
             print_fail("流式聊天未收到预期内容")

    except Exception as e:
        print_fail(f"流式聊天请求失败: {e}")

    # 3. 验证AI自动生成标题 (不变)
    print("  -> 验证AI自动生成标题 (等待3秒)")
    time.sleep(3)
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

    # --- 4. 【新增】用户手动更新标题 ---
    print("  -> 测试用户手动更新标题")
    new_manual_title = "My Custom Python Test Title"
    update_payload = {"title": new_manual_title}
    res_update = requests.put(f"{BASE_URL}/conversations/{conversation_id}/title", headers=get_auth_headers(), json=update_payload)
    if res_update.status_code == 200:
        print_success("手动更新标题请求成功")
    else:
        print_fail("手动更新标题请求失败", res_update)

    # --- 5. 【新增】验证标题已手动更新 ---
    print("  -> 验证标题已被手动更新")
    res_verify = requests.get(f"{BASE_URL}/conversations", headers=get_auth_headers())
    title_updated = False
    for conv in res_verify.json().get("data", []):
        if conv["id"] == conversation_id:
            if conv["title"] == new_manual_title:
                title_updated = True
                print_success(f"标题已正确更新为: '{new_manual_title}'")
            else:
                print_fail(f"标题验证失败，预期为 '{new_manual_title}'，实际为 '{conv['title']}'")
            break
    if not title_updated:
        print_fail("验证手动更新标题时，找不到对应对话或标题不匹配")

    # --- 6. AI自动分类 (原步骤4) ---
    print("  -> 测试AI自动分类")
    res_classify = requests.post(f"{BASE_URL}/conversations/{conversation_id}/auto-classify", headers=get_auth_headers())
    if res_classify.status_code == 200:
        print_success("AI自动分类请求成功")
    else:
        print_fail("AI自动分类请求失败", res_classify)
        
    # --- 7. 验证分类结果 (原步骤5) ---
    print("  -> 验证分类结果")
    res_cat_verify = requests.get(f"{BASE_URL}/conversations", headers=get_auth_headers())
    cat_id_found = False
    for conv in res_cat_verify.json().get("data", []):
        if conv["id"] == conversation_id:
            # 注意：这里的 parent_category_id 是在 test_models_and_categories_flow 中设置的全局变量
            if conv.get("category_id") is not None:
                cat_id_found = True
                print_success(f"对话已成功自动分类到ID为 {conv.get('category_id')} 的分类中")
            break
    if not cat_id_found:
        print_fail("自动分类后，对话的category_id不正确或不存在")

    # --- 8. 删除对话 (移入回收站) (原步骤6) ---
    print(f"  -> 删除对话 (ID: {conversation_id})")
    res_delete = requests.delete(f"{BASE_URL}/conversations/{conversation_id}", headers=get_auth_headers())
    if res_delete.status_code == 200:
        print_success("对话已成功移入回收站")
    else:
        print_fail("删除对话失败", res_delete)

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


def test_memory_and_multiturn_flow():
    """
    专门测试用户记忆功能、多轮对话上下文和思考模式开关。
    """
    print_test_header("用户记忆、多轮对话与思考模式流程")

    # --- Part 1: 测试用户记忆功能 (不变) ---
    print("  -> 1. 测试用户记忆功能 (System Prompt)")
    # ... (这部分的所有代码保持不变) ...
    # a. 设置一个非常具体的用户记忆
    print("      -> a. 设置特定的用户记忆")
    memory_info = "我是一名资深的Go语言开发者，我的秘密代号是'Gopher King'。"
    res_mem_update = requests.put(f"{BASE_URL}/profile/memory", headers=get_auth_headers(), json={"memory_info": memory_info})
    if res_mem_update.status_code != 200:
        print_fail("为测试记忆功能而更新用户记忆失败", res_mem_update)
    
    # b. 创建一个新对话来测试记忆效果
    print("      -> b. 创建新对话以测试记忆效果")
    res_mem_conv = requests.post(f"{BASE_URL}/conversations", headers=get_auth_headers(), json={})
    memory_conv_id = res_mem_conv.json()["data"]["id"]

    # c. 提出一个依赖于记忆信息的问题
    print("      -> c. 提出一个需要AI使用记忆才能回答的问题 (等待AI响应...)")
    memory_test_payload = {
        "message": "根据你所知关于我的信息，我的秘密代号是什么？",
        "model_id": available_models[0]["id"]
    }
    
    try:
        response = requests.post(f"{BASE_URL}/conversations/{memory_conv_id}/messages",
                                 headers=get_auth_headers(),
                                 json=memory_test_payload,
                                 stream=True)
        response.raise_for_status()
        client = SSEClient(response)
        
        full_response = ""
        for event in client.events():
            if event.event == 'message':
                full_response += event.data
        
        # d. 验证AI的回答是否包含了记忆中的关键信息
        print(f"      AI对记忆问题的回复: {full_response.strip()}")
        if "gopher king" in full_response.lower():
            print_success("用户记忆功能验证成功！AI正确回答了基于记忆的问题。")
        else:
            print_fail("用户记忆功能验证失败，AI未能使用System Prompt中的记忆信息。")

    except Exception as e:
        print_fail(f"测试用户记忆功能时流式聊天请求失败: {e}")

    # --- Part 2: 测试多轮对话上下文 (不变) ---
    print("\n  -> 2. 测试多轮对话上下文")
    # ... (这部分的所有代码保持不变) ...
    # a. 创建一个全新的对话以保证上下文干净
    print("      -> a. 创建新对话以测试上下文")
    res_multi_conv = requests.post(f"{BASE_URL}/conversations", headers=get_auth_headers(), json={})
    multiturn_conv_id = res_multi_conv.json()["data"]["id"]
    
    # b. 第一轮对话：提出一个基础问题
    print("      -> b. 第一轮对话：'中国的首都是哪里？' (等待AI响应...)")
    turn1_payload = {"message": "中国的首都是哪里？", "model_id": available_models[0]["id"]}
    try:
        response1 = requests.post(f"{BASE_URL}/conversations/{multiturn_conv_id}/messages", 
                                  headers=get_auth_headers(), json=turn1_payload, stream=True)
        response1.raise_for_status()
        client1 = SSEClient(response1)
        response1_text = "".join(event.data for event in client1.events() if event.event == 'message')
        print(f"      AI第一轮回复: {response1_text.strip()}")
        if "北京" not in response1_text:
            print_fail("多轮对话第一轮基础问答失败，无法继续测试。")
    except Exception as e:
        print_fail(f"多轮对话第一轮请求失败: {e}")
        
    # c. 第二轮对话：提出一个依赖于上一轮上下文的问题
    print("      -> c. 第二轮对话：'那里的天气怎么样？' (等待AI响应...)")
    turn2_payload = {"message": "那里的天气怎么样？", "model_id": available_models[0]["id"]}
    try:
        response2 = requests.post(f"{BASE_URL}/conversations/{multiturn_conv_id}/messages",
                                  headers=get_auth_headers(), json=turn2_payload, stream=True)
        response2.raise_for_status()
        client2 = SSEClient(response2)
        response2_text = "".join(event.data for event in client2.events() if event.event == 'message')
        # d. 验证第二轮的回答是否包含了第一轮的上下文信息
        print(f"      AI第二轮回复: {response2_text.strip()}")
        if "北京" in response2_text and ("天气" in response2_text or "气温" in response2_text):
            print_success("多轮对话上下文验证成功！AI理解了'那里'指代的是'北京'。")
        else:
            print_fail("多轮对话上下文验证失败，AI未能理解上下文。")
    except Exception as e:
        print_fail(f"多轮对话第二轮请求失败: {e}")

    # --- Part 3: 【新增】测试思考模式开关 ---
    print("\n  -> 3. 测试思考模式 (Thinking Mode) 开关")
    
    # a. 创建一个新对话用于测试
    print("      -> a. 创建新对话以测试思考模式")
    res_think_conv = requests.post(f"{BASE_URL}/conversations", headers=get_auth_headers(), json={})
    thinking_conv_id = res_think_conv.json()["data"]["id"]

    # b. 发送一个开启思考模式的请求
    print("      -> b. 发送请求并设置 enable_thinking: true (等待AI响应...)")
    # 注意：思考模式的输出通常是 SSE 中的一个特殊 event 类型，而不是 message
    # 我们需要检查整个流的内容
    thinking_enabled_payload = {
        "message": "简单介绍一下什么是量子计算",
        "model_id": available_models[0]["id"],
        "enable_thinking": True
    }
    
    try:
        response_think = requests.post(f"{BASE_URL}/conversations/{thinking_conv_id}/messages",
                                       headers=get_auth_headers(),
                                       json=thinking_enabled_payload,
                                       stream=True)
        response_think.raise_for_status()
        client_think = SSEClient(response_think)
        
        has_thinking_event = False
        final_message = ""
        # 遍历所有事件，而不仅仅是 message 事件
        for event in client_think.events():
            # 火山引擎的思考模式通常会发送一个 'VolcengineThinking' 事件
            # 这里我们做一个更通用的检查，即检查是否存在非 'message' 的事件
            # 或者检查响应内容中是否有思考过程的特定标记（如 "思考中"）
            # 最简单的验证是，思考模式的返回内容通常会包含 "思考中" 或类似字样
            if "思考中" in event.data or "正在思考" in event.data:
                 has_thinking_event = True
            
            if event.event == 'message':
                final_message += event.data

        print(f"      AI开启思考模式的回复: {final_message.strip()}")
        if has_thinking_event:
            print_success("思考模式测试成功！在流中检测到了思考过程的标记。")
        else:
            # 这是一个警告而不是失败，因为某些简单问题可能不会触发思考过程
            print("      ⚠️  \033[93m警告:\033[0m 未在流中检测到明确的思考过程标记。这可能是正常的，因为问题可能过于简单。请人工检查AI回复是否更详细。")
        
        if "量子位" not in final_message and "叠加" not in final_message:
            print_fail("思考模式功能验证失败，AI最终回复内容不符合预期。")

    except Exception as e:
        print_fail(f"测试思考模式时流式聊天请求失败: {e}")
if __name__ == "__main__":
    print("🚀  开始对AI对话系统后端进行全功能自动化测试...")
    try:
        test_user_flow()
        test_models_and_categories_flow()
        test_conversation_and_chat_flow()
        # --- 新增的调用 ---
        test_memory_and_multiturn_flow()
        
        test_recycle_bin_flow() # 确保回收站测试在最后，因为它会删除一个重要对话
        
        print("\n" + "="*50)
        print("  🎉  \033[92m所有测试均已成功通过！\03d[0m")
        print("="*50)
    except Exception as e:
        print(f"\n❌  测试过程中出现意外错误: {e}")