import requests
import json
import random
import string

# --- Configuration ---
BASE_URL = "http://localhost:8080/api/v1"

# --- Helper Functions (same as before) ---
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
    """Registers and logs in a user, returning their token."""
    requests.post(f"{BASE_URL}/register", json={"username": username, "password": password})
    login_res = requests.post(f"{BASE_URL}/login", json={"username": username, "password": password})
    return login_res.json()["data"]["token"]

# --- Test Workflow ---

def test_cascade_delete_for_categories():
    """Tests that deleting a parent category also deletes all its descendants."""
    print_test_header("多级分类的级联删除")

    # 1. SETUP: Create a user and a nested category structure
    username = f"cascade_user_{random_string()}"
    password = "password123"
    token = register_and_login(username, password)
    headers = get_auth_headers(token)
    
    print(f"  -> 准备测试环境: 创建用户 '{username}'")

    # Create a "safe" category that should NOT be deleted
    res_safe_cat = requests.post(f"{BASE_URL}/categories", headers=headers, json={"name": "Safe Category"})
    safe_category_id = res_safe_cat.json()["data"]["id"]

    # Create Parent -> Child -> Grandchild structure
    res_parent = requests.post(f"{BASE_URL}/categories", headers=headers, json={"name": "Parent"})
    parent_id = res_parent.json()["data"]["id"]
    
    res_child = requests.post(f"{BASE_URL}/categories", headers=headers, json={"name": "Child", "parent_id": parent_id})
    child_id = res_child.json()["data"]["id"]

    res_grandchild = requests.post(f"{BASE_URL}/categories", headers=headers, json={"name": "Grandchild", "parent_id": child_id})
    grandchild_id = res_grandchild.json()["data"]["id"]
    
    print(f"  -> 创建了三级分类: Parent({parent_id}) -> Child({child_id}) -> Grandchild({grandchild_id})")
    print(f"  -> 创建了安全分类: Safe Category({safe_category_id})")

    # Create a conversation linked to the deepest child (Grandchild)
    res_conv_to_delete = requests.post(f"{BASE_URL}/conversations", headers=headers, json={"category_id": grandchild_id})
    conv_to_delete_id = res_conv_to_delete.json()["data"]["id"]
    
    # Create a conversation linked to the safe category
    res_safe_conv = requests.post(f"{BASE_URL}/conversations", headers=headers, json={"category_id": safe_category_id})
    safe_conv_id = res_safe_conv.json()["data"]["id"]

    print(f"  -> 创建并关联了对话: ConvToDelete({conv_to_delete_id}) -> Grandchild")
    print(f"  -> 创建并关联了对话: SafeConv({safe_conv_id}) -> Safe Category")

    # 2. ACTION: Delete the top-level parent category
    print(f"  -> 执行操作: 删除顶级分类 'Parent' (ID: {parent_id})")
    res_delete = requests.delete(f"{BASE_URL}/categories/{parent_id}", headers=headers)
    if res_delete.status_code != 200:
        print_fail("删除顶级分类的请求失败", res_delete)
    
    # 3. VERIFICATION: Check the results
    print("  -> 开始验证结果...")
    
    # Verify categories are deleted
    res_cats_after = requests.get(f"{BASE_URL}/categories", headers=headers)
    categories_after = res_cats_after.json().get("data", [])
    
    existing_cat_ids = {cat['id'] for cat in categories_after}
    deleted_ids = {parent_id, child_id, grandchild_id}

    if existing_cat_ids.isdisjoint(deleted_ids):
        print_success("级联删除验证: Parent, Child, 和 Grandchild 分类均已被删除")
    else:
        print_fail(f"级联删除失败，不应存在的分类ID: {existing_cat_ids.intersection(deleted_ids)}")

    if safe_category_id in existing_cat_ids:
        print_success("安全验证: 'Safe Category' 未受影响，依然存在")
    else:
        print_fail("'Safe Category' 被错误地删除了")

    # Verify conversation links are updated
    res_convs_after = requests.get(f"{BASE_URL}/conversations", headers=headers)
    convs_after = res_convs_after.json().get("data", [])
    
    conv_to_delete_updated = next((c for c in convs_after if c['id'] == conv_to_delete_id), None)
    safe_conv_updated = next((c for c in convs_after if c['id'] == safe_conv_id), None)

    if conv_to_delete_updated and conv_to_delete_updated.get("category_id") is None:
        print_success("关联对话验证: 被删除分类下的对话 category_id 已被正确设为 null")
    else:
        print_fail(f"关联对话的 category_id 未被正确更新: {conv_to_delete_updated}")

    if safe_conv_updated and safe_conv_updated.get("category_id") == safe_category_id:
        print_success("安全验证: 安全对话的分类关联未受影响")
    else:
        print_fail(f"安全对话的分类关联被错误地修改: {safe_conv_updated}")


if __name__ == "__main__":
    print("🚀  开始对AI对话系统后端进行【多级分类级联删除】自动化测试...")
    test_cascade_delete_for_categories()
    print("\n" + "="*50)
    print("  🎉  \033[92m多级分类级联删除测试已成功通过！\033[0m")
    print("="*50)