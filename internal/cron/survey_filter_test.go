package cron

import (
	"testing"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
)

func TestSurveyFilterCategories(t *testing.T) {
	t.Parallel()

	eligiblePaper := arxiv.Paper{
		Title:      "Survey of X",
		Categories: []string{"math.OC", "cs.AI"},
	}
	if !isEligibleSurvey(eligiblePaper, topicList, surveyKeywords) {
		t.Fatalf("expected paper with target category to be eligible")
	}

	ineligiblePaper := arxiv.Paper{
		Title:      "Survey of X",
		Categories: []string{"math.OC"},
	}
	if isEligibleSurvey(ineligiblePaper, topicList, surveyKeywords) {
		t.Fatalf("expected paper without target categories to be ineligible")
	}
}

func TestSurveyFilterKeywords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		paper    arxiv.Paper
		eligible bool
	}{
		{
			name: "case-insensitive keyword in title",
			paper: arxiv.Paper{
				Title:      "A SURVEY of Y",
				Categories: []string{"cs.AI"},
			},
			eligible: true,
		},
		{
			name: "keyword in abstract",
			paper: arxiv.Paper{
				Title:      "Methods",
				Abstract:   "This paper is a review of Z",
				Categories: []string{"cs.AI"},
			},
			eligible: true,
		},
		{
			name: "title-only match when abstract empty",
			paper: arxiv.Paper{
				Title:      "Taxonomy of W",
				Abstract:   "",
				Categories: []string{"cs.AI"},
			},
			eligible: true,
		},
		{
			name: "hyphenated phrase matches keyword",
			paper: arxiv.Paper{
				Title:      "State-of-the-art methods",
				Categories: []string{"cs.AI"},
			},
			eligible: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isEligibleSurvey(tt.paper, topicList, surveyKeywords)
			if got != tt.eligible {
				t.Fatalf("isEligibleSurvey() = %v, want %v", got, tt.eligible)
			}
		})
	}
}

func TestSurveyKeywordList(t *testing.T) {
	t.Parallel()

	required := []string{"survey", "review", "state of the art", "taxonomy"}
	present := make(map[string]bool, len(surveyKeywords))
	for _, keyword := range surveyKeywords {
		present[keyword] = true
	}

	for _, keyword := range required {
		if !present[keyword] {
			t.Fatalf("missing required keyword %q", keyword)
		}
	}
}
