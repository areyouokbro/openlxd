// 迁移管理相关函数

// 加载迁移任务列表
function loadMigrationTasks() {
    fetch('/api/migration/tasks', {
        headers: {
            'X-API-Hash': localStorage.getItem('api_key')
        }
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            displayMigrationTasks(result.data || []);
        } else {
            showMessage('加载迁移任务失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('加载迁移任务失败', 'error');
    });
}

// 显示迁移任务列表
function displayMigrationTasks(tasks) {
    const tbody = document.getElementById('migration-tasks-tbody');
    if (!tbody) return;
    
    if (tasks.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" style="text-align: center;">暂无迁移任务</td></tr>';
        return;
    }
    
    tbody.innerHTML = tasks.map(task => `
        <tr>
            <td>${task.id}</td>
            <td>${task.container_name}</td>
            <td>${task.source_host}</td>
            <td>${task.target_host}</td>
            <td>${task.migration_type === 'cold' ? '离线迁移' : '在线迁移'}</td>
            <td><span class="badge badge-${getStatusClass(task.status)}">${getStatusText(task.status)}</span></td>
            <td>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: ${task.progress}%">${task.progress}%</div>
                </div>
            </td>
            <td>
                <button class="btn btn-sm" onclick="viewMigrationLogs(${task.id})">日志</button>
                ${task.status === 'running' || task.status === 'pending' ? 
                    `<button class="btn btn-sm btn-warning" onclick="cancelMigration(${task.id})">取消</button>` : ''}
                ${task.status === 'completed' ? 
                    `<button class="btn btn-sm btn-danger" onclick="rollbackMigration(${task.id})">回滚</button>` : ''}
            </td>
        </tr>
    `).join('');
}

// 获取状态样式类
function getStatusClass(status) {
    const statusMap = {
        'pending': 'info',
        'running': 'warning',
        'completed': 'success',
        'failed': 'danger',
        'cancelled': 'secondary',
        'rollback': 'danger'
    };
    return statusMap[status] || 'secondary';
}

// 获取状态文本
function getStatusText(status) {
    const statusMap = {
        'pending': '等待中',
        'running': '进行中',
        'completed': '已完成',
        'failed': '失败',
        'cancelled': '已取消',
        'rollback': '已回滚'
    };
    return statusMap[status] || status;
}

// 创建迁移任务
function createMigrationTask() {
    const containerName = document.getElementById('migration-container').value;
    const targetHost = document.getElementById('migration-target-host').value;
    const migrationType = document.getElementById('migration-type').value;
    
    if (!containerName || !targetHost) {
        showMessage('请填写完整信息', 'error');
        return;
    }
    
    fetch('/api/migration/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': localStorage.getItem('api_key')
        },
        body: JSON.stringify({
            container_name: containerName,
            target_host: targetHost,
            migration_type: migrationType
        })
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            showMessage('迁移任务已创建', 'success');
            closeMigrationModal();
            loadMigrationTasks();
            // 每5秒刷新一次任务列表
            setTimeout(() => {
                const interval = setInterval(() => {
                    loadMigrationTasks();
                    // 检查是否还有进行中的任务
                    fetch('/api/migration/tasks', {
                        headers: { 'X-API-Hash': localStorage.getItem('api_key') }
                    })
                    .then(r => r.json())
                    .then(res => {
                        if (res.code === 200) {
                            const running = (res.data || []).some(t => 
                                t.status === 'running' || t.status === 'pending'
                            );
                            if (!running) {
                                clearInterval(interval);
                            }
                        }
                    });
                }, 5000);
            }, 1000);
        } else {
            showMessage('创建迁移任务失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('创建迁移任务失败', 'error');
    });
}

// 查看迁移日志
function viewMigrationLogs(taskId) {
    fetch(`/api/migration/logs?task_id=${taskId}`, {
        headers: {
            'X-API-Hash': localStorage.getItem('api_key')
        }
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            displayMigrationLogs(result.data || []);
        } else {
            showMessage('加载日志失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('加载日志失败', 'error');
    });
}

// 显示迁移日志
function displayMigrationLogs(logs) {
    const modal = document.getElementById('migration-logs-modal');
    const tbody = document.getElementById('migration-logs-tbody');
    
    if (!modal || !tbody) return;
    
    if (logs.length === 0) {
        tbody.innerHTML = '<tr><td colspan="3" style="text-align: center;">暂无日志</td></tr>';
    } else {
        tbody.innerHTML = logs.map(log => `
            <tr>
                <td>${new Date(log.created_at).toLocaleString()}</td>
                <td><span class="badge badge-${getLogLevelClass(log.level)}">${log.level}</span></td>
                <td>${log.message}</td>
            </tr>
        `).join('');
    }
    
    modal.style.display = 'block';
}

// 获取日志级别样式类
function getLogLevelClass(level) {
    const levelMap = {
        'info': 'info',
        'warning': 'warning',
        'error': 'danger'
    };
    return levelMap[level] || 'secondary';
}

// 取消迁移
function cancelMigration(taskId) {
    if (!confirm('确定要取消这个迁移任务吗？')) {
        return;
    }
    
    fetch('/api/migration/cancel', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': localStorage.getItem('api_key')
        },
        body: JSON.stringify({ task_id: taskId })
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            showMessage('任务已取消', 'success');
            loadMigrationTasks();
        } else {
            showMessage('取消失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('取消失败', 'error');
    });
}

// 回滚迁移
function rollbackMigration(taskId) {
    if (!confirm('确定要回滚这个迁移吗？这将在源主机上重新创建容器。')) {
        return;
    }
    
    fetch('/api/migration/rollback', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': localStorage.getItem('api_key')
        },
        body: JSON.stringify({ task_id: taskId })
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            showMessage('迁移已回滚', 'success');
            loadMigrationTasks();
        } else {
            showMessage('回滚失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('回滚失败', 'error');
    });
}

// 加载远程主机列表
function loadRemoteHosts() {
    fetch('/api/migration/hosts', {
        headers: {
            'X-API-Hash': localStorage.getItem('api_key')
        }
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            displayRemoteHosts(result.data || []);
            updateHostSelectors(result.data || []);
        } else {
            showMessage('加载远程主机失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('加载远程主机失败', 'error');
    });
}

// 显示远程主机列表
function displayRemoteHosts(hosts) {
    const tbody = document.getElementById('remote-hosts-tbody');
    if (!tbody) return;
    
    if (hosts.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" style="text-align: center;">暂无远程主机配置</td></tr>';
        return;
    }
    
    tbody.innerHTML = hosts.map(host => `
        <tr>
            <td>${host.name}</td>
            <td>${host.address}:${host.port}</td>
            <td>${host.protocol}</td>
            <td><span class="badge badge-${host.status === 'active' ? 'success' : 'secondary'}">${host.status === 'active' ? '活跃' : '非活跃'}</span></td>
            <td>${host.description || '-'}</td>
            <td>
                <button class="btn btn-sm btn-danger" onclick="deleteRemoteHost(${host.id})">删除</button>
            </td>
        </tr>
    `).join('');
}

// 更新主机选择器
function updateHostSelectors(hosts) {
    const selector = document.getElementById('migration-target-host');
    if (!selector) return;
    
    selector.innerHTML = '<option value="">请选择目标主机</option>' + 
        hosts.map(host => `<option value="${host.name}">${host.name} (${host.address})</option>`).join('');
}

// 添加远程主机
function addRemoteHost() {
    const name = document.getElementById('host-name').value;
    const address = document.getElementById('host-address').value;
    const port = document.getElementById('host-port').value || 8443;
    const description = document.getElementById('host-description').value;
    
    if (!name || !address) {
        showMessage('请填写主机名称和地址', 'error');
        return;
    }
    
    fetch('/api/migration/host/create', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': localStorage.getItem('api_key')
        },
        body: JSON.stringify({
            name: name,
            address: address,
            port: parseInt(port),
            protocol: 'https',
            description: description
        })
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            showMessage('远程主机已添加', 'success');
            closeHostModal();
            loadRemoteHosts();
        } else {
            showMessage('添加远程主机失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('添加远程主机失败', 'error');
    });
}

// 删除远程主机
function deleteRemoteHost(hostId) {
    if (!confirm('确定要删除这个远程主机配置吗？')) {
        return;
    }
    
    fetch('/api/migration/host/delete', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'X-API-Hash': localStorage.getItem('api_key')
        },
        body: JSON.stringify({ id: hostId })
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            showMessage('远程主机已删除', 'success');
            loadRemoteHosts();
        } else {
            showMessage('删除失败: ' + result.msg, 'error');
        }
    })
    .catch(error => {
        console.error('Error:', error);
        showMessage('删除失败', 'error');
    });
}

// 打开创建迁移任务模态框
function openMigrationModal() {
    const modal = document.getElementById('create-migration-modal');
    if (modal) {
        modal.style.display = 'block';
        // 加载容器列表
        loadContainersForMigration();
    }
}

// 关闭创建迁移任务模态框
function closeMigrationModal() {
    const modal = document.getElementById('create-migration-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

// 打开添加远程主机模态框
function openHostModal() {
    const modal = document.getElementById('add-host-modal');
    if (modal) {
        modal.style.display = 'none';
        document.getElementById('host-form').reset();
    }
}

// 关闭添加远程主机模态框
function closeHostModal() {
    const modal = document.getElementById('add-host-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

// 关闭日志模态框
function closeLogsModal() {
    const modal = document.getElementById('migration-logs-modal');
    if (modal) {
        modal.style.display = 'none';
    }
}

// 加载容器列表用于迁移
function loadContainersForMigration() {
    fetch('/api/system/containers', {
        headers: {
            'X-API-Hash': localStorage.getItem('api_key')
        }
    })
    .then(response => response.json())
    .then(result => {
        if (result.code === 200) {
            const selector = document.getElementById('migration-container');
            if (selector) {
                selector.innerHTML = '<option value="">请选择容器</option>' + 
                    (result.data || []).map(c => 
                        `<option value="${c.name}">${c.name} (${c.status})</option>`
                    ).join('');
            }
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
}

// 初始化迁移管理页面
function initMigration() {
    loadMigrationTasks();
    loadRemoteHosts();
    
    // 每30秒自动刷新任务列表
    setInterval(() => {
        if (document.getElementById('tab-migration').style.display !== 'none') {
            loadMigrationTasks();
        }
    }, 30000);
}
