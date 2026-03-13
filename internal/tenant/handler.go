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

// ListPublic 无需认证，仅返回租户 ID 和名称，供登录页下拉
func (h *Handler) ListPublic(c *gin.Context) {
	tenants, _, err := h.svc.List(1, 100)
	if err != nil {
		middleware.InternalError(c, err.Error())
		return
	}
	type item struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	list := make([]item, 0, len(tenants))
	for _, t := range tenants {
		if t.Status == 1 {
			list = append(list, item{ID: t.ID, Name: t.Name})
		}
	}
	middleware.OK(c, list)
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
