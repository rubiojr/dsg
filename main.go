package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	_ "embed"

	"time"

	"github.com/rubiojr/dsg/internal/datahub"
	"github.com/rubiojr/dsg/internal/log"
	storage "github.com/rubiojr/dsg/internal/storage/sqlite"
	"github.com/sashabaranov/go-openai"
	"github.com/urfave/cli/v2"
)

//go:embed tdata/schema.json
var trainingDataset string

func main() {
	app := &cli.App{
		Name:  "dsg",
		Usage: "AI assisted DataHub dataset generator",
		Commands: []*cli.Command{
			{
				Name:   "add-term",
				Usage:  "Add a glossary term to DataHub",
				Action: runAddGlossaryTerm,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "datahub-gms-url",
						EnvVars: []string{"DATAHUB_GMS_URL"},
						Usage:   "DataHub URL",
						Value:   "https://api.datahub.io",
					},
					&cli.StringFlag{
						Name:    "datahub-gms-token",
						EnvVars: []string{"DATAHUB_GMS_TOKEN"},
						Usage:   "DataHub token",
					},
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Glossary Term name",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "urn",
						Usage: "Glossary Term URN",
					},
					&cli.StringFlag{
						Name:     "definition",
						Usage:    "Glossary Term definition",
						Required: false,
					},
				},
			},
			{
				Name:      "post-history-file",
				Usage:     "Create a dataset from a JSON history file",
				ArgsUsage: "FILE",
				Action:    runFromHistoryFile,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "datahub-gms-url",
						EnvVars: []string{"DATAHUB_GMS_URL"},
						Usage:   "DataHub URL",
						Value:   "https://api.datahub.io",
					},
					&cli.StringFlag{
						Name:    "datahub-gms-token",
						EnvVars: []string{"DATAHUB_GMS_TOKEN"},
						Usage:   "DataHub token",
					},
				},
			},
			{
				Name:      "from-json",
				Usage:     "Create a dataset from a JSON file",
				ArgsUsage: "FILE",
				Action:    runFromJSON,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "datahub-gms-url",
						EnvVars: []string{"DATAHUB_GMS_URL"},
						Usage:   "DataHub URL",
						Value:   "https://api.datahub.io",
					},
					&cli.StringFlag{
						Name:    "datahub-gms-token",
						EnvVars: []string{"DATAHUB_GMS_TOKEN"},
						Usage:   "DataHub token",
					},
					&cli.StringFlag{
						Name:     "entity-type",
						Usage:    "Entity type to send (dataset, glossaryTerm, tag, etc)",
						Required: true,
					},
				},
			},
			{
				Name:      "post",
				Usage:     "Post a previously saved response to DataHub",
				ArgsUsage: "HISTORY_ID",
				Action:    runPostHistory,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "datahub-gms-url",
						EnvVars: []string{"DATAHUB_GMS_URL"},
						Usage:   "DataHub URL",
						Value:   "https://api.datahub.io",
					},
					&cli.StringFlag{
						Name:    "datahub-gms-token",
						EnvVars: []string{"DATAHUB_GMS_TOKEN"},
						Usage:   "DataHub token",
					},
				},
			},
			{
				Name:   "generate",
				Usage:  "Generate a new dataset",
				Action: runGenerate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "api-key",
						EnvVars:  []string{"OPENAI_API_KEY"},
						Usage:    "OpenAI API key",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "api-base",
						EnvVars: []string{"OPENAI_API_BASE"},
						Usage:   "OpenAI API base URL (for Azure OpenAI)",
						Value:   "https://api.openai.com/v1",
					},
					&cli.StringFlag{
						Name:    "model",
						EnvVars: []string{"OPENAI_MODEL"},
						Usage:   "OpenAI model to use",
						Value:   "gpt-4o",
					},
					&cli.BoolFlag{
						Name:    "azure",
						EnvVars: []string{"OPENAI_USE_AZURE"},
						Usage:   "Use Azure OpenAI",
						Value:   false,
					},
					&cli.StringFlag{
						Name:    "azure-deployment",
						EnvVars: []string{"AZURE_OPENAI_DEPLOYMENT"},
						Usage:   "Azure OpenAI deployment name (required when using Azure)",
					},
					&cli.StringFlag{
						Name:    "azure-api-version",
						EnvVars: []string{"AZURE_OPENAI_API_VERSION"},
						Usage:   "Azure OpenAI API version",
						Value:   "2023-05-15",
					},
					&cli.StringFlag{
						Name:    "datahub-gms-url",
						EnvVars: []string{"DATAHUB_GMS_URL"},
						Usage:   "DataHub URL",
						Value:   "https://api.datahub.io",
					},
					&cli.StringFlag{
						Name:    "datahub-gms-token",
						EnvVars: []string{"DATAHUB_GMS_TOKEN"},
						Usage:   "DataHub token",
					},
					&cli.BoolFlag{
						Name:  "stdout",
						Usage: "Write the generated datasets to stdout",
					},
					&cli.BoolFlag{
						Name:  "skip-post",
						Usage: "Do not post the datasets to DataHub",
						Value: false,
					},
					&cli.IntFlag{
						Name:  "prompt-from",
						Usage: "Post using the prompt from history",
						Value: -1,
					},
				},
			},
			{
				Name:   "history",
				Usage:  "View generation history",
				Action: runListHistory,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Usage:   "Limit the number of entries",
						Value:   10,
					},
					&cli.IntFlag{
						Name:    "offset",
						Aliases: []string{"o"},
						Usage:   "Offset for pagination",
						Value:   0,
					},
					&cli.BoolFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "Output in JSON format",
						Value:   false,
					},
				},
			},
			{
				Name:      "show",
				Usage:     "Show details of a specific history entry",
				ArgsUsage: "HISTORY_ID",
				Action:    runShowHistory,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "json",
						Aliases: []string{"j"},
						Usage:   "Output in JSON format",
						Value:   false,
					},
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a specific history entry",
				ArgsUsage: "HISTORY_ID",
				Action:    runDeleteHistory,
			},
			{
				Name:   "clear",
				Usage:  "Clear all history entries",
				Action: runClearHistory,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Skip confirmation",
						Value:   false,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func getResponse(id int64) (*storage.Response, error) {
	db, err := storage.NewSQLiteStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize history database: %w", err)
	}
	defer db.Close()

	resp, err := db.GetResponse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get history entry: %w", err)
	}
	return resp, nil
}

func runGenerate(c *cli.Context) error {
	apiKey := c.String("api-key")
	apiBase := c.String("api-base")
	model := c.String("model")
	useAzure := c.Bool("azure")
	azureDeployment := c.String("azure-deployment")
	azureAPIVersion := c.String("azure-api-version")
	datahubURL := c.String("datahub-gms-url")
	datahubToken := c.String("datahub-gms-token")
	toStdout := c.Bool("stdout")
	skipPost := c.Bool("skip-post")
	fromHistory := c.Int64("prompt-from")

	// Validate Azure arguments
	if useAzure && azureDeployment == "" {
		return fmt.Errorf("azure-deployment is required when using Azure OpenAI")
	}

	// Create a temporary file for the prompt
	tmpfile, err := os.CreateTemp("", "XXXXXprompt")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	log.Debugf("Writing temp prompt file to %s...\n", tmpfile.Name())

	var userInput string
	if fromHistory > -1 {
		fmt.Println("Loading prompt from history...")
		resp, err := getResponse(fromHistory)
		if err != nil {
			return fmt.Errorf("error getting response from history: %w", err)
		}
		userInput = resp.Prompt
		fmt.Println("\n>> " + strings.TrimSpace(userInput))
	} else {
		fmt.Println("Write the input for AI, hit Enter+Ctrl-D when finished:")
		fmt.Println()
		userInput, err = readUserInput()
		if err != nil {
			return fmt.Errorf("error reading user input: %w", err)
		}
	}

	// Construct the prompt
	prompt := fmt.Sprintf(`Given a reference json schema like:

%s

Give me another schema taking into account:

%s

If a schema name is provided, set schemaName to the name provided. If not, replace @@@REPLACE_ME@@@ with %d.
Do not explain anything. Return only the required JSON. Do not format the response as markdown.`, trainingDataset, userInput, time.Now().UnixMilli())

	// Write the prompt to the temp file
	if _, err := tmpfile.WriteString(prompt); err != nil {
		return fmt.Errorf("error writing to temp file: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %w", err)
	}

	fmt.Println()
	fmt.Println("Understood! generating DataHub datasets...")
	fmt.Println("Processing input and generating the dataset (may take a while)...")

	// Initialize the OpenAI client
	var client *openai.Client
	if useAzure {
		config := openai.DefaultAzureConfig(apiKey, azureDeployment)
		config.APIVersion = azureAPIVersion
		config.BaseURL = apiBase
		client = openai.NewClientWithConfig(config)
	} else {
		config := openai.DefaultConfig(apiKey)
		config.BaseURL = apiBase
		client = openai.NewClientWithConfig(config)
	}

	// Create chat completion request
	responseFile := tmpfile.Name() + ".response.json"
	responseData, err := sendOpenAIRequest(client, model, prompt)
	if err != nil {
		return fmt.Errorf("error sending request to OpenAI: %w", err)
	}

	// Write the response to a file
	if err := os.WriteFile(responseFile, []byte(responseData), 0644); err != nil {
		return fmt.Errorf("error writing response to file: %w", err)
	}
	defer os.Remove(responseFile)

	// Parse the JSON response
	var jsonResponse []map[string]interface{}
	if err := json.Unmarshal([]byte(responseData), &jsonResponse); err != nil {
		return fmt.Errorf("error parsing JSON response: %w", err)
	}

	// Extract schema information
	var schemaName, schemaURN, datasetName string
	if len(jsonResponse) > 0 {
		if metadata, ok := jsonResponse[0]["schemaMetadata"].(map[string]interface{}); ok {
			if value, ok := metadata["value"].(map[string]interface{}); ok {
				if name, ok := value["schemaName"].(string); ok {
					schemaName = name
				}
			}
		}
		if urn, ok := jsonResponse[0]["urn"].(string); ok {
			schemaURN = urn
		}
		if datasetKey, ok := jsonResponse[0]["datasetKey"].(map[string]interface{}); ok {
			if value, ok := datasetKey["value"].(map[string]interface{}); ok {
				if name, ok := value["name"].(string); ok {
					datasetName = name
				}
			}
		}
	}

	// Save to history database
	db, err := storage.NewSQLiteStorage()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize history database: %v\n", err)
	} else {
		defer db.Close()
		id, err := db.SaveResponse(userInput, responseData, schemaName, schemaURN, datasetName)
		if err != nil {
			fmt.Printf("Warning: Failed to save to history: %v\n", err)
		} else {
			log.Debugf("Response saved to history with ID: %d\n", id)
		}
	}

	if toStdout {
		fmt.Println("Generated JSON:")
		fmt.Println()
		fmt.Println(responseData)
		fmt.Println()
	}

	if skipPost {
		return nil
	}

	// Execute post-dataset command
	log.Debug("posting the dataset")
	dh := datahub.NewClient(datahubURL, datahubToken)
	count, err := dh.PostEntity("dataset", responseData)
	if err != nil {
		return fmt.Errorf("error posting datasets: %w", err)
	}

	fmt.Println("🤖 finished!")
	if count > 1 {
		fmt.Printf("%d datasets created! ☑", count)
	} else {
		fmt.Println()
		fmt.Println("Dataset info")
		fmt.Println("-------------")
		fmt.Printf("Schema URN: %s\n", schemaURN)
		fmt.Printf("Schema Name: %s\n", schemaName)
		fmt.Println()
		fmt.Println("Dataset created! ☑")
	}

	return nil
}

func runListHistory(c *cli.Context) error {
	limit := c.Int("limit")
	offset := c.Int("offset")
	outputJSON := c.Bool("json")

	db, err := storage.NewSQLiteStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize history database: %w", err)
	}
	defer db.Close()

	responses, err := db.ListResponses(limit, offset)
	if err != nil {
		return fmt.Errorf("failed to list history: %w", err)
	}

	if outputJSON {
		jsonData, err := json.MarshalIndent(responses, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}

	if len(responses) == 0 {
		fmt.Println("No history entries found.")
		return nil
	}

	fmt.Printf("%-6s %-20s %-40s %-30s\n", "ID", "DATE", "SCHEMA NAME", "DATASET NAME")
	fmt.Println(strings.Repeat("-", 100))
	for _, resp := range responses {
		fmt.Printf("%-6d %-20s %-40s %-30s\n",
			resp.ID,
			resp.CreatedAt.Format("2006-01-02 15:04:05"),
			truncateString(resp.SchemaName, 38),
			truncateString(resp.DatasetName, 28))
	}

	return nil
}

func runShowHistory(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("history ID is required")
	}

	id, err := strconv.ParseInt(c.Args().Get(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid history ID: %w", err)
	}

	outputJSON := c.Bool("json")

	db, err := storage.NewSQLiteStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize history database: %w", err)
	}
	defer db.Close()

	resp, err := db.GetResponse(id)
	if err != nil {
		return fmt.Errorf("failed to get history entry: %w", err)
	}

	if outputJSON {
		jsonData, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
		return nil
	}

	fmt.Println("History Entry Details")
	fmt.Println("---------------------")
	fmt.Printf("ID:          %d\n", resp.ID)
	fmt.Printf("Created At:  %s\n", resp.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Schema Name: %s\n", resp.SchemaName)
	fmt.Printf("Schema URN:  %s\n", resp.SchemaURN)
	fmt.Printf("Dataset:     %s\n", resp.DatasetName)
	fmt.Println()
	fmt.Println("Prompt:")
	fmt.Println("-------")
	fmt.Println(resp.Prompt)
	fmt.Println()
	fmt.Println("Response:")
	fmt.Println("---------")

	// Try to pretty print the JSON response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(resp.Response), "", "  "); err == nil {
		fmt.Println(prettyJSON.String())
	} else {
		fmt.Println(resp.Response)
	}

	return nil
}

func runDeleteHistory(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("history ID is required")
	}

	id, err := strconv.ParseInt(c.Args().Get(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid history ID: %w", err)
	}

	db, err := storage.NewSQLiteStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize history database: %w", err)
	}
	defer db.Close()

	// First get the entry to confirm it exists
	_, err = db.GetResponse(id)
	if err != nil {
		return fmt.Errorf("failed to find history entry: %w", err)
	}

	// Confirm deletion
	fmt.Printf("Are you sure you want to delete history entry %d? (y/N): ", id)
	reader := bufio.NewReader(os.Stdin)
	confirm, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	confirm = strings.TrimSpace(strings.ToLower(confirm))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("Deletion cancelled.")
		return nil
	}

	if err := db.DeleteResponse(id); err != nil {
		return fmt.Errorf("failed to delete history entry: %w", err)
	}

	fmt.Printf("History entry %d deleted successfully.\n", id)
	return nil
}

func runClearHistory(c *cli.Context) error {
	force := c.Bool("force")

	db, err := storage.NewSQLiteStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize history database: %w", err)
	}
	defer db.Close()

	if !force {
		// Confirm deletion
		fmt.Print("Are you sure you want to clear all history entries? This action cannot be undone. (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		confirm, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		confirm = strings.TrimSpace(strings.ToLower(confirm))
		if confirm != "y" && confirm != "yes" {
			fmt.Println("Clear operation cancelled.")
			return nil
		}
	}

	if err := db.ClearHistory(); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	fmt.Println("All history entries have been cleared.")
	return nil
}

// Helper function to truncate strings for display
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func runPostHistory(c *cli.Context) error {
	if c.NArg() == 0 {
		return fmt.Errorf("history ID is required")
	}

	id, err := strconv.ParseInt(c.Args().Get(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid history ID: %w", err)
	}

	datahubURL := c.String("datahub-gms-url")
	datahubToken := c.String("datahub-gms-token")

	db, err := storage.NewSQLiteStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize history database: %w", err)
	}
	defer db.Close()

	resp, err := db.GetResponse(id)
	if err != nil {
		return fmt.Errorf("failed to get history entry: %w", err)
	}

	fmt.Printf("Sending datasets (ID: %d) to DataHub...\n", resp.ID)

	// Execute post-dataset command
	dh := datahub.NewClient(datahubURL, datahubToken)
	count, err := dh.PostEntity("dataset", resp.Response)
	if err != nil {
		return fmt.Errorf("error posting dataset: %w", err)
	}

	if count > 1 {
		fmt.Printf("%d datasets successfully sent to DataHub!\n", count)
	} else {
		fmt.Println("Dataset successfully sent to DataHub!")
		fmt.Println()
		fmt.Println("Dataset info")
		fmt.Println("-------------")
		fmt.Printf("Schema URN: %s\n", resp.SchemaURN)
		fmt.Printf("Schema Name: %s\n", resp.SchemaName)
		fmt.Printf("Dataset Name: %s\n", resp.DatasetName)
	}

	return nil
}

func readUserInput() (string, error) {
	// Read user input
	reader := bufio.NewReader(os.Stdin)
	var userInput strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
		userInput.WriteString(line)
	}

	return userInput.String(), nil
}

func runAddGlossaryTerm(c *cli.Context) error {
	name := c.String("name")
	urn := c.String("urn")
	if urn == "" {
		urn = "urn:li:glossaryTerm:" + name
	}
	definition := c.String("definition")

	datahubURL := c.String("datahub-gms-url")
	datahubToken := c.String("datahub-gms-token")

	dh := datahub.NewClient(datahubURL, datahubToken)
	gTerm := datahub.GlossaryTerm{
		URN: urn,
		Info: datahub.GlossaryTermInfo{
			Value: datahub.GlossaryTermValue{
				Name:       name,
				Definition: definition,
				Source:     "INTERNAL",
			},
		},
	}

	terms := []datahub.GlossaryTerm{gTerm}
	payload, err := json.Marshal(terms)
	if err != nil {
		return fmt.Errorf("error encoding glossary term to JSON: %w", err)
	}

	_, err = dh.PostEntity("glossaryTerm", string(payload))
	if err != nil {
		return fmt.Errorf("error adding glossary term: %w", err)
	}

	fmt.Println("Glossary term successfully added to DataHub!")
	return nil
}

func runFromJSON(c *cli.Context) error {
	filePath := c.Args().First()
	entityType := c.String("entity-type")

	if filePath == "" {
		return errors.New("file path is required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// if entity-type is dataset it'll be an array of Dataset objects
	var datasets []datahub.Dataset
	var glossaryTerms []datahub.GlossaryTerm
	var entities interface{}

	switch entityType {
	case "dataset":
		err = json.Unmarshal(data, &datasets)
		entities = datasets
	case "glossaryTerm":
		err = json.Unmarshal(data, &glossaryTerms)
		entities = glossaryTerms
	default:
		return fmt.Errorf("unsupported entity type: %s", entityType)
	}

	if err != nil {
		return fmt.Errorf("error decoding JSON: %w", err)
	}

	datahubURL := c.String("datahub-gms-url")
	datahubToken := c.String("datahub-gms-token")

	dh := datahub.NewClient(datahubURL, datahubToken)
	jblob, err := json.MarshalIndent(entities, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding datasets to JSON: %w", err)
	}

	count, err := dh.PostEntity(entityType, string(jblob))
	if err != nil {
		return fmt.Errorf("error adding datasets: %w", err)
	}

	fmt.Printf("%d entities successfully created in DataHub!\n", count)
	return nil
}

type HistoryItem struct {
	ID       int64  `json:"ID"`
	Prompt   string `json:"Prompt"`
	Response string `json:"Response"`
}

func runFromHistoryFile(c *cli.Context) error {
	filePath := c.Args().First()

	if filePath == "" {
		return errors.New("file path is required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// if entity-type is dataset it'll be an array of Dataset objects
	var historyItem HistoryItem

	err = json.Unmarshal(data, &historyItem)
	if err != nil {
		return fmt.Errorf("error decoding JSON: %w", err)
	}

	var datasets []datahub.Dataset
	err = json.Unmarshal([]byte(historyItem.Response), &datasets)
	if err != nil {
		return fmt.Errorf("error decoding JSON: %w", err)
	}

	datahubURL := c.String("datahub-gms-url")
	datahubToken := c.String("datahub-gms-token")

	dh := datahub.NewClient(datahubURL, datahubToken)
	jblob, err := json.MarshalIndent(datasets, "", "  ")
	if err != nil {
		return fmt.Errorf("error encoding datasets to JSON: %w", err)
	}

	count, err := dh.PostEntity("dataset", string(jblob))
	if err != nil {
		return fmt.Errorf("error adding datasets: %w", err)
	}

	fmt.Printf("%d entities successfully created in DataHub!\n", count)
	return nil
}
