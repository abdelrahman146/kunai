package codebase

import (
	"context"
	"github.com/abdelrahman146/kunai/internal/ai"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/vectorstores"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"time"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Launch an interactive REPL with project context via Ollama",
	Long: `Starts a REPL where each question is sent to a local Ollama model
with the current project context prepended. Use 'exit', 'quit' or Ctrl+C to quit.`,
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
	chatCmd.Flags().StringVarP(&chatCmdParams.Model, "model", "m", "mistral", "Specify the LLM model")
	chatCmd.Flags().StringVarP(&chatCmdParams.EmbedModel, "embed-model", "e", "nomic-embed-text", "Specify the embedding model")
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
	docs, err := ai.ScanProject(chatCmdParams.ContextDir)
	if err != nil {
		return err
	}
	if err = ai.StoreDocuments(ctx, docs, store); err != nil {
		return err
	}
	// Create Conversational Retrieval
	retriever := vectorstores.ToRetriever(store, chatCmdParams.MaxRelevantDocs)
	convMem := memory.NewConversationBuffer(memory.WithReturnMessages(true))
	qaChain := chains.NewConversationalRetrievalQAFromLLM(llm, retriever, convMem)

	// Prerpare Rendering
	var renderer *glamour.TermRenderer
	isTTY := terminal.IsTerminal(int(os.Stdout.Fd()))
	style := "notty"
	if isTTY {
		style = "dark"
		r, renderErr := glamour.NewTermRenderer(
			glamour.WithStandardStyle(style),
			glamour.WithWordWrap(utils.TerminalWidth()),
		)
		if renderErr == nil {
			renderer = r
		}
	}
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
		// Send Answer
		if renderer != nil {
			if out, err := renderer.Render(answer); err != nil {
				return answer, err
			} else {
				return out, nil
			}
		}
		return answer, nil
	})
	return nil
}
