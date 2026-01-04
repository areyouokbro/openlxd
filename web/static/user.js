// 用户管理模块

// 全局变量
let currentUser = null;
let users = [];

// 初始化用户管理
export function initUserManagement() {
    loadCurrentUser();
    loadUsers();
    setupEventListeners();
}

// 加载当前用户信息
async function loadCurrentUser() {
    const token = localStorage.getItem('auth_token');
    if (!token) {
        window.location.href = '/login.html';
        return;
    }

    try {
        const response = await fetch('/api/v1/users/profile', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to load user profile');
        }

        const result = await response.json();
        currentUser = result.data;
        
        // 更新UI
        updateUserInfo();
        
        // 如果不是管理员，隐藏管理员功能
        if (!currentUser || currentUser.role !== 'admin') {
            document.getElementById('users-tab')?.classList.add('hidden');
        }
    } catch (error) {
        console.error('Error loading user profile:', error);
        localStorage.removeItem('auth_token');
        window.location.href = '/login.html';
    }
}

// 更新用户信息显示
function updateUserInfo() {
    const userInfoElement = document.getElementById('user-info');
    if (userInfoElement && currentUser) {
        userInfoElement.innerHTML = `
            <div class="user-profile">
                <span class="user-name">${currentUser.username}</span>
                <span class="user-role">${currentUser.role === 'admin' ? '管理员' : '用户'}</span>
                <button onclick="logout()" class="btn-logout">退出</button>
            </div>
        `;
    }

    // 显示API密钥
    const apiKeyElement = document.getElementById('api-key-display');
    if (apiKeyElement && currentUser) {
        apiKeyElement.innerHTML = `
            <div class="api-key-section">
                <label>API 密钥：</label>
                <input type="text" value="${currentUser.api_key}" readonly class="api-key-input">
                <button onclick="regenerateAPIKey()" class="btn-regenerate">重新生成</button>
            </div>
        `;
    }
}

// 加载用户列表（管理员）
async function loadUsers() {
    if (!currentUser || currentUser.role !== 'admin') {
        return;
    }

    const token = localStorage.getItem('auth_token');
    try {
        const response = await fetch('/api/v1/users/list', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to load users');
        }

        const result = await response.json();
        users = result.data;
        renderUsersTable();
    } catch (error) {
        console.error('Error loading users:', error);
        showNotification('加载用户列表失败', 'error');
    }
}

// 渲染用户列表表格
function renderUsersTable() {
    const tbody = document.getElementById('users-table-body');
    if (!tbody) return;

    if (users.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="empty-state">暂无用户</td></tr>';
        return;
    }

    tbody.innerHTML = users.map(user => `
        <tr>
            <td>${user.id}</td>
            <td>${user.username}</td>
            <td>${user.email}</td>
            <td><span class="badge badge-${user.role === 'admin' ? 'primary' : 'secondary'}">${user.role === 'admin' ? '管理员' : '用户'}</span></td>
            <td><span class="badge badge-${user.status === 'active' ? 'success' : 'warning'}">${getStatusText(user.status)}</span></td>
            <td>
                <button onclick="editUser(${user.id})" class="btn-icon" title="编辑">✏️</button>
                <button onclick="toggleUserStatus(${user.id}, '${user.status}')" class="btn-icon" title="${user.status === 'active' ? '暂停' : '激活'}">
                    ${user.status === 'active' ? '⏸️' : '▶️'}
                </button>
            </td>
        </tr>
    `).join('');
}

// 获取状态文本
function getStatusText(status) {
    const statusMap = {
        'active': '活跃',
        'suspended': '暂停',
        'deleted': '已删除'
    };
    return statusMap[status] || status;
}

// 切换用户状态
async function toggleUserStatus(userId, currentStatus) {
    const newStatus = currentStatus === 'active' ? 'suspended' : 'active';
    const token = localStorage.getItem('auth_token');

    try {
        const response = await fetch('/api/v1/users/status', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: userId,
                status: newStatus
            })
        });

        if (!response.ok) {
            throw new Error('Failed to update user status');
        }

        showNotification('用户状态更新成功', 'success');
        loadUsers();
    } catch (error) {
        console.error('Error updating user status:', error);
        showNotification('更新用户状态失败', 'error');
    }
}

// 编辑用户
function editUser(userId) {
    const user = users.find(u => u.id === userId);
    if (!user) return;

    // 显示编辑对话框
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <h3>编辑用户</h3>
            <form id="edit-user-form">
                <div class="form-group">
                    <label>用户名：</label>
                    <input type="text" value="${user.username}" readonly>
                </div>
                <div class="form-group">
                    <label>邮箱：</label>
                    <input type="email" value="${user.email}" readonly>
                </div>
                <div class="form-group">
                    <label>角色：</label>
                    <select id="user-role" name="role">
                        <option value="user" ${user.role === 'user' ? 'selected' : ''}>用户</option>
                        <option value="admin" ${user.role === 'admin' ? 'selected' : ''}>管理员</option>
                    </select>
                </div>
                <div class="form-actions">
                    <button type="submit" class="btn-primary">保存</button>
                    <button type="button" onclick="closeModal()" class="btn-secondary">取消</button>
                </div>
            </form>
        </div>
    `;

    document.body.appendChild(modal);

    // 绑定表单提交事件
    document.getElementById('edit-user-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const role = document.getElementById('user-role').value;
        await updateUserRole(userId, role);
        closeModal();
    });
}

// 更新用户角色
async function updateUserRole(userId, role) {
    const token = localStorage.getItem('auth_token');

    try {
        const response = await fetch('/api/v1/users/role', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                user_id: userId,
                role: role
            })
        });

        if (!response.ok) {
            throw new Error('Failed to update user role');
        }

        showNotification('用户角色更新成功', 'success');
        loadUsers();
    } catch (error) {
        console.error('Error updating user role:', error);
        showNotification('更新用户角色失败', 'error');
    }
}

// 重新生成API密钥
async function regenerateAPIKey() {
    if (!confirm('确定要重新生成API密钥吗？旧密钥将失效。')) {
        return;
    }

    const token = localStorage.getItem('auth_token');

    try {
        const response = await fetch('/api/v1/users/regenerate-key', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error('Failed to regenerate API key');
        }

        const result = await response.json();
        currentUser = result.data;
        updateUserInfo();
        showNotification('API密钥重新生成成功', 'success');
    } catch (error) {
        console.error('Error regenerating API key:', error);
        showNotification('重新生成API密钥失败', 'error');
    }
}

// 退出登录
function logout() {
    localStorage.removeItem('auth_token');
    window.location.href = '/login.html';
}

// 关闭模态框
function closeModal() {
    const modal = document.querySelector('.modal');
    if (modal) {
        modal.remove();
    }
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
    // 可以添加其他事件监听器
}

// 导出函数供全局使用
window.toggleUserStatus = toggleUserStatus;
window.editUser = editUser;
window.regenerateAPIKey = regenerateAPIKey;
window.logout = logout;
window.closeModal = closeModal;
