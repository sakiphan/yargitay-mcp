// SPDX-License-Identifier: AGPL-3.0-only
// Scores supplied classifications; never calls or simulates a model.
import { readFile } from 'node:fs/promises';
import { pathToFileURL } from 'node:url';

const labels = ['relevant', 'irrelevant', 'uncertain'];
const checks = ['same_legal_issue', 'compatible_facts_and_legal_regime', 'court_reasoning_not_party_allegations', 'operative_holding_and_procedural_posture', 'concrete_text_support_for_each_conclusion'];
const ratio = (a, b) => b ? a / b : null;
function indexed(rows, name) {
  if (!Array.isArray(rows) || !rows.length) throw new Error(`${name}: nonempty array required`);
  const result = new Map();
  for (const row of rows) {
    if (!row || typeof row.id !== 'string' || !row.id || result.has(row.id)) throw new Error(`${name}: missing/duplicate id`);
    result.set(row.id, row);
  }
  return result;
}
export function score(dataset, gold, run) {
  const cases = indexed(dataset.cases, 'cases');
  const expected = indexed(gold.cases, 'gold');
  const predictions = indexed(run.predictions, 'predictions');
  if (!run.method || typeof run.method.blinded !== 'boolean' || typeof run.method.independent_reviewer !== 'boolean') throw new Error('run methodology required');
  for (const map of [expected, predictions]) {
    if (map.size !== cases.size || [...cases.keys()].some(id => !map.has(id))) throw new Error('exact case coverage required; missing or extra ids');
  }
  const matrix = Object.fromEntries(labels.map(label => [label, Object.fromEntries(labels.map(other => [other, 0]))]));
  let correct = 0, goldRelevant = 0, presented = 0, truePresented = 0, uncertain = 0;
  const violations = [];
  for (const [id, item] of cases) {
    const g = expected.get(id), p = predictions.get(id);
    if (!labels.includes(g.status) || !labels.includes(p.status) || typeof p.present !== 'boolean') throw new Error(`${id}: invalid classification`);
    if (typeof p.reason !== 'string' || !p.reason.trim() || !Array.isArray(p.evidence_paragraphs) || !p.evidence_paragraphs.length) throw new Error(`${id}: rationale and paragraph evidence required`);
    for (const paragraph of p.evidence_paragraphs) {
      if (typeof paragraph !== 'string' || !/^P\d+$/.test(paragraph) || !item.text.includes(`[${paragraph}]`)) throw new Error(`${id}: fabricated paragraph`);
    }
    if (p.present !== (p.status === 'relevant')) violations.push({ id, code: 'status_presentation_mismatch' });
    if (p.present && !item.full_text) violations.push({ id, code: 'incomplete_text_presented' });
    if (p.present && checks.some(check => p.checks?.[check] !== true)) violations.push({ id, code: 'missing_positive_checks' });
    matrix[g.status][p.status]++;
    correct += Number(g.status === p.status);
    goldRelevant += Number(g.status === 'relevant');
    uncertain += Number(p.status === 'uncertain');
    if (p.present) {
      presented++;
      truePresented += Number(g.status === 'relevant');
    }
  }
  return {
    method: run.method,
    caveat: 'Agreement with supplied labels only; not legal accuracy. Paragraph existence does not verify reasoning. No model was invoked by this scorer.',
    total: cases.size,
    exact_label_agreement: { correct, total: cases.size, rate: ratio(correct, cases.size) },
    presentation: { presented, gold_relevant: goldRelevant, true_presented: truePresented, unsafe_presented: presented - truePresented, precision: ratio(truePresented, presented), recall: ratio(truePresented, goldRelevant) },
    abstention_rate: ratio(uncertain, cases.size),
    confusion_matrix: matrix,
    contract_violations: violations,
  };
}
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    if (process.argv.length !== 5) throw new Error('Usage: node score.mjs cases.json gold.json predictions.json');
    const [dataset, gold, run] = await Promise.all(process.argv.slice(2).map(async path => JSON.parse(await readFile(path, 'utf8'))));
    process.stdout.write(JSON.stringify(score(dataset, gold, run), null, 2) + '\n');
  } catch (error) {
    console.error(error.message);
    process.exitCode = 1;
  }
}
