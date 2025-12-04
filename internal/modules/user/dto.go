package user

// RegisterReq 注册请求参数
type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32" example:"alice"`
	Password string `json:"password" binding:"required,min=6" example:"password121212"`
	Email    string `json:"email" binding:"required,email" example:"alice@example.com"`
}

// LoginReq 登录请求参数
type LoginReq struct {
	Username string `json:"username" binding:"required" example:"johndoe"`
	Password string `json:"password" binding:"required" example:"secret123"`
}

// LoginResp 登录成功返回的数据
type LoginResp struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1Ni..."`
	ExpiresIn   int    `json:"expires_in" example:"86400"` // Token 有效期(秒)
	UserID      int64  `json:"user_id" example:"1"`
	Username    string `json:"username" example:"johndoe"`
}

// ProfileResp 用户个人信息返回
type ProfileResp struct {
	ID        int64  `json:"id" example:"1"`
	Username  string `json:"username" example:"johndoe"`
	Email     string `json:"email" example:"john@example.com"`
	CreatedAt string `json:"created_at" example:"2023-10-01 12:00:00"`
}
