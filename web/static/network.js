// ========== IP地址池管理 ==========

async function loadIPPool() {
    const tbody = document.getElementById('ippool-tbody');
    tbody.innerHTML = '<tr><td colspan="7" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/network/ippool');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const ipAddresses = data.data.ip_addresses || [];
    
    // 更新统计信息
    document.getElementById('ipv4-available').textContent = `IPv4: ${data.data.ipv4_available}`;
    document.getElementById('ipv6-available').textContent = `IPv6: ${data.data.ipv6_available}`;
    
    if (ipAddresses.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">暂无IP地址</td></tr>';
        return;
    }
    
    tbody.innerHTML = ipAddresses.map(ip => `
        <tr>
            <td><strong>${ip.ip}</strong></td>
            <td><span class="badge badge-${ip.type === 'ipv4' ? 'primary' : 'info'}">${ip.type.toUpperCase()}</span></td>
            <td><span class="status-badge status-${ip.status === 'available' ? 'success' : 'warning'}">${ip.status}</span></td>
            <td>${ip.container_id || '-'}</td>
            <td>${ip.gateway || '-'}</td>
            <td>${ip.netmask || '-'}</td>
            <td>
                ${ip.status === 'used' ? 
                    `<button class="btn btn-sm btn-warning" onclick="releaseIP(${ip.id})">释放</button>` : 
                    `<button class="btn btn-sm btn-danger" onclick="deleteIP(${ip.id})">删除</button>`
                }
            </td>
        </tr>
    `).join('');
}

function showAddIPRangeModal() {
    const html = `
        <div class="modal active" id="add-iprange-modal">
            <div class="modal-content">
                <h2>添加IP地址段</h2>
                <form id="add-iprange-form">
                    <div class="form-group">
                        <label>起始IP</label>
                        <input type="text" name="start_ip" placeholder="例如: 192.168.1.100" required>
                    </div>
                    <div class="form-group">
                        <label>结束IP</label>
                        <input type="text" name="end_ip" placeholder="例如: 192.168.1.200" required>
                    </div>
                    <div class="form-group">
                        <label>网关</label>
                        <input type="text" name="gateway" placeholder="例如: 192.168.1.1" required>
                    </div>
                    <div class="form-group">
                        <label>子网掩码</label>
                        <input type="text" name="netmask" placeholder="例如: 255.255.255.0" required>
                    </div>
                    <div class="form-group">
                        <label>类型</label>
                        <select name="type" required>
                            <option value="ipv4">IPv4</option>
                            <option value="ipv6">IPv6</option>
                        </select>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">添加</button>
                        <button type="button" class="btn btn-secondary" onclick="closeModal('add-iprange-modal')">取消</button>
                    </div>
                </form>
            </div>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', html);
    
    document.getElementById('add-iprange-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData);
        
        const result = await apiRequest('/api/network/ippool', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        if (result && result.code === 200) {
            showAlert('IP地址段添加成功', 'success');
            closeModal('add-iprange-modal');
            loadIPPool();
        } else {
            showAlert(result.msg || '添加失败', 'error');
        }
    });
}

async function deleteIP(ipID) {
    if (!confirm('确定要删除这个IP地址吗？')) return;
    
    const result = await apiRequest(`/api/network/ippool?id=${ipID}`, {
        method: 'DELETE'
    });
    
    if (result && result.code === 200) {
        showAlert('IP地址删除成功', 'success');
        loadIPPool();
    } else {
        showAlert(result.msg || '删除失败', 'error');
    }
}

// ========== 端口映射管理 ==========

async function loadPortMappings() {
    const tbody = document.getElementById('portmapping-tbody');
    tbody.innerHTML = '<tr><td colspan="8" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/network/portmapping');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="8" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const mappings = data.data || [];
    
    if (mappings.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" class="empty-state">暂无端口映射</td></tr>';
        return;
    }
    
    tbody.innerHTML = mappings.map(m => `
        <tr>
            <td>${m.container_id}</td>
            <td>${m.container_ip}</td>
            <td><span class="badge badge-${m.protocol === 'tcp' ? 'primary' : 'info'}">${m.protocol.toUpperCase()}</span></td>
            <td><strong>${m.external_port}</strong></td>
            <td>${m.internal_port}</td>
            <td>${m.description || '-'}</td>
            <td><span class="status-badge status-${m.status === 'active' ? 'running' : 'stopped'}">${m.status}</span></td>
            <td>
                <button class="btn btn-sm btn-danger" onclick="deletePortMapping(${m.id})">删除</button>
            </td>
        </tr>
    `).join('');
}

function showAddPortMappingModal() {
    const html = `
        <div class="modal active" id="add-portmapping-modal">
            <div class="modal-content">
                <h2>添加端口映射</h2>
                <form id="add-portmapping-form">
                    <div class="form-group">
                        <label>容器ID</label>
                        <input type="number" name="container_id" required>
                    </div>
                    <div class="form-group">
                        <label>容器IP</label>
                        <input type="text" name="container_ip" placeholder="例如: 10.0.0.100" required>
                    </div>
                    <div class="form-group">
                        <label>协议</label>
                        <select name="protocol" required>
                            <option value="tcp">TCP</option>
                            <option value="udp">UDP</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>映射类型</label>
                        <select name="type" id="mapping-type" onchange="updateMappingForm()" required>
                            <option value="single">单端口映射</option>
                            <option value="range">端口段映射</option>
                            <option value="random">随机端口</option>
                        </select>
                    </div>
                    <div class="form-group" id="external-port-group">
                        <label>外部端口</label>
                        <input type="number" name="external_port" min="1" max="65535">
                    </div>
                    <div class="form-group">
                        <label>内部端口</label>
                        <input type="number" name="internal_port" min="1" max="65535" required>
                    </div>
                    <div class="form-group" id="count-group" style="display:none;">
                        <label>端口数量</label>
                        <input type="number" name="count" min="1" max="1000">
                    </div>
                    <div class="form-group">
                        <label>说明</label>
                        <input type="text" name="description">
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">添加</button>
                        <button type="button" class="btn btn-secondary" onclick="closeModal('add-portmapping-modal')">取消</button>
                    </div>
                </form>
            </div>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', html);
    
    document.getElementById('add-portmapping-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData);
        
        // 转换数字类型
        data.container_id = parseInt(data.container_id);
        data.external_port = parseInt(data.external_port) || 0;
        data.internal_port = parseInt(data.internal_port);
        data.count = parseInt(data.count) || 0;
        
        const result = await apiRequest('/api/network/portmapping', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        if (result && result.code === 200) {
            showAlert('端口映射添加成功', 'success');
            closeModal('add-portmapping-modal');
            loadPortMappings();
        } else {
            showAlert(result.msg || '添加失败', 'error');
        }
    });
}

function updateMappingForm() {
    const type = document.getElementById('mapping-type').value;
    const externalPortGroup = document.getElementById('external-port-group');
    const countGroup = document.getElementById('count-group');
    
    if (type === 'random') {
        externalPortGroup.style.display = 'none';
        countGroup.style.display = 'none';
    } else if (type === 'range') {
        externalPortGroup.style.display = 'block';
        countGroup.style.display = 'block';
    } else {
        externalPortGroup.style.display = 'block';
        countGroup.style.display = 'none';
    }
}

async function deletePortMapping(mappingID) {
    if (!confirm('确定要删除这个端口映射吗？')) return;
    
    const result = await apiRequest(`/api/network/portmapping?id=${mappingID}`, {
        method: 'DELETE'
    });
    
    if (result && result.code === 200) {
        showAlert('端口映射删除成功', 'success');
        loadPortMappings();
    } else {
        showAlert(result.msg || '删除失败', 'error');
    }
}

// ========== 反向代理管理 ==========

async function loadProxies() {
    const tbody = document.getElementById('proxy-tbody');
    tbody.innerHTML = '<tr><td colspan="7" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/network/proxy');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const proxies = data.data || [];
    
    if (proxies.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">暂无反向代理</td></tr>';
        return;
    }
    
    tbody.innerHTML = proxies.map(p => `
        <tr>
            <td>${p.container_id}</td>
            <td><strong>${p.domain}</strong></td>
            <td>${p.target_ip}</td>
            <td>${p.target_port}</td>
            <td>${p.ssl ? '<span class="badge badge-success">✓ HTTPS</span>' : '<span class="badge badge-secondary">HTTP</span>'}</td>
            <td><span class="status-badge status-${p.status === 'active' ? 'running' : 'stopped'}">${p.status}</span></td>
            <td>
                ${!p.ssl ? `<button class="btn btn-sm btn-info" onclick="updateProxySSL(${p.id})">配置SSL</button>` : ''}
                <button class="btn btn-sm btn-danger" onclick="deleteProxy(${p.id})">删除</button>
            </td>
        </tr>
    `).join('');
}

function showAddProxyModal() {
    const html = `
        <div class="modal active" id="add-proxy-modal">
            <div class="modal-content">
                <h2>添加反向代理</h2>
                <form id="add-proxy-form">
                    <div class="form-group">
                        <label>容器ID</label>
                        <input type="number" name="container_id" required>
                    </div>
                    <div class="form-group">
                        <label>域名</label>
                        <input type="text" name="domain" placeholder="例如: example.com" required>
                    </div>
                    <div class="form-group">
                        <label>目标IP</label>
                        <input type="text" name="target_ip" placeholder="例如: 10.0.0.100" required>
                    </div>
                    <div class="form-group">
                        <label>目标端口</label>
                        <input type="number" name="target_port" min="1" max="65535" value="80" required>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" name="ssl" id="ssl-checkbox" onchange="toggleSSLFields()">
                            启用SSL/HTTPS
                        </label>
                    </div>
                    <div id="ssl-fields" style="display:none;">
                        <div class="form-group">
                            <label>证书路径</label>
                            <input type="text" name="cert_path" placeholder="/path/to/cert.pem">
                        </div>
                        <div class="form-group">
                            <label>私钥路径</label>
                            <input type="text" name="key_path" placeholder="/path/to/key.pem">
                        </div>
                    </div>
                    <div class="form-actions">
                        <button type="submit" class="btn btn-primary">添加</button>
                        <button type="button" class="btn btn-secondary" onclick="closeModal('add-proxy-modal')">取消</button>
                    </div>
                </form>
            </div>
        </div>
    `;
    
    document.body.insertAdjacentHTML('beforeend', html);
    
    document.getElementById('add-proxy-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const formData = new FormData(e.target);
        const data = Object.fromEntries(formData);
        
        // 转换数字和布尔类型
        data.container_id = parseInt(data.container_id);
        data.target_port = parseInt(data.target_port);
        data.ssl = document.getElementById('ssl-checkbox').checked;
        
        const result = await apiRequest('/api/network/proxy', {
            method: 'POST',
            body: JSON.stringify(data)
        });
        
        if (result && result.code === 200) {
            showAlert('反向代理添加成功', 'success');
            closeModal('add-proxy-modal');
            loadProxies();
        } else {
            showAlert(result.msg || '添加失败', 'error');
        }
    });
}

function toggleSSLFields() {
    const sslFields = document.getElementById('ssl-fields');
    const sslCheckbox = document.getElementById('ssl-checkbox');
    sslFields.style.display = sslCheckbox.checked ? 'block' : 'none';
}

async function deleteProxy(proxyID) {
    if (!confirm('确定要删除这个反向代理吗？')) return;
    
    const result = await apiRequest(`/api/network/proxy?id=${proxyID}`, {
        method: 'DELETE'
    });
    
    if (result && result.code === 200) {
        showAlert('反向代理删除成功', 'success');
        loadProxies();
    } else {
        showAlert(result.msg || '删除失败', 'error');
    }
}

// 辅助函数
function closeModal(modalId) {
    const modal = document.getElementById(modalId);
    if (modal) {
        modal.remove();
    }
}
