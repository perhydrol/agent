### 1. 统一响应规范 (Standard Response)

所有非流式接口，必须遵循此格式。

```go
// pkg/response/response.go

type Response struct {
    Code int         `json:"code"`           // 业务码: 0=成功, >0=错误
    Msg  string      `json:"msg"`            // 提示信息
    Data interface{} `json:"data,omitempty"` // 数据载荷
}

// 分页响应包装
type PageResult struct {
    List      interface{} `json:"list"`
    Total     int64       `json:"total"`
    Page      int         `json:"page"`
    PageSize  int         `json:"page_size"`
}
```

---

### 2. 模块一：认证 (Auth Module)

最基础的 JWT 鉴权。

**URL Prefix:** `/api/v1/auth`

```go
// internal/modules/user/delivery/dto.go

// 注册请求
type RegisterReq struct {
    Username string `json:"username" binding:"required,min=3,max=32"`
    Password string `json:"password" binding:"required,min=6"`
    Email    string `json:"email" binding:"required,email"`
}

// 登录请求
type LoginReq struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// 登录响应
type LoginResp struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"` // 秒
    UserID      int64  `json:"user_id"`
    Username    string `json:"username"`
}
```

*   `POST /register`: 注册
*   `POST /login`: 登录

---

### 3. 模块二：保险产品 (Product Module)

展示你的半结构化数据设计 (`Features` 字段)。

**URL Prefix:** `/api/v1/products`

```go
// internal/modules/product/delivery/dto.go

// 产品列表查询参数 (Query Params)
type ProductListReq struct {
    Page     int    `form:"page,default=1"`
    PageSize int    `form:"page_size,default=10"`
    Category string `form:"category"` // e.g., "travel", "health"
}

// 产品详情响应 (VO)
type ProductResp struct {
    ID          int64           `json:"id"`
    Name        string          `json:"name"`
    Category    string          `json:"category"`
    BasePrice   float64         `json:"base_price"`
    Description string          `json:"description"`
    // Highlights: 使用 map[string]interface{} 或 json.RawMessage 处理动态字段
    Features    json.RawMessage `json:"features"` 
}
```

*   `GET /`: 获取列表 (支持分页)
*   `GET /:id`: 获取详情

---

### 4. 模块三：AI 顾问 (Agent Module)

这里我们需要两套接口：一套普通 JSON，一套 SSE 流式。

**URL Prefix:** `/api/v1/agent`

```go
// internal/modules/agent/delivery/dto.go

// 聊天请求
type ChatReq struct {
    SessionID string `json:"session_id" binding:"required"` // 前端生成 UUID，保持对话上下文
    Query     string `json:"query" binding:"required,max=1000"`
    Stream    bool   `json:"stream"` // 是否开启流式 (可选)
}

// 普通模式响应 (非流式)
type ChatResp struct {
    Answer      string   `json:"answer"`       // AI 的回答 (Markdown)
    References  []string `json:"references"`   // RAG 引用的条款片段 (增强可信度)
    UsedTokens  int      `json:"used_tokens"`  // (可选) Token 消耗
}

// --- SSE 流式响应结构 (Event Stream) ---
// event: message
// data: {"chunk": "建议", "done": false}
// data: {"chunk": "购买", "done": false}
// data: {"chunk": "", "done": true, "references": [...]}
type StreamChunk struct {
    Chunk      string   `json:"chunk"`
    Done       bool     `json:"done"`
    References []string `json:"references,omitempty"`
}
```

*   `POST /chat`: 发送消息，获取完整回复 (简单调试用)。
*   `GET /chat/stream`: **(高级)** SSE 接口。Query 参数: `?session_id=xxx&query=xxx`。响应 Header 为 `Content-Type: text/event-stream`。

---

### 5. 模块四：订单与异步处理 (Order Module)

展示状态机和异步 Worker 的接口。

**URL Prefix:** `/api/v1/orders`

```go
// internal/modules/order/delivery/dto.go

// 下单请求
type CreateOrderReq struct {
    ProductID int64 `json:"product_id" binding:"required"`
    // 真实场景还需要投保人信息，MVP 简化，假设用当前登录用户的信息
}

// 订单详情响应
type OrderResp struct {
    ID           string  `json:"id"` // UUID
    ProductName  string  `json:"product_name"`
    TotalAmount  float64 `json:"total_amount"`
    Status       string  `json:"status"`        // Pending, Paid, Underwriting, Active
    PolicyNumber string  `json:"policy_number"` // 只有 Active 状态才有
    CreatedAt    string  `json:"created_at"`
}

// 支付请求 (Mock)
type PaymentReq struct {
    PaymentMethod string `json:"payment_method" binding:"required"` // "wechat", "alipay"
}
```

*   `POST /`: 创建订单 -> 返回 Status: "Pending"
*   `POST /:id/pay`: 支付 -> **触发异步任务** -> 立即返回 Status: "Paid" (提示: 正在核保)
*   `GET /:id`: 查询详情 -> **前端轮询** -> 直到 Status 变为 "Active"

---

### 6. Swagger 整合策略

既然没有前端，你需要让 Swagger 尽可能好用。在你的 `main.go` 和 `handler` 代码上方，你需要加上这样的注释（使用 `swaggo/swag`）：

```go
// @Summary 与 AI 顾问对话
// @Description 基于 RAG 检索保险条款并回答问题
// @Tags Agent
// @Accept json
// @Produce json
// @Param request body dto.ChatReq true "用户问题"
// @Success 200 {object} response.Response{data=dto.ChatResp}
// @Router /api/v1/agent/chat [post]
func (h *AgentHandler) Chat(c *gin.Context) { ... }
```

---

### 7. 总结：你需要在代码里做的事

1.  **定义包结构**：在 `internal/modules/{module}/delivery` 下创建 `dto.go`，把上面的 Struct 放进去。
2.  **Gin Binding**：利用 `binding:"required"` 等标签，让 Gin 帮你做参数校验（省去很多 `if` 判断）。
3.  **Router Group**：
    ```go
    v1 := r.Group("/api/v1")
    {
        auth := v1.Group("/auth")
        auth.POST("/login", authHandler.Login)

        // 需要登录的接口
        api := v1.Group("/", middleware.JWTAuth())
        {
            api.POST("/orders", orderHandler.Create)
            api.POST("/agent/chat", agentHandler.Chat)
        }
    }
    ```

这个 API 设计既满足了 MVP 的需求，又包含了 **SSE 流式** 和 **异步轮询** 这两个能体现高级架构能力的点。

**准备好开始写第一行代码了吗？我们建议先从 User 和 Product 模块开始，热热身。**