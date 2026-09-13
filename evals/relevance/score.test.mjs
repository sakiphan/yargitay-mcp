// SPDX-License-Identifier: AGPL-3.0-only
import assert from 'node:assert/strict';
import { test } from 'node:test';
import { score } from './score.mjs';

function fixture() {
  const cases = ['A', 'B', 'C'].map(id => ({ id, full_text: id !== 'C', text: '[P1] Synthetic statement.' }));
  const gold = cases.map((c, i) => ({ id: c.id, status: ['relevant', 'irrelevant', 'uncertain'][i] }));
  const method = { blinded: false, independent_reviewer: false, mode: 'scorer_unit_test' };
  const predictions = gold.map(g => ({ ...g, present: g.status === 'relevant', reason: 'Synthetic rationale.', evidence_paragraphs: ['P1'], checks: Object.fromEntries(['same_legal_issue', 'compatible_facts_and_legal_regime', 'court_reasoning_not_party_allegations', 'operative_holding_and_procedural_posture', 'concrete_text_support_for_each_conclusion'].map(k => [k, true])) }));
  return [{ cases }, { cases: gold }, { method, predictions }];
}
test('counts label agreement and presentation separately', () => {
  const result = score(...fixture());
  assert.equal(result.exact_label_agreement.correct, 3);
  assert.equal(result.presentation.precision, 1);
  assert.equal(result.presentation.recall, 1);
  assert.deepEqual(result.contract_violations, []);
});
test('always rejecting cannot look like perfect retrieval; undefined precision is null', () => {
  const args = fixture();
  args[2].predictions.forEach(p => { p.status = 'irrelevant'; p.present = false; });
  const result = score(...args);
  assert.equal(result.exact_label_agreement.correct, 1);
  assert.equal(result.presentation.precision, null);
  assert.equal(result.presentation.recall, 0);
});
test('counts unsafe presentation and incomplete-text violations', () => {
  const args = fixture();
  args[2].predictions[2].present = true;
  const result = score(...args);
  assert.equal(result.presentation.unsafe_presented, 1);
  assert.ok(result.contract_violations.some(v => v.code === 'incomplete_text_presented'));
  assert.ok(result.contract_violations.some(v => v.code === 'status_presentation_mismatch'));
});
test('rejects missing, duplicate and fabricated ids instead of shrinking denominator', () => {
  for (const mutate of [r => r.predictions.pop(), r => r.predictions.push(r.predictions[0]), r => { r.predictions[0].id = 'invented'; }]) {
    const args = fixture(); mutate(args[2]); assert.throws(() => score(...args));
  }
});
test('requires honest methodology and actual paragraph ids', () => {
  const args = fixture(); args[2].predictions[0].evidence_paragraphs = ['P99'];
  assert.throws(() => score(...args));
  const args2 = fixture(); delete args2[2].method;
  assert.throws(() => score(...args2));
});
test('positive results require all review checks', () => {
  const args = fixture(); delete args[2].predictions[0].checks;
  assert.equal(score(...args).contract_violations[0].code, 'missing_positive_checks');
});
