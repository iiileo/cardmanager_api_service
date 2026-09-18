# 卡管家 API 接口文档（Postman）

本地默认：`http://localhost:8080`（见 `internal/config/config.yaml` 的 `server.address`）。

> **维护约定**：修改 HTTP 接口时须同步 **`postman/README.md`**，并通过 **Postman MCP（`user-postman`）** 更新云端 Collection / Environment（见 `.cursor/rules/postman-api-sync.mdc`）。**仓库内不再存放 Collection / Environment JSON。**

## Postman 云端（唯一维护处）

| 项 | 名称 | UID |
|----|------|-----|
| Workspace | 卡管家 | `5a4fa48c-2a8a-4137-ac70-40a746b9ca3b` |
| Collection | 卡管家 API | `50595618-af109a3a-5dd8-4330-8622-03ae4239ecf1` |
| Environment | Local | `50595618-d01c6f27-b78b-4db8-80a9-5385265b1eb7` |

在 Postman 客户端登录同一账号 → 打开 Workspace **卡管家** → 同步 Collection **卡管家 API**，**右上角环境必须选 Local**（否则 Post-response 写的 token 不会保存）。

### 鉴权说明

- **仅 4 个公开请求** 使用 `No Auth`：`/healthz`、发验证码、登录、刷新令牌。
- **其余请求** 在 Postman 里显式设为 **Bearer Token** `{{access_token}}`（云端 `auth: null` 不会继承 Collection，故逐请求配置）。
- Collection 级 Bearer 仍保留，与请求级 Bearer 一致。
- 门店内接口另加 Header **`X-Store-Id: {{store_id}}`**。

### Post-response 脚本（Postman Scripts 页）

各请求 **Post-response** 内联脚本（不依赖 Collection 公共函数）：解析 `{ code, message, data }`，`code === 0` 时 `pm.environment.set(...)`。

| 请求 | 写入变量 |
|------|----------|
| 发送短信验证码 | `sms_code` ← `data.dev_code` |
| 验证码登录 | `access_token`、`refresh_token`、`user_id` |
| 刷新令牌 | `access_token`、`refresh_token` |
| 发送绑定手机验证码 | `bind_sms_code` |
| 创建门店 | `store_id`、`invite_code` |
| 新建卡种 | `product_id` |
| 开卡 | `member_id`、`card_id`、`pack_item_id_1/2` |
| 充值 / 扣除 | `ledger_id` |

### 推荐联调顺序

1. **认证 / 发送短信验证码**（Post-response 写入 `sms_code`）
2. **认证 / 验证码登录**（Post-response 写入 `access_token`、`refresh_token`）
3. **门店 / 创建门店**（Post-response 写入 `store_id`）
4. **首页与经营数据 / 首页统计** 或 **经营数据概览**
5. **卡种 → 会员 / 开卡 → 会员卡 / 充值或扣除**

---

## 通用约定

### 响应包络

成功 HTTP 200：

```json
{
  "code": 0,
  "message": "ok",
  "data": { }
}
```

业务错误：`code != 0`，`data` 多为 `null`。

### 鉴权

| Header | 说明 |
|--------|------|
| `Authorization: Bearer {{access_token}}` | 除发验证码、登录、刷新令牌外必填 |
| `X-Store-Id: {{store_id}}` | **门店内**接口必填（见接口目录「需门店」列） |

金额字段单位为**分**（整数）。流水日期 `from` / `to` 为 **`YYYY-MM-DD`**，按服务器**本地时区**自然日。

### 分页（会员列表、流水类列表）

| Query | 默认 | 上限 |
|-------|------|------|
| `page` | `1` | — |
| `page_size` | `20` | `100` |

列表 `data` 含：`list`、`total`、`page`、`page_size`、`has_more`（`page * page_size < total`）。

### 列表附带汇总 `include_stats`

仅**列表**接口：`1` / `true` / `yes` / `on` → 响应带 `stats`（与同路径 `/stats` 结构一致）。

适用：`GET /api/v1/ledger`、`/recharges`、`/consumes`。

---

## Postman 环境变量（云端 Environment **Local**）

下表为含义与自动写入来源（在 Postman 中编辑，Agent 用 MCP `updateEnvironment` / `putEnvironment` 同步）：

| 变量 | 默认 / 来源 |
|------|-------------|
| `baseUrl` | `http://localhost:8080` |
| `phone` | `13800138000` |
| `sms_code` | 发码 Post-response → `data.dev_code` |
| `access_token` / `refresh_token` | 登录或刷新 Post-response |
| `user_id` | 登录 Post-response → `data.user.id` |
| `store_id` / `invite_code` | 创建门店 Post-response |
| `product_id` | 新建卡种 Post-response |
| `member_id` / `card_id` | 开卡 Post-response |
| `pack_item_id_1` / `pack_item_id_2` | 开卡 Post-response（套餐子项） |
| `ledger_id` | 充值/消费 Post-response |
| `new_phone` / `bind_sms_code` | 改手机号流程 |

手动设置示例：`pm.environment.set("access_token", pm.response.json().data.access_token)`（Collection 中已内置常用 Tests）。

---

## 接口目录

与 Postman Collection 分组一致。路径为实际 URL；`:id` 等为路径参数。

| 模块 | 方法 | 路径 | 需登录 | 需门店 | 说明 |
|------|------|------|:------:|:------:|------|
| 系统 | GET | `/healthz` | | | 健康检查 |
| 认证 | POST | `/api/v1/auth/sms/send` | | | 发送短信验证码 |
| 认证 | POST | `/api/v1/auth/login/sms` | | | 验证码登录 |
| 认证 | POST | `/api/v1/auth/token/refresh` | | | 刷新令牌 |
| 认证 | POST | `/api/v1/auth/logout` | ✓ | | 退出登录 |
| 认证 | GET | `/api/v1/auth/me` | ✓ | | 当前用户 |
| 认证 | PATCH | `/api/v1/auth/me` | ✓ | | 修改昵称 |
| 认证 | PATCH | `/api/v1/auth/me/phone` | ✓ | | 修改手机号 |
| 认证 | DELETE | `/api/v1/auth/me` | ✓ | | 注销账号（删除相关数据） |
| 业态 | GET | `/api/v1/biz-types` | ✓ | | 业态列表 |
| 门店 | GET | `/api/v1/stores` | ✓ | | 我的门店列表 |
| 门店 | POST | `/api/v1/stores` | ✓ | | 创建门店 |
| 门店 | GET | `/api/v1/stores/invite/:code` | ✓ | | 邀请码预览门店 |
| 门店 | POST | `/api/v1/stores/join` | ✓ | | 申请加入门店 |
| 门店 | GET | `/api/v1/stores/:id` | ✓ | | 门店详情 |
| 门店 | PATCH | `/api/v1/stores/:id` | ✓ | | 更新门店资料 |
| 门店 | GET | `/api/v1/stores/:id/invite-code` | ✓ | | 查看邀请码 |
| 门店 | POST | `/api/v1/stores/:id/invite-code/refresh` | ✓ | | 刷新邀请码 |
| 首页与经营数据 | GET | `/api/v1/home/stats` | ✓ | ✓ | 首页统计 |
| 首页与经营数据 | GET | `/api/v1/stats/overview` | ✓ | ✓ | 经营数据概览 |
| 员工 | GET | `/api/v1/staff` | ✓ | ✓ | 在职员工列表 |
| 员工 | GET | `/api/v1/staff/applications` | ✓ | ✓ | 待审核入店申请（老板） |
| 员工 | POST | `/api/v1/staff/applications/:id/approve` | ✓ | ✓ | 同意入店（老板） |
| 员工 | POST | `/api/v1/staff/applications/:id/reject` | ✓ | ✓ | 拒绝入店（老板） |
| 卡种 | GET | `/api/v1/card-products` | ✓ | ✓ | 卡种列表 |
| 卡种 | POST | `/api/v1/card-products` | ✓ | ✓ | 新建卡种（老板） |
| 卡种 | DELETE | `/api/v1/card-products/:id` | ✓ | ✓ | 删除卡种（老板） |
| 会员 | GET | `/api/v1/members` | ✓ | ✓ | 会员列表 |
| 会员 | POST | `/api/v1/members/cards` | ✓ | ✓ | 开卡 |
| 会员 | GET | `/api/v1/members/:id` | ✓ | ✓ | 会员详情 |
| 会员 | GET | `/api/v1/members/:id/cards` | ✓ | ✓ | 会员名下卡列表 |
| 会员卡 | GET | `/api/v1/cards/:id` | ✓ | ✓ | 卡详情 |
| 会员卡 | POST | `/api/v1/cards/:id/recharge` | ✓ | ✓ | 充值 |
| 会员卡 | POST | `/api/v1/cards/:id/consume` | ✓ | ✓ | 消费/扣次 |
| 流水 | GET | `/api/v1/ledger` | ✓ | ✓ | 流水列表 |
| 流水 | GET | `/api/v1/ledger/stats` | ✓ | ✓ | 流水汇总 |
| 流水 | GET | `/api/v1/ledger/:id` | ✓ | ✓ | 流水详情 |
| 流水 | GET | `/api/v1/recharges` | ✓ | ✓ | 充值记录列表 |
| 流水 | GET | `/api/v1/recharges/stats` | ✓ | ✓ | 充值汇总 |
| 流水 | GET | `/api/v1/recharges/:id` | ✓ | ✓ | 充值详情 |
| 流水 | GET | `/api/v1/consumes` | ✓ | ✓ | 消费记录列表 |
| 流水 | GET | `/api/v1/consumes/stats` | ✓ | ✓ | 消费汇总 |
| 流水 | GET | `/api/v1/consumes/:id` | ✓ | ✓ | 消费详情 |
| 通知 | GET | `/api/v1/notify/settings` | ✓ | ✓ | 通知设置列表（老板） |
| 通知 | GET | `/api/v1/notify/settings/open` | ✓ | ✓ | 开卡通知详情（老板） |
| 通知 | PUT | `/api/v1/notify/settings/open` | ✓ | ✓ | 更新开卡通知（老板） |

本地调试（`access_log.ui_enabled: true`）：

| 模块 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 调试 | GET | `/debug/access-logs` | 访问日志页面 |
| 调试 | GET | `/debug/access-logs/api?limit=100` | 访问日志 JSON |

---

## 认证

### 发送验证码

`POST {{baseUrl}}/api/v1/auth/sms/send`

```json
{ "phone": "{{phone}}", "scene": "login" }
```

改手机号场景：`scene` 为 `bind_phone`（以代码为准）。

短信通道由配置 `auth.sms_provider` 选择（`dev` / `spug`），实现见 `internal/sms.Sender`：

| 配置 | 行为 |
|------|------|
| `sms_provider: dev`（或 `sms_dev_mode: true`） | 不真正发短信；响应可带 `data.dev_code` |
| `sms_provider: spug` | [Spug 推送助手](https://push.spug.cc/guide/sms)：`POST https://push.spug.cc/sms/<模板编码>`，body `to`/`code`，带有效期模板时另传 `number`（分钟） |

生产请设 `sms_dev_mode: false`、`sms_provider: spug`，并填写控制台复制的 `sms_spug_template_code`（勿写入客户端）。

Postman **Post-response**（云端「发送短信验证码」）：`code === 0` 时把 `data.dev_code` 写入 `sms_code`（仅开发模式有该字段）。

### 验证码登录

`POST {{baseUrl}}/api/v1/auth/login/sms`

```json
{ "phone": "{{phone}}", "code": "{{sms_code}}" }
```

成功时 `data`（`TokenResponse`）示例字段：

| 字段 | 说明 |
|------|------|
| `access_token` | 访问令牌（Collection Bearer 使用） |
| `refresh_token` | 刷新令牌 |
| `access_expires_in` / `refresh_expires_in` | 秒 |
| `user.id` / `user.phone` / `user.nickname` | 当前用户 |

Postman **Post-response**（「验证码登录」）：`code === 0` 时写入 `access_token`、`refresh_token`、`user_id`。

### 刷新令牌

`POST {{baseUrl}}/api/v1/auth/token/refresh`

```json
{ "refresh_token": "{{refresh_token}}" }
```

`data` 结构与登录相同（无 `user` 时仅更新 token 字段）。

Postman **Post-response**（「刷新令牌」）：`code === 0` 时覆盖 `access_token`、`refresh_token`。

### 修改手机号

`PATCH {{baseUrl}}/api/v1/auth/me/phone`

```json
{ "phone": "13900139000", "code": "123456" }
```

### 注销账号

`DELETE {{baseUrl}}/api/v1/auth/me`

需登录。请求体必须显式确认：

```json
{ "confirm": true }
```

成功后删除当前用户相关数据，包括：

- 名下门店及其业务数据（会员、卡、卡种、流水、通知设置、员工关系等）
- 在他人门店留下的流水操作人改挂该店老板（不删对方门店数据）
- 本用户的员工身份、刷新令牌、OAuth 绑定、短信验证码记录、用户本身

注销后 access / refresh token 均失效，需重新注册登录。

---

## 业态

`GET {{baseUrl}}/api/v1/biz-types`

面向**服务行业**：老板/员工用 App 记会员卡（储值/次卡），客户无需操作 App。

| code | 名称 | 典型场景 |
|------|------|----------|
| `beauty` | 美业 | 美发、美甲、护肤、纹绣 |
| `spa` | 养生保健 | 推拿、理疗、足疗、按摩 |
| `fitness` | 健身运动 | 瑜伽、普拉提、健身房私教 |
| `education` | 教育培训 | 兴趣班、驾校、早教 |
| `pet` | 宠物服务 | 洗护、美容、寄养 |
| `auto` | 汽车服务 | 洗车、美容、保养 |
| `photo` | 摄影写真 | 写真套餐、跟拍次卡 |
| `other` | 其他服务 | 未单独列出的服务店 |

创建门店时 `biz_type` 传上表 `code`。旧值 `tea`（茶饮）、`retail`（零售）已停用，不再出现在列表。

---

## 首页统计

```http
GET {{baseUrl}}/api/v1/home/stats
GET {{baseUrl}}/api/v1/home/stats?month=2026-08
```

| 字段 | 含义 |
|------|------|
| `today_recharge` / `today_consume` | 今日充值 / 储值消费（`consume_value`） |
| `today_new_members` | 今日新建会员 |
| `today_txn_count` | 今日笔数（充值+消费，不含开卡） |
| `month_*` | 默认当月 1 日至今；`month=YYYY-MM` 为指定自然月 |
| `store_balance` | 在店储值卡余额合计 |

---

## 经营数据（统计 Tab）

```http
GET {{baseUrl}}/api/v1/stats/overview?days=7
GET {{baseUrl}}/api/v1/stats/overview?days=30
```

| Query | 默认 | 说明 |
|-------|------|------|
| `days` | `7` | 仅 `7` 或 `30`；含今日的最近 N 天 |

`data` 含：`range_days`、`from`、`to`、`summary`（充值/消费/开卡/复购及环比）、`daily_recharge[]`（每项 `date` 为 `YYYY-MM-DD`、`amount` 分）、`card_type_mix`、`recharge_rank[]`（Top10）。字段定义见 `internal/api/dto/stats.go`。

---

## 会员

```http
GET {{baseUrl}}/api/v1/members?q=张三&page=1&page_size=20
```

`q`：姓名、拼音、手机号（含尾号）。

---

## 流水 / 充值 / 消费

### `GET /api/v1/ledger` 的 `kind` / `type`

| `kind` | 范围 |
|--------|------|
| （省略） | 全部（含开卡） |
| `txn` | 充值 + 消费（不含开卡） |
| `recharge` / `consume` | 子集 |

| `type` | `recharge` · `consume_value` · `consume_count` · `consume_pack` · `open` |

公共 Query：`member_id`、`card_id`、`from`、`to`、`page`、`page_size`、`card_type`（`value|count|pack`）、`include_stats`。

`GET /api/v1/ledger/stats`：**未传 `kind` 时默认按 `txn` 汇总**。

`/recharges`、`/consumes` 路径已限定类型，仍可用 `type` 细筛；列表支持 `include_stats=1`。

---

## 通知（仅老板）

`event`：`open` · `recharge` · `consume` · `count`

PUT body：`enabled`、`notify_boss`、`wechat`、`app`（启用时至少一种渠道）。

---

## 开卡约定

- 同店同会员：**储值卡 / 次卡各一张**；重复开卡返回已有卡（`ledger_id` 可能为空）
- **套餐卡**可多张

---

## 建议联调顺序

1. 发验证码 → 登录 → 保存 `access_token`
2. 业态列表 → 创建门店 → 设置 `X-Store-Id`
3. 首页统计 / 经营数据概览
4. 新建卡种 → 开卡 → 充值或消费
5. 流水列表（`kind=txn` + `include_stats=1` 验证汇总）

## 仓库内文档

| 文件 | 用途 |
|------|------|
| `README.md` | 接口说明（本文档） |

Collection / Environment **仅维护在 Postman 云端**；Agent 通过 **`user-postman` MCP** 更新（Workspace / Collection / Environment UID 见 `.cursor/rules/postman-api-sync.mdc`）。
