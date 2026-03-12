package host

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sshive/sshive/internal/auth"
	"github.com/sshive/sshive/pkg/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler() *Handler {
	return &Handler{svc: NewService()}
}

func (h *Handler) List(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.svc.List(tenantID, page, pageSize)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, gin.H{"total": total, "list": list})
}

func (h *Handler) Create(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	host, err := h.svc.Create(tenantID, req)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, host)
}

func (h *Handler) Update(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid id")
		return
	}
	var req UpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	host, err := h.svc.Update(tenantID, id, req)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, host)
}

func (h *Handler) Delete(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid id")
		return
	}
	if err := h.svc.Delete(tenantID, id); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, nil)
}
