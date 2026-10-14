async function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const btn = document.getElementById('loginBtn');

    btn.innerText = '登录中...';
    btn.disabled = true;

    try {
        const response = await fetch('/api/v1/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();
        if (response.ok && data.token) {
            localStorage.setItem('token', data.token);
            window.location.href = '/dashboard';
        } else {
            alert('登录失败: ' + (data.error || '未知错误'));
            btn.innerText = '登 录';
            btn.disabled = false;
        }
    } catch (err) {
        console.error(err);
        alert('网络请求失败，请检查后端服务是否启动');
        btn.innerText = '登 录';
        btn.disabled = false;
    }
}