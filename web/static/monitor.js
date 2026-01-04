// ========== 实时监控 ==========

async function loadMonitorData() {
    // 加载当前系统指标
    loadCurrentSystemMetrics();
    
    // 加载容器资源统计
    loadResourceStats();
}

async function loadCurrentSystemMetrics() {
    const data = await apiRequest('/api/monitor/system/current');
    
    if (!data || data.code !== 200) {
        return;
    }
    
    const metric = data.data;
    
    // 更新系统指标卡片
    document.getElementById('monitor-cpu').textContent = metric.cpu_usage ? metric.cpu_usage.toFixed(2) + '%' : '-';
    document.getElementById('monitor-memory').textContent = metric.memory_usage ? metric.memory_usage.toFixed(2) + '%' : '-';
    document.getElementById('monitor-disk').textContent = metric.disk_usage ? metric.disk_usage.toFixed(2) + '%' : '-';
    document.getElementById('monitor-load').textContent = metric.load_average_1 ? metric.load_average_1.toFixed(2) : '-';
}

async function loadResourceStats() {
    const tbody = document.getElementById('monitor-tbody');
    tbody.innerHTML = '<tr><td colspan="6" class="loading">加载中...</td></tr>';
    
    const data = await apiRequest('/api/monitor/stats');
    
    if (!data || data.code !== 200) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state">加载失败</td></tr>';
        return;
    }
    
    const stats = data.data || [];
    
    if (stats.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state">暂无监控数据</td></tr>';
        return;
    }
    
    tbody.innerHTML = stats.map(s => `
        <tr>
            <td><strong>${s.container_name}</strong></td>
            <td>${s.cpu_usage_avg ? s.cpu_usage_avg.toFixed(2) : '0'}% / ${s.cpu_usage_max ? s.cpu_usage_max.toFixed(2) : '0'}%</td>
            <td>${s.memory_usage_avg ? s.memory_usage_avg.toFixed(2) : '0'}% / ${s.memory_usage_max ? s.memory_usage_max.toFixed(2) : '0'}%</td>
            <td>${s.disk_usage_avg ? s.disk_usage_avg.toFixed(2) : '0'}% / ${s.disk_usage_max ? s.disk_usage_max.toFixed(2) : '0'}%</td>
            <td>${formatBytes(s.network_rx_total || 0)} / ${formatBytes(s.network_tx_total || 0)}</td>
            <td>
                <span class="badge badge-info">IPv4: ${s.ipv4_count || 0}</span>
                <span class="badge badge-info">IPv6: ${s.ipv6_count || 0}</span>
                <span class="badge badge-success">端口: ${s.port_mapping_count || 0}</span>
                <span class="badge badge-primary">代理: ${s.proxy_count || 0}</span>
            </td>
        </tr>
    `).join('');
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 自动刷新监控数据（每30秒）
let monitorInterval = null;

function startMonitorAutoRefresh() {
    if (monitorInterval) {
        clearInterval(monitorInterval);
    }
    monitorInterval = setInterval(() => {
        const activeTab = document.querySelector('.tab-content.active');
        if (activeTab && activeTab.id === 'tab-monitor') {
            loadMonitorData();
        }
    }, 30000); // 30秒
}

function stopMonitorAutoRefresh() {
    if (monitorInterval) {
        clearInterval(monitorInterval);
        monitorInterval = null;
    }
}

// 页面加载时启动自动刷新
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', startMonitorAutoRefresh);
} else {
    startMonitorAutoRefresh();
}
