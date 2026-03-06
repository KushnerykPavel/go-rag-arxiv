package arxiv

import "time"

// Paper represents an arXiv academic paper with its metadata.
type Paper struct {
	ArxivID     string
	Title       string
	Authors     []string
	Abstract    string
	PublishedAt time.Time
	Categories  []string
	PDFURL      string
}
