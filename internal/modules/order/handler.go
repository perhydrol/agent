package order

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/perhydrol/insurance-agent-backend/pkg/response"
	"go.uber.org/zap"
)

type OHandler struct {
	svc Service
}

func NewOHandler(svc Service) *OHandler {
	return &OHandler{svc: svc}
}

// CreateOrder 创建订单
// @Summary 创建订单
// @Description 创建订单
// @Tags Order
// @Accept json
// @Produce json
// @Param request body CreateOrderReq true "创建订单请求"
// @Success 200 {object} response.Response{data=domain.Order}
// @Router /orders [post]
func (h *OHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to bind json in CreateOrder",
			zap.Error(err),
		)
		response.Error(c, errno.ErrBadRequest.WithCause(err))
		return
	}

	userID := c.GetInt64("userID")

	orderResp, err := h.svc.CreateOrder(c, userID, req.ProductID)
	if err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to CreateOrder",
			zap.Int64("user_id", userID),
			zap.Int64("product_id", req.ProductID),
			zap.Error(err),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, orderResp)
}

// PayOrder 支付订单
// @Summary 支付订单
// @Description 支付订单
// @Tags Order
// @Accept json
// @Produce json
// @Param id path string true "订单ID"
// @Success 200 {object} response.Response
// @Router /orders/{id}/pay [post]
func (h *OHandler) PayOrder(c *gin.Context) {
	idStr := c.Param("id")
	orderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, errno.ErrBadRequest.WithCause(err))
		return
	}

	userID := c.GetInt64("userID")

	if err := h.svc.PayOrder(c, userID, orderID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
