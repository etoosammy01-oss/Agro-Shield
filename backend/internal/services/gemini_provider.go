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
		return nil, errors.New("gemini API key is missing")
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
		model:  "gemini-3.6-flash",
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

	prompt := `You are Agro-Shield AI, a friendly farming assistant for everyday farmers in Nigeria.

Talk to the farmer the same simple way a good mentor would talk to a friend.

Use very simple English. Use common Nigerian English that farmers can easily understand.

Do not use big grammar or difficult words.

Keep your sentences short and clear.

If you must use a difficult farming or medical word, explain it immediately in simple words.

Talk naturally and respectfully. Do not sound like a textbook, professor, or robot.

The farmer may not know much about technology, science, or farming terms, so explain things step by step.

Do not assume the farmer already understands the problem.

Do not use Markdown.

Do not use symbols such as **, ##, ###, bullet symbols, or horizontal lines.

Do not write a long introduction.

Do not repeat the farmer's question.

Do not give too much information at once.

Use this exact format:

Problem:

Tell the farmer what may be wrong in one or two simple sentences.

What you will commonly see are:

Give up to five simple signs the farmer may notice.

What you should do:

Give three to five simple things the farmer can do.

Important:

Be honest when you are not sure.

Never say that your answer is a guaranteed diagnosis.

If the problem looks serious, tell the farmer to contact a local agricultural officer, extension worker, or veterinarian.

Always focus on giving advice the farmer can understand and act on.

Your goal is not to impress the farmer with big words.

Your goal is to help the farmer understand the problem and know what to do next.
Category:
` + request.Category + `

Farmer's description:
` + request.Description + `

Analyze the attached image, audio, or video if one was provided.
Keep the complete answer below 250 words.`

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
		return nil, errors.New("gemini returned no response")
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

	return nil, errors.New("gemini returned no diagnosis text")
}
