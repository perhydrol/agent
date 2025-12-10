package product

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/perhydrol/insurance-agent-backend/pkg/errno"
	"github.com/perhydrol/insurance-agent-backend/pkg/logger"
	"github.com/perhydrol/insurance-agent-backend/pkg/response"
	"go.uber.org/zap"
)

type PHandler struct {
	srv Service
}

func NewPHandler(srv Service) *PHandler {
	return &PHandler{srv: srv}
}

// ListProducts 获取产品列表
// @Summary 获取产品列表
// @Description 获取产品列表，支持分页和分类筛选
// @Tags Product
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Param category query string false "分类"
// @Success 200 {object} response.Response{data=PListResp}
// @Router /products [get]
func (h *PHandler) ListProducts(c *gin.Context) {
	var req PListReq
	if err := c.ShouldBindJSON(req); err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to bind query in ListProducts", zap.Error(err))
		response.Error(c, errno.ErrBadRequest.WithCause(err))
		return
	}

	products, total, err := h.srv.ListProducts(c, req.Page, req.PageSize, req.Category)
	if err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to ListProducts", zap.Error(err))
		response.Error(c, errno.ErrNotFound.WithCause(err))
		return
	}

	respList := make([]PResp, 0, len(products))
	for _, pr := range products {
		if pr == nil {
			logger.NewContext(c.Request.Context()).Error(
				"failed to process product due to nil value",    // 更明确的日志消息
				zap.Error(errors.New("product pointer is nil")), // 创建一个具体的错误对象
			)
			continue
		}
		respList = append(respList, *pr)
	}
	resp := PListResp{List: respList, Total: total}
	response.Page(c, resp)
}

// GetProduct 获取产品详情
// @Summary 获取产品详情
// @Description 根据ID获取产品详情
// @Tags Product
// @Accept json
// @Produce json
// @Param id path string true "产品ID"
// @Success 200 {object} response.Response{data=PResp}
// @Router /products/{id} [get]
func (h *PHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to parse productID in GetProduct", zap.String("id", idStr), zap.Error(err))
		response.Error(c, errno.ErrBadRequest.WithCause(err))
		return
	}

	product, err := h.srv.GetProduct(c, id)
	if err != nil {
		logger.NewContext(c.Request.Context()).Error("failed to GetProduct", zap.Int64("id", id), zap.Error(err))
		response.Error(c, err)
		return
	}

	if product == nil {
		response.Error(c, errno.ErrNotFound)
		return
	}

	response.Success(c, product)
}
