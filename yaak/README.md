# Yaak API 文档

工作区：**卡管家** · 环境：**Local**

## 变量如何填充（Request Chaining）

Yaak **不支持** Postman 式 post-response 脚本。官方做法是环境变量里用 `response.body.path`，从上游请求响应自动取值：

```text
${[ response.body.path(request='请求ID', path='$.data.xxx') ]}
```

| 变量 | 来源请求 | JSONPath |
|------|----------|----------|
| `baseUrl` | 静态 | `http://localhost:8080` |
| `phone` | 静态 | 默认 `13800138000` |
| `sms_code` | 发送短信验证码 | `$.data.dev_code` |
| `access_token` | 验证码登录 | `$.data.access_token` |
| `refresh_token` | 验证码登录 | `$.data.refresh_token` |
| `user_id` | 验证码登录 | `$.data.user.id` |
| `store_id` | 创建门店 | `$.data.id` |
| `invite_code` | 创建门店 | `$.data.invite_code` |
| `product_id` | 新建卡种 | `$.data.id` |
| `member_id` | 开卡 | `$.data.member.id` |
| `card_id` | 开卡 | `$.data.card.id` |
| `pack_item_id_1/2` | 开卡 | `$.data.card.items[0/1].id` |
| `ledger_id` | 扣除 | `$.data.ledger_id` |

请求里引用变量用 Yaak 语法：`${[ access_token ]}`（不是 `{{access_token}}`）。

首次引用时若上游还没响应，Yaak 会按 **When no responses** 自动先发上游请求。

## 分组（9 / 32 请求）

| 分组 | 说明 |
|------|------|
| 系统 | 健康检查 |
| 认证 | 登录 / Token / 当前用户 |
| 业态 | 业态列表 |
| 门店 | 门店 |
| 员工 | 默认 `X-Store-Id` + Bearer |
| 卡种 | 同上 |
| 会员 | 同上 |
| 持卡 | 同上 |
| 流水 | 同上 |

## 联调步骤

1. 选中环境 **Local**
2. `认证 / 发送短信验证码`（会写入 `sms_code`）
3. `认证 / 验证码登录`（会写入 token）
4. `业态 / 业态列表` → `门店 / 创建门店`（写入 `store_id`）
5. `卡种 / 新建卡种` → `会员 / 开卡` → `持卡 / 扣除`

点环境变量上的蓝色标签可改「来源请求 / JSONPath / 发送行为」。

若界面未刷新变量，切换一次环境或重启 Yaak。
