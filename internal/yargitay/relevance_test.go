// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"testing"
	"time"
)

func TestFullTextDeliveryIsNotLegalApproval(t *testing.T) {
	d := Document{ID: "123", Text: "DAVA KONUSU: sentetik savunma hakkı iddiası. HÜKÜM: başka bir usul meselesi.", Markdown: "sentetik", FetchedAt: time.Now()}
	text := "text"
	limit := 20
	first, e := ChunkDocument(d, DocumentRequest{DocumentID: "123", OutputFormat: &text, MaxChars: &limit})
	if e != nil {
		t.Fatal(e)
	}
	last, e := ChunkDocument(d, DocumentRequest{DocumentID: "123", OutputFormat: &text, Offset: first.NextOffset})
	if e != nil {
		t.Fatal(e)
	}
	full, e := ChunkDocument(d, DocumentRequest{DocumentID: "123", OutputFormat: &text})
	if e != nil {
		t.Fatal(e)
	}
	for _, part := range []Chunk{first, last, full} {
		if part.Review.Status != "not_assessed" || part.Review.Evaluator != "calling_model" || part.Review.PresentationRule != "only_after_client_review" || len(part.Review.Checks) != 5 {
			t.Fatal("unreviewed document approved", part.Review)
		}
	}
	if first.Review.FullDocumentInThisResponse || last.Review.FullDocumentInThisResponse || !full.Review.FullDocumentInThisResponse {
		t.Fatal("last chunk mistaken for full reading")
	}
}
