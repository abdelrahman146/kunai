package codebase

import (
	"context"
	"github.com/abdelrahman146/kunai/internal/ai"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/prompts"
	"time"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Launch an interactive multi-task REPL with project context via Ollama",
	Long: `Starts a REPL where each question is sent to a local Ollama model
with the current project context and tasks prepended. Tasks can include:
- Adding new features
- Code analysis & reviews
- Explaining code & workflows
- Refactoring & optimization
- Enhancements & bug fixes
- Writing tests & documentation
Use 'exit', 'quit' or Ctrl+C to quit.`,
	RunE: runChatCmd,
}

var chatCmdParams struct {
	ContextDir         string
	Model              string
	EmbedModel         string
	MaxRelevantDocs    int
	MaxChatHistoryDocs int
	OllamaBaseURL      string
	VectorStoreURL     string
}

func init() {
	chatCmd.Flags().StringVarP(&chatCmdParams.ContextDir, "context-dir", "c", "", "Specify the context directory")
	chatCmd.Flags().StringVarP(&chatCmdParams.Model, "model", "m", "gemma3:12b", "Specify the LLM model")
	chatCmd.Flags().StringVarP(&chatCmdParams.EmbedModel, "embed-model", "e", "bge-m3", "Specify the embedding model")
	chatCmd.Flags().IntVar(&chatCmdParams.MaxChatHistoryDocs, "max-history-docs", 10, "Specify the max history docs")
	chatCmd.Flags().IntVar(&chatCmdParams.MaxRelevantDocs, "max-relevant-docs", 10, "Specify the max relevant docs")
	chatCmd.Flags().StringVar(&chatCmdParams.OllamaBaseURL, "ollama-url", "http://localhost:11434", "Ollama base url")
	chatCmd.Flags().StringVar(&chatCmdParams.VectorStoreURL, "vector-store-url", "postgres://postgres:postgres@localhost:5432/kunai", "Postgres Vector Store URL")
}

func runChatCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	var err error
	// Resolve context directory
	if chatCmdParams.ContextDir == "" {
		chatCmdParams.ContextDir, err = utils.FindRepoRoot()
	} else {
		chatCmdParams.ContextDir, err = utils.GetAbsPath(chatCmdParams.ContextDir)
	}
	if err != nil {
		return err
	}
	// Initialize LLM, embedder, and vector store
	llm, err := ai.NewModel(chatCmdParams.OllamaBaseURL, chatCmdParams.Model)
	if err != nil {
		return err
	}
	emb, err := ai.NewEmbedder(chatCmdParams.OllamaBaseURL, chatCmdParams.EmbedModel)
	if err != nil {
		return err
	}
	store, err := ai.NewStore(ctx, chatCmdParams.VectorStoreURL, emb)
	if err != nil {
		return err
	}

	// scan project and embed vectors
	docs, err := ai.ScanProject(chatCmdParams.ContextDir)
	if err != nil {
		return err
	}
	if err = ai.StoreDocuments(ctx, docs, store); err != nil {
		return err
	}

	basePrompt := chatCmdBasePrompt()
	historyPrompt := chatCmdHistoryPrompt()
	qaChain, convMem := ai.NewConversationRetriever(store, llm, chatCmdParams.MaxRelevantDocs, basePrompt, historyPrompt)

	// Start REPL
	utils.RunREPL(func(input string) (response any, err error) {
		// Retrieve relevant docs
		inputs := map[string]any{
			"question": input, // your prompt
		}
		memVars, err := convMem.LoadMemoryVariables(ctx, nil)
		if err != nil {
			return "", err
		}
		inputs["history"] = memVars["history"]
		var answer string

		utils.RunWithSpinner("Thinking...", func() {
			tc, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			out, outErr := qaChain.Call(tc, inputs)
			err = outErr
			if outErr != nil {
				return
			}
			answer = out["text"].(string)
		})
		if err != nil {
			return "", err
		}
		saveIn := map[string]any{"question": input}
		saveOut := map[string]any{"text": answer}
		if err := convMem.SaveContext(ctx, saveIn, saveOut); err != nil {
			return "", err
		}
		// return answer to be presented
		return answer, nil
	})
	return nil
}

func chatCmdBasePrompt() prompts.PromptTemplate {
	basePrompt := prompts.PromptTemplate{
		Template: `
SYSTEM: You are an AI assistant expert in software development and engineering, your task is to answer development questions EXCLUSIVELY for this codebase.  
1. If a question is not answerable from the context, reply exactly:  
   "I can't answer this because it is outside the context."  
2. Otherwise, answer concisely.

Context:
{{.context}}

User’s question:
{{.question}}

Answer:`,
		InputVariables: []string{"context", "question"},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	return basePrompt
}

func chatCmdHistoryPrompt() prompts.PromptTemplate {
	historyPrompt := prompts.PromptTemplate{
		Template: `
Given the conversation so far:
{{.chat_history}}

Rewrite the follow-up question so it can be understood on its own:
{{.question}}`,
		InputVariables: []string{"chat_history", "question"},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
	return historyPrompt
}
