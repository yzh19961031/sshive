package ai

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type ShellRequest struct {
	Prompt string `json:"prompt" binding:"required"`
	OSInfo string `json:"os_info"`
}

// Shell 接收自然语言，SSE 流式返回 Shell 命令
func (h *Handler) Shell(c *gin.Context) {
	var req ShellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")

	chunk := make(chan string, 32)
	errCh := make(chan error, 1)

	go func() {
		errCh <- h.svc.StreamShellCommand(c.Request.Context(), req.Prompt, req.OSInfo, chunk)
		close(chunk)
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case text, ok := <-chunk:
			if !ok {
				fmt.Fprintf(w, "event: done\ndata: \n\n")
				return false
			}
			fmt.Fprintf(w, "data: %s\n\n", text)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
