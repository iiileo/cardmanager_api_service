# Postman

工作区：**卡管家**  
Collection：**卡管理 API**  
Environment：**Local**（`baseUrl=http://localhost:8080`）

## 联调步骤

1. 打开 Postman，选环境 **Local**
2. `Auth / 发送短信验证码` → `Auth / 验证码登录`（自动写入 `access_token`）
3. `业态列表`（`GET /api/v1/biz-types`）拉取可选行业，创建门店时填返回的 `code`
4. 其他接口走 Collection Bearer `{{access_token}}`，无需手填
5. `Stores / 创建门店` 会写入 `store_id`；Staff 接口自动带 `X-Store-Id`

## 脚本说明

| 请求 | Tests 脚本 |
|------|-----------|
| 验证码登录 | 写入 `access_token` / `refresh_token` / `user_id` |
| 刷新 Token | 更新双 token |
| 创建门店 | 写入 `store_id` / `invite_code` |

## 本地文件

如需离线导入，可从 Postman 导出 Collection / Environment 到本目录。
