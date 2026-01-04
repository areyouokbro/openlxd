// 主题管理
const ThemeManager = {
    // 当前主题
    currentTheme: localStorage.getItem('theme') || 'light',
    
    // 初始化主题
    init() {
        // 应用保存的主题
        this.applyTheme(this.currentTheme);
        
        // 创建主题切换按钮
        this.createThemeToggle();
    },
    
    // 应用主题
    applyTheme(theme) {
        document.documentElement.setAttribute('data-theme', theme);
        this.currentTheme = theme;
        localStorage.setItem('theme', theme);
        
        // 更新按钮图标
        const toggleBtn = document.getElementById('theme-toggle');
        if (toggleBtn) {
            toggleBtn.textContent = theme === 'dark' ? '☀️' : '🌙';
            toggleBtn.title = theme === 'dark' ? '切换到亮色主题' : '切换到暗色主题';
        }
    },
    
    // 切换主题
    toggle() {
        const newTheme = this.currentTheme === 'light' ? 'dark' : 'light';
        this.applyTheme(newTheme);
    },
    
    // 创建主题切换按钮
    createThemeToggle() {
        const userInfo = document.querySelector('.user-info');
        if (!userInfo) return;
        
        const toggleBtn = document.createElement('button');
        toggleBtn.id = 'theme-toggle';
        toggleBtn.className = 'btn btn-secondary';
        toggleBtn.textContent = this.currentTheme === 'dark' ? '☀️' : '🌙';
        toggleBtn.title = this.currentTheme === 'dark' ? '切换到亮色主题' : '切换到暗色主题';
        toggleBtn.onclick = () => this.toggle();
        
        // 插入到用户信息区域的第一个位置
        userInfo.insertBefore(toggleBtn, userInfo.firstChild);
    }
};

// 页面加载时初始化主题
document.addEventListener('DOMContentLoaded', () => {
    ThemeManager.init();
});
