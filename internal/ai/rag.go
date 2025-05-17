package ai

import (
	"context"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

func NewModel(serverURL, model string) (*ollama.LLM, error) {
	return ollama.New(ollama.WithModel(model), ollama.WithServerURL(serverURL))
}

func NewEmbedder(serverURL string, model string) (*embeddings.EmbedderImpl, error) {
	llm, err := ollama.New(ollama.WithModel(model), ollama.WithServerURL(serverURL))
	if err != nil {
		return nil, err
	}
	return embeddings.NewEmbedder(llm)
}

func NewStore(ctx context.Context, storeURL string, embedder embeddings.Embedder) (*pgvector.Store, error) {
	s, err := pgvector.New(ctx, pgvector.WithConnectionURL(storeURL), pgvector.WithEmbedder(embedder), pgvector.WithPreDeleteCollection(true))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func StoreDocuments(ctx context.Context, docs []schema.Document, store *pgvector.Store) error {
	if len(docs) == 0 {
		return nil
	}
	_, err := store.AddDocuments(ctx, docs)
	return err
}

//func RetrieveRelevantDocuments(ctx context.Context, store *pgvector.Store, prompt string, numDocuments int) ([]schema.Document, error) {
//	opts := []vectorstores.Option{vectorstores.WithScoreThreshold(0.8)}
//	retriever := vectorstores.ToRetriever(store, numDocuments, opts...)
//	return retriever.GetRelevantDocuments(ctx, prompt)
//}
//
//func GenerateHistory(ctx context.Context, relevantDocs []schema.Document, chatHistoryDocs []schema.Document) (*memory.ChatMessageHistory, error) {
//	history := memory.NewChatMessageHistory()
//	if len(relevantDocs) > 0 {
//		for _, doc := range relevantDocs {
//			if err := history.AddMessage(ctx, llms.SystemChatMessage{Content: doc.PageContent}); err != nil {
//				return nil, err
//			}
//		}
//	}
//	if len(chatHistoryDocs) > 0 {
//		for _, doc := range chatHistoryDocs {
//			role := doc.Metadata["role"].(string)
//			switch role {
//			case "user":
//				_ = history.AddUserMessage(ctx, doc.PageContent)
//			case "assistant":
//				_ = history.AddAIMessage(ctx, doc.PageContent)
//			}
//		}
//	}
//	return history, nil
//}
//
//func Execute(ctx context.Context, history *memory.ChatMessageHistory, llm llms.Model, prompt string) (string, error) {
//	conversation := memory.NewConversationBuffer(memory.WithChatHistory(history))
//	executor := agents.NewExecutor(agents.NewConversationalAgent(llm, nil), agents.WithMemory(conversation))
//	options := []chains.ChainCallOption{chains.WithTemperature(0.8)}
//
//	res, err := chains.Run(ctx, executor, prompt, options...)
//	if err != nil {
//		return "", err
//	}
//	return res, nil
//}
