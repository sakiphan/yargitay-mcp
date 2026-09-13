// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

// A contract for the calling model, NOT an automated legal classifier.
// Retrieval and even full-text delivery never change the status to "relevant".
type RelevanceReview struct {
	Status                     string   `json:"status" jsonschema:"not_assessed: sunucu hukuki uygunluk değerlendirmesi yapmadı"`
	Evaluator                  string   `json:"evaluator" jsonschema:"calling_model: mevcut istemci modeli; ek API çağrısı yok"`
	PresentationRule           string   `json:"presentation_rule" jsonschema:"only_after_client_review: aday kayıtları doğrudan kullanıcıya emsal olarak listelemeyin"`
	Checks                     []string `json:"required_checks"`
	FullDocumentInThisResponse bool     `json:"full_document_in_this_response" jsonschema:"Yalnızca bu yanıtın tam metni içerip içermediği; modelin okuduğunu veya uygunluğu kanıtlamaz"`
}

func PendingRelevanceReview(full bool) RelevanceReview {
	return RelevanceReview{Status: "not_assessed", Evaluator: "calling_model", PresentationRule: "only_after_client_review", Checks: []string{"same_legal_issue", "compatible_facts_and_legal_regime", "court_reasoning_not_party_allegations", "operative_holding_and_procedural_posture", "concrete_text_support_for_each_conclusion"}, FullDocumentInThisResponse: full}
}
