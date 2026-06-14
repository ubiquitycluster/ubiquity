package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ubiquitycluster/ubiquity/pkg/provision"
)

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

var aiCmd = &cobra.Command{
	Use:   "ai [prompt]",
	Short: "Local diagnostics using Ollama",
	Long: `Sends cluster state and logs to a local diagnostics Ollama LLM for troubleshooting.
Ollama is not the production AI serving layer; production inference readiness is
provided by NVIDIA NIM Operator-backed services and checked through ubiquity health --ai.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		userPrompt := ""
		if len(args) > 0 {
			userPrompt = strings.Join(args, " ")
		}

		systemPrompt := "Analyze this cluster state:\n"
		state, _ := provision.LoadState()
		if state != nil {
			systemPrompt += state.Summary()
		}
		if userPrompt != "" {
			systemPrompt += "\nUser question: " + userPrompt
		}

		body := ollamaRequest{
			Model:  "llama3.2",
			Prompt: systemPrompt,
			Stream: false,
		}
		jsonBody, _ := json.Marshal(body)

		resp, err := http.Post("http://localhost:11434/api/generate", "application/json", strings.NewReader(string(jsonBody)))
		if err != nil {
			return fmt.Errorf("Ollama not available at localhost:11434 — is it running? %w", err)
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		var result ollamaResponse
		json.Unmarshal(respBody, &result)
		fmt.Println(result.Response)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
}
