# Postman

工作区：**卡管家**  
Collection：**卡管理 API（分组）**（推荐）/ 或导入本地 `card-manager-api.postman_collection.json`  
Environment：**Local**（`baseUrl=http://localhost:8080`）

## 分组结构

| 分组 | 说明 |
|------|------|
| System | 健康检查 |
| Auth | 登录 / Token / 当前用户 |
| BizTypes | 业态列表 |
| Stores | 门店 |
| Staff | 员工（需 `X-Store-Id`） |
| CardProducts | 卡种 |
| Members | 会员 / 开卡 |
| Cards | 卡详情 / 充值 / 扣除 |
| Ledger | 流水 |

## 联调步骤

1. 打开 Postman，选环境 **Local**，用 **卡管理 API（分组）**
2. `Auth / 发送短信验证码` → `Auth / 验证码登录`（自动写入 `access_token`）
3. `BizTypes / 业态列表` 拉行业 code
4. `Stores / 创建门店` → `store_id`
5. `CardProducts / 新建卡种` → `product_id` → `Members / 开卡` → `Cards / 扣除`

## 本地文件

- `postman/card-manager-api.postman_collection.json`：完整分组 Collection，可直接 Import
- 如需离线导入 Environment，从 Postman 导出 **Local** 到本目录

## 扣除（套餐多项）

`Cards / 扣除（支持套餐多项）`：

```json
{
  "items": [
    { "item_id": "{{pack_item_id_1}}", "times": 1 },
    { "item_id": "{{pack_item_id_2}}", "times": 1 }
  ],
  "remark": "一次扣多项"
}
```
