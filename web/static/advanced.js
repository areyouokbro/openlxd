// 高级功能管理 JavaScript

// ==================== 快照管理 ====================

// 加载容器列表到快照选择器
function loadSnapshotContainers() {
    fetch('/api/system/containers', {
        headers: { 'X-API-Hash': getAPIKey() }
    })
    .then(res => res.json())
    .then(data => {
        const select = document.getElementById('snapshot-container-select');
        select.innerHTML = '<option value="">选择容器...</option>';
        
        if (data.data && data.data.length > 0) {
            data.data.forEach(container => {
                const option = document.createElement('option');
                option.value = container.name;
                option.textContent = container.name;
                select.appendChild(option);
            });
        }
    })
    .catch(err => {
        console.error('加载容器列表失败:', err);
    });
}

// 加载快照列表
function loadSnapshots() {
    const containerName = document.getElementById('snapshot-container-select').value;
    const tbody = document.getElementById('snapshot-tbody');
    
    if (!containerName) {
        tbody.innerHTML = '<tr><td colspan="5" class="loading">请选择容器...</td></tr>';
        return;
    }
    
    tbody.innerHTML = '<tr><td colspan="5" class="loading">加载中...</td></tr>';
    
    fetch(`/api/snapshots?container=${containerName}`, {
        headers: { 'X-API-Hash': getAPIKey() }
    })
    .then(res => res.json())
    .then(data => {
        if (data.data && data.data.length > 0) {
            tbody.innerHTML = data.data.map(snap => `
                <tr>
                    <td>${snap.name}</td>
                    <td>${new Date(snap.created_at).toLocaleString('zh-CN')}</td>
                    <td>${snap.stateful ? '是' : '否'}</td>
                    <td>${formatSize(snap.size || 0)}</td>
                    <td class="action-buttons">
                        <button class="btn btn-sm btn-primary" onclick="restoreSnapshot('${containerName}', '${snap.name}')">恢复</button>
                        <button class="btn btn-sm btn-danger" onclick="deleteSnapshot('${containerName}', '${snap.name}')">删除</button>
                    </td>
                </tr>
            `).join('');
        } else {
            tbody.innerHTML = '<tr><td colspan="5" class="empty">暂无快照</td></tr>';
        }
    })
    .catch(err => {
        console.error('加载快照列表失败:', err);
        tbody.innerHTML = '<tr><td colspan="5" class="error">加载失败</td></tr>';
    });
}

// 显示创建快照模态框
function showCreateSnapshotModal() {
    const containerName = document.getElementById('snapshot-container-select').value;
    
    if (!containerName) {
        showAlert('请先选择容器', 'warning');
        return;
    }
    
    const modal = `
        <div class="modal active" id="create-snapshot-modal">
            <div class="modal-content">
                <h2>📸 创建快照</h2>
                <form onsubmit="createSnapshot(event)">
                    <div class="form-group">
                        <label>容器名称</label>
                        <input type="text" value="${containerName}" disabled>
                    </div>
                    <div class="form-group">
                        <label>快照名称（可选）</label>
                        <input type="text" id="snapshot-name" placeholder="留空自动生成">
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="snapshot-stateful">
                            有状态快照（保存内存状态）
                        </label>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">创建</button>
                        <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                    </div>
                </form>
            </div>
        </div>
    `;
    
    document.getElementById('modal-container').innerHTML = modal;
}

// 创建快照
function createSnapshot(event) {
    event.preventDefault();
    
    const containerName = document.getElementById('snapshot-container-select').value;
    const snapshotName = document.getElementById('snapshot-name').value;
    const stateful = document.getElementById('snapshot-stateful').checked;
    
    fetch(`/api/snapshots?container=${containerName}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': getAPIKey()
        },
        body: JSON.stringify({
            snapshot_name: snapshotName,
            stateful: stateful
        })
    })
    .then(res => res.json())
    .then(data => {
        if (data.code === 200) {
            showAlert('快照创建成功', 'success');
            closeModal();
            loadSnapshots();
        } else {
            showAlert(data.message || '快照创建失败', 'error');
        }
    })
    .catch(err => {
        console.error('创建快照失败:', err);
        showAlert('创建快照失败', 'error');
    });
}

// 恢复快照
function restoreSnapshot(containerName, snapshotName) {
    if (!confirm(`确定要将容器 ${containerName} 恢复到快照 ${snapshotName} 吗？\n\n注意：这将覆盖当前容器状态！`)) {
        return;
    }
    
    fetch(`/api/snapshots?container=${containerName}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': getAPIKey()
        },
        body: JSON.stringify({
            snapshot_name: snapshotName
        })
    })
    .then(res => res.json())
    .then(data => {
        if (data.code === 200) {
            showAlert('快照恢复成功', 'success');
            loadSnapshots();
        } else {
            showAlert(data.message || '快照恢复失败', 'error');
        }
    })
    .catch(err => {
        console.error('恢复快照失败:', err);
        showAlert('恢复快照失败', 'error');
    });
}

// 删除快照
function deleteSnapshot(containerName, snapshotName) {
    if (!confirm(`确定要删除快照 ${snapshotName} 吗？`)) {
        return;
    }
    
    fetch(`/api/snapshots?container=${containerName}&snapshot=${snapshotName}`, {
        method: 'DELETE',
        headers: { 'X-API-Hash': getAPIKey() }
    })
    .then(res => res.json())
    .then(data => {
        if (data.code === 200) {
            showAlert('快照删除成功', 'success');
            loadSnapshots();
        } else {
            showAlert(data.message || '快照删除失败', 'error');
        }
    })
    .catch(err => {
        console.error('删除快照失败:', err);
        showAlert('删除快照失败', 'error');
    });
}

// ==================== 克隆管理 ====================

// 显示克隆模态框
function showCloneModal() {
    fetch('/api/system/containers', {
        headers: { 'X-API-Hash': getAPIKey() }
    })
    .then(res => res.json())
    .then(data => {
        const containers = data.data || [];
        const containerOptions = containers.map(c => 
            `<option value="${c.name}">${c.name}</option>`
        ).join('');
        
        const modal = `
            <div class="modal active" id="clone-modal">
                <div class="modal-content">
                    <h2>📋 克隆容器</h2>
                    <form onsubmit="cloneContainer(event)">
                        <div class="form-group">
                            <label>源容器</label>
                            <select id="clone-source" required onchange="loadSnapshotsForClone()">
                                <option value="">选择源容器...</option>
                                ${containerOptions}
                            </select>
                        </div>
                        <div class="form-group">
                            <label>
                                <input type="checkbox" id="clone-from-snapshot" onchange="toggleSnapshotSelect()">
                                从快照克隆
                            </label>
                        </div>
                        <div class="form-group" id="snapshot-select-group" style="display: none;">
                            <label>选择快照</label>
                            <select id="clone-snapshot">
                                <option value="">请先选择源容器...</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label>目标容器名称</label>
                            <input type="text" id="clone-target" required placeholder="新容器名称">
                        </div>
                        <div class="form-actions">
                            <button type="submit" class="btn btn-primary">开始克隆</button>
                            <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                        </div>
                    </form>
                </div>
            </div>
        `;
        
        document.getElementById('modal-container').innerHTML = modal;
    })
    .catch(err => {
        console.error('加载容器列表失败:', err);
        showAlert('加载容器列表失败', 'error');
    });
}

// 切换快照选择显示
function toggleSnapshotSelect() {
    const checked = document.getElementById('clone-from-snapshot').checked;
    const group = document.getElementById('snapshot-select-group');
    group.style.display = checked ? 'block' : 'none';
    
    if (checked) {
        loadSnapshotsForClone();
    }
}

// 加载快照列表用于克隆
function loadSnapshotsForClone() {
    const containerName = document.getElementById('clone-source').value;
    const select = document.getElementById('clone-snapshot');
    
    if (!containerName) {
        select.innerHTML = '<option value="">请先选择源容器...</option>';
        return;
    }
    
    select.innerHTML = '<option value="">加载中...</option>';
    
    fetch(`/api/snapshots?container=${containerName}`, {
        headers: { 'X-API-Hash': getAPIKey() }
    })
    .then(res => res.json())
    .then(data => {
        if (data.data && data.data.length > 0) {
            select.innerHTML = '<option value="">选择快照...</option>' +
                data.data.map(snap => 
                    `<option value="${snap.name}">${snap.name} (${new Date(snap.created_at).toLocaleString('zh-CN')})</option>`
                ).join('');
        } else {
            select.innerHTML = '<option value="">该容器没有快照</option>';
        }
    })
    .catch(err => {
        console.error('加载快照列表失败:', err);
        select.innerHTML = '<option value="">加载失败</option>';
    });
}

// 克隆容器
function cloneContainer(event) {
    event.preventDefault();
    
    const sourceContainer = document.getElementById('clone-source').value;
    const targetContainer = document.getElementById('clone-target').value;
    const fromSnapshot = document.getElementById('clone-from-snapshot').checked;
    const snapshotName = fromSnapshot ? document.getElementById('clone-snapshot').value : '';
    
    if (fromSnapshot && !snapshotName) {
        showAlert('请选择快照', 'warning');
        return;
    }
    
    const requestBody = {
        source_container: sourceContainer,
        target_container: targetContainer
    };
    
    if (fromSnapshot) {
        requestBody.snapshot_name = snapshotName;
    }
    
    fetch('/api/clone', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': getAPIKey()
        },
        body: JSON.stringify(requestBody)
    })
    .then(res => res.json())
    .then(data => {
        if (data.code === 200) {
            showAlert('容器克隆成功', 'success');
            closeModal();
        } else {
            showAlert(data.message || '容器克隆失败', 'error');
        }
    })
    .catch(err => {
        console.error('克隆容器失败:', err);
        showAlert('克隆容器失败', 'error');
    });
}

// ==================== DNS 设置 ====================

// 加载容器列表到DNS选择器
function loadDNSContainers() {
    fetch('/api/system/containers', {
        headers: { 'X-API-Hash': getAPIKey() }
    })
    .then(res => res.json())
    .then(data => {
        const select = document.getElementById('dns-container-select');
        select.innerHTML = '<option value="">选择容器...</option>';
        
        if (data.data && data.data.length > 0) {
            data.data.forEach(container => {
                const option = document.createElement('option');
                option.value = container.name;
                option.textContent = container.name;
                select.appendChild(option);
            });
        }
    })
    .catch(err => {
        console.error('加载容器列表失败:', err);
    });
}

// 加载DNS配置
function loadDNSConfig() {
    const containerName = document.getElementById('dns-container-select').value;
    const textarea = document.getElementById('dns-servers');
    
    if (!containerName) {
        textarea.value = '';
        textarea.placeholder = '请先选择容器...';
        return;
    }
    
    textarea.value = '加载中...';
    
    fetch(`/api/dns?container=${containerName}`, {
        headers: { 'X-API-Hash': getAPIKey() }
    })
    .then(res => res.json())
    .then(data => {
        if (data.data && data.data.dns_servers) {
            textarea.value = data.data.dns_servers.join('\n');
        } else {
            textarea.value = '';
            textarea.placeholder = '未配置 DNS 服务器';
        }
    })
    .catch(err => {
        console.error('加载DNS配置失败:', err);
        textarea.value = '';
        textarea.placeholder = '加载失败';
        showAlert('加载DNS配置失败', 'error');
    });
}

// 保存DNS配置
function saveDNSConfig() {
    const containerName = document.getElementById('dns-container-select').value;
    const dnsText = document.getElementById('dns-servers').value;
    
    if (!containerName) {
        showAlert('请先选择容器', 'warning');
        return;
    }
    
    // 解析DNS服务器列表
    const dnsServers = dnsText.split('\n')
        .map(line => line.trim())
        .filter(line => line.length > 0);
    
    if (dnsServers.length === 0) {
        showAlert('请至少输入一个DNS服务器', 'warning');
        return;
    }
    
    fetch(`/api/dns?container=${containerName}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': getAPIKey()
        },
        body: JSON.stringify({
            dns_servers: dnsServers
        })
    })
    .then(res => res.json())
    .then(data => {
        if (data.code === 200) {
            showAlert('DNS配置保存成功', 'success');
        } else {
            showAlert(data.message || 'DNS配置保存失败', 'error');
        }
    })
    .catch(err => {
        console.error('保存DNS配置失败:', err);
        showAlert('保存DNS配置失败', 'error');
    });
}

// ==================== 工具函数 ====================

// 格式化大小
function formatSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', function() {
    // 初始化容器选择器
    loadSnapshotContainers();
    loadDNSContainers();
});
