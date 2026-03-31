package cron

import (
	"strings"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
)

var surveyKeywords = []string{"survey", "review", "state of the art", "taxonomy"}

func normalizeForMatch(s string) string {
	normalized := strings.ToLower(s)
	normalized = strings.ReplaceAll(normalized, "-", " ")
	return strings.Join(strings.Fields(normalized), " ")
}

func hasAnyCategory(categories, topics []string) bool {
	for _, category := range categories {
		for _, topic := range topics {
			if category == topic {
				return true
			}
		}
	}
	return false
}

func matchesSurveyKeyword(title, abstract string, keywords []string) bool {
	text := normalizeForMatch(title)
	if abstract != "" {
		text = text + " " + normalizeForMatch(abstract)
	}

	for _, keyword := range keywords {
		if strings.Contains(text, normalizeForMatch(keyword)) {
			return true
		}
	}

	return false
}

func isEligibleSurvey(p arxiv.Paper, topics, keywords []string) bool {
	if !hasAnyCategory(p.Categories, topics) {
		return false
	}

	return matchesSurveyKeyword(p.Title, p.Abstract, keywords)
}
