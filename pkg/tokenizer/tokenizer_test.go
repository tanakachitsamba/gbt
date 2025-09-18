package tokenizer

import "testing"

func TestTokenIDs(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		encodingName string
		want         []int
	}{
		{
			name:         "cl100k_base",
			text:         "Hello, world!",
			encodingName: "cl100k_base",
			want:         []int{9906, 11, 1917, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TokenIDs(tt.text, tt.encodingName)
			if err != nil {
				t.Fatalf("TokenIDs returned error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("TokenIDs returned %d tokens, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("TokenIDs[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}

	t.Run("slice isolation", func(t *testing.T) {
		tokens, err := TokenIDs("Hello, world!", "cl100k_base")
		if err != nil {
			t.Fatalf("TokenIDs returned error: %v", err)
		}
		tokens[0] = 0
		tokensAgain, err := TokenIDs("Hello, world!", "cl100k_base")
		if err != nil {
			t.Fatalf("TokenIDs returned error: %v", err)
		}
		if tokensAgain[0] == 0 {
			t.Fatalf("TokenIDs should return copy of tokens, got shared slice")
		}
	})
}

func TestTokenCounts(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		encodingName string
		want         int
	}{
		{
			name:         "cl100k_base",
			text:         "Hello, world!",
			encodingName: "cl100k_base",
			want:         4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TokenCount(tt.text, tt.encodingName)
			if err != nil {
				t.Fatalf("TokenCount returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("TokenCount returned %d, want %d", got, tt.want)
			}
		})
	}
}

func TestModelTokenization(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		model      string
		want       int
		wantTokens []int
	}{
		{
			name:       "gpt-3.5-turbo",
			text:       "Hello, world!",
			model:      "gpt-3.5-turbo",
			want:       4,
			wantTokens: []int{9906, 11, 1917, 0},
		},
		{
			name:       "gpt-4",
			text:       "The quick brown fox jumps over the lazy dog.",
			model:      "gpt-4",
			want:       10,
			wantTokens: []int{791, 4062, 14198, 39935, 35308, 927, 279, 16053, 5679, 13},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := TokenIDsForModel(tt.text, tt.model)
			if err != nil {
				t.Fatalf("TokenIDsForModel returned error: %v", err)
			}
			if len(tokens) != len(tt.wantTokens) {
				t.Fatalf("TokenIDsForModel returned %d tokens, want %d", len(tokens), len(tt.wantTokens))
			}
			for i := range tokens {
				if tokens[i] != tt.wantTokens[i] {
					t.Fatalf("TokenIDsForModel[%d] = %d, want %d", i, tokens[i], tt.wantTokens[i])
				}
			}

			count, err := TokenCountForModel(tt.text, tt.model)
			if err != nil {
				t.Fatalf("TokenCountForModel returned error: %v", err)
			}
			if count != tt.want {
				t.Fatalf("TokenCountForModel returned %d, want %d", count, tt.want)
			}
		})
	}
}
