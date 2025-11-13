package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/gogf/gf/v2/frame/g"
	v1 "github.com/wangle201210/go-rag/server/api/rag/v1"
	"github.com/wangle201210/wachat/backend/config"
)

// GoRagAPIResponse go-rag 标准 API 响应结构（泛型）
type GoRagAPIResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// RAGServiceImpl 提供 RAG (Retrieval Augmented Generation) 功能
// 通过 HTTP 调用 go-rag 服务器的 RESTful API
type RAGServiceImpl struct {
	ctx        context.Context
	config     *config.RAGConfig
	httpClient *http.Client
	baseURL    string
}

// NewRAGService 创建 RAG 服务（通过 HTTP API）
func NewRAGService(ctx context.Context, cfg *config.RAGConfig, aiCfg *config.AIConfig) (*RAGServiceImpl, error) {
	if cfg == nil || !cfg.Enabled {
		g.Log().Info(ctx, "RAG service is disabled")
		return &RAGServiceImpl{
			ctx:    ctx,
			config: cfg,
		}, nil
	}

	// 检查 go-rag 服务器配置
	if cfg.Server == nil || cfg.Server.Address == "" {
		g.Log().Info(ctx, "RAG service: go-rag server is not configured, RAG functions will be unavailable")
		return &RAGServiceImpl{
			ctx:    ctx,
			config: cfg,
		}, nil
	}

	// 构建 base URL
	baseURL := fmt.Sprintf("http://localhost%s/api", cfg.Server.Address)
	g.Log().Infof(ctx, "RAG service will use go-rag server at: %s", baseURL)

	return &RAGServiceImpl{
		ctx:        ctx,
		config:     cfg,
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}, nil
}

// callAPI 通用的 HTTP API 调用方法（泛型）
func (r *RAGServiceImpl) callAPI(ctx context.Context, method, path string, reqBody interface{}, respData interface{}) error {
	url := fmt.Sprintf("%s%s", r.baseURL, path)

	var req *http.Request
	var err error

	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call go-rag API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("go-rag API error: %s, body: %s", resp.Status, string(body))
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// 先解析通用响应结构检查错误
	var commonResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &commonResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if commonResp.Code != 0 {
		return fmt.Errorf("go-rag API error: %s", commonResp.Message)
	}

	// 解析完整响应到目标结构
	if err := json.Unmarshal(body, respData); err != nil {
		return fmt.Errorf("failed to decode response data: %w", err)
	}
	g.Log().Debugf(ctx, "RAG service response: %s", string(body))
	return nil
}

// IsEnabled 检查 RAG 服务是否启用
func (r *RAGServiceImpl) IsEnabled() bool {
	return r.config != nil && r.config.Enabled && r.baseURL != ""
}

// CheckHealth 检查 RAG 服务是否健康（检测端口）
func (r *RAGServiceImpl) CheckHealth() error {
	if !r.IsEnabled() {
		return fmt.Errorf("RAG service is not enabled")
	}

	if r.config.Server == nil || r.config.Server.Address == "" {
		return fmt.Errorf("server address not configured")
	}

	// 解析地址（如 ":8000"）
	address := r.config.Server.Address
	if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}

	// 尝试连接（快速超时，避免阻塞对话）
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return fmt.Errorf("cannot connect to go-rag server: %w", err)
	}
	conn.Close()
	return nil
}

// isHealthy 内部方法：检查服务是否健康（不抛错，只返回 true/false）
func (r *RAGServiceImpl) isHealthy() bool {
	return r.CheckHealth() == nil
}

// GetKnowledgeBases 获取所有知识库列表
// 直接返回 go-rag API 的完整响应数据
func (r *RAGServiceImpl) GetKnowledgeBases(ctx context.Context) (*v1.KBGetListRes, error) {
	if !r.IsEnabled() {
		return nil, fmt.Errorf("RAG service is not enabled")
	}

	// 调用 go-rag API: GET /v1/kb
	var apiResp GoRagAPIResponse[v1.KBGetListRes]
	if err := r.callAPI(ctx, "GET", "/v1/kb", nil, &apiResp); err != nil {
		return nil, err
	}

	g.Log().Debugf(ctx, "Retrieved %d knowledge bases", len(apiResp.Data.List))
	return &apiResp.Data, nil
}

// Retrieve 从知识库检索相关文档
// knowledgeName: 知识库名称（必填）
// scoreThreshold: 分数阈值，范围 0-2，建议使用 1.3-1.5
// 返回值直接使用 go-rag 的 schema.Document，与 go-rag 保持一致
func (r *RAGServiceImpl) Retrieve(ctx context.Context, query string, knowledgeName string, scoreThreshold float64) ([]*schema.Document, error) {
	if !r.IsEnabled() {
		return nil, fmt.Errorf("RAG service is not enabled")
	}

	if knowledgeName == "" {
		return nil, fmt.Errorf("knowledgeName is required")
	}

	if scoreThreshold == 0 {
		scoreThreshold = 0.2 // go-rag 默认阈值
	}

	topK := r.config.TopK
	if topK == 0 {
		topK = 5
	}

	g.Log().Infof(ctx, "Retrieving documents for query: %s, knowledge: %s", query, knowledgeName)

	// 调用 go-rag API: POST /v1/retriever
	reqBody := map[string]interface{}{
		"question":       query,
		"top_k":          topK,
		"score":          scoreThreshold,
		"knowledge_name": knowledgeName,
	}

	var apiResp GoRagAPIResponse[v1.RetrieverRes]
	if err := r.callAPI(ctx, "POST", "/v1/retriever", reqBody, &apiResp); err != nil {
		return nil, err
	}

	// 直接返回 go-rag 的 Document 列表
	g.Log().Debugf(ctx, "Retrieved %d documents", len(apiResp.Data.Document))
	for i, doc := range apiResp.Data.Document {
		g.Log().Debugf(ctx, "Document[%d]: ID=%s, Score=%.2f, Content=%s", i, doc.ID, doc.Score(), doc.Content)
	}

	return apiResp.Data.Document, nil
}

// RetrieveDocuments 检索文档并返回原始文档列表
// 用于在 UI 中展示检索到的文档
func (r *RAGServiceImpl) RetrieveDocuments(ctx context.Context, query string) ([]*schema.Document, error) {
	if !r.IsEnabled() {
		return nil, nil // 如果未启用，返回空列表
	}

	// 健康检查：如果服务不健康，直接返回空列表，不影响对话
	if !r.isHealthy() {
		g.Log().Debug(ctx, "RAG service is not healthy, skipping retrieval")
		return nil, nil
	}

	// 检查是否配置了默认知识库
	if r.config.DefaultKnowledgeBase == "" {
		g.Log().Debug(ctx, "DefaultKnowledgeBase not configured, skipping RAG retrieval")
		return nil, nil
	}

	// 使用默认知识库和阈值
	results, err := r.Retrieve(ctx, query, r.config.DefaultKnowledgeBase, 1.3)
	if err != nil {
		g.Log().Warningf(ctx, "Failed to retrieve documents: %v", err)
		return nil, nil // 检索失败返回空列表，不影响对话
	}

	return results, nil
}

// RetrieveWithContext 检索文档并返回上下文信息
// 这可以用于增强 AI 的回答
func (r *RAGServiceImpl) RetrieveWithContext(ctx context.Context, query string) (string, error) {
	// 先检查服务是否健康，不健康直接返回空
	if !r.IsEnabled() || !r.isHealthy() {
		g.Log().Debug(ctx, "RAG service not available, skipping context retrieval")
		return "", nil
	}

	results, err := r.RetrieveDocuments(ctx, query)
	if err != nil || len(results) == 0 {
		return "", nil
	}

	// 将检索到的文档组合成上下文
	contextStr := "以下是相关的知识库信息：\n\n"
	for i, doc := range results {
		contextStr += fmt.Sprintf("%d. [相关度: %.2f] %s\n\n", i+1, doc.Score(), doc.Content)
	}
	g.Log().Debugf(ctx, "Generated context: %s", contextStr)
	return contextStr, nil
}

// IndexDocumentFromURI 从文件或 URL 索引文档到知识库
// uri: 文档地址，可以是文件路径（pdf、html、md等）或网址
// knowledgeName: 知识库名称（必填）
// 返回值直接使用 go-rag 的 IndexerRes，包含所有生成的文档 ID
func (r *RAGServiceImpl) IndexDocumentFromURI(ctx context.Context, uri string, knowledgeName string) (*v1.IndexerRes, error) {
	if !r.IsEnabled() {
		return nil, fmt.Errorf("RAG service is not enabled")
	}

	if knowledgeName == "" {
		return nil, fmt.Errorf("knowledgeName is required")
	}

	g.Log().Infof(ctx, "Indexing document from URI: %s to knowledge base: %s", uri, knowledgeName)

	// 调用 go-rag API: POST /v1/indexer
	// 注意：这个 API 使用 multipart/form-data，但我们简化使用 JSON
	reqBody := map[string]interface{}{
		"url":            uri,
		"knowledge_name": knowledgeName,
	}

	var apiResp GoRagAPIResponse[v1.IndexerRes]
	if err := r.callAPI(ctx, "POST", "/v1/indexer", reqBody, &apiResp); err != nil {
		return nil, err
	}

	g.Log().Infof(ctx, "Successfully indexed document, generated %d chunks", len(apiResp.Data.DocIDs))
	return &apiResp.Data, nil
}

// IndexText 直接索引文本内容到知识库
// content: 要索引的文本内容
// knowledgeName: 知识库名称，如果为空则使用配置中的 indexName
// 注意：go-rag HTTP API 不直接支持文本索引，需要通过文件上传
func (r *RAGServiceImpl) IndexText(ctx context.Context, content string, knowledgeName string, metadata map[string]any) (*v1.IndexerRes, error) {
	if !r.IsEnabled() {
		return nil, fmt.Errorf("RAG service is not enabled")
	}

	// go-rag HTTP API 不支持直接索引文本
	// 需要先将文本保存为文件再通过 /v1/indexer 上传
	return nil, fmt.Errorf("IndexText is not supported when using go-rag HTTP API, please use IndexDocumentFromURI instead")
}

// DeleteDocument 删除文档
// documentID: 文档在数据库中的 ID (go-rag documents 表的 ID)
func (r *RAGServiceImpl) DeleteDocument(ctx context.Context, documentID string) error {
	if !r.IsEnabled() {
		return fmt.Errorf("RAG service is not enabled")
	}

	g.Log().Infof(ctx, "Deleting document: %s", documentID)

	// 调用 go-rag API: DELETE /v1/documents?document_id={id}
	path := fmt.Sprintf("/v1/documents?document_id=%s", documentID)

	var apiResp GoRagAPIResponse[v1.DocumentsDeleteRes]
	if err := r.callAPI(ctx, "DELETE", path, nil, &apiResp); err != nil {
		return err
	}

	g.Log().Infof(ctx, "Successfully deleted document: %s", documentID)
	return nil
}
