package prompt

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Repository defines the persistence operations required by the prompt evaluation
// service. Implementations are expected to provide DuckDB backed storage so that
// research context, evaluation metadata, and model transcripts can be recorded
// for future analysis.
type Repository interface {
	// ListResearchContext returns the stored research snippets that are
	// relevant for the provided model and optional topics. The limit controls
	// how many records are returned and should be treated as a best effort
	// maximum.
	ListResearchContext(ctx context.Context, targetModel string, topics []string, limit int) ([]ResearchContext, error)
	// SavePromptEvaluation persists the result of an evaluation run including
	// the raw model output and any research references that informed the
	// critique.
	SavePromptEvaluation(ctx context.Context, record PromptEvaluationRecord) error
}

// ResearchContext represents a single item of stored research that can be fed
// into the prompt evaluator to give it additional grounding when critiquing a
// prompt.
type ResearchContext struct {
	ID          string
	TargetModel string
	Topic       string
	Content     string
	Source      string
	CreatedAt   time.Time
}

// ResearchReference records which research items informed a particular
// evaluation. This allows downstream consumers to reconstruct the provenance of
// the resulting prompt improvements.
type ResearchReference struct {
	ContextID string
	Topic     string
	Content   string
	Source    string
}

// PromptEvaluationRecord captures everything needed to persist an evaluation
// run for later analysis.
type PromptEvaluationRecord struct {
	EvaluationID    string
	TargetModel     string
	EvaluationModel string
	OriginalPrompt  string
	ImprovedPrompt  string
	Critique        string
	Scores          map[string]float64
	Suggestions     []string
	RawModelOutput  string
	Metadata        map[string]string
	References      []ResearchReference
	CreatedAt       time.Time
}

// DuckDBRepository implements the Repository interface using a DuckDB
// connection. It lazily creates the required tables if they do not already
// exist.
type DuckDBRepository struct {
	db *sql.DB
}

// NewDuckDBRepository constructs a DuckDB backed repository. The provided
// database connection should remain open for the lifetime of the repository.
func NewDuckDBRepository(db *sql.DB) (*DuckDBRepository, error) {
	if db == nil {
		return nil, errors.New("duckdb connection is required")
	}

	repo := &DuckDBRepository{db: db}
	if err := repo.ensureSchema(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *DuckDBRepository) ensureSchema(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS prompt_research_contexts (
                        id TEXT PRIMARY KEY,
                        target_model TEXT,
                        topic TEXT,
                        content TEXT NOT NULL,
                        source TEXT,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		`CREATE TABLE IF NOT EXISTS prompt_evaluations (
                        id TEXT PRIMARY KEY,
                        target_model TEXT NOT NULL,
                        evaluation_model TEXT NOT NULL,
                        original_prompt TEXT NOT NULL,
                        improved_prompt TEXT NOT NULL,
                        critique TEXT,
                        scores JSON,
                        suggestions JSON,
                        raw_model_output TEXT,
                        metadata JSON,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );`,
		`CREATE TABLE IF NOT EXISTS prompt_evaluation_references (
                        evaluation_id TEXT NOT NULL,
                        context_id TEXT,
                        topic TEXT,
                        content TEXT,
                        source TEXT,
                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                        FOREIGN KEY (evaluation_id) REFERENCES prompt_evaluations(id)
                );`,
	}

	for _, stmt := range stmts {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("create duckdb schema: %w", err)
		}
	}
	return nil
}

// ListResearchContext fetches research snippets for the provided model/topic
// combination ordered by recency.
func (r *DuckDBRepository) ListResearchContext(ctx context.Context, targetModel string, topics []string, limit int) ([]ResearchContext, error) {
	if limit <= 0 {
		limit = 5
	}

	var builder strings.Builder
	builder.WriteString(`SELECT id, target_model, topic, content, source, created_at
                FROM prompt_research_contexts
                WHERE 1 = 1`)

	args := make([]any, 0)
	if targetModel != "" {
		builder.WriteString(" AND (target_model = ? OR target_model IS NULL OR target_model = '')")
		args = append(args, targetModel)
	}
	if len(topics) > 0 {
		builder.WriteString(" AND topic IN (" + placeholders(len(topics)) + ")")
		for _, topic := range topics {
			args = append(args, topic)
		}
	}
	builder.WriteString(" ORDER BY created_at DESC LIMIT ?")
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("query research context: %w", err)
	}
	defer rows.Close()

	contexts := make([]ResearchContext, 0)
	for rows.Next() {
		var rc ResearchContext
		if err := rows.Scan(&rc.ID, &rc.TargetModel, &rc.Topic, &rc.Content, &rc.Source, &rc.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan research context: %w", err)
		}
		contexts = append(contexts, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate research context: %w", err)
	}
	return contexts, nil
}

// SavePromptEvaluation stores the evaluation row and associated references.
func (r *DuckDBRepository) SavePromptEvaluation(ctx context.Context, record PromptEvaluationRecord) error {
	if record.EvaluationID == "" {
		return errors.New("evaluation id is required")
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}

	scoresJSON, err := json.Marshal(record.Scores)
	if err != nil {
		return fmt.Errorf("marshal scores: %w", err)
	}
	suggestionsJSON, err := json.Marshal(record.Suggestions)
	if err != nil {
		return fmt.Errorf("marshal suggestions: %w", err)
	}
	var metadataJSON []byte
	if record.Metadata != nil {
		metadataJSON, err = json.Marshal(record.Metadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
                INSERT INTO prompt_evaluations (
                        id, target_model, evaluation_model, original_prompt, improved_prompt,
                        critique, scores, suggestions, raw_model_output, metadata, created_at
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `,
		record.EvaluationID,
		record.TargetModel,
		record.EvaluationModel,
		record.OriginalPrompt,
		record.ImprovedPrompt,
		record.Critique,
		string(scoresJSON),
		string(suggestionsJSON),
		record.RawModelOutput,
		nullableJSONString(metadataJSON),
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert prompt evaluation: %w", err)
	}

	if len(record.References) > 0 {
		stmt, prepErr := tx.PrepareContext(ctx, `
                        INSERT INTO prompt_evaluation_references (
                                evaluation_id, context_id, topic, content, source
                        ) VALUES (?, ?, ?, ?, ?)
                `)
		if prepErr != nil {
			return fmt.Errorf("prepare evaluation reference insert: %w", prepErr)
		}
		defer stmt.Close()

		for _, ref := range record.References {
			contextID := any(nil)
			if ref.ContextID != "" {
				contextID = ref.ContextID
			}
			if _, err = stmt.ExecContext(ctx, record.EvaluationID, contextID, ref.Topic, ref.Content, ref.Source); err != nil {
				return fmt.Errorf("insert evaluation reference: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit prompt evaluation: %w", err)
	}
	return nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	items := make([]string, n)
	for i := range items {
		items[i] = "?"
	}
	return strings.Join(items, ",")
}

func nullableJSONString(data []byte) any {
	if len(data) == 0 {
		return nil
	}
	return string(data)
}
