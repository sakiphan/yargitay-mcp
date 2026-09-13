// SPDX-License-Identifier: AGPL-3.0-only
package server

import (
	_ "embed"
	"strings"
)

// Prompts are bundled in the binary; no runtime file reads or working-directory dependency.
//
//go:embed prompts/response.txt
var responsePrompt string

//go:embed prompts/search_review.txt
var searchReviewPrompt string

//go:embed prompts/document_review.txt
var documentReviewPrompt string

//go:embed prompts/query_planning.txt
var queryPlanningPrompt string

//go:embed prompts/query_examples.json
var queryExamples string

var (
	responseInstructions       = strings.TrimSpace(responsePrompt) + "\n\n" + strings.TrimSpace(queryPlanningPrompt) + "\n\nÖrnek planlar (otomatik çalıştırılmaz; sonuç garantisi değildir):\n" + strings.TrimSpace(queryExamples)
	searchReviewInstructions   = " " + strings.TrimSpace(searchReviewPrompt)
	documentReviewInstructions = " " + strings.TrimSpace(documentReviewPrompt)
)
