package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/openai/openai-go"
)

type transcriptionClient interface {
	CreateTranscription(context.Context, openai.AudioTranscriptionNewParams) (*openai.Transcription, error)
}

type transcriptionResult struct {
	index int
	file  string
	text  string
	err   error
}

func transcribeFile(c transcriptionClient, ctx context.Context, audioFile string, index int, wg *sync.WaitGroup, results chan<- transcriptionResult) {
	defer wg.Done()

	file, err := os.Open(audioFile)
	if err != nil {
		results <- transcriptionResult{index: index, file: audioFile, err: err}
		return
	}
	defer file.Close()

	params := openai.AudioTranscriptionNewParams{
		Model: openai.AudioModelWhisper1,
		File:  file,
	}

	resp, err := c.CreateTranscription(ctx, params)
	if err != nil {
		results <- transcriptionResult{index: index, file: audioFile, err: err}
		return
	}

	results <- transcriptionResult{index: index, file: audioFile, text: resp.Text}
}

func whisper(c transcriptionClient, ctx context.Context) ([]string, error) {
	audioDir := "audios"
	targetExtension := ".mp3"

	entries, err := os.ReadDir(audioDir)
	if err != nil {
		return nil, fmt.Errorf("error reading audio directory: %w", err)
	}

	audioFiles := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.EqualFold(filepath.Ext(name), targetExtension) {
			audioFiles = append(audioFiles, filepath.Join(audioDir, name))
		}
	}

	transcriptions := make([]string, len(audioFiles))
	results := make(chan transcriptionResult, len(audioFiles))
	var wg sync.WaitGroup
	for idx, audioFile := range audioFiles {
		wg.Add(1)
		go transcribeFile(c, ctx, audioFile, idx, &wg, results)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []string
	for res := range results {
		if res.err != nil {
			fmt.Printf("Transcription error for %s: %v\n", res.file, res.err)
			errs = append(errs, fmt.Sprintf("%s: %v", res.file, res.err))
			continue
		}
		transcriptions[res.index] = res.text
	}

	for i, transcription := range transcriptions {
		if transcription == "" {
			continue
		}
		fmt.Printf("Transcription %d: %s\n", i+1, transcription)
	}

	if len(errs) > 0 {
		return transcriptions, fmt.Errorf(strings.Join(errs, "; "))
	}

	return transcriptions, nil
}
