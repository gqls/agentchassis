# HANDOFF 2026-08-31 — bugs_open/399, CTA label ↔ destination agreement

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_399_cta_label_agreement/`
**Bug:** `bugs_open/399_HANDOFF_2026-08-25_the_framework_computes_the_right_cta_destination_stores_it_beside_the_wrong_copy_and_never_compares_them.md`
**Register:** LNK-040 (`docs/agent_docs/docs026_concept_register/register/link-management.md`)
**Council:** `e9bda035-5ad7-4a27-8d4f-613bd03abe05` — APPROVED at round 3, 12 of 15 seats.
**Supersedes:** `HANDOFF_2026-08-26_continue_here.md` (its §1 — "the single unfinished thing" — is DONE).

---

## 1. What changed today: the canary passed and the instrument is now unbiased

The 08-26 handoff left exactly one decision, with a written rule: *records from **both**
`page-build-handler` and `page-rerender` → canary passed, apply `645`.* Both produced records. `645`
is applied.

| | |
|---|---|
| demand since `643` armed (22:17:08Z, 08-26) | **214** CTA saves `[MEASURED 08-31 15:03Z]` |
| `page-build-handler` | **61** records · 34 contradicts · 59 ambiguous |
| `page-rerender` | **83** records · 1 contradicts · 124 ambiguous |
| `645` applied | **2026-08-31 15:09:38Z**, all **6** save steps armed |
| pipeline health | 7 `save_sections` failures, **7/7** unrelated `OWNED_PAGE_GUARD`, 0 other |
| `misdirected_cta` volume across the roll (§6 owed) | shift found, **explained**, burden discharged |

Nothing in this lane is now blocked, held, or waiting on a condition.

---

## 2. READ THIS FIRST — the one thing that is easy to get wrong from here

**The rate is readable now. Bind it to `645`'s `applied_at`, not to "the last 14 days".**

```sql
SELECT date_trunc('day', occurred_at)::date AS day, count(*) AS pages_recorded,
       sum((context->>'contradicts')::int) AS contradictions,
       sum((context->>'ambiguous')::int)   AS ambiguous,
       count(DISTINCT agent_type)          AS producing_agents
FROM agent_error_log
WHERE error_code = 'CTA_LABEL_MISMATCH'
  AND occurred_at > (SELECT applied_at FROM schema_migrations
                     WHERE filename = '645_audit_cta_label_agreement_remaining_writers.sql')
GROUP BY 1 ORDER BY 1;
```

**145 records are banked from BEFORE that timestamp** (`page-build-handler` 61, `page-rerender` 84).
They came from a two-of-six-writer instrument and carry precisely the fleet-wide bias the staging
existed to prevent. A 14-day window averages across the boundary and quietly reinstates it.

⚠ **And the first forward days are still not a rate.** Four writers were armed minutes ago; the 391
lane's re-resolve burst has not landed yet. Read early records as *which producers fired*, not as a
percentage. Give it a full cycle across all six.

> ### ⚠ THE ARMING CENSUS LOST ITS CONTROL WHEN `645` APPLIED
>
> While `645` was held, the expected answer was **mixed** (2 armed, 4 not) and the mixture WAS the
> control — an all-false result meant the wrong spelling. Now the expected answer is all-true, and
> all-true cannot be told from a predicate that matches anything. **Carry two known-false types:**
>
> ```sql
> SELECT a.type, (a.default_config::text LIKE '%audit_cta_label_agreement%') AS armed
> FROM agent_definitions a
> WHERE a.is_active AND NOT COALESCE(a.is_snapshot,false) AND a.deleted_at IS NULL
>   AND a.type IN ('page-build-handler','page-rerender','page-rebuild','pageflow-builder',
>                  'site-work-orchestrator','tool-recreation-handler',
>                  'content-writer','council-gate');   -- last two MUST read false
> ```
> Six true and two false is the answer that means something. Six true alone is not.
> (The older trap still holds: the key is `audit_cta_label_agreement`, the Go file is
> `cta_label_audit.go` — `LIKE '%cta_label_audit%'` is false on every writer, armed included.)

---

## 3. §6's owed comparison — DONE, and how it nearly went wrong

The guardian's standing caution asked for a post-roll comparison of `misdirected_cta` finding volume,
*with the burden against this change*. There is a real shift:

```
08-24: 35 · 08-25: 19 · 08-26: 16 · 08-27: 10 · 08-28..08-31: 0
```

**Explained on three independent grounds, none of them the extraction:**

1. **The host agent is not dormant.** Sibling checks in the same `completeness-discovery-agent`
   `checks` array filed **38** (`head_essentials_missing`) and **36** (`canonical_mismatch`) over
   08-28..08-31. So the agent runs — this control does NOT exonerate on its own, which is why 2 and 3
   matter.
2. **The population shrank.** Same census, same predicate as 08-26: **186 of ~1,192 (15.6%)** →
   **126 of 1,779 (7.1%)**, 23 sites. Pairs grew **40%** while mismatches fell **60 in absolute
   terms** — repair, not dilution by new good pages. The 389/391 lanes have been draining exactly
   this population.
3. **The survivors mostly cannot refile.** The convicting subset was only **13 of 186** on 08-26
   (~9 scaled to today), and **99** `cta_names_unknown_destination` items sit in `needs_human_review`
   — non-terminal, so the dedup index suppresses a refile of those `item_key`s.

**`[INFERRED]`, not proven.** The decisive test — feed the check a page known to convict and watch it
convict — was **not** run. If someone wants this closed properly, that is the test.

> ### ⚠ THE LITERAL FORM OF THIS QUERY RETURNS A FALSE ZERO, TWICE OVER
> The 08-26 handoff said "compare `misdirected_cta` finding volume". Done literally,
> `WHERE item_type='misdirected_cta'` returns **zero rows in all of history** — the check is *named*
> `misdirected_cta` (`check_misdirected_cta.go:64`) and *files*
> `item_type='cta_names_unknown_destination'` (`:352`). And `site_work_items` is a **rolling window**:
> closed rows move to `site_work_items_archive`, so even the right type reads zero over any window
> whose findings were dealt with. Either mistake alone produces a clean zero that looks exactly like
> the broken detector the guardian warned to expect. Both are now in `LANDMINES.md`.

---

## 4. State of the work — nothing held

| piece | state |
|---|---|
| `datahelpers.JudgeCTALabel` + `cta_label_agreement.go` | **LIVE** (fleet `v1.0.1349`) |
| `ctaClassifyAnchor` as a thin adaptor | **LIVE**, detector tests unchanged |
| `actions/cta_label_audit.go` write-time pass | **LIVE**, firing on both primary writers |
| Migration `643` (two primary writers) | **APPLIED** 2026-08-26 22:17:08Z |
| Migration `645` (remaining four) | **APPLIED** 2026-08-31 15:09:38Z |
| Council | APPROVED round 3 |
| Canary | **PASSED** — both producers |
| §6 owed comparison | **DISCHARGED** (§3 above) |
| Candidate 5 — somebody reads the rate | **OPEN, UNOWNED** ← the only real gap left |

⚠ **`645` is a shared number.** `645_design_critique_agent.sql` (another lane) was applied
2026-08-26 14:21Z. The ledger keys on **filename** and already carries two `648`s, so this is
precedent, not a defect — but **resolve by filename, never by number**.

---

## 5. What is actually left

**One thing, and it is the same thing it has always been: nobody reads the record.**

The instrument is now complete, unbiased and trustworthy — which removes the last excuse and leaves
the gap in full view. `bugs_open/399` fix candidate 5 (a per-site terminal-state instrument) is OPEN
and unowned; the owner scoped it out of this thread deliberately. `bugs_open/410` is the general
class ("seams that complete green and land somewhere nobody reads"); if that lane produces a
surfacing mechanism, this record is **one query** from using it — the query is §2 above.

The reading obligation is a monthly promise in `RUNBOOK_cta_label_agreement.md` and LNK-040. A
promise is not a mechanism, and this handoff will not pretend otherwise.

---

## 6. Blind spots — unchanged, and still stated

Three, all unchanged from 08-26 and all still true. It is a **page-identity** test: *does this copy
name the page it links to?* and nothing else.

1. **The label-locked class (`bugs_open/391`)** — framework picked the destination *and* told the
   writer to name it: copy and destination agree, button still wrong. Pinned by
   `TestJudgeCTALabelIsBlindToTheLabelLockedDefect`, which passes and is wrong on purpose.
2. **Destination-KIND copy** — "Book a discovery call", "Write to <address>". 95 of the 186 named no
   page at all. The `CTALabelSilence` reason code is the seam a kind-check hangs on, **in the 391
   lane's gate, not inside `JudgeCTALabel`** (RFC_047 §9 — they asked, and they were right).
   A live record sampled today is exactly this: `"Browse all tools"` →
   `/tools/css-specificity-calculator/index.html`, `no_opinion` / `ambiguous`.
3. **The third writer** — `ApplySectionEditAction` writes `content_data` directly, never through
   `SavePageSectionsAction`. Live; CTA exposure 3 of 144 `section_edit` items. Accepted residual, not
   closed coverage.

---

## 7. Neighbours

| lane | what to know |
|---|---|
| `bugfix_389_cta_relevance` (bug 391) | **ACTIVE** (commits today). Their re-resolve burst will **dominate** `page-rerender` records when it lands — read it as their repair, not as a rate. They name the window in their NOTES so it can be excluded rather than reverse-engineered. |
| `bugs_open/389` | Their candidate-1 verifier should **call** `JudgeCTALabel`, not fork it. It does not solve their remit-scoping problem. |
| `cta_target_content_pass` | 173 of 186 mismatches are a **copy** problem — their remit. |
| `bugs_open/410` | The surfacing class this record inherits; §5. |
| `bugs_closed/023` | Same title, closed 2026-07-25 without building the comparison. 399 is its reopening. |

---

## 8. Commits

Today: `645` discharged + applied, the §6 comparison, and the docs. Earlier: `08afad7cd` mechanism ·
`4d248421a` docs+contribs · `cd71277e5` SUMMARY · `00cf81437` council round 2 · `373dbfb70` approval ·
`a7115ab31` **643 applied** · `b1190467c` silence reason code · `cd6cb3cc5` two corrections from 391.

**RUNBOOK is the one to open second** — every query carries its gotcha, and the post-`645` controls
are in it now.
