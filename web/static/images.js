// 镜像市场模块

let localImages = [];
let remoteImages = [];
let selectedDistribution = 'all';

// 初始化镜像市场
export function initImageMarket() {
    loadLocalImages();
    loadRemoteImages();
    setupEventListeners();
}

// 加载本地镜像
async function loadLocalImages() {
    const token = localStorage.getItem('auth_token');
    
    try {
        const response = await fetch('/api/v1/images/list', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to load local images');
        }

        const result = await response.json();
        localImages = result.data || [];
        renderLocalImages();
    } catch (error) {
        console.error('Error loading local images:', error);
        showNotification('加载本地镜像失败', 'error');
    }
}

// 加载远程镜像
async function loadRemoteImages() {
    const token = localStorage.getItem('auth_token');
    
    try {
        const response = await fetch('/api/v1/images/remote', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to load remote images');
        }

        const result = await response.json();
        remoteImages = result.data || [];
        renderRemoteImages();
        renderDistributionFilter();
    } catch (error) {
        console.error('Error loading remote images:', error);
        showNotification('加载远程镜像失败', 'error');
    }
}

// 渲染本地镜像
function renderLocalImages() {
    const container = document.getElementById('local-images-container');
    if (!container) return;

    if (localImages.length === 0) {
        container.innerHTML = '<div class="empty-state">暂无本地镜像</div>';
        return;
    }

    container.innerHTML = localImages.map(image => `
        <div class="image-card">
            <div class="image-header">
                <h4>${image.alias}</h4>
                <span class="badge badge-${getStatusColor(image.status)}">${getStatusText(image.status)}</span>
            </div>
            <div class="image-body">
                <p><strong>发行版：</strong>${image.distribution}</p>
                <p><strong>版本：</strong>${image.release}</p>
                <p><strong>架构：</strong>${image.architecture}</p>
                <p><strong>大小：</strong>${formatSize(image.size)}</p>
                ${image.imported_at ? `<p><strong>导入时间：</strong>${formatDate(image.imported_at)}</p>` : ''}
            </div>
            <div class="image-actions">
                <button onclick="createFromImage('${image.alias}')" class="btn-primary btn-sm" ${image.status !== 'imported' ? 'disabled' : ''}>
                    创建容器
                </button>
                <button onclick="deleteLocalImage('${image.alias}')" class="btn-danger btn-sm">
                    删除
                </button>
            </div>
        </div>
    `).join('');
}

// 渲染远程镜像
function renderRemoteImages() {
    const container = document.getElementById('remote-images-container');
    if (!container) return;

    // 过滤镜像
    let filtered = remoteImages;
    if (selectedDistribution !== 'all') {
        filtered = remoteImages.filter(img => img.distribution === selectedDistribution);
    }

    if (filtered.length === 0) {
        container.innerHTML = '<div class="empty-state">没有找到镜像</div>';
        return;
    }

    container.innerHTML = filtered.map(image => {
        const isImported = localImages.some(local => local.alias === image.alias);
        
        return `
            <div class="image-card">
                <div class="image-header">
                    <h4>${image.alias}</h4>
                    ${isImported ? '<span class="badge badge-success">已导入</span>' : ''}
                </div>
                <div class="image-body">
                    <p><strong>描述：</strong>${image.description}</p>
                    <p><strong>架构：</strong>${image.architecture}</p>
                </div>
                <div class="image-actions">
                    <button onclick="importImage('${image.alias}', '${image.architecture}')" 
                            class="btn-primary btn-sm" 
                            ${isImported ? 'disabled' : ''}>
                        ${isImported ? '已导入' : '导入镜像'}
                    </button>
                </div>
            </div>
        `;
    }).join('');
}

// 渲染发行版过滤器
function renderDistributionFilter() {
    const filterContainer = document.getElementById('distribution-filter');
    if (!filterContainer) return;

    // 获取所有发行版
    const distributions = ['all', ...new Set(remoteImages.map(img => img.distribution))];

    filterContainer.innerHTML = `
        <label>发行版：</label>
        <select id="distribution-select" onchange="filterByDistribution(this.value)">
            ${distributions.map(dist => `
                <option value="${dist}" ${dist === selectedDistribution ? 'selected' : ''}>
                    ${dist === 'all' ? '全部' : dist.charAt(0).toUpperCase() + dist.slice(1)}
                </option>
            `).join('')}
        </select>
    `;
}

// 按发行版过滤
function filterByDistribution(distribution) {
    selectedDistribution = distribution;
    renderRemoteImages();
}

// 导入镜像
async function importImage(alias, architecture) {
    const token = localStorage.getItem('auth_token');

    if (!confirm(`确定要导入镜像 ${alias} 吗？这可能需要几分钟时间。`)) {
        return;
    }

    try {
        const response = await fetch('/api/v1/images/import', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                alias: alias,
                architecture: architecture
            })
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to import image');
        }

        showNotification('镜像导入已开始，请稍候...', 'success');
        
        // 5秒后刷新本地镜像列表
        setTimeout(() => {
            loadLocalImages();
        }, 5000);
    } catch (error) {
        console.error('Error importing image:', error);
        showNotification(`导入镜像失败: ${error.message}`, 'error');
    }
}

// 删除本地镜像
async function deleteLocalImage(alias) {
    const token = localStorage.getItem('auth_token');

    if (!confirm(`确定要删除镜像 ${alias} 吗？`)) {
        return;
    }

    try {
        const response = await fetch(`/api/v1/images/${alias}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to delete image');
        }

        showNotification('镜像删除成功', 'success');
        loadLocalImages();
        loadRemoteImages();
    } catch (error) {
        console.error('Error deleting image:', error);
        showNotification('删除镜像失败', 'error');
    }
}

// 从镜像创建容器
function createFromImage(alias) {
    // 切换到容器管理标签并填充镜像字段
    const containersTab = document.querySelector('[data-tab="containers"]');
    if (containersTab) {
        containersTab.click();
    }

    // 填充镜像字段
    setTimeout(() => {
        const imageInput = document.getElementById('container-image');
        if (imageInput) {
            imageInput.value = alias;
        }
    }, 100);
}

// 同步镜像
async function syncImages() {
    const token = localStorage.getItem('auth_token');

    try {
        const response = await fetch('/api/v1/images/sync', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to sync images');
        }

        const result = await response.json();
        showNotification(`同步成功，新增 ${result.data.synced} 个镜像`, 'success');
        loadLocalImages();
    } catch (error) {
        console.error('Error syncing images:', error);
        showNotification('同步镜像失败', 'error');
    }
}

// 获取状态颜色
function getStatusColor(status) {
    const colorMap = {
        'available': 'secondary',
        'downloading': 'warning',
        'imported': 'success',
        'failed': 'danger'
    };
    return colorMap[status] || 'secondary';
}

// 获取状态文本
function getStatusText(status) {
    const textMap = {
        'available': '可用',
        'downloading': '下载中',
        'imported': '已导入',
        'failed': '失败'
    };
    return textMap[status] || status;
}

// 格式化大小
function formatSize(bytes) {
    if (!bytes) return 'N/A';
    
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024;
        unitIndex++;
    }

    return `${size.toFixed(2)} ${units[unitIndex]}`;
}

// 格式化日期
function formatDate(dateString) {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN');
}

// 显示通知
function showNotification(message, type = 'info') {
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    document.body.appendChild(notification);

    setTimeout(() => {
        notification.remove();
    }, 3000);
}

// 设置事件监听器
function setupEventListeners() {
    // 刷新按钮
    const refreshLocalBtn = document.getElementById('refresh-local-images');
    if (refreshLocalBtn) {
        refreshLocalBtn.addEventListener('click', loadLocalImages);
    }

    const refreshRemoteBtn = document.getElementById('refresh-remote-images');
    if (refreshRemoteBtn) {
        refreshRemoteBtn.addEventListener('click', loadRemoteImages);
    }

    // 同步按钮
    const syncBtn = document.getElementById('sync-images');
    if (syncBtn) {
        syncBtn.addEventListener('click', syncImages);
    }
}

// 导出函数供全局使用
window.filterByDistribution = filterByDistribution;
window.importImage = importImage;
window.deleteLocalImage = deleteLocalImage;
window.createFromImage = createFromImage;
window.syncImages = syncImages;
