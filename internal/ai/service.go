package ai

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/sshive/sshive/internal/config"
)

type Service struct {
	client *anthropic.Client
	model  string
}

func NewService() *Service {
	cfg := config.C.AI
	model := cfg.Model
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))
	return &Service{client: &client, model: model}
}

const systemPrompt = `你是一个 Linux 运维助手。用户会用自然语言描述想要做的操作，你需要返回对应的 Shell 命令。

规则：
- 只返回命令本身，不要解释，不要 markdown 代码块
- 如果需要多条命令，用换行分隔
- 命令必须安全，不要包含 rm -rf / 等破坏性操作
- 根据用户提供的 OS 信息调整命令语法
- 如果需求不明确，返回最常见的安全写法`

// StreamShellCommand 流式返回 Shell 命令，通过 chunk channel 发送每个文本块
func (s *Service) StreamShellCommand(ctx context.Context, prompt, osInfo string, chunk chan<- string) error {
	userMsg := prompt
	if osInfo != "" {
		userMsg = fmt.Sprintf("[目标主机 OS: %s]\n\n%s", osInfo, prompt)
	}

	stream := s.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: 512,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userMsg)),
		},
	})

	for stream.Next() {
		event := stream.Current()
		if event.Type == "content_block_delta" && event.Delta.Text != "" {
			select {
			case chunk <- event.Delta.Text:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return stream.Err()
}
