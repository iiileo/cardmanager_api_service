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

## 分组

| 分组 | 说明 |
|------|------|
| 系统 | 健康检查 |
| 认证 | 登录 / Token / 当前用户 |
| 业态 | 业态列表 |
| 门店 | 门店 |
| 首页 | `GET /api/v1/home/stats` 今日充值/消费/新开会员/在店余额/笔数 |
| 员工 | 默认 `X-Store-Id` + Bearer |
| 卡种 | 同上 |
| 会员 | 列表 `q` 支持姓名 / 全拼 / 首拼 / 手机号（含尾号） |
| 持卡 | 同上 |
| 流水 | 全量流水（兼容） |
| 充值记录 | 列表 / 统计 / 详情 |
| 消费记录 | 列表 / 统计 / 详情（套餐含 `items`） |

## 首页统计

```http
GET /api/v1/home/stats
```

需 Bearer + `X-Store-Id`。金额单位：分；「今日」按服务器本地时区 0 点～24 点。

| 字段 | 含义 |
|------|------|
| `today_recharge` | 今日充值金额（`recharge` 流水合计） |
| `today_consume` | 今日储值消费金额（`consume_value` 绝对值合计） |
| `today_new_members` | 今日新建会员数 |
| `store_balance` | 在店储值卡余额合计 |
| `today_txn_count` | 今日笔数（充值 + 各类消费，不含开卡） |

## 充值 / 消费（同一列表）

同一页展示充值+消费，用统一流水：

```http
GET /api/v1/ledger?kind=txn
```

- `kind=txn`：充值+消费（不含开卡）
- 前端用 `type` 区分（`recharge` / `consume_*`）
- 套餐消费带 `items[]`
- 汇总：`GET /api/v1/ledger/stats`

只要一侧时用 `/recharges` 或 `/consumes`。

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/ledger?kind=txn` | 充值+消费混合列表 |
| GET | `/api/v1/ledger/stats` | 混合汇总 |
| GET | `/api/v1/recharges` | 仅充值 |
| GET | `/api/v1/recharges/stats` | 充值汇总 |
| GET | `/api/v1/recharges/:id` | 充值详情 |
| GET | `/api/v1/consumes` | 仅消费 |
| GET | `/api/v1/consumes/stats` | 消费汇总 + `by_pack_item` |
| GET | `/api/v1/consumes/:id` | 消费详情 |

公共 Query：`member_id`、`card_id`、`from`、`to`、`page`、`page_size`、`card_type`  
另：`kind=txn|recharge|consume`，`type=recharge|consume_value|consume_count|consume_pack`

流水写入时会快照 `card_type`，统计按类型/套餐项目聚合，便于后续报表。

## 联调步骤

1. 选中环境 **Local**
2. `认证 / 发送短信验证码`（会写入 `sms_code`）
3. `认证 / 验证码登录`（会写入 token）
4. `业态 / 业态列表` → `门店 / 创建门店`（写入 `store_id`）
5. `卡种 / 新建卡种` → `会员 / 开卡` → `持卡 / 扣除`

点环境变量上的蓝色标签可改「来源请求 / JSONPath / 发送行为」。

若界面未刷新变量，切换一次环境或重启 Yaak。

## 开卡约定

- 同门店同会员：**储值卡 / 次卡各只能开一张**；再次开卡返回已有卡（`ledger_id` 为空）
- **套餐卡**可开多张
