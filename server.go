package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// APIServer orchestrates HTTP request handling and response rendering.
type APIServer struct {
	openai *OpenAIWrapper
}

// NewAPIServer constructs a new server instance.
func NewAPIServer(openai *OpenAIWrapper) *APIServer {
	return &APIServer{openai: openai}
}

func (s *APIServer) handleCreateResponse(w http.ResponseWriter, r *http.Request) {
	var req ResponseRequestV1
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateResponseRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := s.openai.CreateResponse(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *APIServer) handleCreateThread(w http.ResponseWriter, r *http.Request) {
	var req ThreadRequestV1
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := s.openai.CreateThread(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (s *APIServer) handleCreateAssistant(w http.ResponseWriter, r *http.Request) {
	var req AssistantRequestV1
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateAssistantRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := s.openai.CreateAssistant(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (s *APIServer) handleCreateVectorStore(w http.ResponseWriter, r *http.Request) {
	var req VectorStoreRequestV1
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validateVectorStoreRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := s.openai.CreateVectorStore(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (s *APIServer) handleOpenAPIDocument(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpec)
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	if decoder.More() {
		return errors.New("invalid JSON payload: multiple documents not supported")
	}

	return nil
}

func validateResponseRequest(req ResponseRequestV1) error {
	if req.Model == "" {
		return errors.New("model is required")
	}

	if req.Temperature != nil {
		if *req.Temperature < 0 || *req.Temperature > 2 {
			return errors.New("temperature must be between 0 and 2")
		}
	}

	if req.Instructions == "" && len(req.Input) == 0 {
		return errors.New("instructions or input is required")
	}

	for i, block := range req.Input {
		if block.Role == "" {
			return fmt.Errorf("input[%d].role is required", i)
		}
		if len(block.Content) == 0 {
			return fmt.Errorf("input[%d].content must include at least one entry", i)
		}
		for j, content := range block.Content {
			if content.Text == "" {
				return fmt.Errorf("input[%d].content[%d].text is required", i, j)
			}
		}
	}

	return nil
}

func validateAssistantRequest(req AssistantRequestV1) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Model == "" {
		return errors.New("model is required")
	}
	return nil
}

func validateVectorStoreRequest(req VectorStoreRequestV1) error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

func handleServiceError(w http.ResponseWriter, err error) {
	// Treat all validation errors as Bad Request, others as Internal Server Error.
	// If validation functions return errors, they should be considered client errors.
	writeError(w, http.StatusBadRequest, err.Error())
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
