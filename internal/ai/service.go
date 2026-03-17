package ai

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	openai "github.com/openai/openai-go"
	oaoption "github.com/openai/openai-go/option"
	"github.com/sshive/sshive/internal/config"
)

type Message struct {
	Role    string `json:"role"` // user | assistant
	Content string `json:"content"`
}

type Service struct {
	provider string
	model    string
	// Anthropic
	anthropicClient *anthropic.Client
	// OpenAI-compatible
	openaiClient *openai.Client
}

func NewService() *Service {
	cfg := config.C.AI
	provider := cfg.Provider
	if provider == "" {
		provider = "anthropic"
	}
	svc := &Service{provider: provider, model: cfg.Model}

	switch provider {
	case "openai":
		opts := []oaoption.RequestOption{oaoption.WithAPIKey(cfg.APIKey)}
		if cfg.BaseURL != "" {
			opts = append(opts, oaoption.WithBaseURL(cfg.BaseURL))
		}
		c := openai.NewClient(opts...)
		svc.openaiClient = &c
		if svc.model == "" {
			svc.model = "gpt-4o-mini"
		}
	default: // anthropic
		c := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))
		svc.anthropicClient = &c
		if svc.model == "" {
			svc.model = "claude-haiku-4-5-20251001"
		}
	}
	return svc
}

const systemPrompt = `你是一个 Linux 运维助手。用户会用自然语言描述想要做的操作，你需要返回对应的 Shell 命令。

规则：
- 只返回命令本身，不要解释，不要 markdown 代码块
- 如果需要多条命令，用换行分隔
- 命令必须安全，不要包含 rm -rf / 等破坏性操作
- 根据用户提供的 OS 信息调整命令语法
- 如果需求不明确，返回最常见的安全写法`

// StreamShellCommand 流式返回 Shell 命令，支持多轮对话历史
func (s *Service) StreamShellCommand(ctx context.Context, messages []Message, osInfo string, chunk chan<- string) error {
	// 注入 OS 信息到第一条用户消息
	if osInfo != "" && len(messages) > 0 {
		msgs := make([]Message, len(messages))
		copy(msgs, messages)
		msgs[0].Content = fmt.Sprintf("[目标主机 OS: %s]\n\n%s", osInfo, msgs[0].Content)
		messages = msgs
	}
	switch s.provider {
	case "openai":
		return s.streamOpenAI(ctx, messages, chunk)
	default:
		return s.streamAnthropic(ctx, messages, chunk)
	}
}

func (s *Service) streamAnthropic(ctx context.Context, messages []Message, chunk chan<- string) error {
	params := make([]anthropic.MessageParam, 0, len(messages))
	for _, m := range messages {
		if m.Role == "user" {
			params = append(params, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		} else {
			params = append(params, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}
	stream := s.anthropicClient.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: 512,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages:  params,
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

func (s *Service) streamOpenAI(ctx context.Context, messages []Message, chunk chan<- string) error {
	params := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
	}
	for _, m := range messages {
		if m.Role == "user" {
			params = append(params, openai.UserMessage(m.Content))
		} else {
			params = append(params, openai.AssistantMessage(m.Content))
		}
	}
	stream := s.openaiClient.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Model:     openai.ChatModel(s.model),
		Messages:  params,
		MaxTokens: openai.Int(512),
	})
	for stream.Next() {
		delta := stream.Current().Choices[0].Delta.Content
		if delta != "" {
			select {
			case chunk <- delta:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	return stream.Err()
}
