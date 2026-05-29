package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file - make sure it exists in this folder!")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	config := &genai.GenerateContentConfig{
		Temperature: genai.Ptr(float32(0.0)),
	}

	basePrompt := `
		You are a strict backend data-transformation service for AttractionTickets.
	Your job is to parse natural language holiday requests into a strict JSON schema.

	CRITICAL RULES:
	1. Your entire response must be only the raw JSON object, starting with { and ending with }.
	2. Do NOT use markdown, code blocks, or any other formatting.
	3. If a value cannot be inferred, assign it null.
	4. First, output a "reasoning" key detailing your extraction steps, followed by the "data" payload.

	### EXAMPLES ###

	Input: "Looking to take my wife and 2 kids to Disney World Orlando sometime early next month, probably for a week."
	Output:
	{
		"reasoning": "User mentions wife (1 adult) and 2 kids, totaling 2 adults and 2 children. Destination is Disney World Orlando (theme_park). Early next month translates to the next calendar month. A week means 7 days.",
		"data": {
			"destination": "Disney World Orlando",
			"ticket_type": "theme_park",
			"party_size": {
				"adults": 2,
				"children": 2
			},
			"travel_date_start": "2026-06-01",
			"duration_days": 7
		}
	}

	Input: [Insert current dynamic user input here]
	Output:`

	fmt.Print("Enter your request: ")
	reader := bufio.NewReader(os.Stdin)
	userInput, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Failed to read from console: %v", err)
	}
	userInput = strings.TrimSpace(userInput)

	prompt := strings.Replace(basePrompt, "[Insert current dynamic user input here]", userInput, 1)

	result, err := client.Models.GenerateContent(
		ctx, "gemini-3-flash-preview",
		genai.Text(prompt),
		config,
	)
	if err != nil {
		log.Fatalf("Error generating content: %v", err)
	}

	responseText := result.Text()

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(responseText), &jsonData); err != nil {
		log.Printf("Failed to parse JSON from model response. Printing raw text instead. Error: %v\n", err)
		fmt.Println(responseText)
		return
	}

	prettyJSON, _ := json.MarshalIndent(jsonData, "", "  ")

	fmt.Println(string(prettyJSON))

}
