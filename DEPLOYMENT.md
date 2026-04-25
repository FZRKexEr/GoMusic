# GoMusic 部署文档

## 部署步骤

当前服务不依赖 MySQL、Redis 或其他持久化存储。

### 1. 构建后端项目

在项目根目录执行：

```bash
go build
```

### 2. 启动前端项目

进入前端目录并安装依赖：

```bash
cd static
yarn install
```

**本地开发**：

```bash
yarn serve
```

**生产部署**：

```bash
yarn build
```

### 3. 配置后端请求地址

编辑 `static/src/App.vue`，根据环境修改后端 URL：

```diff
    // 本地开发：使用 http://127.0.0.1:8081
+   const resp = await axios.post('http://127.0.0.1:8081/songlist' + queryParams, params, {
    // 生产部署：替换为你的域名
-   const resp = await axios.post('https://your-domain.com/songlist' + queryParams, params, {
```

- **本地开发**：使用 `http://127.0.0.1:8081`
- **生产部署**：将 URL 改为你的域名，如 `https://your-domain.com`

### 4. 访问应用

**本地开发**：

- 后端：运行编译后的二进制文件 `./GoMusic`
- 前端：访问 `http://localhost:8080`

**生产部署**：

- 将构建后的前端文件（`static/dist`）部署到 Web 服务器
- 配置反向代理将后端 API 请求转发到 Go 服务
- 通过域名访问应用
