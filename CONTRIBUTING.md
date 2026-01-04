# 贡献指南

感谢您考虑为 OpenLXD 做出贡献！

## 🤝 如何贡献

### 报告 Bug

如果您发现了 Bug，请在 [Issues](https://github.com/areyouokbro/openlxd/issues) 页面创建一个新的 Issue，并包含以下信息：

- **清晰的标题**：简洁描述问题
- **详细描述**：说明问题的具体表现
- **复现步骤**：列出重现问题的步骤
- **期望行为**：描述您期望的正确行为
- **实际行为**：描述实际发生的情况
- **环境信息**：
  - 操作系统版本
  - Go 版本
  - LXD 版本
  - OpenLXD 版本
- **日志输出**：相关的错误日志

### 提出新功能

如果您有新功能的想法，请先在 [Discussions](https://github.com/areyouokbro/openlxd/discussions) 中讨论，或者创建一个 Feature Request Issue。

### 提交代码

1. **Fork 仓库**
   ```bash
   # 在 GitHub 上点击 Fork 按钮
   ```

2. **克隆您的 Fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/openlxd.git
   cd openlxd
   ```

3. **创建特性分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

4. **进行修改**
   - 编写代码
   - 添加测试
   - 更新文档

5. **提交更改**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

6. **推送到您的 Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **创建 Pull Request**
   - 在 GitHub 上打开您的 Fork
   - 点击 "New Pull Request"
   - 填写 PR 描述

## 📝 代码规范

### Go 代码风格

- 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 指南
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码质量
- 添加必要的注释

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型：**
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建/工具相关

**示例：**
```
feat(api): add container restart endpoint

Add a new API endpoint to restart containers with optional timeout parameter.

Closes #123
```

## 🧪 测试

在提交 PR 之前，请确保：

1. **代码可以编译**
   ```bash
   go build -o openlxd cmd/main.go
   ```

2. **通过所有测试**
   ```bash
   go test ./...
   ```

3. **代码格式正确**
   ```bash
   gofmt -s -w .
   ```

4. **无 lint 警告**
   ```bash
   golangci-lint run
   ```

## 📚 文档

如果您的更改影响到用户使用方式，请同时更新相关文档：

- `README.md` - 主要功能和快速开始
- `INSTALL.md` - 安装和部署指南
- `docs/api_reference.md` - API 接口文档
- `docs/plugin_integration.md` - 插件集成指南

## 🔍 代码审查

所有的 Pull Request 都需要经过代码审查。审查者可能会：

- 提出修改建议
- 要求添加测试
- 要求更新文档
- 讨论实现方案

请耐心等待审查，并及时回应审查意见。

## 📋 Pull Request 检查清单

提交 PR 前，请确认：

- [ ] 代码遵循项目的代码规范
- [ ] 添加了必要的测试
- [ ] 所有测试都通过
- [ ] 更新了相关文档
- [ ] 提交信息符合规范
- [ ] PR 描述清晰完整

## 🎯 开发环境设置

### 必需工具

- Go 1.18+
- Git
- Make (可选)

### 推荐工具

- [golangci-lint](https://golangci-lint.run/) - 代码检查
- [air](https://github.com/cosmtrek/air) - 热重载
- VS Code + Go 扩展

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/areyouokbro/openlxd.git
cd openlxd

# 安装依赖
go mod download

# 运行开发服务器
go run cmd/main.go

# 或使用 air 热重载
air
```

## 💬 社区

- **GitHub Discussions**: https://github.com/areyouokbro/openlxd/discussions
- **Issues**: https://github.com/areyouokbro/openlxd/issues

## 📜 行为准则

请遵守我们的行为准则，尊重所有贡献者。我们致力于提供一个友好、安全和欢迎的环境。

## 📄 许可证

通过贡献代码，您同意您的贡献将在 MIT 许可证下发布。

---

再次感谢您的贡献！🎉
