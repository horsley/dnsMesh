# 快速使用指南

## 🎯 5 分钟上手

### Step 1: 启动系统

```bash
# 方式 1: 使用启动脚本（推荐）
./start.sh

# 方式 2: 使用 Docker Compose
cd frontend && npm install && npm run build
cd ..
docker-compose up -d
```

### Step 2: 登录系统

1. 打开浏览器访问: http://localhost:8080
2. 使用默认账户登录:
   - 用户名: `admin`
   - 密码: `admin123`

⚠️ **重要**: 首次登录后请立即修改密码！

### Step 3: 添加 DNS Provider

#### 如果你使用 Cloudflare

1. 点击 **"+ 添加 Provider"**
2. 选择 **Cloudflare**
3. 填写信息:
   - **API Key**: 你的 Global API Key
     - 获取方式: Cloudflare Dashboard → My Profile → API Tokens → Global API Key
   - **Email**: 你的 Cloudflare 账户邮箱

#### 如果你使用腾讯云 DNSPod

1. 点击 **"+ 添加 Provider"**
2. 选择 **腾讯云 DNSPod**
3. 填写信息:
   - **Secret ID**: AKID...
   - **Secret Key**: 你的密钥
     - 获取方式: 腾讯云控制台 → 访问管理 → API密钥管理 → 新建密钥

4. 点击 **"连接并同步"**

### Step 4: 智能导入记录

系统会自动分析你的 DNS 记录并识别服务器：

```
💡 发现 2 个可能的服务器

⭐ hk-01.mydomain.com → 1.2.3.4
   🎯 高置信度：域名格式匹配 + 5 个 CNAME 引用

   关联域名:
   ├─ app1.example.com → hk-01.mydomain.com (CNAME)
   ├─ app2.example.com → hk-01.mydomain.com (CNAME)
   └─ api.another.com → 1.2.3.4 (A 同IP)
```

1. 勾选你想要管理的记录
2. 对于服务器记录，可以修改名称和地区
3. 点击 **"导入选中的记录"**

### Step 5: 快速添加新域名

现在你已经导入了服务器，可以快速添加新的应用域名：

1. 找到目标服务器（如 hk-01）
2. 点击服务器卡片底部的 **"+ 快速添加域名"**
3. 输入完整域名，例如: `new-app.example.com`
4. 选择记录类型（默认 CNAME）
5. 点击 **"添加"**

✅ 完成！系统会自动：
- 在 DNS 提供商创建记录
- 保存到本地数据库
- 显示在服务器下方

## 💡 使用技巧

### 1. 域名命名建议

为了让系统更好地识别服务器，建议使用以下格式：

```
地域代码-数字.域名
```

例如:
- `hk-01.mydomain.com` - 香港服务器 1
- `us-02.mydomain.com` - 美国服务器 2
- `sg-1.mydomain.com` - 新加坡服务器 1

支持的地域代码:
- `hk` - 香港
- `us` - 美国
- `sg` - 新加坡
- `jp` - 日本
- `kr` - 韩国
- `de` - 德国
- `uk` - 英国
- `cn` - 中国
- 等等...

### 2. 两种添加方式

**方式 1: CNAME 记录（推荐）**
```
app.example.com → CNAME → hk-01.mydomain.com
```
优点: 服务器 IP 变更时只需修改一条记录

**方式 2: A 记录（直接指向）**
```
app.example.com → A → 1.2.3.4
```
优点: 减少一次 DNS 查询

### 3. 添加备注

为重要的域名添加备注，方便后续管理：

```
blog.example.com
备注：主博客，重要！
```

### 4. 定期同步

如果你在其他地方（如 DNS 提供商后台）修改了记录，可以：

1. 删除当前 Provider
2. 重新添加并同步
3. 重新导入记录

## 🎨 界面说明

### 主界面结构

```
┌─────────────────────────────────────────┐
│ DNSMesh        [设置] [退出登录]      │
├─────────────────────────────────────────┤
│ [+ 添加 Provider]                       │
├─────────────────────────────────────────┤
│ ☁️ Cloudflare                           │
│                                         │
│   🖥️ hk-01  香港  hk-01.mydomain.com   │
│   ├─ app1.example.com → CNAME          │
│   ├─ app2.example.com → CNAME          │
│   └─ [+ 快速添加域名]                  │
│                                         │
│ 🌐 腾讯云 DNSPod                        │
│   ...                                   │
└─────────────────────────────────────────┘
```

### 操作说明

- **添加 Provider**: 顶部工具栏
- **添加域名到服务器**: 服务器卡片底部
- **删除记录**: 记录右侧的"删除"按钮
- **查看详情**: 鼠标悬停查看完整信息

## 🔧 常见问题

### Q1: 添加 Provider 时提示连接失败？

**A**: 检查以下几点:
1. API credentials 是否正确
2. 网络连接是否正常
3. 对于 Cloudflare，确保使用的是 **Global API Key** 而不是 API Token
4. 对于腾讯云，确保 Secret ID 和 Secret Key 正确配对

### Q2: 为什么我的域名没有被识别为服务器？

**A**: 系统通过以下方式识别服务器（优先级从高到低）:
1. 域名格式匹配（如 `hk-01.domain.com`）
2. 被 2+ 个域名 CNAME 引用
3. 3+ 个域名使用同一 IP

如果不满足这些条件，可以在导入时手动标记为服务器。

### Q3: 如何修改已有记录？

**A**: 当前版本暂不支持编辑，请:
1. 删除旧记录
2. 重新添加新记录

### Q4: 数据存储在哪里？

**A**: 数据存储在本地 SQLite 数据库中，包括:
- DNS 记录信息
- Provider credentials（加密存储）
- 操作日志

DNS 提供商那边的记录不会被删除，只是本地不再管理。

### Q5: 删除 Provider 会删除 DNS 记录吗？

**A**: 不会！删除 Provider 只会:
- 删除本地数据库中的记录
- 不会影响 DNS 提供商的实际记录

### Q6: 支持哪些 DNS 记录类型？

**A**: 目前仅支持:
- A 记录（域名指向 IP）
- CNAME 记录（域名指向域名）

其他类型（MX, TXT, SRV 等）暂不支持。

### Q7: 可以管理多个 DNS 提供商吗？

**A**: 可以！你可以同时添加:
- 多个 Cloudflare 账户
- 多个腾讯云账户
- Cloudflare + 腾讯云 混合使用

## 📊 使用示例

### 场景 1: 新部署应用到 hk-01 服务器

```
1. 在 Portainer/Dokploy 部署应用
2. 回到域名管理系统
3. 找到 hk-01 服务器
4. 点击"+ 快速添加域名"
5. 输入: new-app.example.com
6. 点击"添加"
7. 完成！
```

### 场景 2: 服务器更换 IP

```
1. 删除服务器的 A 记录
2. 重新添加 A 记录，指向新 IP
3. 关联的 CNAME 记录自动生效（因为指向域名而非 IP）
```

### 场景 3: 批量导入现有域名

```
1. 添加 Provider
2. 系统自动同步所有记录
3. 查看智能分析结果
4. 勾选要管理的记录
5. 点击"导入"
6. 完成！
```

## 🎓 最佳实践

### 1. 命名规范

建议使用一致的命名规范：

```
服务器域名:
  hk-01.mydomain.com
  us-02.mydomain.com

应用域名:
  app-name.example.com
  api.example.com
  admin.example.com
```

### 2. 备注规范

为重要记录添加清晰的备注：

```
✅ 好的备注:
  "主博客，高流量"
  "测试环境，可随时删除"
  "客户 A 的专用域名"

❌ 不好的备注:
  "test"
  "aaa"
  ""
```

### 3. 定期检查

建议定期（如每月）:
- 检查是否有遗漏的记录
- 删除不再使用的记录
- 更新备注信息

### 4. 安全建议

- 定期更换管理员密码
- 定期轮换 API credentials
- 定期备份数据库
- 仅在安全网络环境下使用

## 📞 获取帮助

如果遇到问题：

1. 查看 **DEPLOYMENT.md** - 部署和故障排查
2. 查看 **PROJECT_SUMMARY.md** - 完整项目文档
3. 运行测试脚本: `./test.sh`
4. 查看日志: `docker-compose logs -f backend`

---

🎉 **开始使用吧！** 整个过程不超过 5 分钟。
