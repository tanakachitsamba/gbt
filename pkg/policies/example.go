package policies

import (
	"context"
	"errors"
	"fmt"
)

// Input captures the information passed between agents during orchestration.
type Input struct {
	Prompt       string
	Conversation []Message
	Temperature  float32
	Model        string
}

// Message records a single exchange within the shared conversation thread.
type Message struct {
	Role    string
	Content string
}

// Agent represents a single specialised role within the orchestration pipeline.
type Agent struct {
	Name   string
	Prompt func(Input) string
}

// ConversationThread stores the chronological history of agent outputs.
type ConversationThread struct {
	Messages []Message
}

// ResponseGenerator abstracts the model invocation used by getResponse.
type ResponseGenerator interface {
	Generate(ctx context.Context, in Input) (string, error)
}

// getResponse requests a completion from the underlying response generator.
func getResponse(ctx context.Context, generator ResponseGenerator, in Input) (string, error) {
	if generator == nil {
		return "", errors.New("response generator is required")
	}
	return generator.Generate(ctx, in)
}

// RunAgents executes each agent in sequence, preserving the conversation history.
func RunAgents(ctx context.Context, generator ResponseGenerator, agents []Agent, seed Input) (ConversationThread, error) {
	thread := ConversationThread{}
	current := seed

	for _, agent := range agents {
		if agent.Prompt == nil {
			return thread, fmt.Errorf("agent %q is missing a prompt function", agent.Name)
		}

		prompt := agent.Prompt(current)
		response, err := getResponse(ctx, generator, Input{
			Prompt:       prompt,
			Conversation: append(thread.Messages, Message{Role: agent.Name, Content: prompt}),
			Temperature:  current.Temperature,
			Model:        current.Model,
		})
		if err != nil {
			return thread, fmt.Errorf("agent %q failed: %w", agent.Name, err)
		}

		thread.Messages = append(thread.Messages,
			Message{Role: agent.Name, Content: prompt},
			Message{Role: fmt.Sprintf("%s_response", agent.Name), Content: response},
		)

		current.Conversation = thread.Messages
		current.Prompt = response
	}

	return thread, nil
}

// MockGenerator is a simple in-memory ResponseGenerator used for examples and tests.
type MockGenerator struct {
	Responses []string
	index     int
}

// Generate implements the ResponseGenerator interface.
func (m *MockGenerator) Generate(_ context.Context, _ Input) (string, error) {
	if m == nil || len(m.Responses) == 0 {
		return "", errors.New("no responses configured")
	}
	if m.index >= len(m.Responses) {
		return "", errors.New("exhausted mock responses")
	}
	response := m.Responses[m.index]
	m.index++
	return response, nil
}

// ExampleRunAgents demonstrates how a simple policy chain can be executed.
func ExampleRunAgents() {
	ctx := context.Background()
	generator := &MockGenerator{Responses: []string{
		"Summarised brief",
		"Plan reviewed",
		"Execution ready",
	}}

	agents := []Agent{
		{
			Name: "summariser",
			Prompt: func(in Input) string {
				return fmt.Sprintf("Summarise: %s", in.Prompt)
			},
		},
		{
			Name: "criticiser",
			Prompt: func(in Input) string {
				return fmt.Sprintf("Critique: %s", in.Prompt)
			},
		},
		{
			Name: "executor",
			Prompt: func(in Input) string {
				return fmt.Sprintf("Execute: %s", in.Prompt)
			},
		},
	}

	thread, err := RunAgents(ctx, generator, agents, Input{Prompt: "Draft a release plan", Temperature: 0.7, Model: "gpt-4"})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	for _, message := range thread.Messages {
		fmt.Printf("%s: %s\n", message.Role, message.Content)
	}
	// Output:
	// summariser: Summarise: Draft a release plan
	// summariser_response: Summarised brief
	// criticiser: Critique: Summarised brief
	// criticiser_response: Plan reviewed
	// executor: Execute: Plan reviewed
	// executor_response: Execution ready
}
