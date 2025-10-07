let currentConversation = null;

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', async () => {
    await loadUserData();
    await loadConversations();
    setupEventListeners();
});

// 加载用户数据
async function loadUserData() {
    try {
        const response = await fetch('/api/user');
        const result = await response.json();

        if (result.success) {
            const user = result.data;
            currentConversation = user.conversation;

            // 更新模型选择
            document.getElementById('modelSelect').value = user.currentModel;

            // 加载消息
            loadMessages(user.conversation.messages);

            // 更新对话标题
            document.getElementById('conversationTitle').textContent = user.conversation.title;
        }
    } catch (error) {
        console.error('加载用户数据失败:', error);
    }
}

// 加载对话列表
async function loadConversations() {
    try {
        const response = await fetch('/api/conversations');
        const result = await response.json();

        if (result.success) {
            displayConversations(result.data.conversations);
        }
    } catch (error) {
        console.error('加载对话列表失败:', error);
    }
}

// 显示对话列表
function displayConversations(conversations) {
    const container = document.getElementById('conversations');
    container.innerHTML = '';

    conversations.forEach(conv => {
        const div = document.createElement('div');
        div.className = `conversation-item ${conv.id === currentConversation?.id ? 'active' : ''}`;
        div.innerHTML = `
            <div><strong>${conv.title}</strong></div>
            <small>${conv.messages.length} 条消息</small>
        `;

        div.addEventListener('click', () => switchConversation(conv.id));
        container.appendChild(div);
    });
}

// 加载消息
function loadMessages(messages) {
    const container = document.getElementById('chatMessages');
    container.innerHTML = '';

    messages.forEach(msg => {
        addMessageToChat(msg.role, msg.content);
    });

    // 滚动到底部
    container.scrollTop = container.scrollHeight;
}

// 添加消息到聊天界面
function addMessageToChat(role, content) {
    const container = document.getElementById('chatMessages');
    const messageDiv = document.createElement('div');
    messageDiv.className = `message-bubble message-${role === 'user' ? 'user' : 'ai'}`;
    messageDiv.textContent = content;
    container.appendChild(messageDiv);

    // 滚动到底部
    container.scrollTop = container.scrollHeight;
}

// 发送消息
document.getElementById('messageForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const input = document.getElementById('messageInput');
    const content = input.value.trim();

    if (!content) return;

    // 添加用户消息到界面
    addMessageToChat('user', content);
    input.value = '';

    try {
        const response = await fetch('/api/message', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ content })
        });

        const result = await response.json();

        if (result.success) {
            // 添加AI响应到界面
            addMessageToChat('assistant', result.data.response);

            // 更新对话信息
            currentConversation = result.data.conversation;
            document.getElementById('conversationTitle').textContent = currentConversation.title;

            // 重新加载对话列表
            await loadConversations();
        } else {
            addMessageToChat('assistant', `错误: ${result.message}`);
        }
    } catch (error) {
        addMessageToChat('assistant', '网络错误，请重试');
    }
});

// 切换对话
async function switchConversation(conversationId) {
    try {
        const response = await fetch(`/api/conversation/${conversationId}`, {
            method: 'PUT'
        });

        const result = await response.json();

        if (result.success) {
            await loadUserData();
            await loadConversations();
        }
    } catch (error) {
        console.error('切换对话失败:', error);
    }
}

// 新建对话
document.getElementById('newConversationBtn').addEventListener('click', async () => {
    const title = prompt('请输入新对话标题:', '新对话');
    if (!title) return;

    try {
        const formData = new FormData();
        formData.append('title', title);

        const response = await fetch('/api/conversation', {
            method: 'POST',
            body: formData
        });

        const result = await response.json();

        if (result.success) {
            await loadUserData();
            await loadConversations();
        }
    } catch (error) {
        console.error('创建对话失败:', error);
    }
});

// 切换模型
document.getElementById('modelSelect').addEventListener('change', async (e) => {
    const model = e.target.value;

    try {
        const formData = new FormData();
        formData.append('model', model);

        const response = await fetch('/api/model', {
            method: 'PUT',
            body: formData
        });

        const result = await response.json();

        if (!result.success) {
            alert(result.message);
            // 恢复原值
            e.target.value = document.getElementById('modelSelect').getAttribute('data-previous-value') || 'basic';
        }
    } catch (error) {
        console.error('切换模型失败:', error);
        // 恢复原值
        e.target.value = document.getElementById('modelSelect').getAttribute('data-previous-value') || 'basic';
    }
});

// 退出登录
document.getElementById('logoutBtn').addEventListener('click', async () => {
    try {
        const response = await fetch('/api/logout', {
            method: 'POST'
        });

        const result = await response.json();

        if (result.success) {
            window.location.href = '/';
        }
    } catch (error) {
        console.error('退出失败:', error);
    }
});

// 设置事件监听器
function setupEventListeners() {
    // 保存模型选择框的原始值
    const modelSelect = document.getElementById('modelSelect');
    modelSelect.setAttribute('data-previous-value', modelSelect.value);
}

// 自动调整输入框高度
function autoResizeTextarea() {
    const textarea = document.getElementById('messageInput');
    textarea.style.height = 'auto';
    textarea.style.height = textarea.scrollHeight + 'px';
}

// 添加输入框事件监听
document.getElementById('messageInput')?.addEventListener('input', autoResizeTextarea);

// 添加键盘快捷键
document.getElementById('messageInput')?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        document.getElementById('messageForm').dispatchEvent(new Event('submit'));
    }
});

// 显示打字指示器
function showTypingIndicator() {
    const container = document.getElementById('chatMessages');
    const indicator = document.createElement('div');
    indicator.id = 'typingIndicator';
    indicator.className = 'message-bubble message-ai typing-indicator';
    indicator.innerHTML = '<div class="typing-dots"><span></span><span></span><span></span></div>';
    container.appendChild(indicator);
    container.scrollTop = container.scrollHeight;
}

// 隐藏打字指示器
function hideTypingIndicator() {
    const indicator = document.getElementById('typingIndicator');
    if (indicator) {
        indicator.remove();
    }
}