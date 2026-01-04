// 终端控制台管理
const TerminalManager = {
    currentContainer: '',
    commandHistory: [],
    historyIndex: -1,
    
    init() {
        this.loadContainers();
        this.setupEventListeners();
    },
    
    async loadContainers() {
        try {
            const data = await apiRequest('/api/system/containers');
            const select = document.getElementById('terminal-container-select');
            select.innerHTML = '<option value="">请选择容器</option>';
            
            if (data && data.data) {
                data.data.forEach(container => {
                    const option = document.createElement('option');
                    option.value = container.name;
                    option.textContent = `${container.name} (${container.status})`;
                    select.appendChild(option);
                });
            }
        } catch (error) {
            console.error('加载容器列表失败:', error);
        }
    },
    
    setupEventListeners() {
        const input = document.getElementById('terminal-input');
        if (input) {
            input.addEventListener('keydown', (e) => {
                if (e.key === 'Enter') {
                    this.executeCommand();
                } else if (e.key === 'ArrowUp') {
                    e.preventDefault();
                    this.navigateHistory(-1);
                } else if (e.key === 'ArrowDown') {
                    e.preventDefault();
                    this.navigateHistory(1);
                }
            });
        }
    },
    
    selectContainer(containerName) {
        this.currentContainer = containerName;
        const output = document.getElementById('terminal-output');
        if (output) {
            output.innerHTML = '';
            this.addOutput(`已连接到容器: ${containerName}`, 'info');
            this.addOutput('提示: 输入命令后按 Enter 执行，使用 ↑↓ 键浏览历史命令', 'info');
        }
        document.getElementById('terminal-input').focus();
    },
    
    async executeCommand() {
        const input = document.getElementById('terminal-input');
        const command = input.value.trim();
        
        if (!command) return;
        
        if (!this.currentContainer) {
            showAlert('请先选择容器', 'error');
            return;
        }
        
        // 添加到历史记录
        this.commandHistory.push(command);
        this.historyIndex = this.commandHistory.length;
        
        // 显示命令
        this.addOutput(`$ ${command}`, 'command');
        
        // 清空输入框
        input.value = '';
        
        try {
            const response = await apiRequest('/api/advanced/exec', {
                method: 'POST',
                body: JSON.stringify({
                    container: this.currentContainer,
                    command: command
                })
            });
            
            if (response && response.code === 200) {
                const output = response.data.output || '';
                if (output) {
                    this.addOutput(output, 'output');
                } else {
                    this.addOutput('(命令执行成功，无输出)', 'success');
                }
            } else {
                this.addOutput(`错误: ${response.message || '命令执行失败'}`, 'error');
            }
        } catch (error) {
            this.addOutput(`错误: ${error.message}`, 'error');
        }
    },
    
    addOutput(text, type = 'output') {
        const output = document.getElementById('terminal-output');
        const line = document.createElement('div');
        line.className = `terminal-line terminal-${type}`;
        line.textContent = text;
        output.appendChild(line);
        
        // 自动滚动到底部
        output.scrollTop = output.scrollHeight;
    },
    
    navigateHistory(direction) {
        if (this.commandHistory.length === 0) return;
        
        this.historyIndex += direction;
        
        if (this.historyIndex < 0) {
            this.historyIndex = 0;
        } else if (this.historyIndex >= this.commandHistory.length) {
            this.historyIndex = this.commandHistory.length;
            document.getElementById('terminal-input').value = '';
            return;
        }
        
        document.getElementById('terminal-input').value = this.commandHistory[this.historyIndex];
    },
    
    clearTerminal() {
        document.getElementById('terminal-output').innerHTML = '';
        if (this.currentContainer) {
            this.addOutput(`已连接到容器: ${this.currentContainer}`, 'info');
        }
    },
    
    async quickCommand(cmd) {
        const input = document.getElementById('terminal-input');
        input.value = cmd;
        await this.executeCommand();
    }
};
