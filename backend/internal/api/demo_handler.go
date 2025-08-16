package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/jjckrbbt/catalyst/backend/internal/apps/demo"
	"github.com/labstack/echo/v4"
	"github.com/pgvector/pgvector-go"
)

// --- Struct Definitions ---

type PlannerResponse struct {
	ToolCalls []ToolCall `json:"tool_calls"`
}
type ToolCall struct {
	ToolName  string            `json:"tool"`
	Arguments map[string]string `json:"arguments"`
}

type HybridContext struct {
	MissionFacts    *demo.VwApolloMissionFact
	KnowledgeChunks []demo.FindSimilarMissionKnowledgeRow
}

type EmbeddingRequest struct {
	Text string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

type LLMRequestBody struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type DemoHandler struct {
	queries             *demo.Queries
	logger              *slog.Logger
	httpClient          *http.Client
	embeddingServiceURL string
	plannerTemplate     *template.Template
	synthesizerTemplate *template.Template
	openAIAPIKey        string
}

func NewDemoHandler(q *demo.Queries, logger *slog.Logger, apiKey string) (*DemoHandler, error) {
	plannerTmpl, err := template.ParseFiles("backend/configs/prompts/planner_prompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse planner template: %w", err)
	}

	synthesizerTmpl, err := template.ParseFiles("backend/configs/prompts/synthesizer_prompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse synthesizer template: %w", err)
	}

	return &DemoHandler{
		queries:             q,
		logger:              logger.With("component", "demo_handler"),
		httpClient:          &http.Client{Timeout: 30 * time.Second},
		embeddingServiceURL: "http://localhost:5001/embed",
		plannerTemplate:     plannerTmpl,
		synthesizerTemplate: synthesizerTmpl,
		openAIAPIKey:	     apiKey,
	}, nil
}

func (h *DemoHandler) HandleHybridQuery(c echo.Context) error {
	ctx := c.Request().Context()
	reqLogger := h.logger.With("request_id", c.Get("requestID"))

	question := c.FormValue("question")
	if question == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "form value 'question' is required")
	}
	reqLogger.InfoContext(ctx, "Received new hybrid query", "question", question)

	plan, err := h.getExecutionPlan(ctx, question)
	if err != nil {
		reqLogger.ErrorContext(ctx, "Failed to get execution plan", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error planning query")
	}
	reqLogger.InfoContext(ctx, "LLM generated execution plan", "plan", plan)

	contextData, err := h.getContextFromPlan(ctx, plan)
	if err != nil {
		reqLogger.ErrorContext(ctx, "Failed to get context from plan", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error executing plan")
	}
	reqLogger.InfoContext(ctx, "Successfully executed plan and gathered context")

	finalAnswer, err := h.synthesizeAnswer(ctx, question, contextData)
	if err != nil {
		reqLogger.ErrorContext(ctx, "Failed to synthesize final answer", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error synthesizing answer")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"answer": finalAnswer})
}

func (h *DemoHandler) getExecutionPlan(ctx context.Context, question string) ([]ToolCall, error) {
	var promptBuffer bytes.Buffer
	if err := h.plannerTemplate.Execute(&promptBuffer, map[string]string{"UserQuestion": question}); err != nil {
		return nil, fmt.Errorf("failed to execute planner template: %w", err)
	}

	llmResponseContent, err := h.callLLM(ctx, promptBuffer.String(), true)
	if err != nil {
		return nil, err
	}
	
	cleanedJSON := strings.TrimSpace(llmResponseContent)
	cleanedJSON = strings.TrimPrefix(cleanedJSON, "```json")
	cleanedJSON = strings.TrimSuffix(cleanedJSON, "```")
	
	var plannerResponse PlannerResponse
	if err := json.Unmarshal([]byte(cleanedJSON), &plannerResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool call plan from LLM content: %w. Raw content: %s", err, llmResponseContent)
	}

	return plannerResponse.ToolCalls, nil
}

func (h *DemoHandler) getContextFromPlan(ctx context.Context, plan []ToolCall) (*HybridContext, error) {
	var hybridCtx HybridContext
	reqLogger := h.logger.With("plan_execution", true)
	
	for _, toolCall := range plan {
		switch toolCall.ToolName {
		case "get_mission_facts":
			missionName, ok := toolCall.Arguments["mission_name"]
			if !ok {
				return nil, fmt.Errorf("missing 'mission_name' argument for get_mission_facts")
			}
			facts, err := h.queries.GetMissionFacts(ctx, missionName)
			if err != nil {
				reqLogger.ErrorContext(ctx, "Failed to execute GetMissionFacts", "error", err, "mission_name", missionName)
				continue 
			}
			hybridCtx.MissionFacts = &facts
			reqLogger.InfoContext(ctx,"Executed tool: get_mission_facts", "result", facts)

		case "find_mission_context":
			searchQuery, ok := toolCall.Arguments["search_query"]
			if !ok {
				return nil, fmt.Errorf("missing 'search_query' argument for find_mission_context")
			}
			embedding, err := h.getEmbedding(ctx, searchQuery)
			if err != nil {
				return nil, fmt.Errorf("failed to get embedding for context search: %w", err)
			}
			chunks, err := h.queries.FindSimilarMissionKnowledge(ctx, pgvector.NewVector(embedding))
			if err != nil {
				reqLogger.ErrorContext(ctx, "Failed to execute FindSimilarMissionKnowledge", "error", err, "search_query", searchQuery)
				continue
			}
			hybridCtx.KnowledgeChunks = chunks
			reqLogger.InfoContext(ctx, "Executed tool: find_mission_context", "results_found", len(chunks))
		}
	}
	return &hybridCtx, nil
}

func (h *DemoHandler) synthesizeAnswer(ctx context.Context, question string, context *HybridContext) (string, error) {
	h.logger.InfoContext(ctx, "Synthesizing final answer from hybrid context...")
	
	factsJSON, _ := json.MarshalIndent(context.MissionFacts, "", "  ")

	var chunksText []string
	for _, chunk := range context.KnowledgeChunks {
		chunksText = append(chunksText, chunk.ChunkText)
	}

	templateData := struct {
		UserQuestion    string
		StructuredData  string
		NarrativeChunks []string
	}{
		UserQuestion:    question,
		StructuredData:  string(factsJSON),
		NarrativeChunks: chunksText,
	}
	
	var promptBuffer bytes.Buffer
	if err := h.synthesizerTemplate.Execute(&promptBuffer, templateData); err != nil {
		return "", fmt.Errorf("failed to execute synthesizer template: %w", err)
	}
	
	finalAnswer, err := h.callLLM(ctx, promptBuffer.String(), false)
	if err != nil {
		return "", err
	}

	return finalAnswer, nil
}

// CORRECTED: getEmbedding now correctly marshals the request and handles errors.
func (h *DemoHandler) getEmbedding(ctx context.Context, textToEmbed string) ([]float32, error) {
	reqBody, err := json.Marshal(EmbeddingRequest{Text: textToEmbed})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", h.embeddingServiceURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding service returned non-OK status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("failed to decode embedding response: %w", err)
	}

	return embeddingResp.Embedding, nil
}

func (h *DemoHandler) callLLM(ctx context.Context, prompt string, useJSONMode bool) (string, error) {
	apiKey := h.openAIAPIKey
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	payload := LLMRequestBody{
		Model: "gpt-4o",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}
	if useJSONMode {
		payload.ResponseFormat = &ResponseFormat{Type: "json_object"}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create OpenAI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API returned non-OK status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var llmResponse LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResponse); err != nil {
		return "", fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	if len(llmResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from OpenAI")
	}

	return llmResponse.Choices[0].Message.Content, nil
}
