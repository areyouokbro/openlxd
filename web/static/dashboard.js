// API Key 管理
const API_KEY = localStorage.getItem('api_key');

if (!API_KEY) {
    window.location.href = '/admin/login';
}

// API 请求封装
async function apiRequest(url, options = {}) {
    const defaultOptions = {
        headers: {
            'X-API-Hash': API_KEY,
            'Content-Type': 'application/json'
        }
    };
    
    const response = await fetch(url, { ...defaultOptions, ...options });
    const data = await response.json();
    
    if (response.status === 401) {
        showAlert('登录已过期，请重新登录', 'error');
        setTimeout(() => {
            window.location.href = '/admin/login';
        }, 2000);
        return null;
    }
    
    return data;
}

// 显示提示信息
function showAlert(message, type = 'success') {
    const container = document.getElementById('alert-container');
    const alert = document.createElement('div');
    alert.className = `alert alert-${type}`;
    alert.textContent = message;
    container.appendChild(alert);
    
    setTimeout(() => {
        alert.remove();
    }, 5000);
}

// 标签页切换
function switchTab(tabName) {
    // 更新按钮状态
    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    
    // 更新内容显示
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(`tab-${tabName}`).classList.add('active');
    
    // 加载对应数据
    switch(tabName) {
        case 'containers':
            loadContainers();
            break;
        case 'ippool':
            loadIPPool();
            break;
        case 'portmapping':
            loadPortMappings();
            break;
        case 'proxy':
            loadProxies();
            break;
        case 'monitor':
            loadMonitorData();
            break;
        case 'snapshot':
            loadSnapshots();
            break;
        case 'clone':
            // 克隆页面不需要加载数据
            break;
        case 'dns':
            loadDNSConfig();
            break;
        case 'quota':
            loadQuotaList();
            break;
        case 'migration':
            initMigration();
            break;
        case 'images':
            loadImages();
            break;
        case 'storage':
            loadStorage();
            break;
        case 'networks':
            loadNetworks();
            break;
        case 'port-forwards':
            loadPortForwards();
            break;
    }
}

// ========== 容器管理 ==========

async function loadContainers() {
    const tbody = document.getElementById('containers-tbody');
    tbody.innerHTML = '<tr><td colspan="8" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/system/containers');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="8" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const containers = data.data || [];
    
    if (containers.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" class="empty-state">暂无容器</td></tr>';
        return;
    }
    
    tbody.innerHTML = containers.map(c => `
        <tr>
            <td><strong>${c.hostname || '-'}</strong></td>
            <td>${c.ip || '-'}</td>
            <td>${c.image || '-'}</td>
            <td>
                <span class="status-badge status-${c.status === 'Running' ? 'running' : 'stopped'}">
                    ${c.status === 'Running' ? '运行中' : '已停止'}
                </span>
            </td>
            <td>${c.cpu || '-'}</td>
            <td>${c.memory || '-'}</td>
            <td>${formatTraffic(c.traffic_up, c.traffic_down)}</td>
            <td>
                <div class="action-buttons">
                    ${c.status === 'Running' ? 
                        `<button class="btn btn-warning btn-sm" onclick="containerAction('${c.hostname}', 'stop')">停止</button>
                         <button class="btn btn-warning btn-sm" onclick="containerAction('${c.hostname}', 'restart')">重启</button>` :
                        `<button class="btn btn-success btn-sm" onclick="containerAction('${c.hostname}', 'start')">启动</button>`
                    }
                    <button class="btn btn-danger btn-sm" onclick="deleteContainer('${c.hostname}')">删除</button>
                </div>
            </td>
        </tr>
    `).join('');
    
    // 更新统计
    const total = containers.length;
    const running = containers.filter(c => c.status === 'Running').length;
    document.getElementById('total-containers').textContent = total;
    document.getElementById('running-containers').textContent = running;
    document.getElementById('stopped-containers').textContent = total - running;
    
    loadSystemStats();
}

async function loadSystemStats() {
    const data = await apiRequest('/api/system/stats');
    if (data && data.code === 200) {
        document.getElementById('system-load').textContent = data.data.load || '-';
    }
}

function formatTraffic(up, down) {
    const formatBytes = (bytes) => {
        if (!bytes) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
    };
    return `↑ ${formatBytes(up)} / ↓ ${formatBytes(down)}`;
}

async function containerAction(hostname, action) {
    const actionText = {start: '启动', stop: '停止', restart: '重启'}[action];
    if (!confirm(`确定要${actionText}容器 ${hostname} 吗？`)) return;
    
    const result = await apiRequest(`/api/system/containers/${hostname}/action`, {
        method: 'POST',
        body: JSON.stringify({ action })
    });
    
    if (result && result.code === 200) {
        showAlert(`容器${actionText}成功`, 'success');
        loadContainers();
    } else {
        showAlert(result?.msg || `${actionText}失败`, 'error');
    }
}

async function deleteContainer(hostname) {
    if (!confirm(`⚠️ 确定要删除容器 ${hostname} 吗？此操作不可恢复！`)) return;
    
    const result = await apiRequest(`/api/system/containers/${hostname}`, {
        method: 'DELETE'
    });
    
    if (result && result.code === 200) {
        showAlert('容器删除成功', 'success');
        loadContainers();
    } else {
        showAlert(result?.msg || '删除失败', 'error');
    }
}

// ========== 镜像管理 ==========

async function loadImages() {
    const tbody = document.getElementById('images-tbody');
    tbody.innerHTML = '<tr><td colspan="5" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/system/images');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="5" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const images = data.data || [];
    
    if (images.length === 0) {
        tbody.innerHTML = '<tr><td colspan="5" class="empty-state">暂无镜像</td></tr>';
        return;
    }
    
    tbody.innerHTML = images.map(img => `
        <tr>
            <td><strong>${img.alias || '-'}</strong></td>
            <td><code>${img.fingerprint?.substring(0, 12) || '-'}</code></td>
            <td>${img.size || '-'}</td>
            <td>${img.created || '-'}</td>
            <td>
                <button class="btn btn-danger btn-sm" onclick="deleteImage('${img.fingerprint}')">删除</button>
            </td>
        </tr>
    `).join('');
}

async function deleteImage(fingerprint) {
    if (!confirm('确定要删除此镜像吗？')) return;
    
    const result = await apiRequest('/api/system/images/delete', {
        method: 'POST',
        body: JSON.stringify({ fingerprint })
    });
    
    if (result && result.code === 200) {
        showAlert('镜像删除成功', 'success');
        loadImages();
    } else {
        showAlert(result?.msg || '删除失败', 'error');
    }
}

// ========== 存储池管理 ==========

async function loadStorage() {
    const tbody = document.getElementById('storage-tbody');
    tbody.innerHTML = '<tr><td colspan="6" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/system/storage');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const storage = data.data || [];
    
    if (storage.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state">暂无存储池</td></tr>';
        return;
    }
    
    tbody.innerHTML = storage.map(s => {
        const usage = s.used && s.size ? Math.round((parseFloat(s.used) / parseFloat(s.size)) * 100) : 0;
        return `
            <tr>
                <td><strong>${s.name}</strong></td>
                <td>${s.driver}</td>
                <td>${s.size || '-'}</td>
                <td>${s.used || '-'}</td>
                <td>${usage}%</td>
                <td>
                    <button class="btn btn-secondary btn-sm">详情</button>
                </td>
            </tr>
        `;
    }).join('');
}

// ========== 网络管理 ==========

async function loadNetworks() {
    const tbody = document.getElementById('networks-tbody');
    tbody.innerHTML = '<tr><td colspan="6" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/system/networks');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const networks = data.data || [];
    
    if (networks.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state">暂无网络</td></tr>';
        return;
    }
    
    tbody.innerHTML = networks.map(n => `
        <tr>
            <td><strong>${n.name}</strong></td>
            <td>${n.type}</td>
            <td>${n.ipv4 || '-'}</td>
            <td>${n.ipv6 || '-'}</td>
            <td>${n.managed ? '是' : '否'}</td>
            <td>
                <button class="btn btn-secondary btn-sm">详情</button>
            </td>
        </tr>
    `).join('');
}

// ========== 端口转发管理 ==========

async function loadPortForwards() {
    const tbody = document.getElementById('forwards-tbody');
    tbody.innerHTML = '<tr><td colspan="7" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/system/port-forwards');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const forwards = data.data || [];
    
    if (forwards.length === 0) {
        tbody.innerHTML = '<tr><td colspan="7" class="empty-state">暂无转发规则</td></tr>';
        return;
    }
    
    tbody.innerHTML = forwards.map(f => `
        <tr>
            <td><strong>${f.container}</strong></td>
            <td>${f.public_ip}</td>
            <td>${f.public_port}</td>
            <td>${f.private_port}</td>
            <td>${f.protocol.toUpperCase()}</td>
            <td>${f.interface}</td>
            <td>
                <button class="btn btn-danger btn-sm" onclick="deletePortForward(${f.id})">删除</button>
            </td>
        </tr>
    `).join('');
}

async function deletePortForward(id) {
    if (!confirm('确定要删除此转发规则吗？')) return;
    
    const result = await apiRequest('/api/system/port-forwards/delete', {
        method: 'POST',
        body: JSON.stringify({ id })
    });
    
    if (result && result.code === 200) {
        showAlert('转发规则删除成功', 'success');
        loadPortForwards();
    } else {
        showAlert(result?.msg || '删除失败', 'error');
    }
}

// ========== 模态框管理 ==========

function showModal(modalType) {
    const modalContainer = document.getElementById('modal-container');
    
    const modals = {
        'create-container': createContainerModal(),
        'download-image': downloadImageModal(),
        'create-storage': createStorageModal(),
        'create-network': createNetworkModal(),
        'create-forward': createPortForwardModal()
    };
    
    modalContainer.innerHTML = modals[modalType] || '';
    document.querySelector('.modal').classList.add('active');
}

function closeModal() {
    document.querySelectorAll('.modal').forEach(m => m.classList.remove('active'));
}

function createContainerModal() {
    return `
        <div class="modal active">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>创建新容器</h2>
                </div>
                <form onsubmit="submitCreateContainer(event)">
                    <div class="form-group">
                        <label>主机名 *</label>
                        <input type="text" name="hostname" required placeholder="例如: web-server-01">
                    </div>
                    <div class="form-group">
                        <label>镜像 *</label>
                        <select name="image" required>
                            <option value="">请选择镜像</option>
                            <option value="ubuntu:22.04">Ubuntu 22.04</option>
                            <option value="ubuntu:20.04">Ubuntu 20.04</option>
                            <option value="debian:11">Debian 11</option>
                            <option value="debian:12">Debian 12</option>
                            <option value="centos:7">CentOS 7</option>
                            <option value="rocky:8">Rocky Linux 8</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>CPU 核心数</label>
                        <input type="number" name="cpu" value="2" min="1" max="32">
                    </div>
                    <div class="form-group">
                        <label>内存 (GB)</label>
                        <input type="number" name="memory" value="2" min="1" max="128">
                    </div>
                    <div class="form-group">
                        <label>磁盘 (GB)</label>
                        <input type="number" name="disk" value="20" min="10" max="1000">
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                        <button type="submit" class="btn btn-primary">创建</button>
                    </div>
                </form>
            </div>
        </div>
    `;
}

function downloadImageModal() {
    return `
        <div class="modal active">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>下载镜像</h2>
                </div>
                <form onsubmit="submitDownloadImage(event)">
                    <div class="form-group">
                        <label>镜像服务器</label>
                        <select name="server">
                            <option value="https://images.linuxcontainers.org">images.linuxcontainers.org</option>
                            <option value="https://mirrors.tuna.tsinghua.edu.cn/lxc-images">清华大学镜像站</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>镜像别名 *</label>
                        <input type="text" name="alias" required placeholder="例如: ubuntu/22.04">
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                        <button type="submit" class="btn btn-primary">下载</button>
                    </div>
                </form>
            </div>
        </div>
    `;
}

function createStorageModal() {
    return `
        <div class="modal active">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>创建存储池</h2>
                </div>
                <form onsubmit="submitCreateStorage(event)">
                    <div class="form-group">
                        <label>名称 *</label>
                        <input type="text" name="name" required placeholder="例如: pool1">
                    </div>
                    <div class="form-group">
                        <label>驱动 *</label>
                        <select name="driver" required>
                            <option value="dir">dir (目录)</option>
                            <option value="zfs">ZFS</option>
                            <option value="btrfs">Btrfs</option>
                            <option value="lvm">LVM</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>大小</label>
                        <input type="text" name="size" placeholder="例如: 100GB">
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                        <button type="submit" class="btn btn-primary">创建</button>
                    </div>
                </form>
            </div>
        </div>
    `;
}

function createNetworkModal() {
    return `
        <div class="modal active">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>创建网络</h2>
                </div>
                <form onsubmit="submitCreateNetwork(event)">
                    <div class="form-group">
                        <label>名称 *</label>
                        <input type="text" name="name" required placeholder="例如: lxdbr1">
                    </div>
                    <div class="form-group">
                        <label>类型 *</label>
                        <select name="type" required>
                            <option value="bridge">Bridge (桥接)</option>
                            <option value="macvlan">Macvlan</option>
                            <option value="physical">Physical (物理)</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>IPv4 网段</label>
                        <input type="text" name="ipv4" placeholder="例如: 10.0.1.1/24">
                    </div>
                    <div class="form-group">
                        <label>IPv6 网段</label>
                        <input type="text" name="ipv6" placeholder="例如: fd42:1::/64">
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                        <button type="submit" class="btn btn-primary">创建</button>
                    </div>
                </form>
            </div>
        </div>
    `;
}

function createPortForwardModal() {
    return `
        <div class="modal active">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>添加端口转发规则</h2>
                </div>
                <form onsubmit="submitCreatePortForward(event)">
                    <div class="form-group">
                        <label>容器名称 *</label>
                        <input type="text" name="container" required placeholder="例如: web-01">
                    </div>
                    <div class="form-group">
                        <label>公网 IP/域名 *</label>
                        <input type="text" name="public_ip" required placeholder="例如: 156.246.90.151 或 example.com">
                    </div>
                    <div class="form-group">
                        <label>公网端口 *</label>
                        <input type="number" name="public_port" required min="1" max="65535" placeholder="例如: 8080">
                    </div>
                    <div class="form-group">
                        <label>容器端口 *</label>
                        <input type="number" name="private_port" required min="1" max="65535" placeholder="例如: 80">
                    </div>
                    <div class="form-group">
                        <label>协议 *</label>
                        <select name="protocol" required>
                            <option value="tcp">TCP</option>
                            <option value="udp">UDP</option>
                            <option value="both">TCP + UDP</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>绑定网卡 *</label>
                        <input type="text" name="interface" required value="eth0" placeholder="例如: eth0">
                        <small style="color: #999;">公网流量入口网卡</small>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="btn btn-secondary" onclick="closeModal()">取消</button>
                        <button type="submit" class="btn btn-primary">创建</button>
                    </div>
                </form>
            </div>
        </div>
    `;
}

// ========== 表单提交 ==========

async function submitCreateContainer(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        hostname: formData.get('hostname'),
        image: formData.get('image'),
        cpu: parseInt(formData.get('cpu')),
        memory: parseInt(formData.get('memory')) * 1024,
        disk: parseInt(formData.get('disk'))
    };
    
    const result = await apiRequest('/api/system/containers', {
        method: 'POST',
        body: JSON.stringify(data)
    });
    
    if (result && result.code === 200) {
        showAlert('容器创建成功！', 'success');
        closeModal();
        loadContainers();
    } else {
        showAlert(result?.msg || '创建失败', 'error');
    }
}

async function submitDownloadImage(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        alias: formData.get('alias'),
        server: formData.get('server')
    };
    
    const result = await apiRequest('/api/system/images/download', {
        method: 'POST',
        body: JSON.stringify(data)
    });
    
    if (result && result.code === 200) {
        showAlert('镜像下载已启动，请稍候...', 'success');
        closeModal();
        setTimeout(loadImages, 3000);
    } else {
        showAlert(result?.msg || '下载失败', 'error');
    }
}

async function submitCreateStorage(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        name: formData.get('name'),
        driver: formData.get('driver'),
        size: formData.get('size')
    };
    
    const result = await apiRequest('/api/system/storage/create', {
        method: 'POST',
        body: JSON.stringify(data)
    });
    
    if (result && result.code === 200) {
        showAlert('存储池创建成功！', 'success');
        closeModal();
        loadStorage();
    } else {
        showAlert(result?.msg || '创建失败', 'error');
    }
}

async function submitCreateNetwork(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        name: formData.get('name'),
        type: formData.get('type'),
        ipv4: formData.get('ipv4'),
        ipv6: formData.get('ipv6')
    };
    
    const result = await apiRequest('/api/system/networks/create', {
        method: 'POST',
        body: JSON.stringify(data)
    });
    
    if (result && result.code === 200) {
        showAlert('网络创建成功！', 'success');
        closeModal();
        loadNetworks();
    } else {
        showAlert(result?.msg || '创建失败', 'error');
    }
}

async function submitCreatePortForward(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = {
        container: formData.get('container'),
        public_ip: formData.get('public_ip'),
        public_port: parseInt(formData.get('public_port')),
        private_port: parseInt(formData.get('private_port')),
        protocol: formData.get('protocol'),
        interface: formData.get('interface')
    };
    
    const result = await apiRequest('/api/system/port-forwards/create', {
        method: 'POST',
        body: JSON.stringify(data)
    });
    
    if (result && result.code === 200) {
        showAlert('端口转发规则创建成功！', 'success');
        closeModal();
        loadPortForwards();
    } else {
        showAlert(result?.msg || '创建失败', 'error');
    }
}

// 退出登录
function logout() {
    localStorage.removeItem('api_key');
    window.location.href = '/admin/login';
}

// 页面加载时初始化
loadContainers();

// 自动刷新（每30秒）
setInterval(() => {
    const activeTab = document.querySelector('.tab-content.active').id.replace('tab-', '');
    switch(activeTab) {
        case 'containers':
            loadContainers();
            break;
        case 'images':
            loadImages();
            break;
        case 'storage':
            loadStorage();
            break;
        case 'networks':
            loadNetworks();
            break;
        case 'port-forwards':
            loadPortForwards();
            break;
    }
}, 30000);
