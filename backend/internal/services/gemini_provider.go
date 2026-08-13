package services

import (
	"context"
	"errors"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
}

func NewGeminiProvider(apiKey string) (*GeminiProvider, error) {
	if apiKey == "" {
		return nil, errors.New("Gemini API key is missing")
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &GeminiProvider{
		client: client,
		model:  "gemini-2.5-pro",
	}, nil
}

func (p *GeminiProvider) Analyze(
	request AIRequest,
) (*DiagnosisResult, error) {

	if request.Category == "" &&
		request.Description == "" &&
		len(request.Image) == 0 &&
		len(request.Audio) == 0 &&
		len(request.Video) == 0 {
		return nil, errors.New("no information was provided")
	}

	prompt := `You are Agro-Shield AI, an agricultural assistant.

Analyze the farmer's request and provide useful farming guidance.

Category:
` + request.Category + `

Farmer's description:
` + request.Description + `

Give:
1. The likely problem, condition, or answer.
2. The signs or information that support your answer.
3. Practical recommended actions.

If the information or media is unclear, say that the result is uncertain.
Do not claim that your answer is a guaranteed diagnosis.`

	parts := []*genai.Part{
		{Text: prompt},
	}

	if len(request.Image) > 0 {
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				Data:     request.Image,
				MIMEType: request.ImageType,
			},
		})
	}

	if len(request.Audio) > 0 {
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				Data:     request.Audio,
				MIMEType: request.AudioType,
			},
		})
	}

	if len(request.Video) > 0 {
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				Data:     request.Video,
				MIMEType: request.VideoType,
			},
		})
	}

	contents := []*genai.Content{
		{
			Parts: parts,
		},
	}

	result, err := p.client.Models.GenerateContent(
		context.Background(),
		p.model,
		contents,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result.Candidates) == 0 {
		return nil, errors.New("Gemini returned no response")
	}

	for _, candidate := range result.Candidates {
		if candidate.Content == nil {
			continue
		}

		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				return &DiagnosisResult{
					Result: part.Text,
				}, nil
			}
		}
	}

	return nil, errors.New("Gemini returned no diagnosis text")
}
