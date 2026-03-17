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
	Messages []Message `json:"messages" binding:"required,min=1"`
	OSInfo   string    `json:"os_info"`
}

type SQLRequest struct {
	Messages  []Message `json:"messages" binding:"required,min=1"`
	DBContext string    `json:"db_context"` // 如 "MySQL · sshive 数据库"
}

// SQL 接收多轮对话历史，SSE 流式返回 SQL 查询
func (h *Handler) SQL(c *gin.Context) {
	var req SQLRequest
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
		errCh <- h.svc.StreamSQLCommand(c.Request.Context(), req.Messages, req.DBContext, chunk)
		close(chunk)
	}()

	var streamErr error
	c.Stream(func(w io.Writer) bool {
		select {
		case text, ok := <-chunk:
			if !ok {
				if streamErr != nil {
					fmt.Fprintf(w, "event: error\ndata: %s\n\n", streamErr.Error())
				} else {
					fmt.Fprintf(w, "event: done\ndata: \n\n")
				}
				return false
			}
			fmt.Fprintf(w, "data: %s\n\n", text)
			return true
		case err := <-errCh:
			streamErr = err
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

// Shell 接收多轮对话历史，SSE 流式返回 Shell 命令
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
		errCh <- h.svc.StreamShellCommand(c.Request.Context(), req.Messages, req.OSInfo, chunk)
		close(chunk)
	}()

	var streamErr error
	c.Stream(func(w io.Writer) bool {
		select {
		case text, ok := <-chunk:
			if !ok {
				if streamErr != nil {
					fmt.Fprintf(w, "event: error\ndata: %s\n\n", streamErr.Error())
				} else {
					fmt.Fprintf(w, "event: done\ndata: \n\n")
				}
				return false
			}
			fmt.Fprintf(w, "data: %s\n\n", text)
			return true
		case err := <-errCh:
			streamErr = err
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
