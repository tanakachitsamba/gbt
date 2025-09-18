package tokenizer

import (
	"fmt"
	"sync"

	tiktoken "github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"
)

var (
	loaderOnce sync.Once

	encodingMu    sync.RWMutex
	encodingCache = make(map[string]*tiktoken.Tiktoken)

	modelMu    sync.RWMutex
	modelCache = make(map[string]*tiktoken.Tiktoken)
)

func ensureLoader() {
	loaderOnce.Do(func() {
		tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	})
}

func encodingForName(name string) (*tiktoken.Tiktoken, error) {
	ensureLoader()

	encodingMu.RLock()
	enc, ok := encodingCache[name]
	encodingMu.RUnlock()
	if ok {
		return enc, nil
	}

	encodingMu.Lock()
	defer encodingMu.Unlock()

	if enc, ok := encodingCache[name]; ok {
		return enc, nil
	}

	enc, err := tiktoken.GetEncoding(name)
	if err != nil {
		return nil, fmt.Errorf("get encoding %q: %w", name, err)
	}
	encodingCache[name] = enc
	return enc, nil
}

func encodingForModel(model string) (*tiktoken.Tiktoken, error) {
	ensureLoader()

	modelMu.RLock()
	enc, ok := modelCache[model]
	modelMu.RUnlock()
	if ok {
		return enc, nil
	}

	modelMu.Lock()
	defer modelMu.Unlock()

	if enc, ok := modelCache[model]; ok {
		return enc, nil
	}

	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return nil, fmt.Errorf("encoding for model %q: %w", model, err)
	}
	modelCache[model] = enc
	return enc, nil
}

// TokenIDs returns the token identifiers for the provided text using the
// specified encoding name.
func TokenIDs(text, encodingName string) ([]int, error) {
	enc, err := encodingForName(encodingName)
	if err != nil {
		return nil, err
	}

	tokens := enc.Encode(text, nil, nil)
	result := make([]int, len(tokens))
	copy(result, tokens)
	return result, nil
}

// TokenCount returns the number of tokens for the provided text using the
// specified encoding name.
func TokenCount(text, encodingName string) (int, error) {
	tokens, err := TokenIDs(text, encodingName)
	if err != nil {
		return 0, err
	}
	return len(tokens), nil
}

// TokenIDsForModel returns the token identifiers for the provided text using
// the encoding associated with the supplied OpenAI model name.
func TokenIDsForModel(text, model string) ([]int, error) {
	enc, err := encodingForModel(model)
	if err != nil {
		return nil, err
	}

	tokens := enc.Encode(text, nil, nil)
	result := make([]int, len(tokens))
	copy(result, tokens)
	return result, nil
}

// TokenCountForModel returns the number of tokens for the provided text using
// the encoding associated with the supplied OpenAI model name.
func TokenCountForModel(text, model string) (int, error) {
	enc, err := encodingForModel(model)
	if err != nil {
		return 0, err
	}
	return len(enc.Encode(text, nil, nil)), nil
}
