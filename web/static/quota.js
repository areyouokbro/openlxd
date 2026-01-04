// ========== 配额管理 ==========

async function loadQuotas() {
    const tbody = document.getElementById('quota-tbody');
    tbody.innerHTML = '<tr><td colspan="9" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/quota');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="9" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const quotas = data.data || [];
    
    if (quotas.length === 0) {
        tbody.innerHTML = '<tr><td colspan="9" class="empty-state">暂无配额设置</td></tr>';
        return;
    }
    
    tbody.innerHTML = quotas.map(q => `
        <tr>
            <td><strong>${q.container_id}</strong></td>
            <td>${q.ipv4_quota === -1 ? '无限制' : q.ipv4_quota}</td>
            <td>${q.ipv6_quota === -1 ? '无限制' : q.ipv6_quota}</td>
            <td>${q.port_mapping_quota === -1 ? '无限制' : q.port_mapping_quota}</td>
            <td>${q.proxy_quota === -1 ? '无限制' : q.proxy_quota}</td>
            <td>${q.traffic_quota === -1 ? '无限制' : q.traffic_quota}</td>
            <td>${q.traffic_used || 0}</td>
            <td><span class="badge badge-${q.on_exceed === 'stop' ? 'danger' : q.on_exceed === 'limit' ? 'warning' : 'info'}">${q.on_exceed}</span></td>
            <td>
                <button class="btn btn-sm btn-primary" onclick="showEditQuotaModal(${q.container_id})">编辑</button>
                <button class="btn btn-sm btn-warning" onclick="resetTraffic(${q.container_id})">重置流量</button>
                <button class="btn btn-sm btn-danger" onclick="deleteQuota(${q.container_id})">删除</button>
            </td>
        </tr>
    `).join('');
}

function showSetQuotaModal() {
    const html = `
        <div class="modal active" id="set-quota-modal">
            <div class="modal-content">
                <h2>设置配额</h2>
                <form id="set-quota-form">
                    <div class="form-group">
                        <label>容器ID</label>
                        <input type="number" name="container_id" required>
                    </div>
                    <div class="form-group">
                        <label>IPv4配额 (-1表示无限制)</label>
                        <input type="number" name="ipv4_quota" value="-1" required>
                    </div>
                    <div class="form-group">
                        <label>IPv6配额 (-1表示无限制)</label>
                        <input type="number" name="ipv6_quota" value="-1" required>
                    </div>
                    <div class="form-group">
                        <label>端口映射配额 (-1表示无限制)</label>
                        <input type="number" name="port_mapping_quota" value="-1" required>
                    </div>
                    <div class="form-group">
                        <label>反向代理配额 (-1表示无限制)</label>
                        <input type="number" name="proxy_quota" value="-1" required>
                    </div>
                    <div class="form-group">
                        <label>流量配额(GB) (-1表示无限制)</label>
                        <input type="number" name="traffic_quota" value="-1" required>
                    </div>
                    <div class="form-group">
                        <label>超限处理</label>
                        <select name="on_exceed" required>
                            <option value="warn">警告</option>
                            <option value="limit">限制</option>
                            <option value="stop">停止容器</option>
                        </select>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">设置</button>
                        <button type="button" class="btn btn-secondary" onclick="closeModal('set-quota-modal')">取消</button>
                    </div>
                </form>
            </div>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', html);
    
    document.getElementById('set-quota-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData);
        
        // 转换数字类型
        data.container_id = parseInt(data.container_id);
        data.ipv4_quota = parseInt(data.ipv4_quota);
        data.ipv6_quota = parseInt(data.ipv6_quota);
        data.port_mapping_quota = parseInt(data.port_mapping_quota);
        data.proxy_quota = parseInt(data.proxy_quota);
        data.traffic_quota = parseInt(data.traffic_quota);
        
        const result = await apiRequest('/api/quota', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        if (result && result.code === 200) {
            showAlert('配额设置成功', 'success');
            closeModal('set-quota-modal');
            loadQuotas();
        } else {
            showAlert(result.message || '设置失败', 'error');
        }
    });
}

function showEditQuotaModal(containerID) {
    // 获取当前配额信息
    apiRequest(`/api/quota?container_id=${containerID}`).then(data => {
        if (!data || data.code !== 200) {
            showAlert('获取配额信息失败', 'error');
            return;
        }
        
        const quota = data.data;
        
        const html = `
            <div class="modal active" id="edit-quota-modal">
                <div class="modal-content">
                    <h2>编辑配额 - 容器 ${containerID}</h2>
                    <form id="edit-quota-form">
                        <input type="hidden" name="container_id" value="${containerID}">
                        <div class="form-group">
                            <label>IPv4配额 (-1表示无限制)</label>
                            <input type="number" name="ipv4_quota" value="${quota.ipv4_quota}" required>
                        </div>
                        <div class="form-group">
                            <label>IPv6配额 (-1表示无限制)</label>
                            <input type="number" name="ipv6_quota" value="${quota.ipv6_quota}" required>
                        </div>
                        <div class="form-group">
                            <label>端口映射配额 (-1表示无限制)</label>
                            <input type="number" name="port_mapping_quota" value="${quota.port_mapping_quota}" required>
                        </div>
                        <div class="form-group">
                            <label>反向代理配额 (-1表示无限制)</label>
                            <input type="number" name="proxy_quota" value="${quota.proxy_quota}" required>
                        </div>
                        <div class="form-group">
                            <label>流量配额(GB) (-1表示无限制)</label>
                            <input type="number" name="traffic_quota" value="${quota.traffic_quota}" required>
                        </div>
                        <div class="form-group">
                            <label>超限处理</label>
                            <select name="on_exceed" required>
                                <option value="warn" ${quota.on_exceed === 'warn' ? 'selected' : ''}>警告</option>
                                <option value="limit" ${quota.on_exceed === 'limit' ? 'selected' : ''}>限制</option>
                                <option value="stop" ${quota.on_exceed === 'stop' ? 'selected' : ''}>停止容器</option>
                            </select>
                        </div>
                        <div class="form-actions">
                            <button type="submit" class="btn btn-primary">更新</button>
                            <button type="button" class="btn btn-secondary" onclick="closeModal('edit-quota-modal')">取消</button>
                        </div>
                    </form>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', html);
        
        document.getElementById('edit-quota-form').addEventListener('submit', async (e) => {
            e.preventDefault();
            const formData = new FormData(e.target);
            const updates = {};
            
            for (let [key, value] of formData.entries()) {
                if (key !== 'container_id') {
                    updates[key] = isNaN(value) ? value : parseInt(value);
                }
            }
            
            const result = await apiRequest('/api/quota', {
                method: 'PUT',
                body: JSON.stringify({
                    container_id: containerID,
                    updates: updates
                })
            });
            
            if (result && result.code === 200) {
                showAlert('配额更新成功', 'success');
                closeModal('edit-quota-modal');
                loadQuotas();
            } else {
                showAlert(result.message || '更新失败', 'error');
            }
        });
    });
}

async function deleteQuota(containerID) {
    if (!confirm(`确定要删除容器 ${containerID} 的配额设置吗？`)) return;
    
    const result = await apiRequest(`/api/quota?container_id=${containerID}`, {
        method: 'DELETE'
    });
    
    if (result && result.code === 200) {
        showAlert('配额删除成功', 'success');
        loadQuotas();
    } else {
        showAlert(result.message || '删除失败', 'error');
    }
}

async function resetTraffic(containerID) {
    if (!confirm(`确定要重置容器 ${containerID} 的流量统计吗？`)) return;
    
    const result = await apiRequest(`/api/quota/reset-traffic?container_id=${containerID}`, {
        method: 'POST'
    });
    
    if (result && result.code === 200) {
        showAlert('流量重置成功', 'success');
        loadQuotas();
    } else {
        showAlert(result.message || '重置失败', 'error');
    }
}

// 辅助函数
function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.remove();
    }
}
