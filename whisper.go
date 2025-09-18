package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type transcriptionClient interface {
	CreateTranscription(context.Context, openai.AudioRequest) (openai.AudioResponse, error)
}

type transcriptionResult struct {
	index int
	file  string
	text  string
	err   error
}

func transcribeFile(c transcriptionClient, ctx context.Context, audioFile string, index int, wg *sync.WaitGroup, results chan<- transcriptionResult) {
	defer wg.Done()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: audioFile,
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		results <- transcriptionResult{index: index, file: audioFile, err: err}
		return
	}

	results <- transcriptionResult{index: index, file: audioFile, text: resp.Text}
}

func whisper(c transcriptionClient, ctx context.Context) ([]string, error) {

	// Define the audio files directory and the audio format
	audioDir := "audios/"
	audioFormat := ".mp3"

	// Read the audio files from the directory
	files, err := ioutil.ReadDir(audioDir)
	if err != nil {
		return nil, fmt.Errorf("error reading audio directory: %w", err)
	}

	// Filter audio files based on the audio format
	var audioFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), audioFormat) {
			audioFiles = append(audioFiles, audioDir+file.Name())
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
