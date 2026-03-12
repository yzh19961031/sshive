package audit

import (
	"log/slog"

	"github.com/sshive/sshive/internal/model"
	"github.com/sshive/sshive/pkg/cast"
)

type EventType int

const (
	EventOutput EventType = iota
	EventInput
	EventCommand
)

type Event struct {
	Type    EventType
	Content string
}

// AsyncWriter 异步接收审计事件，不阻塞主 SSH 链路
type AsyncWriter struct {
	sessionID int64
	repo      *Repo
	castW     *cast.Writer
	ch        chan Event
	done      chan struct{}
}

func NewAsyncWriter(sessionID int64, castWriter *cast.Writer) *AsyncWriter {
	w := &AsyncWriter{
		sessionID: sessionID,
		repo:      &Repo{},
		castW:     castWriter,
		ch:        make(chan Event, 256),
		done:      make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *AsyncWriter) run() {
	defer close(w.done)
	for evt := range w.ch {
		switch evt.Type {
		case EventOutput:
			if w.castW != nil {
				w.castW.WriteOutput(0, evt.Content)
			}
			if err := w.repo.WriteLog(&model.SessionLog{
				SessionID: w.sessionID,
				Type:      "output",
				Content:   evt.Content,
			}); err != nil {
				slog.Warn("audit write output log failed", "err", err)
			}
		case EventInput:
			if w.castW != nil {
				w.castW.WriteInput(0, evt.Content)
			}
			if err := w.repo.WriteLog(&model.SessionLog{
				SessionID: w.sessionID,
				Type:      "input",
				Content:   evt.Content,
			}); err != nil {
				slog.Warn("audit write input log failed", "err", err)
			}
		case EventCommand:
			if err := w.repo.WriteCommand(&model.SessionCommand{
				SessionID: w.sessionID,
				Command:   evt.Content,
			}); err != nil {
				slog.Warn("audit write command failed", "err", err)
			}
		}
	}
}

func (w *AsyncWriter) SendOutput(data string) {
	select {
	case w.ch <- Event{Type: EventOutput, Content: data}:
	default:
		slog.Warn("audit channel full, dropping output event")
	}
}

func (w *AsyncWriter) SendInput(data string) {
	select {
	case w.ch <- Event{Type: EventInput, Content: data}:
	default:
	}
}

func (w *AsyncWriter) SendCommand(cmd string) {
	select {
	case w.ch <- Event{Type: EventCommand, Content: cmd}:
	default:
	}
}

// Close 等待所有事件处理完毕后关闭
func (w *AsyncWriter) Close() {
	close(w.ch)
	<-w.done
	if w.castW != nil {
		w.castW.Close()
	}
}
