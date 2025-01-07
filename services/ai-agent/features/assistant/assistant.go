package assistant

import (
	"context"

	"github.com/henomis/lingoose/assistant"
	"github.com/henomis/lingoose/llm/openai"
	"github.com/henomis/lingoose/thread"
)

type Assistant struct {
	aiAssistant *assistant.Assistant
}

// Function for creating a new aiAssistant instance
func New() *Assistant {
	openaiLLM := assistant.New(openai.New().
		WithTemperature(0).
		WithModel(openai.GPT4))

	return &Assistant{aiAssistant: openaiLLM}
}

// Function to start a conversation thread
func (a *Assistant) CreateThread(systemMessage, userMessage string) *thread.Thread {
	return thread.New().AddMessage(thread.NewSystemMessage().AddContent(
		thread.NewTextContent(systemMessage),
	)).AddMessage(thread.NewUserMessage().AddContent(
		thread.NewTextContent(userMessage),
	))
}

// Function to gather responses from LLM
func (a *Assistant) ResponseThread(ctx context.Context, th *thread.Thread) (string, error) {
	err := a.aiAssistant.RunWithThread(ctx, th)

	if err != nil {
		return "", err
	}

	return th.LastMessage().Contents[0].Data.(string), nil
}
