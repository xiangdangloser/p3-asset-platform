const token = localStorage.getItem('token');
if (!token) window.location.href = '/login';

function logout() {
    localStorage.removeItem('token');
    window.location.href = '/login';
}

async function startScan() {
    const text = document.getElementById('targets').value;
    const targets = text.split('\n').map(t => t.trim()).filter(t => t);
    if (targets.length === 0) return alert('请输入至少一个目标');

    const btn = document.getElementById('scanBtn');
    const info = document.getElementById('taskInfo');
    
    btn.disabled = true;
    btn.innerText = '任务下发中...';
    info.innerText = '';

    try {
        const res = await fetch('/api/v1/scan', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token },
            body: JSON.stringify({ targets })
        });
        
        if (res.status === 401) return logout();

        const data = await res.json();
        if (res.ok) {
            info.innerText = `任务已下发 (ID: ${data.task_id})，正在后台扫描中...`;
            pollTask(data.task_id);
        } else {
            alert('下发失败: ' + (data.error || '未知错误'));
            btn.disabled = false;
            btn.innerText = '开始扫描';
        }
    } catch (err) {
        console.error(err);
        alert('网络错误');
        btn.disabled = false;
        btn.innerText = '开始扫描';
    }
}

function pollTask(taskId) {
    const info = document.getElementById('taskInfo');
    const btn = document.getElementById('scanBtn');

    const interval = setInterval(async () => {
        const res = await fetch(`/api/v1/tasks/${taskId}`, {
            headers: { 'Authorization': 'Bearer ' + token }
        });
        const data = await res.json();
        
        if (data.data && data.data.status === 'completed') {
            clearInterval(interval);
            info.innerText = '扫描完成！';
            info.className = 'text-sm text-green-400 font-medium';
            btn.disabled = false;
            btn.innerText = '开始扫描';
            loadAssets();
        } else {
            info.innerText = `任务 (ID: ${taskId}) 正在扫描中，请稍候...`;
        }
    }, 2000);
}

async function loadAssets() {
    const res = await fetch('/api/v1/assets', {
        headers: { 'Authorization': 'Bearer ' + token }
    });
    const data = await res.json();
    const tbody = document.getElementById('resultTable');
    tbody.innerHTML = '';

    if (!data.data || data.data.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="py-4 text-center text-gray-500">暂无数据</td></tr>';
        return;
    }

    data.data.forEach(asset => {
        // ⭐ 新增：根据状态码赋予颜色
        let statusColor = 'text-gray-300';
        if (asset.status_code === 200) statusColor = 'text-green-400 font-bold';
        else if (asset.status_code === 0) statusColor = 'text-red-400 font-bold';
        else statusColor = 'text-yellow-400 font-bold';

        tbody.innerHTML += `
            <tr class="border-b border-gray-700 hover:bg-gray-750 transition">
                <td class="py-3">${asset.target}</td>
                <td class="py-3 ${statusColor}">${asset.status_code}</td>
                <td class="py-3">${asset.server || '-'}</td>
                <td class="py-3 text-gray-400">${asset.title || '-'}</td>
                <td class="py-3 text-blue-400 font-medium">${asset.fingerprint || '-'}</td>
            </tr>
        `;
    });
}

// 页面加载时自动拉一次数据
loadAssets();