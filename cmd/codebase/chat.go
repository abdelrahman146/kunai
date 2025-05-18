package codebase

import (
	"context"
	"fmt"
	"github.com/abdelrahman146/kunai/internal/ai"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/prompts"
	"path/filepath"
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
	chatCmd.Flags().StringVarP(&chatCmdParams.Model, "model", "m", "deepseek-r1:14b", "Specify the LLM model")
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
	utils.RunWithSpinner(fmt.Sprintf("Scanning: %s", filepath.Base(chatCmdParams.ContextDir)), func() {
		err = ai.ScanProject(chatCmdParams.ContextDir, 4000, 200, store)
	})
	if err != nil {
		return err
	}
	basePrompt := chatCmdBasePrompt()
	historyPrompt := chatCmdHistoryPrompt()
	qaChain, convMem := ai.NewConversationRetriever(store, llm, chatCmdParams.MaxRelevantDocs, basePrompt, historyPrompt)
	fmt.Println("Ready! You can now ask questions about this project.")
	// Start REPL
	utils.RunREPL(func(input string) (response any, err error) {
		memVars, err := convMem.LoadMemoryVariables(ctx, nil)
		if err != nil {
			return "", err
		}
		inputs := map[string]any{
			"question": input,
			"history":  memVars["history"],
		}
		var answer string

		utils.RunWithSpinner("Thinking...", func() {
			tc, cancel := context.WithTimeout(ctx, 5*time.Minute)
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
	return prompts.PromptTemplate{
		Template: `
SYSTEM: You are a universal code assistant. You can handle tasks such as:
- Adding or enhancing features
- Code reviews and analysis
- Explaining code and flows
- Refactoring and optimization
- Performance improvements
- Architectural suggestions
Answer using ONLY the provided code context; if none applies, reply exactly:
"I can't answer this because it is outside the context."

CODE CONTEXT:
{{.context}}

USER QUESTION:
{{.question}}`,
		InputVariables: []string{"context", "question"},
		TemplateFormat: prompts.TemplateFormatGoTemplate,
	}
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
