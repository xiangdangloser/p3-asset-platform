async function register() {
    const username = document.getElementById('username').value.trim();
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirmPassword').value;
    const btn = document.getElementById('registerBtn');

    // 1. 前端基础校验
    if (!username || !password) {
        return alert('用户名和密码不能为空');
    }
    if (password !== confirmPassword) {
        return alert('两次输入的密码不一致，请重新输入');
    }

    btn.innerText = '注册中...';
    btn.disabled = true;

    try {
        // 2. 调用后端注册接口
        const response = await fetch('/api/v1/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();
        
        if (response.ok) {
            alert('注册成功！即将跳转到登录页');
            window.location.href = '/login';
        } else {
            alert('注册失败: ' + (data.error || '未知错误'));
            btn.innerText = '注 册';
            btn.disabled = false;
        }
    } catch (err) {
        console.error(err);
        alert('网络请求失败，请检查后端服务是否启动');
        btn.innerText = '注 册';
        btn.disabled = false;
    }
}