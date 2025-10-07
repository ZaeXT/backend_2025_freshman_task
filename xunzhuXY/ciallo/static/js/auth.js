// 登录表单处理
document.getElementById('loginForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = {
        username: document.getElementById('username').value,
        password: document.getElementById('password').value
    };

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        const result = await response.json();

        if (result.success) {
            showMessage('登录成功！正在跳转...', 'success');
            setTimeout(() => {
                window.location.href = '/chat';
            }, 1000);
        } else {
            showMessage(result.message, 'error');
        }
    } catch (error) {
        showMessage('网络错误，请重试', 'error');
    }
});

// 注册表单处理
document.getElementById('registerForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();

    const formData = {
        username: document.getElementById('username').value,
        password: document.getElementById('password').value,
        confirm: document.getElementById('confirm').value
    };

    if (formData.password !== formData.confirm) {
        showMessage('两次密码不一致', 'error');
        return;
    }

    if (formData.password.length < 6) {
        showMessage('密码长度至少6位', 'error');
        return;
    }

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(formData)
        });

        const result = await response.json();

        if (result.success) {
            showMessage('注册成功！正在跳转...', 'success');
            setTimeout(() => {
                window.location.href = '/chat';
            }, 1000);
        } else {
            showMessage(result.message, 'error');
        }
    } catch (error) {
        showMessage('网络错误，请重试', 'error');
    }
});

// 显示消息
function showMessage(message, type) {
    const messageDiv = document.getElementById('message');
    messageDiv.textContent = message;
    messageDiv.className = `message ${type}`;
}