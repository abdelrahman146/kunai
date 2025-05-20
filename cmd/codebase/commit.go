package codebase

import (
	"fmt"
	"github.com/abdelrahman146/kunai/internal/ai"
	"github.com/abdelrahman146/kunai/utils"
	"github.com/spf13/cobra"
	"strings"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Create a commit message, and creates a commit on your changes",
	RunE:  runCommitCmd,
}

var commitCmdParams struct {
	OllamaBaseURL string
	Model         string
	Ticket        string
}

func init() {
	commitCmd.Flags().StringVar(&commitCmdParams.Ticket, "ticket", "", "story ticket")
	commitCmd.Flags().StringVarP(&commitCmdParams.Model, "model", "m", "gemma3:12b", "Specify the LLM model")
	commitCmd.Flags().StringVar(&commitCmdParams.OllamaBaseURL, "ollama-url", "http://localhost:11434", "Ollama base url")
}

func runCommitCmd(cmd *cobra.Command, args []string) error {
	var err error
	diff, err := utils.RunCLICommand("git", "diff", "--staged")
	if err != nil {
		return err
	}
	prompt := commitCmdPrompt(diff, commitCmdParams.Ticket)
	var output string
	utils.RunWithSpinner("Generating commit", func() {
		output, err = ai.AskOllama(commitCmdParams.OllamaBaseURL, commitCmdParams.Model, prompt)
	})
	if err != nil {
		return err
	}
	ok, confirmedOutput, err := utils.RequestOutputConfirmation(fmt.Sprintf("generated commit:\n %s", output), output)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Println("Aborting...")
		return nil
	}
	_, err = utils.RunCLICommand("git", "commit", "-m", confirmedOutput)
	if err != nil {
		return err
	}
	fmt.Println("Commit completed")
	return nil
}

func commitCmdPrompt(diff string, ticket string) string {
	prompt := `
You are a Conventional Commits generator. Analyze the following staged git diff and produce exactly one commit message according to https://www.conventionalcommits.org/en/v1.0.0/. 
Do not include any explanation or extra text—only the commit message.

Rules:
1. Determine the <type> from the changes:
   - feat, fix, chore, docs, style, refactor, perf, test
2. Detect a <scope> if the changes are focused on a particular module or file group.
3. If a ticket identifier like ABC-123 or PLA-6435 appears after the TICKET, include it in square brackets after the type/scope.
4. If there is any breaking change (e.g. BREAKING CHANGE: in the diff or a semantic API change), add a '!' after the type or scope, **and** include a BREAKING CHANGE: footer.
5. Write a short, imperative <description> summarizing what changed.
6. If no body or footer is needed beyond the optional BREAKING CHANGE:, omit them.

Format:
<type>[<scope>][!]: [<TICKET>] <description>

[BREAKING CHANGE: Detailed explanation…]

Examples:
1)
feat: allow provided config object to extend other configs

BREAKING CHANGE: extends key in config file is now used for extending other config files

2)
feat!: [SQD-5432] send an email to the customer when a product is shipped

3)
feat(api)!: [FIN-123] send an email to the customer when a product is shipped

4)
chore!: drop support for Node 6

BREAKING CHANGE: use JavaScript features not available in Node 6.

5)
docs: correct spelling of CHANGELOG

6)
feat(lang): [DWQ-1] add Polish language

TICKET: {ticket}

Now, here is the diff to analyze:
{diff}
`
	prompt = strings.Replace(prompt, "{ticket}", ticket, -1)
	prompt = strings.Replace(prompt, "{diff}", diff, -1)
	return prompt
}
