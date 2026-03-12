package tenant

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sshive/sshive/pkg/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler() *Handler {
	return &Handler{svc: NewService()}
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	tenants, total, err := h.svc.List(page, pageSize)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, gin.H{"total": total, "list": tenants})
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	t, err := h.svc.Create(req)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, t)
}
