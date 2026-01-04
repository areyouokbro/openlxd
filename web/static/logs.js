// 日志管理
const LogsManager = {
    currentContainer: '',
    
    // 加载容器日志
    async loadContainerLogs(containerName = '') {
        try {
            let url = '/api/logs/container?limit=50';
            if (containerName) {
                url += `&container=${containerName}`;
            }
            
            const response = await fetch(url, {
                headers: { 'X-API-Hash': API_KEY }
            });
            
            if (!response.ok) throw new Error('获取日志失败');
            
            const data = await response.json();
            this.displayLogs(data.data.logs || [], 'container-logs-list');
        } catch (error) {
            console.error('加载容器日志失败:', error);
            showNotification('加载容器日志失败: ' + error.message, 'error');
        }
    },
    
    // 加载系统日志
    async loadSystemLogs(level = '') {
        try {
            let url = '/api/logs/system?limit=50';
            if (level) {
                url += `&level=${level}`;
            }
            
            const response = await fetch(url, {
                headers: { 'X-API-Hash': API_KEY }
            });
            
            if (!response.ok) throw new Error('获取系统日志失败');
            
            const data = await response.json();
            this.displayLogs(data.data.logs || [], 'system-logs-list');
        } catch (error) {
            console.error('加载系统日志失败:', error);
            showNotification('加载系统日志失败: ' + error.message, 'error');
        }
    },
    
    // 显示日志
    displayLogs(logs, containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;
        
        if (logs.length === 0) {
            container.innerHTML = '<div class="empty-state">暂无日志记录</div>';
            return;
        }
        
        const html = logs.map(log => `
            <div class="log-item log-${log.status || 'info'}">
                <div class="log-time">${new Date(log.created_at).toLocaleString('zh-CN')}</div>
                <div class="log-container">${log.container || '-'}</div>
                <div class="log-action">${log.action || log.operation || '-'}</div>
                <div class="log-message">${log.description || log.message || '-'}</div>
                <div class="log-status">
                    <span class="badge badge-${this.getStatusClass(log.status)}">${log.status || 'info'}</span>
                </div>
            </div>
        `).join('');
        
        container.innerHTML = html;
    },
    
    // 获取状态样式类
    getStatusClass(status) {
        const statusMap = {
            'success': 'success',
            'completed': 'success',
            'error': 'danger',
            'failed': 'danger',
            'warning': 'warning',
            'running': 'info',
            'pending': 'secondary'
        };
        return statusMap[status] || 'secondary';
    }
};

// 容器详情管理
const ContainerDetailManager = {
    currentContainer: null,
    
    // 显示容器详情
    async showDetail(containerName) {
        try {
            const response = await fetch(`/api/container/detail?name=${containerName}`, {
                headers: { 'X-API-Hash': API_KEY }
            });
            
            if (!response.ok) throw new Error('获取容器详情失败');
            
            const data = await response.json();
            this.currentContainer = data.data;
            this.displayDetail();
            
            // 切换到详情标签页
            switchTab('detail');
        } catch (error) {
            console.error('获取容器详情失败:', error);
            showNotification('获取容器详情失败: ' + error.message, 'error');
        }
    },
    
    // 显示详情
    displayDetail() {
        if (!this.currentContainer) return;
        
        const container = this.currentContainer.container;
        const quota = this.currentContainer.quota || {};
        const ipAddresses = this.currentContainer.ip_addresses || [];
        const portMappings = this.currentContainer.port_mappings || [];
        const proxyConfigs = this.currentContainer.proxy_configs || [];
        const recentLogs = this.currentContainer.recent_logs || [];
        
        // 基本信息
        document.getElementById('detail-name').textContent = container.hostname || '-';
        document.getElementById('detail-status').innerHTML = `<span class="badge badge-${container.status === 'Running' ? 'success' : 'secondary'}">${container.status || 'Unknown'}</span>`;
        document.getElementById('detail-image').textContent = container.image || '-';
        document.getElementById('detail-ipv4').textContent = container.ipv4 || '-';
        document.getElementById('detail-ipv6').textContent = container.ipv6 || '-';
        
        // 资源配置
        document.getElementById('detail-cpu').textContent = container.cpus || '-';
        document.getElementById('detail-memory').textContent = (container.memory || 0) + ' MB';
        document.getElementById('detail-disk').textContent = (container.disk || 0) + ' GB';
        
        // 配额信息
        document.getElementById('detail-ipv4-quota').textContent = quota.ipv4_quota === -1 ? '无限制' : quota.ipv4_quota;
        document.getElementById('detail-port-quota').textContent = quota.port_mapping_quota === -1 ? '无限制' : quota.port_mapping_quota;
        document.getElementById('detail-proxy-quota').textContent = quota.proxy_quota === -1 ? '无限制' : quota.proxy_quota;
        
        // IP 地址列表
        const ipList = document.getElementById('detail-ip-list');
        ipList.innerHTML = ipAddresses.length > 0 
            ? ipAddresses.map(ip => `<div class="detail-item">${ip.ip} (${ip.type})</div>`).join('')
            : '<div class="empty-state">暂无 IP 地址</div>';
        
        // 端口映射列表
        const portList = document.getElementById('detail-port-list');
        portList.innerHTML = portMappings.length > 0
            ? portMappings.map(pm => `<div class="detail-item">${pm.external_port} → ${pm.internal_port} (${pm.protocol})</div>`).join('')
            : '<div class="empty-state">暂无端口映射</div>';
        
        // 反向代理列表
        const proxyList = document.getElementById('detail-proxy-list');
        proxyList.innerHTML = proxyConfigs.length > 0
            ? proxyConfigs.map(pc => `<div class="detail-item">${pc.domain} → ${pc.target_ip}:${pc.target_port}</div>`).join('')
            : '<div class="empty-state">暂无反向代理</div>';
        
        // 最近操作
        const logsList = document.getElementById('detail-logs-list');
        logsList.innerHTML = recentLogs.length > 0
            ? recentLogs.map(log => `
                <div class="log-item">
                    <span class="log-time">${new Date(log.created_at).toLocaleString('zh-CN')}</span>
                    <span class="log-action">${log.action}</span>
                    <span class="badge badge-${log.status === 'success' ? 'success' : 'danger'}">${log.status}</span>
                </div>
            `).join('')
            : '<div class="empty-state">暂无操作记录</div>';
    }
};
