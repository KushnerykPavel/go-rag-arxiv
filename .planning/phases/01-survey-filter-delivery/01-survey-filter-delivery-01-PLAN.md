---
phase: 01-survey-filter-delivery
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/cron/survey_filter.go
  - internal/cron/arxiv_fetcher.go
  - internal/cron/survey_filter_test.go
  - internal/cron/arxiv_fetcher_test.go
autonomous: true
requirements:
  - FILT-01
  - FILT-02
  - FILT-03
  - FILT-04
  - FILT-05
must_haves:
  truths:
    - "Only papers in the configured categories (cs.AI, cs.CL) can be delivered."
    - "Papers are delivered only when a survey keyword matches title or abstract, case-insensitive."
    - "Keyword matching tolerates hyphen/space variants (e.g., state-of-the-art)."
    - "Telegram receives only eligible papers; ineligible papers are skipped."
    - "Message formatting for eligible papers matches the current format."
  artifacts:
    - path: "internal/cron/survey_filter.go"
      provides: "Survey eligibility helpers and fixed keyword list"
      contains: "var surveyKeywords"
    - path: "internal/cron/arxiv_fetcher.go"
      provides: "Eligibility gate before sendNotification"
      contains: "isEligibleSurvey(paper, topicList, surveyKeywords)"
    - path: "internal/cron/survey_filter_test.go"
      provides: "Unit coverage for category + keyword eligibility"
      contains: "TestSurvey"
    - path: "internal/cron/arxiv_fetcher_test.go"
      provides: "Fetch loop filtering + format regression coverage"
      contains: "TestFetchPapersFiltersBeforeSend"
  key_links:
    - from: "internal/cron/arxiv_fetcher.go"
      to: "internal/cron/survey_filter.go"
      via: "isEligibleSurvey gate before sendNotification"
      pattern: "if !isEligibleSurvey\\(paper, topicList, surveyKeywords\\) \\{"
---

<objective>
Add a survey-only eligibility filter in the cron fetch loop with fixed keywords and tests that lock in eligibility rules and formatting.

Purpose: Ensure only survey/review papers in configured categories reach Telegram without changing output formatting.
Output: Survey filter helper + tests + eligibility gate in fetch loop.
</objective>

<execution_context>
@/Users/pavelkushneryk/Documents/vsprojects/go-rag-arxiv/.codex/get-shit-done/workflows/execute-plan.md
@/Users/pavelkushneryk/Documents/vsprojects/go-rag-arxiv/.codex/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/REQUIREMENTS.md
@.planning/phases/01-survey-filter-delivery/01-CONTEXT.md
@.planning/phases/01-survey-filter-delivery/01-RESEARCH.md
@internal/cron/arxiv_fetcher.go
@internal/client/arxiv/paper.go

<interfaces>
From internal/client/arxiv/paper.go:
```go
type Paper struct {
	ArxivID     string
	Title       string
	Authors     []string
	Abstract    string
	PublishedAt time.Time
	Categories  []string
	PDFURL      string
}
```

From internal/cron/arxiv_fetcher.go:
```go
func (f *ArxivFetcher) FetchPapers(ctx context.Context)
func (f *ArxivFetcher) sendNotification(ctx context.Context, topic string, paper arxiv.Paper)
func formatPaper(topic string, p arxiv.Paper) string
```
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add survey filter tests (eligibility, keywords, formatting)</name>
  <files>internal/cron/survey_filter_test.go, internal/cron/arxiv_fetcher_test.go</files>
  <read_first>
    .planning/PROJECT.md
    .planning/ROADMAP.md
    .planning/REQUIREMENTS.md
    internal/cron/arxiv_fetcher.go
    internal/client/arxiv/paper.go
    internal/wrappers/ratelimit.go
    internal/app/config_test.go
  </read_first>
  <action>
    Create internal/cron/survey_filter_test.go with table-driven tests that call isEligibleSurvey and validate:
    1. Category gate: paper with Categories ["math.OC", "cs.AI"] and title "Survey of X" is eligible; paper with Categories ["math.OC"] and title "Survey of X" is not.
    2. Case-insensitive keyword match: title "A SURVEY of Y" matches.
    3. Abstract match: title "Methods" with abstract "This paper is a review of Z" matches.
    4. Title-only allowed when abstract is empty: title "Taxonomy of W" with Abstract "" matches.
    5. Hyphen/space normalization: title "State-of-the-art methods" matches keyword "state of the art".
    6. Keyword list contains exact required phrases: "survey", "review", "state of the art", "taxonomy".

    Use arxiv.Paper values directly, and call isEligibleSurvey(paper, topicList, surveyKeywords).
    For keyword list check, iterate surveyKeywords and assert all required phrases are present (exact string match).

    Create internal/cron/arxiv_fetcher_test.go with:
    - TestFetchPapersFiltersBeforeSend: use a fake fetcher returning two papers (one eligible, one ineligible). Use a fake notifier that counts SendHTML calls. Construct ArxivFetcher with a real RateLimiter via wrappers.NewRateLimiter(1). Call FetchPapers with context.Background and assert SendHTML called exactly once and the sent HTML corresponds to the eligible paper.
    - TestFormatPaperUnchanged: create a paper with known values and assert formatPaper(topic, paper) equals:
      "<b>[cs.AI]</b> Survey of AI\n\n<b>Authors:</b> Ada Lovelace, Alan Turing\n<b>Published:</b> 2026-03-30\n<a href=\"https://arxiv.org/pdf/1234.56789.pdf\">PDF</a>"

    For TestFetchPapersFiltersBeforeSend, set eligible paper Title "Survey of AI", Abstract "overview", Categories ["cs.AI"], Authors ["Ada"], PublishedAt time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC), PDFURL "https://arxiv.org/pdf/1234.56789.pdf".
    Set ineligible paper Title "On Transformers", Abstract "method", Categories ["cs.AI"], and no survey keywords.
  </action>
  <acceptance_criteria>
    internal/cron/survey_filter_test.go contains "TestSurveyFilterCategories"
    internal/cron/survey_filter_test.go contains "TestSurveyFilterKeywords"
    internal/cron/survey_filter_test.go contains "TestSurveyKeywordList"
    internal/cron/survey_filter_test.go contains "state of the art"
    internal/cron/arxiv_fetcher_test.go contains "TestFetchPapersFiltersBeforeSend"
    internal/cron/arxiv_fetcher_test.go contains "TestFormatPaperUnchanged"
    internal/cron/arxiv_fetcher_test.go contains "<b>[cs.AI]</b> Survey of AI"
  </acceptance_criteria>
  <verify>
    <automated>go test ./internal/cron -run TestSurvey -v</automated>
  </verify>
  <done>Survey eligibility rules, keyword list, filter gating behavior, and formatting are covered by Go tests in internal/cron.</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Implement survey eligibility helpers and gate notifications</name>
  <files>internal/cron/survey_filter.go, internal/cron/arxiv_fetcher.go</files>
  <read_first>
    .planning/PROJECT.md
    .planning/ROADMAP.md
    .planning/REQUIREMENTS.md
    internal/cron/arxiv_fetcher.go
    internal/client/arxiv/paper.go
    internal/cron/survey_filter_test.go
    internal/cron/arxiv_fetcher_test.go
  </read_first>
  <behavior>
    - Categories: eligibility requires any category in topicList (cs.AI or cs.CL).
    - Keywords: eligibility requires any keyword match in title or abstract, case-insensitive.
    - Normalization: "state of the art" must match "state-of-the-art".
    - Abstract empty: title-only match allowed when Abstract is empty.
  </behavior>
  <action>
    Create internal/cron/survey_filter.go in package cron with:
    - var surveyKeywords = []string{"survey", "review", "state of the art", "taxonomy"}
    - func normalizeForMatch(s string) string:
        * strings.ToLower(s)
        * strings.ReplaceAll(s, "-", " ")
        * strings.Join(strings.Fields(s), " ")
    - func hasAnyCategory(categories, topics []string) bool:
        * return true if any category equals any topic (exact string match)
    - func matchesSurveyKeyword(title, abstract string, keywords []string) bool:
        * text := normalizeForMatch(title)
        * if abstract != "" { text = text + " " + normalizeForMatch(abstract) }
        * for each keyword: if strings.Contains(text, normalizeForMatch(keyword)) return true
        * return false
    - func isEligibleSurvey(p arxiv.Paper, topics, keywords []string) bool:
        * if !hasAnyCategory(p.Categories, topics) return false
        * return matchesSurveyKeyword(p.Title, p.Abstract, keywords)

    Update internal/cron/arxiv_fetcher.go in FetchPapers loop, before f.sendNotification:
      if !isEligibleSurvey(paper, topicList, surveyKeywords) {
        continue
      }
      f.sendNotification(ctx, topic, paper)

    Do not change formatPaper output or sendNotification implementation.
  </action>
  <acceptance_criteria>
    internal/cron/survey_filter.go contains "var surveyKeywords = []string{\"survey\", \"review\", \"state of the art\", \"taxonomy\"}"
    internal/cron/survey_filter.go contains "func normalizeForMatch(s string) string"
    internal/cron/survey_filter.go contains "func isEligibleSurvey(p arxiv.Paper, topics, keywords []string) bool"
    internal/cron/arxiv_fetcher.go contains "if !isEligibleSurvey(paper, topicList, surveyKeywords) {"
    internal/cron/arxiv_fetcher.go contains "f.sendNotification(ctx, topic, paper)"
  </acceptance_criteria>
  <verify>
    <automated>go test ./internal/cron -run TestSurvey -v</automated>
  </verify>
  <done>Fetch loop gates notifications on survey eligibility with fixed keywords; tests pass.</done>
</task>

</tasks>

<verification>
go test ./...
</verification>

<success_criteria>
All FILT-01..FILT-05 behaviors are enforced by code and covered by tests in internal/cron; formatPaper output remains unchanged.
</success_criteria>

<output>
After completion, create `.planning/phases/01-survey-filter-delivery/01-survey-filter-delivery-01-SUMMARY.md`
</output>
