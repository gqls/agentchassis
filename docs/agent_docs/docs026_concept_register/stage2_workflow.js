export const meta = {
  name: 'concept-register-stage2',
  description: 'Stage 2: verify concept-register status signals against real code/DB, adversarially confirm corrections',
  phases: [
    { title: 'Verify', detail: 'one agent per work unit follows verify-later pointers into the repo' },
    { title: 'Adversarial', detail: 'independently re-check every proposed status correction' },
  ],
}

const REPO = '/home/ant/projects/agentchassis'
const REG = `${REPO}/docs/agent_docs/docs026_concept_register/register`
const units = (typeof args === 'string' ? JSON.parse(args) : args) || [] // [{cat, mode:'deep'|'sweep', ids:[...]}]
if (!Array.isArray(units) || units.length === 0) {
  throw new Error(`args did not resolve to a non-empty array (got ${typeof args}, length ${Array.isArray(units) ? units.length : 'n/a'})`)
}

const VERDICT_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['category', 'mode', 'verdicts'],
  properties: {
    category: { type: 'string' },
    mode: { type: 'string' },
    verdicts: {
      type: 'array',
      items: {
        type: 'object',
        additionalProperties: false,
        required: ['id', 'stage1_status', 'verified_status', 'is_correction', 'kind', 'evidence'],
        properties: {
          id: { type: 'string' },
          stage1_status: { type: 'string' },
          verified_status: {
            type: 'string',
            enum: ['deployed', 'partial', 'aspirational', 'superseded', 'abandoned', 'unknown', 'convention', 'unverifiable'],
          },
          is_correction: { type: 'boolean', description: 'true iff verified_status materially differs from stage1_status' },
          kind: {
            type: 'string',
            enum: ['code-artifact', 'convention', 'idea', 'process', 'infra', 'db', 'other'],
            description: 'what sort of thing this concept is — governs whether code-existence is even the right test',
          },
          wired: { type: 'boolean', description: 'deep mode only: mechanism is referenced by a live workflow/SQL/deployment, not just present as code. false if code exists but nothing calls it.' },
          evidence: { type: 'string', description: 'concrete file:line / grep-count evidence. <=280 chars. cite what you actually found.' },
        },
      },
    },
  },
}

const ADV_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['id', 'holds', 'final_status', 'reason'],
  properties: {
    id: { type: 'string' },
    holds: { type: 'boolean', description: 'true if the proposed correction survives adversarial scrutiny' },
    final_status: { type: 'string', enum: ['deployed', 'partial', 'aspirational', 'superseded', 'abandoned', 'unknown', 'convention', 'unverifiable'] },
    reason: { type: 'string', description: '<=240 chars: why the correction holds or is overturned, with counter-evidence if overturned' },
  },
}

const deepPrompt = (u) => `You are a Stage-2 code verifier for a concept register. Repo root: ${REPO}

Your work unit: category file ${REG}/${u.cat}, DEEP verification of these concept IDs:
${u.ids.join(', ')}

For EACH id:
1. Read its \`### <id> — ...\` block in ${REG}/${u.cat}. Note its status, status-evidence, what, and especially verify-later (named files, DB tables, workflow names, actions, env vars).
2. Follow the verify-later pointers into the repo with grep/find/read. Determine ground truth:
   - Does each named file/dir/table/action actually EXIST? (use grep -rn / find / ls under ${REPO})
   - Is the mechanism WIRED into a live path — referenced by a workflow definition, agent_definitions SQL, a registry, or a k8s/terraform deployment — or is it dead code that exists but nothing invokes? Set \`wired\` accordingly.
   - If the doc signal was 'partial', decide which way it resolves: fully deployed, still genuinely partial, or actually aspirational (built but unused / never wired).
   - If 'unknown', try to resolve to a concrete status from code.
3. Classify \`kind\`: is this a concrete code-artifact/infra/db thing (existence is testable), or a convention/idea/process (code-existence is NOT the right test → verified_status 'convention' or 'unverifiable', is_correction false unless the doc claimed a built artifact that is absent)?
4. Set verified_status. is_correction=true ONLY if it materially differs from stage1 status (e.g. partial→aspirational, unknown→deployed). Same-status confirmation is is_correction=false.

Evidence must be concrete: cite file:line or a grep hit-count. "0 hits across .go/.sql" is valid evidence of absence. Be terse. Do NOT modify any file. Return the schema object.`

const sweepPrompt = (u) => `You are a Stage-2 existence-sweep verifier for a concept register. Repo root: ${REPO}

Your work unit: category file ${REG}/${u.cat}, EXISTENCE SWEEP of these concept IDs (all tagged 'deployed' in stage 1 — you are hunting FALSE POSITIVES):
${u.ids.join(', ')}

The specific failure mode you hunt: a plan/design document narrated its own design in the PRESENT TENSE, so a never-built thing was tagged 'deployed'. (Confirmed example: a second cluster "va001" that exists only in archived prose, no kubeconfig/overlay/terraform.)

For EACH id:
1. Read its \`### <id>\` block in ${REG}/${u.cat}. Extract every CONCRETE named artifact from verify-later/what: file paths, DB tables, service names, k8s deployments, API endpoints, registry keys, env vars.
2. Classify \`kind\` FIRST. Many 'deployed' concepts are CONVENTIONS, design principles, ideas, or processes (e.g. "layout: technical-precise", a naming rule, a doctrine) — code-existence is NOT the right test. For those: verified_status='convention', is_correction=false, kind='convention'/'idea'/'process'. DO NOT flag conventions as false positives — that is noise.
3. ONLY for concepts that make a concrete claim of a BUILT code/infra/db artifact: check existence with batched greps under ${REPO} (grep -rln across .go/.sql/.yaml/.tf, find, ls). Batch your greps — one command can check many names.
   - If the named artifact EXISTS in code/config/deployment → verified_status='deployed', is_correction=false. (Confirmations are the common case; do not over-correct.)
   - If it exists ONLY in docs/ prose (0 hits in code/config/deployment) → this is a present-tense-plan false positive → verified_status='aspirational' (or 'abandoned' if the idea vanished from later docs), is_correction=true.
4. Evidence: cite the grep hit-count / file:line, e.g. "site_ownership: 0 hits in .sql/.go; only docs/" or "dispatch_actions.go:1 exists". Be terse.

Bias: confirm when the artifact exists; correct ONLY on clear absence-from-code of a concretely-claimed artifact. Do NOT modify any file. Return the schema object.`

phase('Verify')
const results = await pipeline(
  units,
  (u) => agent(u.mode === 'deep' ? deepPrompt(u) : sweepPrompt(u), {
    label: `${u.mode}:${u.cat.replace('.md', '')}`,
    phase: 'Verify',
    schema: VERDICT_SCHEMA,
    effort: u.mode === 'deep' ? 'medium' : 'low',
  }),
  // Adversarial stage: re-check ONLY the corrections this unit proposed.
  (res, u) => {
    if (!res) return { unit: u, verdicts: [], adversarial: [] }
    const corrections = (res.verdicts || []).filter((v) => v.is_correction)
    if (!corrections.length) return { unit: u, verdicts: res.verdicts, adversarial: [] }
    return parallel(corrections.map((c) => () =>
      agent(`Adversarially re-verify a proposed Stage-2 status correction. Repo root: ${REPO}
Concept ${c.id} (category ${u.cat}). Stage-1 said "${c.stage1_status}"; a verifier now proposes "${c.verified_status}".
Its evidence: ${c.evidence}

Try to REFUTE the correction. Read the concept's \`### ${c.id}\` block in ${REG}/${u.cat}, then independently check the repo:
- If the correction says something is absent-from-code, hunt harder for it (alternate names, spellings, subdirs, generated files, terraform, kustomize). If you find it, the correction is WRONG.
- If the correction says something was built/wired, confirm the wiring exists (a live workflow/registry/deployment actually references it), not just that a file exists.
- If the concept is really a convention/idea (not a code artifact), the correction to 'aspirational' may be a miscategorization — say so.
Default to holds=true only if you genuinely cannot refute it. Be terse, cite counter-evidence. Do NOT modify files. Return the schema object.`, {
        label: `adv:${c.id}`,
        phase: 'Adversarial',
        schema: ADV_SCHEMA,
        effort: 'medium',
      }).then((v) => ({ ...c, adversarial: v })).catch(() => ({ ...c, adversarial: null }))
    )).then((adv) => ({ unit: u, verdicts: res.verdicts, adversarial: adv.filter(Boolean) }))
  }
)

// Assemble a flat report. A correction is CONFIRMED only if its adversarial check held.
const clean = results.filter(Boolean)
const confirmedCorrections = []
const overturned = []
let totalVerdicts = 0, totalConfirmedDeployed = 0
for (const r of clean) {
  totalVerdicts += (r.verdicts || []).length
  for (const v of r.verdicts || []) {
    if (!v.is_correction && v.verified_status === v.stage1_status) totalConfirmedDeployed++
  }
  for (const a of r.adversarial || []) {
    const adv = a.adversarial
    if (adv && adv.holds) confirmedCorrections.push({ id: a.id, cat: r.unit.cat, from: a.stage1_status, to: adv.final_status, kind: a.kind, evidence: a.evidence, reason: adv.reason })
    else if (adv && !adv.holds) overturned.push({ id: a.id, cat: r.unit.cat, proposed: a.verified_status, kept: adv.final_status, reason: adv.reason })
    else overturned.push({ id: a.id, cat: r.unit.cat, proposed: a.verified_status, kept: a.stage1_status, reason: 'adversarial check failed to run' })
  }
}

log(`units=${clean.length} verdicts=${totalVerdicts} corrections_confirmed=${confirmedCorrections.length} overturned=${overturned.length}`)

return {
  units_processed: clean.length,
  total_verdicts: totalVerdicts,
  confirmed_corrections: confirmedCorrections,
  overturned_corrections: overturned,
  per_unit: clean.map((r) => ({ cat: r.unit.cat, mode: r.unit.mode, n: (r.verdicts || []).length, corrections: (r.adversarial || []).length })),
}
