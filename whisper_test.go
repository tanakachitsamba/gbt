package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openai/openai-go"
)

type mockTranscriptionClient struct {
	responses map[string]string
	delays    map[string]time.Duration
}

type fileNamer interface {
	Name() string
}

func (m *mockTranscriptionClient) CreateTranscription(ctx context.Context, params openai.AudioTranscriptionNewParams) (*openai.Transcription, error) {
	if params.File == nil {
		return nil, fmt.Errorf("missing file reader")
	}

	var path string
	if named, ok := params.File.(fileNamer); ok {
		path = named.Name()
	}
	if closer, ok := params.File.(io.Seeker); ok {
		_, _ = closer.Seek(0, io.SeekStart)
	}

	if delay, ok := m.delays[path]; ok {
		time.Sleep(delay)
	}

	response, ok := m.responses[path]
	if !ok {
		return nil, fmt.Errorf("unexpected file: %s", path)
	}

	return &openai.Transcription{Text: response}, nil
}

func TestWhisperMultipleFiles(t *testing.T) {
	audioDir := "audios"
	if err := os.MkdirAll(audioDir, 0o755); err != nil {
		t.Fatalf("failed to create audio directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(audioDir); err != nil {
			t.Errorf("failed to clean up audio directory: %v", err)
		}
	})

	fileNames := []string{"b.mp3", "a.mp3", "c.mp3"}
	for _, name := range fileNames {
		path := filepath.Join(audioDir, name)
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			t.Fatalf("failed to create test audio file %s: %v", path, err)
		}
	}

	mockClient := &mockTranscriptionClient{
		responses: map[string]string{
			filepath.Join(audioDir, "a.mp3"): "alpha",
			filepath.Join(audioDir, "b.mp3"): "bravo",
			filepath.Join(audioDir, "c.mp3"): "charlie",
		},
		delays: map[string]time.Duration{
			filepath.Join(audioDir, "a.mp3"): 20 * time.Millisecond,
			filepath.Join(audioDir, "b.mp3"): 5 * time.Millisecond,
			filepath.Join(audioDir, "c.mp3"): 10 * time.Millisecond,
		},
	}

	ctx := context.Background()
	transcriptions, err := whisper(mockClient, ctx)
	if err != nil {
		t.Fatalf("whisper returned an error: %v", err)
	}

	expected := []string{"alpha", "bravo", "charlie"}
	if len(transcriptions) != len(expected) {
		t.Fatalf("expected %d transcriptions, got %d", len(expected), len(transcriptions))
	}

	for i, want := range expected {
		if transcriptions[i] != want {
			t.Errorf("transcription %d mismatch: want %q, got %q", i, want, transcriptions[i])
		}
	}
}
