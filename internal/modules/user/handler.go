package user

import (
	"github.com/gin-gonic/gin"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/perhydrol/insurance-agent-backend/pkg/response"
	"go.uber.org/zap"
)

type UHandler struct {
	src Service
}

func NewUHandler(src Service) *UHandler {
	return &UHandler{src: src}
}

// Register 用户注册
// @Summary 用户注册
// @Description 用户注册接口
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterReq true "注册信息"
// @Success 200 {object} response.Response
// @Router /auth/register [post]
func (h *UHandler) Register(c *gin.Context) {
	var req RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to bind json in Register", zap.Error(err))
		response.Error(c, errno.ErrBadRequest.WithCause(err))
		return
	}

	if err := h.src.Register(c, &req); err != nil {
		logger.Log.Error("failed to Register", zap.String("username", req.Username), zap.Error(err))
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录接口
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginReq true "登录信息"
// @Success 200 {object} response.Response{data=LoginResp}
// @Router /auth/login [post]
func (h *UHandler) Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to bind json in Login", zap.Error(err))
		response.Error(c, errno.ErrBadRequest.WithCause(err))
		return
	}

	resp, err := h.src.Login(c, &req)
	if err != nil {
		logger.NewContext(
			c.Request.Context()).Info("failed to Login",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}
