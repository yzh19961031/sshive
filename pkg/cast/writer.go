package cast

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Header struct {
	Version   int    `json:"version"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Timestamp int64  `json:"timestamp"`
	Title     string `json:"title,omitempty"`
}

type Writer struct {
	mu        sync.Mutex
	f         *os.File
	startTime time.Time
}

func NewWriter(path string, width, height int) (*Writer, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("open cast file: %w", err)
	}
	header := Header{
		Version:   2,
		Width:     width,
		Height:    height,
		Timestamp: time.Now().Unix(),
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		f.Close()
		return nil, err
	}
	if _, err := fmt.Fprintf(f, "%s\n", headerJSON); err != nil {
		f.Close()
		return nil, err
	}
	return &Writer{f: f, startTime: time.Now()}, nil
}

func (w *Writer) WriteOutput(offset time.Duration, data string) {
	w.writeEvent("o", offset, data)
}

func (w *Writer) WriteInput(offset time.Duration, data string) {
	w.writeEvent("i", offset, data)
}

func (w *Writer) writeEvent(eventType string, offset time.Duration, data string) {
	elapsed := offset.Seconds()
	if offset == 0 {
		elapsed = time.Since(w.startTime).Seconds()
	}
	line, _ := json.Marshal([]any{elapsed, eventType, data})
	w.mu.Lock()
	defer w.mu.Unlock()
	fmt.Fprintf(w.f, "%s\n", line)
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.f.Close()
}
