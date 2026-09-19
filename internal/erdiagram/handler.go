package erdiagram

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const defaultUIPath = "/debug/er"

// RegisterRoutes 注册数据库 ER 图页面（Mermaid）。
func RegisterRoutes(r *gin.Engine, uiPath string) {
	if uiPath == "" {
		uiPath = defaultUIPath
	}
	r.GET(uiPath, func(c *gin.Context) {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, pageHTML())
	})
}

func pageHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Database ER</title>
<script type="module">
  import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";
  mermaid.initialize({
    startOnLoad: true,
    theme: "dark",
    securityLevel: "loose",
    er: { layoutDirection: "TB", useMaxWidth: false }
  });
</script>
<style>
  :root {
    --bg: #0b0f14;
    --surface: #121820;
    --border: #2a3444;
    --text: #e8edf4;
    --muted: #8b97a8;
    --accent: #4c8dff;
    --radius: 10px;
  }
  * { box-sizing: border-box; }
  html, body { height: 100%; }
  body {
    font-family: "SF Pro Text", ui-sans-serif, system-ui, sans-serif;
    background: var(--bg);
    color: var(--text);
    margin: 0;
    padding: 1.25rem 1.5rem 1.5rem;
    line-height: 1.45;
    display: flex; flex-direction: column; min-height: 100%;
  }
  .header { margin-bottom: 1rem; flex-shrink: 0; }
  .header h1 {
    font-size: 1.35rem; font-weight: 600; margin: 0 0 .35rem; letter-spacing: -0.02em;
  }
  .header p { margin: 0; color: var(--muted); font-size: .875rem; }
  .toolbar {
    display: flex; gap: .5rem; align-items: center; flex-wrap: wrap;
    margin-bottom: .75rem; flex-shrink: 0;
  }
  .btn {
    background: transparent; color: var(--muted); border: 1px solid var(--border);
    padding: .35rem .7rem; border-radius: 8px; cursor: pointer; font-size: .8rem;
  }
  .btn:hover { color: var(--text); border-color: var(--muted); }
  .zoom-label {
    font-size: .78rem; color: var(--muted); font-variant-numeric: tabular-nums;
    min-width: 3.5rem;
  }
  .viewport {
    flex: 1; min-height: 0;
    border: 1px solid var(--border); border-radius: var(--radius);
    background: var(--surface); overflow: hidden; position: relative;
    cursor: grab; touch-action: none;
    user-select: none; -webkit-user-select: none;
  }
  .viewport.dragging { cursor: grabbing; }
  .canvas {
    position: absolute; left: 0; top: 0;
    transform-origin: 0 0;
    will-change: transform;
  }
  .mermaid { display: inline-block; padding: 1.5rem; margin: 0; }
  .mermaid svg { max-width: none !important; height: auto !important; }
</style>
</head>
<body>
<div class="header">
  <h1>数据库 ER 图</h1>
  <p>滚轮缩放 · 拖拽平移 · 基于 Ent schema（Mermaid erDiagram）</p>
</div>
<div class="toolbar">
  <button type="button" class="btn" id="zoom-out">−</button>
  <span class="zoom-label" id="zoom-label">100%</span>
  <button type="button" class="btn" id="zoom-in">+</button>
  <button type="button" class="btn" id="zoom-reset">重置</button>
</div>
<div class="viewport" id="viewport">
  <div class="canvas" id="canvas">
<pre class="mermaid">
erDiagram
  users["users（用户）"] {
    int64 id PK "主键"
    string phone UK "手机号"
    string nickname "昵称"
    int8 status "状态 1启用"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  oauth_identities["oauth_identities（OAuth 身份）"] {
    int64 id PK "主键"
    string provider "渠道 wechat等"
    string provider_uid "渠道用户ID"
    string union_id "UnionID"
    int64 user_id FK "绑定用户"
    string nickname "渠道昵称"
    string avatar_url "头像"
    json extra "扩展信息"
    time authorized_at "授权时间"
    time bound_at "绑定时间"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  refresh_tokens["refresh_tokens（刷新令牌）"] {
    int64 id PK "主键"
    int64 user_id FK "用户"
    string token_hash UK "令牌哈希"
    time expires_at "过期时间"
    time revoked_at "撤销时间"
    string device_id "设备ID"
    string user_agent "UA"
    string ip "IP"
    time created_at "创建时间"
  }
  sms_codes["sms_codes（短信验证码）"] {
    int64 id PK "主键"
    string phone "手机号"
    string scene "场景 login等"
    string code_hash "验证码哈希"
    time expires_at "过期时间"
    time used_at "使用时间"
    time created_at "创建时间"
  }
  biz_types["biz_types（业态字典）"] {
    int64 id PK "主键"
    string code UK "业态编码"
    string name "业态名称"
    int sort "排序"
    int8 status "状态 1启用"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  stores["stores（门店）"] {
    int64 id PK "主键"
    string name "店名"
    string city "城市"
    string address "地址"
    string open_time "开门 HH:MM"
    string close_time "关门 HH:MM"
    string biz_type "业态编码"
    string invite_code UK "邀请码"
    int64 owner_user_id FK "老板用户"
    int8 status "状态 1正常"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  store_members["store_members（门店员工）"] {
    int64 id PK "主键"
    int64 store_id FK "门店"
    int64 user_id FK "用户"
    string role "角色 owner/staff"
    string status "状态 active/pending"
    string display_name "展示名"
    time joined_at "入职时间"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  store_notify_settings["store_notify_settings（门店通知配置）"] {
    int64 id PK "主键"
    int64 store_id FK "门店"
    string event "事件 open/recharge/consume"
    bool enabled "是否启用"
    bool notify_boss "通知老板"
    bool wechat "微信通知"
    bool app "App通知"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  members["members（会员）"] {
    int64 id PK "主键"
    int64 store_id FK "门店"
    string name "姓名"
    string name_pinyin "姓名全拼"
    string name_initials "姓名首拼"
    string phone "手机号"
    string source "来源"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  card_products["card_products（卡种）"] {
    int64 id PK "主键"
    int64 store_id FK "门店"
    string type "类型 value/count/pack"
    string name "卡种名"
    int price "价格分"
    int times "次卡次数"
    int valid_months "有效月数"
    int8 status "状态 1上架"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  card_product_items["card_product_items（套餐项目定义）"] {
    int64 id PK "主键"
    int64 product_id FK "卡种"
    string name "项目名"
    int times "次数"
    int sort "排序"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  member_cards["member_cards（持卡实例）"] {
    int64 id PK "主键"
    int64 store_id FK "门店"
    int64 member_id FK "会员"
    int64 product_id FK "卡种"
    string type "类型 value/count/pack"
    string name_snapshot "开卡时卡种名"
    int balance "储值余额分"
    int remain_times "次卡剩余"
    time valid_from "有效期起"
    time valid_to "有效期止"
    string status "active/expired/exhausted"
    int64 opened_by FK "开卡操作人"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  card_item_balances["card_item_balances（套餐项目余额）"] {
    int64 id PK "主键"
    int64 card_id FK "持卡"
    int64 product_item_id FK "套餐项目"
    string name_snapshot "项目名快照"
    int remain_times "剩余次数"
    time created_at "创建时间"
    time updated_at "更新时间"
  }
  ledger_entries["ledger_entries（流水账）"] {
    int64 id PK "主键"
    int64 store_id FK "门店"
    int64 member_id FK "会员"
    int64 card_id FK "持卡"
    string type "open/recharge/consume"
    string card_type "卡类型快照"
    int amount "金额分"
    int times "次数"
    string item_name "项目名"
    int balance_after "操作后余额"
    int times_after "操作后次数"
    string remark "备注"
    int64 operator_id FK "操作人"
    time created_at "创建时间"
  }
  ledger_entry_items["ledger_entry_items（流水扣项明细）"] {
    int64 id PK "主键"
    int64 ledger_id FK "流水"
    int64 item_balance_id FK "项目余额"
    int64 product_item_id FK "套餐项目"
    string name_snapshot "项目名快照"
    int times "扣减次数"
    int times_after "扣后剩余"
  }

  users ||--o{ stores : "拥有门店"
  users ||--o{ store_members : "加入门店"
  users ||--o{ oauth_identities : "绑定身份"
  users ||--o{ refresh_tokens : "刷新令牌"
  users ||--o{ ledger_entries : "操作流水"
  users ||--o{ member_cards : "开卡"
  stores ||--o{ store_members : "员工"
  stores ||--o{ store_notify_settings : "通知配置"
  stores ||--o{ members : "会员"
  stores ||--o{ card_products : "卡种"
  stores ||--o{ member_cards : "持卡"
  stores ||--o{ ledger_entries : "流水"
  members ||--o{ member_cards : "持有卡"
  members ||--o{ ledger_entries : "流水"
  card_products ||--o{ card_product_items : "套餐项目"
  card_products ||--o{ member_cards : "开出卡"
  member_cards ||--o{ card_item_balances : "项目余额"
  member_cards ||--o{ ledger_entries : "流水"
  ledger_entries ||--o{ ledger_entry_items : "扣项明细"
  card_product_items ||--o{ card_item_balances : "余额快照"
  card_item_balances ||--o{ ledger_entry_items : "被扣除"
</pre>
  </div>
</div>
<script>
(function () {
  const viewport = document.getElementById("viewport");
  const canvas = document.getElementById("canvas");
  const label = document.getElementById("zoom-label");
  const MIN = 0.15, MAX = 4, STEP = 1.12;
  let scale = 1, tx = 40, ty = 40;
  let dragging = false, lastX = 0, lastY = 0;

  function apply() {
    canvas.style.transform = "translate(" + tx + "px," + ty + "px) scale(" + scale + ")";
    label.textContent = Math.round(scale * 100) + "%";
  }

  function zoomAt(clientX, clientY, factor) {
    const next = Math.min(MAX, Math.max(MIN, scale * factor));
    if (next === scale) return;
    const rect = viewport.getBoundingClientRect();
    const x = clientX - rect.left;
    const y = clientY - rect.top;
    // keep point under cursor stable
    tx = x - (x - tx) * (next / scale);
    ty = y - (y - ty) * (next / scale);
    scale = next;
    apply();
  }

  viewport.addEventListener("wheel", function (e) {
    e.preventDefault();
    const factor = e.deltaY < 0 ? STEP : 1 / STEP;
    zoomAt(e.clientX, e.clientY, factor);
  }, { passive: false });

  viewport.addEventListener("pointerdown", function (e) {
    if (e.button !== 0) return;
    dragging = true;
    lastX = e.clientX;
    lastY = e.clientY;
    viewport.classList.add("dragging");
    viewport.setPointerCapture(e.pointerId);
  });
  viewport.addEventListener("pointermove", function (e) {
    if (!dragging) return;
    tx += e.clientX - lastX;
    ty += e.clientY - lastY;
    lastX = e.clientX;
    lastY = e.clientY;
    apply();
  });
  function endDrag(e) {
    if (!dragging) return;
    dragging = false;
    viewport.classList.remove("dragging");
    try { viewport.releasePointerCapture(e.pointerId); } catch (_) {}
  }
  viewport.addEventListener("pointerup", endDrag);
  viewport.addEventListener("pointercancel", endDrag);

  document.getElementById("zoom-in").onclick = function () {
    const r = viewport.getBoundingClientRect();
    zoomAt(r.left + r.width / 2, r.top + r.height / 2, STEP);
  };
  document.getElementById("zoom-out").onclick = function () {
    const r = viewport.getBoundingClientRect();
    zoomAt(r.left + r.width / 2, r.top + r.height / 2, 1 / STEP);
  };
  document.getElementById("zoom-reset").onclick = function () {
    scale = 1; tx = 40; ty = 40; apply();
  };

  apply();
})();
</script>
</body>
</html>`
}
