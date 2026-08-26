# HANDOFF 2026-08-26 — bugs_open/399, CTA label ↔ destination agreement

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_399_cta_label_agreement/`
**Bug:** `bugs_open/399_HANDOFF_2026-08-25_the_framework_computes_the_right_cta_destination_stores_it_beside_the_wrong_copy_and_never_compares_them.md`
**Register:** LNK-040 (`docs/agent_docs/docs026_concept_register/register/link-management.md`)
**Council:** `e9bda035-5ad7-4a27-8d4f-613bd03abe05` — **APPROVED at round 3**, 12 of 15 seats.

---

## 1. READ THIS FIRST — the single unfinished thing

**The mechanism is built, approved, deployed and ARMED on two of six writers. It has never fired,
and that is currently EXPECTED, not a defect.**

Migration `643` applied **2026-08-26 22:17:08Z**. As of 22:25Z, **zero** CTA-bearing components had
been saved since that moment, so there has been no demand at all. `CTA_LABEL_MISMATCH` count: 0.

> ⚠ **DO NOT read that zero as "it works" or as "it is broken". It is uninformative.**
> I nearly filed it as a defect by anchoring the demand window at 20:40 (when I started work) instead
> of 22:17:08 (when the migration applied). That manufactured 82 pairs of fake demand and ~7 fake
> expected findings. See `WRONG_CALLS.md` 2026-08-26 (evening).

### The exact next check

```sql
-- 1. Is there DEMAND yet? (bound on the migration's OWN applied_at, never on your session)
SELECT (SELECT applied_at FROM schema_migrations WHERE filename LIKE '643_audit_cta%') AS armed_at,
       (SELECT count(*) FROM page_components
          WHERE content_data ? 'cta_target_title'
            AND updated_at > (SELECT applied_at FROM schema_migrations WHERE filename LIKE '643_audit_cta%')
       ) AS cta_saves_since_arming;

-- 2. Has it fired, and from WHICH producers?
SELECT agent_type, count(*), sum((context->>'contradicts')::int) AS contradictions,
       sum((context->>'ambiguous')::int) AS ambiguous, max(occurred_at)
FROM agent_error_log WHERE error_code='CTA_LABEL_MISMATCH' GROUP BY 1 ORDER BY 1;
```

**Decision rule:**
- `cta_saves_since_arming = 0` → nothing to conclude. Wait. Do **not** apply 645.
- saves > 0 **and** records = 0 → *now* investigate. First discriminator below.
- records present from **both** `page-build-handler` **and** `page-rerender` → **canary passed, apply 645.**
- records from only ONE producer → the coverage claim is failing silently. That is the failure the
  six-step census exists to prevent. Investigate before widening.

### If saves > 0 and records = 0, check these in order

1. **Was the writer an armed agent?** Only `page-build-handler` and `page-rerender` are armed.
   `page-rebuild`, `pageflow-builder`, `site-work-orchestrator`, `tool-recreation-handler` are 645's
   job and will legitimately produce nothing.
2. **Is step config read per-run or cached at pod start?** CLAUDE.md says DB config is live
   immediately, but I did not verify it for this seam. ⚠ The `page-rerender` pods restarted at
   **22:17** — the same minute 643 applied — which I could not explain and did not chase.
3. **Do the saved pairs actually convict?** My token census and `JudgeCTALabel` measure different
   things: of 186 heuristic mismatches only **13** name exactly one other page. Expected yield is far
   lower than the heuristic suggests — a handful of saves proves nothing either way.
4. **Grep the right pods.** `-l app=agent-chassis` is the WRONG pod set: the agents run in their own
   deployments (`agent-page-build-handler-*`, `agent-page-rerender-*`). The pass logs
   `"audited CTA label/destination agreement before persist"` — **but only when it finds something**,
   so its absence is not evidence. Use the sibling `"internal links before persist"` as a control:
   if that is zero too, the window simply contains no saves.

---

## 2. State of the work

| piece | state |
|---|---|
| `datahelpers.JudgeCTALabel` + `cta_label_agreement.go` | **LIVE** in chassis `v1.0.1345` |
| `ctaClassifyAnchor` reduced to a thin adaptor | **LIVE**, detector tests unchanged (the extraction proof) |
| `actions/cta_label_audit.go` (the write-time pass) | **LIVE**, opt-in |
| Migration `643` — arms `page-build-handler` + `page-rerender` | **APPLIED 22:17:08Z** |
| Migration `645_..._HOLD.sql` — arms the other four | **HELD** — waiting on the canary |
| Council | **APPROVED** round 3 |
| First record | **NOT YET OBSERVED** ← the whole of §1 |

**Deploy verified at the artefact**, not from git: pod `agent-chassis-5864bf97c5-68t5h`,
`audit_cta_label_agreement` and `CTA_LABEL_MISMATCH` both PRESENT in `/proc/1/exe`, with two negative
controls both ABSENT in the same exec.
⚠ The startup `build provenance` line had **already scrolled out of a 200-line tail ten minutes after
the pod started**. On a busy chassis it is not a usable fallback; the binary probe has no shelf life.

---

## 3. ⚠⚠ DO NOT READ THE MISMATCH RATE UNTIL 645 IS APPLIED

The instrument is armed on two of six writers. A rate measured now reads fleet-wide and is silently
biased by the four it cannot see. **That is the same argument that made this a six-step census in the
first place**, and staging did not dissolve it — it scheduled it. Both migrations and the RUNBOOK
carry this warning beside the query.

Order: `643` (done) → canary proves both producers → apply `645` → **only then** the rate is readable.

---

## 4. What this mechanism CANNOT see — three stated blind spots

It is a **page-identity** test. It answers "does this copy name the page it links to?" and nothing else.

1. **The label-locked class (`bugs_open/391`).** When the framework picked the destination *and* told
   the writer to name it, copy and destination **agree** and the button is still wrong. All three
   buttons the owner originally reported would PASS this check. Pinned by
   `TestJudgeCTALabelIsBlindToTheLabelLockedDefect`, a test that passes and is wrong on purpose.
2. **Destination-KIND copy** — raised by the 391 lane on 2026-08-26 and the reason
   `CTALabelSilence` now exists. Of 186 live mismatches the copy names **no page at all in 95**, and
   much of that bucket expresses a *kind* ("Book a discovery call", "Write to <address>"). They
   measured 23 such contact-intent labels among 41 fields, with a live case on
   `leopardess/careers.html` where re-resolution sends a "Write to …" button to an ROI estimator.
   **The reason code is the seam a kind-check hangs on — in THEIR gate, not inside `JudgeCTALabel`.**
   They asked explicitly that the judge not be widened, and they are right: RFC_047 §9.
3. **The third writer.** `ApplySectionEditAction` writes `page_components.content_data` directly and
   never passes through `SavePageSectionsAction`. **Live** — 144 `section_edit` items, newest
   2026-08-26 — with CTA exposure of **3 of 144**. Accepted residual risk, not closed coverage. If
   that share grows, widening to it is the named follow-up.

---

## 5. The one thing that would make this a fix rather than an instrument

**A record nobody reads is not a fix**, and today the reading obligation is a paper promise in the
RUNBOOK and LNK-040. This inherits the exact failure class `bugs_open/410` was filed for on the same
day. **The deliverable is the RATE, not the row** — nobody should read 155 records; somebody must
notice if 14.6% becomes 30%, or drops to 2% after a prompt change.

If the 410 lane produces a general surfacing mechanism, this record is one query from using it.
`bugs_open/399` fix candidate 5 (per-site terminal-state instrument) remains **OPEN and unowned** —
the owner scoped it out of this thread deliberately.

---

## 6. Residual council objections — checked, not waved through

`editquality` kept three MEDIUMs of the form "undefined symbol, will not compile". **All three are
pre-existing same-package helpers** and the package builds and tests green against HEAD:
`ctaTargetTitleField` (`resolve_internal_links_action.go:634`), `pqStringArray`
(`fixloop_digest_action.go:487`), `readSourceFile` (`render_sitemap_test.go:17`). Artefacts of a seat
reading one file in isolation.

`guardian` asked whether any of the six agent types carries two ACTIVE rows (a live landmine that
would make both the census and the per-type UPDATEs wrong). **Checked: all six carry exactly one**
`[MEASURED 2026-08-26]`, and the migrations `RAISE` if the count moves.

**`guardian`'s standing caution is ACCEPTED and NOT closed:** *unchanged tests prove the cases they
cover, not that the detector's live finding population is unchanged.* The extraction is
behaviour-preserving by construction, but that is a claim only production settles. **Owed:** compare
`misdirected_cta` finding volume across the roll and treat any shift as this change's until shown
otherwise.

---

## 7. Neighbours — all told, all in their own files

| lane | what they hold |
|---|---|
| `bugfix_389_cta_relevance` (bug 391) | **ACTIVE.** Sequencing warning: their step re-resolving label-less fields writes new URLs under old copy → contradictions **by construction**. A burst is expected, not damage. CONTRIB in their dir + pointer in `bugs_open/391`. |
| `bugs_open/389` (308's unowned successor) | Their candidate-1 verifier should call `JudgeCTALabel`, not fork it. ⚠ It does **not** solve their remit-scoping problem (`verifier_coverage_test.go:148-185`, the 1,849-item hold). |
| `cta_target_content_pass` | **173 of 186 mismatches are a copy problem** — their remit, and the worklist they have lacked since 2026-08-15. |
| `bugs_open/410` | This record inherits its failure class; named in LNK-040. |
| `bugs_closed/023` | Same title, closed 2026-07-25 **without building the comparison**. 399 is its reopening. |

---

## 8. Commits (all on `087_towards_multiple_domains`)

`08afad7cd` mechanism · `4d248421a` docs+contribs · `cd71277e5` SUMMARY · `00cf81437` council round 2
fixes · `373dbfb70` approval recorded · `a7115ab31` **643 discharged + applied** · `b1190467c`
silence reason code + WRONG_CALLS.

**Lane docs:** PLAN / RUNBOOK / NOTES / README_where_we_are / SUMMARY_2026-08-26 / this handoff.
**RUNBOOK is the one to open second** — it carries every query with its gotcha attached.
