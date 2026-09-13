// SPDX-License-Identifier: AGPL-3.0-only
package courts

import (
	"github.com/sakiphan/yargitay-mcp/internal/yargitay"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type SearchRequest struct {
	Query           string   `json:"query" jsonschema:"Düz metin ifade; gelişmiş sorgu değil. 3-1000 karakter."`
	RequiredPhrases []string `json:"required_phrases,omitempty" jsonschema:"Yalnızca Danıştay: query ile birlikte bulunması gereken en fazla 5 ek düz metin ifade; her biri 3-200 karakter. Hukuki uygunluk garantisi değildir."`
	DecisionType    string   `json:"decision_type,omitempty"`
	Language        string   `json:"language,omitempty"`
	Page            int      `json:"page,omitempty"`
	PageSize        int      `json:"page_size,omitempty"`
}

func (r *SearchRequest) normalize(source string) error {
	r.Query = strings.TrimSpace(r.Query)
	for _, pair := range [][2]string{{`"`, `"`}, {"“", "”"}} {
		if strings.HasPrefix(r.Query, pair[0]) && strings.HasSuffix(r.Query, pair[1]) && len(r.Query) >= len(pair[0])+len(pair[1]) {
			r.Query = strings.TrimSuffix(strings.TrimPrefix(r.Query, pair[0]), pair[1])
			break
		}
	}
	r.Query = strings.Join(strings.Fields(r.Query), " ")
	if utf8.RuneCountInString(r.Query) < 3 || utf8.RuneCountInString(r.Query) > 1000 || strings.ContainsAny(r.Query, "\"\\“”") {
		return yargitay.ErrInvalid
	}
	if len(r.RequiredPhrases) > 5 || len(r.RequiredPhrases) > 0 && source != "danistay" {
		return yargitay.ErrInvalid
	}
	phrases := make([]string, 0, len(r.RequiredPhrases))
	seen := map[string]bool{strings.ToLowerSpecial(unicode.TurkishCase, r.Query): true}
	total := utf8.RuneCountInString(r.Query)
	for _, phrase := range r.RequiredPhrases {
		// Reuse exactly the same literal normalization; never accept operators.
		single := SearchRequest{Query: phrase}
		if e := single.normalize("danistay"); e != nil {
			return e
		}
		phrase = single.Query
		if utf8.RuneCountInString(phrase) > 200 {
			return yargitay.ErrInvalid
		}
		key := strings.ToLowerSpecial(unicode.TurkishCase, phrase)
		if seen[key] {
			continue
		}
		seen[key] = true
		total += utf8.RuneCountInString(phrase)
		phrases = append(phrases, phrase)
	}
	if total > 1000 {
		return yargitay.ErrInvalid
	}
	r.RequiredPhrases = phrases
	if r.Page == 0 {
		r.Page = 1
	}
	if r.PageSize == 0 {
		r.PageSize = 10
	}
	if r.Page < 1 || r.Page > 1000 || r.PageSize < 1 || r.PageSize > 20 {
		return yargitay.ErrInvalid
	}
	switch source {
	case "aym":
		if r.DecisionType == "" {
			r.DecisionType = "individual_application"
		}
		if r.Language != "" || r.DecisionType != "individual_application" && r.DecisionType != "norm_review" {
			return yargitay.ErrInvalid
		}
	case "danistay":
		if r.Language != "" || r.DecisionType != "" {
			return yargitay.ErrInvalid
		}
	case "aihm":
		if r.DecisionType == "" {
			r.DecisionType = "judgments"
		}
		if r.Language == "" {
			r.Language = "ENG"
		}
		if r.DecisionType != "judgments" && r.DecisionType != "decisions" || r.Language != "ENG" && r.Language != "FRE" && r.Language != "TUR" {
			return yargitay.ErrInvalid
		}
	default:
		return yargitay.ErrInvalid
	}
	return nil
}

type Summary struct {
	ID                string  `json:"id"`
	Source            string  `json:"source"`
	Title             *string `json:"title"`
	Unit              *string `json:"unit"`
	ApplicationNumber *string `json:"application_number"`
	CaseNumber        *string `json:"case_number"`
	DecisionNumber    *string `json:"decision_number"`
	DecisionDate      *string `json:"decision_date"`
	DecisionType      *string `json:"decision_type"`
	Language          *string `json:"language"`
	SourceURL         string  `json:"source_url"`
	ViewURL           *string `json:"view_url"`
}
type SearchResponse struct {
	Review           yargitay.RelevanceReview `json:"relevance_review"`
	Source           string                   `json:"source"`
	Filters          SearchRequest            `json:"filters"`
	QueryMode        string                   `json:"query_mode"`
	RelevanceChecked bool                     `json:"relevance_checked"`
	Total            *int                     `json:"total"`
	Results          []Summary                `json:"results"`
	FetchedAt        time.Time                `json:"fetched_at"`
	Warnings         []yargitay.Warning       `json:"warnings"`
}
type DocumentRequest struct {
	DocumentID   string  `json:"document_id" jsonschema:"Bu kaynağın aramasından dönen kimlik; başka kaynağın kimliğini kullanmayın."`
	Offset       *int    `json:"offset,omitempty"`
	MaxChars     *int    `json:"max_chars,omitempty"`
	OutputFormat *string `json:"output_format,omitempty"`
	ExpectedHash *string `json:"expected_content_sha256,omitempty"`
}

func (r DocumentRequest) Range() yargitay.DocumentRequest {
	return yargitay.DocumentRequest{DocumentID: r.DocumentID, Offset: r.Offset, MaxChars: r.MaxChars, OutputFormat: r.OutputFormat, ExpectedHash: r.ExpectedHash}
}

type Chunk struct {
	yargitay.Chunk
	Source string `json:"source"`
}

var uuidPattern = regexp.MustCompile(`^(bb|norm):[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
var numberPattern = regexp.MustCompile(`^[0-9]{1,32}$`)
var hudocPattern = regexp.MustCompile(`^[0-9]{3}-[0-9]{1,12}$`)

func ValidID(source, id string) bool {
	switch source {
	case "aym":
		return uuidPattern.MatchString(id)
	case "danistay":
		return numberPattern.MatchString(id)
	case "aihm":
		return hudocPattern.MatchString(id)
	}
	return false
}
