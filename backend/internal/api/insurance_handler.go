// backend/internal/api/insurance_handler.go

package api

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"encoding/json"
	"strings"
	"strconv"
	"time"
	"fmt"
	"sort"
	"io"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jjckrbbt/catalyst/backend/internal/apps/insurance"
	"github.com/jjckrbbt/catalyst/backend/internal/repository"
	"github.com/pgvector/pgvector-go"
	"github.com/labstack/echo/v4"
)

// --- Structs for RAG pipeline ---
type PlannerResponse struct {
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	ToolName	string			`json:"tool"`
	Arguments	map[string]string	`json:"arguments"`
}

type InsuranceContext struct {
	ClaimsData		[]insurance.ListClaimsRow
	KnowledgeChunks		[]insurance.SearchInsuranceContextRow
}

type LLMRequestBody struct {
	Model		string		`json:"model"`
	Messages	[]Message	`json:"messages"`
	ResponseFormat	*ResponseFormat	`json:"response_format,omitempty"`
}

type Message struct {
	Role	string	`json:"role"`
	Content	string	`json:"content"`
}

type ResponseFormat struct {
	Type	string	`json:"type"`
}

type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type SearchResult struct {
	Source		string
	Text		string
	SimilarityScore	float32
}

// InsuranceHandler is the handler for our new insurance application module.
type InsuranceHandler struct {
	queries *insurance.Queries
	platformQuerier repository.Querier
	httpClient  *http.Client
	embeddingServiceURL string
	plannerTemplate		*template.Template
	synthesizerTemplate	*template.Template
	openAIAPIKey		string
	logger  *slog.Logger
}

type UpdateClaimRequest struct {
	BusinessStatus string `json:"business_status"`
}

type CreateCommentRequest struct {
	CommentText string `json:"comment_text"`
}

type EmbeddingRequest struct {
	Text string `json:"text"`
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// NewInsuranceHandler creates a new instance of the InsuranceHandler.
func NewInsuranceHandler(q *insurance.Queries, pq repository.Querier, apiKey string, logger *slog.Logger) *InsuranceHandler {
	plannerTmpl, err := template.ParseFiles("backend/configs/prompts/apps/insurance/insurance_planner_prompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse insuance planner template: %w", err)
	}
	
	synthesizerTmpl, err := template.ParseFiles("backend/configs/prompts/apps/insurance/synthesizer_prompt.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse insurance synthesizer template: %w", err)
	}

	return &InsuranceHandler{
		queries: q,
		platformQuerier: pq,
		httpClient:          &http.Client{Timeout: 30 * time.Second},
		embeddingServiceURL: "http://localhost:5001/embed",
		plannerTemplate:	plannerTmpl,
		synthesizerTemplate:	synthesizerTmpl,
		openAIAPIKey:		apiKey,
		logger:  logger.With("component", "insurance_handler"),
	}
}

// HandleListClaims retrieves a paginated and filtered list of insurance claims.
func (h *InsuranceHandler) HandleListClaims(c echo.Context) error {
	ctx := c.Request().Context()

	// --- Parse Pagination and Filtering Parameters ---
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 32)
	if limit <= 0 {
		limit = 50 // Default limit
	}

	page, _ := strconv.ParseInt(c.QueryParam("page"), 10, 32)
	if page <= 0 {
		page = 1 // Default page
	}

	offset := (page - 1) * limit

	// --- Build the Params Struct for sqlc ---
	params := insurance.ListClaimsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
		AdjusterAssigned: pgtype.Text{String: c.QueryParam("adjuster_assigned"), Valid: c.QueryParam("adjuster_assigned") != ""},
		Status:           pgtype.Text{String: c.QueryParam("status"), Valid: c.QueryParam("status") != ""},
		PolicyNumber:     pgtype.Text{String: c.QueryParam("policy_number"), Valid: c.QueryParam("policy_number") != ""},
	}

	// --- Execute the Query ---
	claims, err := h.queries.ListClaims(ctx, params)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to list insurance claims", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve claims")
	}

	h.logger.InfoContext(ctx, "Successfully retrieved claims list", "count", len(claims))
	return c.JSON(http.StatusOK, claims)
}

// HandleListPolicyholders retrieves a paginated and filtered list of policyholders.
func (h *InsuranceHandler) HandleListPolicyholders(c echo.Context) error {
    ctx := c.Request().Context()

    limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 32)
    if limit <= 0 {
        limit = 50
    }
    page, _ := strconv.ParseInt(c.QueryParam("page"), 10, 32)
    if page <= 0 {
        page = 1
    }
    offset := (page - 1) * limit

    params := insurance.ListPolicyholdersParams{
        Limit:          int32(limit),
        Offset:         int32(offset),
        State:          pgtype.Text{String: c.QueryParam("state"), Valid: c.QueryParam("state") != ""},
        CustomerLevel:  pgtype.Text{String: c.QueryParam("customer_level"), Valid: c.QueryParam("customer_level") != ""},
    }

    policyholders, err := h.queries.ListPolicyholders(ctx, params)
    if err != nil {
        h.logger.ErrorContext(ctx, "Failed to list policyholders", "error", err)
        return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve policyholders")
    }

    h.logger.InfoContext(ctx, "Successfully retrieved policyholders list", "count", len(policyholders))
    return c.JSON(http.StatusOK, policyholders)
}

// HandleGetClaimDetails retrieves a single, detailed claim record joined with policyholdeer info
func (h *InsuranceHandler) HandleGetClaimDetails(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid claim ID format")
	}

	claimDetails, err := h.queries.GetClaimDetails(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get claim details", "error", err, "claim_id", id)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve claim details")
	}

	h.logger.InfoContext(ctx, "Successfully retrieved claim details", "claim_id", id)
	return c.JSON(http.StatusOK, claimDetails)
}

// HandleGetClaimStatusHistory retrieves the business status history for a single claim
func (h *InsuranceHandler) HandleGetClaimStatusHistory(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid claim ID format")
	}

	history, err := h.queries.GetClaimStatusHistory(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to get claim status history", "error", err, "claim_id", id)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve claim history")
	}

	type HistoryResponse struct {
		ID		int64		`json:"ID"`
		EventTimestamp	time.Time	`json:"event_timestamp"`
		EventData	json.RawMessage	`json:"event_data"`
		UserName	pgtype.Text	`json:"user_name"`
	}

	response := make([]HistoryResponse, len(history))
	for i, event := range history {
		response[i] = HistoryResponse{
			ID:		event.EventID,
			EventTimestamp:	event.EventTimestamp.Time,
			EventData:	event.EventData,
			UserName:	event.UserName,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// HandleUpdateClaim retrieves claim, updates fields, and inserts back to db
func (h *InsuranceHandler) HandleUpdateClaim(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid claim ID format")
	}

	var req UpdateClaimRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	var userID int64 = 1

	existingItem, err := h.platformQuerier.GetItemForUpdate(ctx, id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Item not found")
	}

	var customProps map[string]interface{}
	if err := json.Unmarshal(existingItem.CustomProperties, &customProps); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse existing item properties")
	}

	oldStatus := customProps["Status"]
	customProps["Status"] = req.BusinessStatus

	updatedCustomProps, err := json.Marshal(customProps)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to serialize updated properties")
	}

	updateParams := repository.UpdateItemParams{
		ID:	id,
		Scope:	existingItem.Scope,
		Status:	existingItem.Status,
		CustomProperties: updatedCustomProps,
	}

	_, err = h.platformQuerier.UpdateItem(ctx, updateParams)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to update item", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update claim")
	}

	eventData := map[string]interface{}{
		"old_status": oldStatus,
		"new_status": req.BusinessStatus,
	}
	eventDataJSON, _ := json.Marshal(eventData)

	eventParams := repository.CreateItemEventParams{
		ItemID:		id, 
		EventType:	"CLAIM_STATUS_CHANGED",
		EventData:	eventDataJSON,
		CreatedBy:	userID,
	}
	
	_, err = h.platformQuerier.CreateItemEvent(ctx, eventParams)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to create status change event", "error", err)
		echo.NewHTTPError(http.StatusInternalServerError, "Failed to create audit event for claim update")
	}

	return c.NoContent(http.StatusNoContent)
}

// HandleListComments retrieves all comments for a specific claim
func (h *InsuranceHandler) HandleListComments(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid claim ID format")
	}

	comments, err := h.platformQuerier.ListCommentsForItem(ctx, id)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to list comments for item", "error", err, "item_id", id)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve comments")
	}

	return c.JSON(http.StatusOK, comments)
}

// HandleCreateComment adds a new comment to a specific claim
func (h *InsuranceHandler) HandleCreateComment(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid claim ID format")
	}

	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil || req.CommentText == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: comment_text is required")
	}

	var userID int64 = 1

	params := repository.CreateCommentParams{
		ItemID:		id, 
		Comment:	req.CommentText,
		UserID:		userID,
	}

	newComment, err := h.platformQuerier.CreateComment(ctx, params)
	if err != nil {
		h.logger.ErrorContext(ctx, "Failed to create comment", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save comment")
	}

	// TODO: in production, make a background job
	embedding, err := h.getEmbedding(ctx, newComment.Comment)
	if err != nil {
		h.logger.WarnContext(ctx, "Failed to generate embedding for comment, but comment was saved", "error", err, "comment_id", newComment.ID)
	} else {
		updateEmbeddingParams := repository.SetCommentEmbeddingParams{
			ID:		newComment.ID,
			Embedding:	pgvector.NewVector(embedding),
		}
		err = h.platformQuerier.SetCommentEmbedding(ctx, updateEmbeddingParams)
		if err != nil {
			h.logger.ErrorContext(ctx, "Failed to save embedding for comment", "error", err, "comment_id", newComment.ID)
		}
	}
	// TODO: parse comment for user mentions and call AddMentionToComment in loop

	return c.JSON(http.StatusCreated, newComment)
}

// getEmbedding gets a vector embedding from python microservice to power similarity search
func (h *InsuranceHandler) getEmbedding(ctx context.Context, textToEmbed string) ([]float32, error) {
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

// --- RAG Handler ---
func (h *InsuranceHandler) HandleInsuranceQuery(c echo.Context) error {
	ctx := c.Request().Context()
	question := c.FormValue("question")
	if question == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "form value 'question' is required")
	}

	plan, err := h.getExecutionPlan(ctx, question)
	if err != nil {
		h.logger.ErrorContext(ctx, "RAG Error: Failed to get execution plan", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error planning query")
	}

	contextData, err := h.getContextFromPlan(ctx, plan)
	if err != nil {
		h.logger.ErrorContext(ctx, "RAG Error: Failed to execute plan", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error executing plan")
	}

	finalAnswer, err := h.synthesizeAnswer(ctx, question, contextData)
	if err != nil {
		h.logger.ErrorContext(ctx, "RAG Error: Failed to synthesize answer", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error synthesizing answer")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"answer": finalAnswer})
}

// --- Helper functions for RAG ---

func (h *InsuranceHandler) getExecutionPlan(ctx context.Context, question string) ([]ToolCall, error) {
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

func (h *InsuranceHandler) getContextFromPlan(ctx context.Context, plan []ToolCall) (*InsuranceContext, error) {
	var insuranceCtx InsuranceContext

	for _, toolCall := range plan {
		switch toolCall.ToolName {
		case "find_context_in_documents":
			searchQuery, ok := toolCall.Arguments["search_query"]
			if !ok {
				return nil, fmt.Errorf("missing 'search_query' argument for find_context_in_documents")
			}
			embedding, err := h.getEmbedding(ctx, searchQuery)
			if err != nil {
				return nil, fmt.Errorf("failed to get embedding for context search: %w", err)
			}

			pgVec :=pgvector.NewVector(embedding)


			knowledgeChunks, err1 := h.queries.SearchKnowledgeChunks(ctx,pgVec))
			comments, err2 := h.queries(SearchComments(ctx, pgVec)
			if err1 != nil { hlogger.ErrorContext(ctx, "Failed to search knowledge chunks", "error", err1)}
			if err2 != nil { hlogger.ErrorContext(ctx, "Failed to search comments", "error", err2)}
			
			var combinedResults []SearchResult
			for _, chunk := range knowledgeChunks {
				combinedResults = append(combinedResults, SearchResults{
					Source:		chunk.Source,
					Text:		chunk.Text,
					SimilarityScore:	chunk.SimilarityScore.(float32),
				})
			}
			for _, comment := range comments {
				combinedResults = append(combinedResults, SearchResults{
					Source:		comment.Source,
					Text:		comment.Text,
					SimilarityScore:	comment.SimilarityScore.(float32),
				})
			}
			sort.Slice(combinedResults, func(i, j int) bool {
				return combinedResults[i].SimilarityScore < combinedResults[j].SimilarityScore
			})
			
			topResults := combinedResults
			if len(topResults) > 5 {
				topResults = topResults[:5]
			}
			var finalChunks []insurance.SearchInsuranceContextRow
			for _, res := range topResults {
				finalChunks = append(finalChunks, insurance.SearchInsuranceContextRow{
					Source:		res.Source,
					Text:		res.Text,
					SimilarityScore:	res.SimilarityScore,
				})
			}
			
			insuranceCtx.KnowledgeChunks = finalChunks
		}
	}
	return &insuranceCtx, nil
}

func (h *InsuranceHandler) synthesizeAnswer(ctx context.Context, question string, context *InsuranceContext) (string, error) {
	h.logger.InfoContext(ctx, "Synthesizing final answer from hybrid context...")
	
	claimsJSON, _ := json.MarshalIndent(context.ClaimsData, "", "  ")

	var chunksText []string
	for _, chunk := range context.KnowledgeChunks {
		chunksText = append(chunksText, chunk.Text)
	}

	templateData := struct {
		UserQuestion    string
		StructuredData  string
		NarrativeChunks []string
	}{
		UserQuestion:    question,
		StructuredData:  string(claimsJSON),
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

func (h *InsuranceHandler) callLLM(ctx context.Context, prompt string, useJSONMode bool) (string, error) {
	h.logger.InfoContext(ctx, "Executing LLM call", "prompt", prompt)

	apiKey := h.openAIAPIKey
	if apiKey == "" {
		return "", fmt.Errorf("OpenAI key is not configured on the handler")
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



