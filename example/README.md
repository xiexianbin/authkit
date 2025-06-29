# example

## 1. 需求分析 (Requirement Analysis)

核心需求是将本地账户系统与第三方 OAuth 系统解耦，并提供统一的登录和绑定流程。

### 主要流程

#### A. OAuth 登录流程 (用户未登录)

1.  **前端请求**: 前端向后端请求特定提供商（如 Google）的登录跳转 URL。
    * `GET /api/v1/oauth/google/login`
2.  **后端响应**: 后端生成 Google 的授权 URL（包含 `client_id`, `redirect_uri`, `scope`, `state`）并返回给前端。
3.  **用户授权**: 前端重定向用户到该 URL，用户在 Google 页面登录并授权。
4.  **回调**: Google 携带 `code` 和 `state` 参数，重定向用户到后端指定的回调 URL。
    * `GET /api/v1/oauth/google/callback?code=...&state=...`
5.  **后端处理**:
    * 验证 `state` 参数，防止 CSRF 攻击。
    * 使用 `code` 向 Google 服务器换取 `access_token`。
    * 使用 `access_token` 获取 Google 用户的基本信息（如 `provider_user_id`, `email`, `name`, `avatar`）。
    * **业务逻辑判断**:
        * **情况1: 该第三方账号已存在 (`oauth_accounts` 表)**: 直接找到关联的本地用户 (`users` 表)，生成 JWT 并返回，登录成功。
        * **情况2: 该第三方账号不存在，但 Email 已在 `users` 表中存在**: 说明用户可能用邮箱密码注册过。为该用户自动绑定这个新的第三方账号（在 `oauth_accounts` 表中创建记录），然后生成 JWT 并返回，登录成功。
        * **情况3: 全新用户**: 在 `users` 表和 `oauth_accounts` 表中同时创建新记录并关联它们，然后生成 JWT 并返回，注册并登录成功。
6.  **前端接收**: 前端收到后端的 JWT，保存下来用于后续的 API 请求。

#### B. OAuth 绑定流程 (用户已登录)

1.  **前提**: 用户已登录，请求头中持有合法的 JWT。
2.  **前端请求**: 用户在个人中心点击“绑定 GitHub 账号”，前端请求 GitHub 的绑定 URL。
    * `GET /api/v1/oauth/github/bind` (需要 JWT 认证)
3.  **后端响应**: 后端生成 GitHub 的授权 URL，流程同上。
4.  **用户授权 & 回调**: 流程同上，回调到 `.../callback`。
5.  **后端处理**:
    * 与登录流程不同的是，后端首先会通过请求中的 JWT 解析出当前登录的 `user_id`。
    * 获取到 GitHub 用户信息后。
    * **业务逻辑判断**:
        * **检查1**: 查询该 GitHub 账号是否已被 **其他** 本地用户绑定。如果是，返回错误（"此 GitHub 账号已被其他用户绑定"）。
        * **检查2**: 如果未被绑定，则在 `oauth_accounts` 表中创建一条新记录，将其 `user_id` 指向当前登录的用户 ID。
6.  **前端接收**: 前端收到成功消息，刷新用户信息界面，显示已绑定的账号。

---

## 2. 数据库设计 (Database Design)

我们将设计两张核心表：`users` (本地用户表) 和 `oauth_accounts` (第三方授权表)。

### 表结构 (GORM Models)

#### `users` 表
存储我们系统内部的用户基础信息。这张表应该与任何特定的 OAuth 提供商无关。

| 字段名 | 类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | 主键, 自增 | 用户唯一ID |
| `created_at` | `datetime` | | 创建时间 |
| `updated_at` | `datetime` | | 更新时间 |
| `deleted_at` | `datetime` | GORM软删除 | 删除时间 |
| `username` | `varchar(100)` | 唯一, Not Null | 用户名 |
| `email` | `varchar(255)` | 唯一, 索引 | 用户邮箱 |
| `password_hash`| `varchar(255)`| Nullable | 密码哈希 (允许仅OAuth用户，无密码) |
| `avatar` | `varchar(255)` | Nullable | 用户头像URL |

#### `oauth_accounts` 表
这张表是关键，它建立了我们 `users` 表和第三方平台用户之间的多对一关系（一个 user 可以绑定多个 oauth account）。

| 字段名 | 类型 | 约束 | 描述 |
| :--- | :--- | :--- | :--- |
| `id` | `uint` | 主键, 自增 | 记录ID |
| `created_at` | `datetime` | | 创建时间 |
| `updated_at` | `datetime` | | 更新时间 |
| `user_id` | `uint` | 外键 (users.id), 索引 | 关联的本地用户ID |
| `provider` | `varchar(50)` | Not Null, 索引 | OAuth提供商 (e.g., "github", "google") |
| `provider_user_id` | `varchar(255)` | Not Null | 用户在第三方平台的唯一ID |
| `access_token` | `text` | Nullable | 第三方访问令牌 (建议加密存储) |
| `refresh_token`| `text` | Nullable | 第三方刷新令牌 (建议加密存储) |
| `expires_at` | `datetime` | Nullable | `access_token` 的过期时间 |

**关系**: `User` has many `OauthAccount`. `OauthAccount` belongs to a `User`.


- http://127.0.0.1:8080/api/v1/oauth/:provider/login
  - http://127.0.0.1:8080/api/v1/oauth/github/login
- http://127.0.0.1:8080/api/v1/oauth/:provider/callback
