// ========== 实时监控 ==========

// Chart.js 图表实例
let cpuChart = null;
let memoryChart = null;
let diskChart = null;
let networkChart = null;

async function loadMonitorData() {
    // 加载当前系统指标
    await loadCurrentSystemMetrics();
    
    // 加载历史数据并绘制图表
    await loadHistoricalData();
    
    // 加载容器资源统计
    await loadResourceStats();
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

async function loadHistoricalData() {
    // 获取最近1小时的数据
    const data = await apiRequest('/api/monitor/system?hours=1');
    
    if (!data || data.code !== 200 || !data.data || data.data.length === 0) {
        console.log('没有历史监控数据');
        return;
    }
    
    const metrics = data.data;
    
    // 提取时间标签和数据
    const labels = metrics.map(m => {
        const date = new Date(m.created_at);
        return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
    });
    
    const cpuData = metrics.map(m => m.cpu_usage);
    const memoryData = metrics.map(m => m.memory_usage);
    const diskData = metrics.map(m => m.disk_usage);
    const networkRxData = metrics.map(m => (m.network_rx_rate || 0) / 1024 / 1024); // 转换为 MB/s
    const networkTxData = metrics.map(m => (m.network_tx_rate || 0) / 1024 / 1024); // 转换为 MB/s
    
    // 绘制图表
    drawCPUChart(labels, cpuData);
    drawMemoryChart(labels, memoryData);
    drawDiskChart(labels, diskData);
    drawNetworkChart(labels, networkRxData, networkTxData);
}

function drawCPUChart(labels, data) {
    const ctx = document.getElementById('cpu-chart');
    if (!ctx) return;
    
    if (cpuChart) {
        cpuChart.destroy();
    }
    
    cpuChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'CPU 使用率 (%)',
                data: data,
                borderColor: 'rgb(75, 192, 192)',
                backgroundColor: 'rgba(75, 192, 192, 0.1)',
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                title: {
                    display: true,
                    text: 'CPU 使用率'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        callback: function(value) {
                            return value + '%';
                        }
                    }
                }
            }
        }
    });
}

function drawMemoryChart(labels, data) {
    const ctx = document.getElementById('memory-chart');
    if (!ctx) return;
    
    if (memoryChart) {
        memoryChart.destroy();
    }
    
    memoryChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: '内存使用率 (%)',
                data: data,
                borderColor: 'rgb(255, 99, 132)',
                backgroundColor: 'rgba(255, 99, 132, 0.1)',
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                title: {
                    display: true,
                    text: '内存使用率'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        callback: function(value) {
                            return value + '%';
                        }
                    }
                }
            }
        }
    });
}

function drawDiskChart(labels, data) {
    const ctx = document.getElementById('disk-chart');
    if (!ctx) return;
    
    if (diskChart) {
        diskChart.destroy();
    }
    
    diskChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: '磁盘使用率 (%)',
                data: data,
                borderColor: 'rgb(255, 205, 86)',
                backgroundColor: 'rgba(255, 205, 86, 0.1)',
                tension: 0.4,
                fill: true
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                title: {
                    display: true,
                    text: '磁盘使用率'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    max: 100,
                    ticks: {
                        callback: function(value) {
                            return value + '%';
                        }
                    }
                }
            }
        }
    });
}

function drawNetworkChart(labels, rxData, txData) {
    const ctx = document.getElementById('network-chart');
    if (!ctx) return;
    
    if (networkChart) {
        networkChart.destroy();
    }
    
    networkChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: '接收速率 (MB/s)',
                    data: rxData,
                    borderColor: 'rgb(54, 162, 235)',
                    backgroundColor: 'rgba(54, 162, 235, 0.1)',
                    tension: 0.4,
                    fill: true
                },
                {
                    label: '发送速率 (MB/s)',
                    data: txData,
                    borderColor: 'rgb(153, 102, 255)',
                    backgroundColor: 'rgba(153, 102, 255, 0.1)',
                    tension: 0.4,
                    fill: true
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                title: {
                    display: true,
                    text: '网络速率'
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return value.toFixed(2) + ' MB/s';
                        }
                    }
                }
            }
        }
    });
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
