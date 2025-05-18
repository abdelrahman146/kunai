package ai

import (
	"context"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
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

func StoreDocuments(ctx context.Context, docs []schema.Document, store vectorstores.VectorStore) error {
	var err error
	utils.RunWithSpinner("Embedding...", func() {
		if len(docs) == 0 {
			return
		}
		_, err = store.AddDocuments(ctx, docs)
	})
	return err
}

func NewConversationRetriever(store vectorstores.VectorStore, llm llms.Model, topK int, basePrompt prompts.PromptTemplate, historyPrompt prompts.PromptTemplate) (chains.ConversationalRetrievalQA, *memory.ConversationBuffer) {
	retriever := vectorstores.ToRetriever(store, topK)
	convMem := memory.NewConversationBuffer(memory.WithReturnMessages(true))
	llmChain := chains.NewLLMChain(llm, basePrompt)
	combineChain := chains.NewStuffDocuments(llmChain)
	condenseChain := chains.NewLLMChain(llm, historyPrompt)
	qaChain := chains.NewConversationalRetrievalQA(combineChain, condenseChain, retriever, convMem)
	return qaChain, convMem
}
