package user

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
	users, total, err := h.svc.List(tenantID, page, pageSize)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, gin.H{"total": total, "list": users})
}

func (h *Handler) Create(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	var req CreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	u, err := h.svc.Create(tenantID, req)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, u)
}

func (h *Handler) AssignRoles(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid user id")
		return
	}
	var req AssignRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.AssignRoles(tenantID, userID, req); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *Handler) Update(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid user id")
		return
	}
	var req UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}
	if err := h.svc.UpdateUser(tenantID, userID, req); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *Handler) Disable(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		middleware.BadRequest(c, "invalid user id")
		return
	}
	if err := h.svc.DisableUser(tenantID, userID); err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, nil)
}

func (h *Handler) GetRoles(c *gin.Context) {
	tenantID := auth.GetTenantID(c.Request.Context())
	roles, err := h.svc.ListRoles(tenantID)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	middleware.OK(c, roles)
}
