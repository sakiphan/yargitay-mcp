// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const BaseURL = "https://karararama.yargitay.gov.tr"

var (
	ErrInvalid     = errors.New("Geçersiz istek; filtreleri, tarihleri, kimliği ve sınırları kontrol edin.")
	ErrSchema      = errors.New("Kaynağın yanıt biçimi beklenenden farklı veya yanıt sınırı aşıldı.")
	ErrAccess      = errors.New("Kaynak erişimi reddetti. Erişim engeli aşılmaz; işlemi durdurun.")
	ErrUpstream    = errors.New("Kaynak uygulaması hata döndürdü. Bu, tek başına erişim engeli veya karar bulunmadığı anlamına gelmez.")
	ErrRedirect    = errors.New("Kaynak yönlendirme döndürdü; güvenlik gereği izlenmedi. Bu, tek başına erişim reddi değildir.")
	ErrUnavailable = errors.New("Kaynağa süre sınırı içinde erişilemedi. Bu, karar bulunmadığı anlamına gelmez.")
	ErrCapacity    = errors.New("İstek kuyruğu dolu; daha sonra yeniden deneyin.")
	ErrRate        = errors.New("Kaynak hız sınırı uyguladı; daha sonra yeniden deneyin.")
	ErrNotFound    = errors.New("Karar kaynağında belge bulunamadı.")
	ErrChanged     = errors.New("Karar içeriği veya çıktı formatı değişti; offset=0 ile baştan başlayın.")
	idPattern      = regexp.MustCompile(`^[0-9]{1,32}$`)
	hashPattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

type SearchRequest struct {
	Query         *string `json:"query,omitempty" jsonschema:"Düz metin ifade; otomatik çift tırnakla aranır. Gelişmiş sorgu veya semantik arama değildir"`
	Unit          *string `json:"unit,omitempty" jsonschema:"list_yargitay_units listesindeki tam birim adı"`
	CaseYear      *int    `json:"case_year,omitempty"`
	CaseFrom      *int    `json:"case_number_from,omitempty"`
	CaseTo        *int    `json:"case_number_to,omitempty"`
	DecisionYear  *int    `json:"decision_year,omitempty"`
	DecisionFrom  *int    `json:"decision_number_from,omitempty"`
	DecisionTo    *int    `json:"decision_number_to,omitempty"`
	DateFrom      *string `json:"decision_date_from,omitempty" jsonschema:"YYYY-MM-DD"`
	DateTo        *string `json:"decision_date_to,omitempty" jsonschema:"YYYY-MM-DD"`
	SortBy        *string `json:"sort_by,omitempty" jsonschema:"decision_date (varsayılan), case_number veya decision_number"`
	SortDirection *string `json:"sort_direction,omitempty" jsonschema:"desc (varsayılan) veya asc"`
	Page          *int    `json:"page,omitempty" jsonschema:"1-1000; varsayılan 1; yalnızca tek sayfa"`
	PageSize      *int    `json:"page_size,omitempty" jsonschema:"1-20; varsayılan 10"`
}

func ptr[T any](v T) *T { return &v }
func value[T any](v *T, fallback T) T {
	if v == nil {
		return fallback
	}
	return *v
}

func (r *SearchRequest) Normalize() error {
	for _, p := range []**string{&r.Query, &r.Unit} {
		if *p != nil {
			s := strings.TrimSpace(**p)
			*p = nil
			if s != "" {
				*p = &s
			}
		}
	}
	if r.Query == nil && r.Unit == nil && r.CaseYear == nil && r.CaseFrom == nil && r.CaseTo == nil && r.DecisionYear == nil && r.DecisionFrom == nil && r.DecisionTo == nil && r.DateFrom == nil && r.DateTo == nil {
		return ErrInvalid
	}
	if r.Query != nil {
		q := *r.Query
		for _, pair := range [][2]string{{`"`, `"`}, {"“", "”"}} {
			if strings.HasPrefix(q, pair[0]) && strings.HasSuffix(q, pair[1]) && len(q) >= len(pair[0])+len(pair[1]) {
				q = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(q, pair[0]), pair[1]))
				break
			}
		}
		q = strings.Join(strings.Fields(q), " ")
		// Do not let input break out of the quoted literal expression.
		if strings.ContainsAny(q, "\"\\“”") || utf8.RuneCountInString(q) < 3 || utf8.RuneCountInString(q) > 1000 {
			return ErrInvalid
		}
		r.Query = &q
	}
	if r.Unit != nil {
		if _, ok := FindUnit(*r.Unit); !ok {
			return errors.New("Birim tanınmıyor; list_yargitay_units ile geçerli birimleri alın.")
		}
	}
	for _, n := range []*int{r.CaseYear, r.DecisionYear} {
		if n != nil && (*n < 1900 || *n > 2100) {
			return ErrInvalid
		}
	}
	for _, n := range []*int{r.CaseFrom, r.CaseTo, r.DecisionFrom, r.DecisionTo} {
		if n != nil && (*n < 1 || *n > 99999999) {
			return ErrInvalid
		}
	}
	if r.CaseFrom != nil && r.CaseTo != nil && *r.CaseFrom > *r.CaseTo {
		return ErrInvalid
	}
	if r.DecisionFrom != nil && r.DecisionTo != nil && *r.DecisionFrom > *r.DecisionTo {
		return ErrInvalid
	}
	for _, d := range []*string{r.DateFrom, r.DateTo} {
		if d != nil {
			if _, err := time.Parse("2006-01-02", *d); err != nil {
				return ErrInvalid
			}
		}
	}
	if r.DateFrom != nil && r.DateTo != nil && *r.DateFrom > *r.DateTo {
		return ErrInvalid
	}
	if r.Page == nil {
		r.Page = ptr(1)
	}
	if r.PageSize == nil {
		r.PageSize = ptr(10)
	}
	if *r.Page < 1 || *r.Page > 1000 || *r.PageSize < 1 || *r.PageSize > 20 {
		return ErrInvalid
	}
	if r.SortBy == nil {
		r.SortBy = ptr("decision_date")
	}
	if r.SortDirection == nil {
		r.SortDirection = ptr("desc")
	}
	if *r.SortBy != "decision_date" && *r.SortBy != "case_number" && *r.SortBy != "decision_number" {
		return ErrInvalid
	}
	if *r.SortDirection != "asc" && *r.SortDirection != "desc" {
		return ErrInvalid
	}
	return nil
}

type Warning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Summary struct {
	ID             string  `json:"id"`
	Unit           *string `json:"unit"`
	CaseNumber     *string `json:"case_number"`
	DecisionNumber *string `json:"decision_number"`
	DecisionDate   *string `json:"decision_date"`
	Snippet        *string `json:"snippet"`
	SourceURL      string  `json:"source_url"`
	ViewURL        *string `json:"view_url"`
}
type SearchResponse struct {
	Review           RelevanceReview `json:"relevance_review"`
	Filters          SearchRequest   `json:"filters" jsonschema:"Normalize edilmiş istenen filtreler; kaynağın uyguladığına dair bağımsız doğrulama değildir"`
	QueryMode        string          `json:"query_mode" jsonschema:"phrase veya metadata_only"`
	RelevanceChecked bool            `json:"relevance_checked" jsonschema:"Arama belge metinlerini okumaz; false konuya uygunluğun doğrulanmadığını belirtir"`
	Total            *int            `json:"total"`
	Page             int             `json:"page"`
	PageSize         int             `json:"page_size"`
	Results          []Summary       `json:"results"`
	Warnings         []Warning       `json:"warnings"`
	Source           string          `json:"source"`
	FetchedAt        time.Time       `json:"fetched_at"`
}
type DocumentRequest struct {
	DocumentID   string  `json:"document_id" jsonschema:"Aramadan dönen sayısal kimlik; 1-32 rakam"`
	Offset       *int    `json:"offset,omitempty" jsonschema:"Unicode kod noktası indeksi; varsayılan 0"`
	MaxChars     *int    `json:"max_chars,omitempty" jsonschema:"1-100000; varsayılan 30000"`
	OutputFormat *string `json:"output_format,omitempty" jsonschema:"markdown (varsayılan) veya text"`
	ExpectedHash *string `json:"expected_content_sha256,omitempty" jsonschema:"Devam parçası için önceki content_sha256"`
}

func (r *DocumentRequest) Normalize() error {
	if !idPattern.MatchString(r.DocumentID) {
		return ErrInvalid
	}
	return r.NormalizeRange()
}

// NormalizeRange validates chunk options independently of a source's ID syntax.
func (r *DocumentRequest) NormalizeRange() error {
	if r.Offset == nil {
		r.Offset = ptr(0)
	}
	if r.MaxChars == nil {
		r.MaxChars = ptr(30000)
	}
	if r.OutputFormat == nil {
		r.OutputFormat = ptr("markdown")
	}
	if *r.Offset < 0 || *r.MaxChars < 1 || *r.MaxChars > 100000 || (*r.OutputFormat != "markdown" && *r.OutputFormat != "text") {
		return ErrInvalid
	}
	if r.ExpectedHash != nil && !hashPattern.MatchString(*r.ExpectedHash) {
		return ErrInvalid
	}
	return nil
}

type Document struct {
	ID        string    `json:"document_id"`
	Text      string    `json:"text"`
	Markdown  string    `json:"markdown"`
	SourceURL string    `json:"source_url"`
	FetchedAt time.Time `json:"fetched_at"`
}
type Chunk struct {
	Review        RelevanceReview `json:"relevance_review"`
	DocumentID    string          `json:"document_id"`
	Content       string          `json:"content"`
	OutputFormat  string          `json:"output_format"`
	Offset        int             `json:"offset"`
	ReturnedChars int             `json:"returned_chars"`
	TotalChars    int             `json:"total_chars"`
	NextOffset    *int            `json:"next_offset"`
	IsTruncated   bool            `json:"is_truncated"`
	SourceURL     string          `json:"source_url"`
	ViewURL       *string         `json:"view_url"`
	FetchedAt     time.Time       `json:"fetched_at"`
	ContentSHA256 string          `json:"content_sha256"`
	Warnings      []Warning       `json:"warnings"`
}
