# HANDOFF — bug 414 — 2026-08-31. **THE BUG IS CLOSED. Nothing is outstanding; §5 is forward work only.**

> **⚠ UPDATED 2026-08-31, ~13:00Z, hours after this file was first written — and the update is the
> point of reading it.** The original header said *"There is exactly ONE thing left; it is §3."* That
> was true when written and stopped being true the same evening: the `copy_quality_two_stage` lane
> rebuilt and deployed the fix as **`v1.0.1350`**, a triggered production run verified it clean, and
> **414 moved to `bugs_closed/` in `de99599fb`.** §3 is kept below as the record of what the last
> blocker was and how it was discharged — it is history now, not a task. A handoff whose first line
> asks for work that is already done is the failure this lane spent a week documenting.

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_414_planted_marker_as_claim/`
Bug: **`bugs_closed/`**`414_HANDOFF_2026-08-26_a_planted_acceptance_marker_is_served_as_a_compliance_claim_and_the_audit_fleet_adopted_it_as_the_sites_identity.md`
(§7a–§7n are the working record; **§7o is the closure**, with every condition and its evidence.)

**Counts carry the date they were counted** (owner ruling 2026-08-22). Everything below was
re-measured on 2026-08-31 after the `v1.0.1349` roll — nothing is inherited from the 08-27 session.

---

# 0. STATE IN ONE PARAGRAPH

**CLOSED 2026-08-31 (`de99599fb`) — fixed, live, and verified in production at every layer.**
A planted instruction ("include the exact phrase: checked against the FCA handbook, rule by rule")
made lendzy.co.uk serve an unverifiable regulatory-diligence claim for 24 days, and the audit fleet
adopted it as the site's identity. Both spec sources are stripped, the canonising audit item is
rejected, the served copy is repaired through the framework, and the class fix — two completeness
patterns, practice-family P6, and a detector that applies the claim rules to the **instruction** —
is **live in the running chassis** (`v1.0.1349`, probed at the binary) and council-APPROVED
(`f4c144ad`, credited by `098`). The last defect found was the detector's **own** — its first live run
convicted a site for a phrase in that site's `would_never_say` list — and *running it* is what found
that, not the unit tests and not a nine-seat council round. Fixed (`dc9ccfda2`), deployed as
`v1.0.1350`, and verified by a triggered production run: **0 of 36 sites, from 7,013 scanned fields,
3 suppressions.** Nothing is outstanding on this bug. §5 is forward work that was never part of it.

---

# 1. IF YOU ARE PICKING THIS UP COLD — three cheap confirmations (5 minutes)

Not a task list any more; these are the checks that prove the closure still holds if you ever need to
re-establish it. All three passed on 2026-08-31.

```bash
# 1a. Are the patterns still live in the running chassis? THREE ARMS, never one.
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
for s in 'everything (on this site' 'diligence overclaim' 'guaranteed (accurate' 'zzz-never-written'; do
  kubectl -n ai-persona-system exec $POD -- grep -aq "$s" /proc/1/exe </dev/null 2>/dev/null \
    && echo "PRESENT $s" || echo "absent  $s"
done
```
Expect the first three **PRESENT** and the fourth **absent**. The fourth is not decoration: without a
must-be-absent arm a promiscuous grep returns the same answer for everything.
⚠ **Do not use the log-stamp route on `agent-chassis`.** Measured 2026-08-31: `kubectl logs` returns
**zero lines** for these pods, so the "ask the service what it is running" recipe is unavailable — and
an empty grep there means *out of range*, never *unstamped*.

```sql
-- 1b. Is lendzy still clean? (a regeneration could in principle re-plant)
SELECT count(*) FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='lendzy.co.uk' AND (pc.content_data::text LIKE '%checked against the FCA handbook%'
   OR pc.rendered_html LIKE '%checked against the FCA handbook%');            -- expect 0
SELECT count(*) FROM site_specs WHERE is_current AND data::text LIKE '%checked against the FCA handbook%';  -- expect 0
```

```bash
# 1c. And at the artefact, which is the only thing that counts
scripts/probe-page-url.sh lendzy.co.uk about tool-affordability-complaint-checker-guide
```

---

# 2. WHAT IS DONE, AND THE EVIDENCE (all re-measured 2026-08-31)

| | evidence |
|---|---|
| **The claim is gone from lendzy** | 0 components carry it in `content_data` **or** `rendered_html` (was 3); both served bodies read **0** (was 2 and 1) at HTTP 200, with an **invented-URL 404 control** so the 200s are real pages |
| **Both spec sources stripped** | `content_direction` 08-26 (row `81ddcc40`), `strategy` 08-27 (row `0326a892`, tail-assert guard, history intact). Fleet census: **0** current specs, any aspect, any site |
| **The canonising audit item rejected** | `052d01b0`, `status='rejected'`, reason on the row. It was one Retry click from regenerating the page around the false claim |
| **Patterns LIVE in the chassis** | `v1.0.1349`, both pods; three-arm binary probe passes (2 new strings present, a pre-existing pattern present, a never-written string absent) |
| **Zero false positives fleet-wide, post-roll** | `cmd/claimscan` over **2,715** components (2026-08-31): **1** BANNED finding — the pre-existing `webdesign.co.uk` "never invents" — and **9** PRACTICE, all pre-existing P1–P5. **Nothing from either new pattern, and nothing from P6** |
| **Council** | `f4c144ad` APPROVED at round 2; `098` lists `fc588e445` as REVIEWED "by correlation, via submitted" |
| **The spec detector is live, correct and running daily** | CronJob `brief-negation-check`, `40 7 * * *`, image **`v1.0.1350`** (carries the false-positive fix `dc9ccfda2`), `doc_notes` heartbeat every run. Triggered production run 2026-08-31: exit 1 (sibling detector's findings, not a refusal), **0 of 36 sites, 3 suppressions** |

**What the framework wrote** (quoted because the gate that would vet it was inert when it was written,
so a human is the only reader it has had): the guide now says every figure and rule reference *"is
given together with the named rule it comes from and a pointer to where you can read that rule
yourself… That does not make the checker infallible… rather than take our word for it."* An
unverifiable claim about our diligence became a verifiable statement of what the site provides.

---

# 3. ~~THE ONE THING LEFT~~ — HOW THE LAST BLOCKER WAS DISCHARGED (history, 2026-08-31)

> **DONE.** The `copy_quality_two_stage` lane bumped `IMAGE_TAG` + the overlay `newTag` in one commit
> (`6d5f7911d`), built from committed HEAD, and applied — the CronJob artefact reads
> `brief-negation-check:v1.0.1350`. A triggered run then exited **1** (findings from the sibling
> detector, not a refusal) and reported **0 of 36 sites, 3 suppressions**. Kept below because the
> reading rules in §3a step 2 are the standing way to read this detector's output.

**The spec detector's first live finding was a FALSE POSITIVE.** *(Fix now deployed — see the DONE
banner above; this paragraph is the record of what it was.)*

On 2026-08-28 07:40 it filed `spec_supplies_claim` at severity **high** against **homegarden.uk**, for
the phrase *"We tested six lawn mowers so you don't have to."* — which sits in that site's
`content_direction.example_phrases.would_never_say` and `briefing.voice.avoid`. **It convicted a site
for a phrase its own spec BANS.** The item is rejected (reason on the row); the cause was a gap in my
reasoning, not the code: I excluded `evidence_base` *because* it stores banned phrases as data, and
did not generalise — a spec's negative-example lists are the identical trap in a different aspect.

Fixed in `dc9ccfda2` (label-context tracked *within* a block; exemption from the field path or the
governing label; suppressions **counted and printed**). Live dry run after the fix: **0 findings
across 36 sites, 3 suppressions**. Before: 1 site, 3 claims.

### 3a. What to do

1. **Rebuild and roll `brief-negation-check`** so its image carries `dc9ccfda2`. The
   `copy_quality_two_stage` lane owns the binary and agreed to fold it into their next image cycle —
   **talk to them first rather than rebuilding under them.**
   ⚠ The overlay **pins `newTag`**: bump it *in the same commit as the rebuild*, then read the
   artefact, never the make target:
   ```bash
   kubectl -n ai-persona-system get cronjob brief-negation-check \
     -o jsonpath='{.spec.jobTemplate.spec.template.spec.containers[0].image}'
   ```
2. **Read the next run** (`40 7 * * *`). Three things must all hold:
   ```sql
   SELECT created_at, body FROM doc_notes WHERE source='brief-negation-check'
   ORDER BY created_at DESC LIMIT 1;
   ```
   - **`N of M sites` — ⚠ CORRECTED 2026-09-02: SAY WHICH ONE. The report contains TWO
     `N of M sites` lines and they legitimately differ**, because the two detectors read
     different surfaces: the negation half (`bugs_open/305`) reports against the WRITER-visible
     surface (`11 of 34` on 2026-09-02) and the spec-claims half (this one) against the union of
     every live agent's surface (`0 of 39`). The line to read is the one **under the
     `# spec_supplies_claim` heading**. As written, this instruction sent a reader to the wrong
     number — the `copy_quality_two_stage` lane hit exactly that and had to untangle it. Original
     text kept below.
   - ~~**`N of M sites`** — `M` must be the whole fleet (36 today)~~, and the per-site `scanned_fields`
     count must be **non-zero**. A zero from a blind scan and a zero from a clean fleet are otherwise
     identical in this report.
   - **`X match(es) suppressed as negative examples`** must be printed. It is expected to be ~3. **A
     rising count with findings at zero means the suppression vocabulary is over-reaching, not that
     the fleet got cleaner** — that is the failure mode the fix could introduce.
   - **No `spec_supplies_claim` item filed** for homegarden.uk (or anyone) unless it is real.
3. **Then close 414.** Move the file, naming **both** paths on the commit, and verify at HEAD rather
   than at the tree — a `git mv` under a pathspec commit otherwise ships a copy:
   ```bash
   git add bugs_closed/414_*.md && git commit bugs_open/414_*.md bugs_closed/414_*.md -m "CLOSE 414 …"
   git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep 414   # exactly ONE line
   ```

### 3b. Why the bug is not closeable before that

The reported defect is fixed and live. But the detector that stops the class recurring is filing a
known-false finding until its fix ships, and the estate's bar is **fixed AND live**. Closing now would
record a mechanism as working while it is demonstrably mis-firing.

---

# 4. TRAPS — read before touching any of this

- **`kubectl logs` on `agent-chassis` returns nothing** (measured 2026-08-31), so the provenance-stamp
  recipe is unavailable. Use the three-arm binary probe in §1a. An empty stamp grep means *out of
  range*, never *unstamped*.
- **The spec detector scans with the PRACTICE family ONLY, and that is measured.** The fleet-wide +
  regulated set over 522 current spec rows gives **21** hits, effectively all false — 15 are the
  estate's own "never invent a person, company, scheme" honesty instructions (the negation guard
  cannot save them: the match *starts* at "never"), and every site's `banned_claims` are stored as
  data, quoting the sentences they forbid. `brief-negation-check`'s own header records this census as
  tried and **withdrawn within hours** on 2026-08-19. **Do not re-run it.**
- **`fleetSurface` does NOT walk steps — it regexes `default_config::text` as one document, and that
  is deliberate.** Exactly **one** live agent nests a `site_specs` ref inside a `sub_workflow` and it
  is **`page-content-writer`**, whose refs all live in its process-sections loop. A step-walking
  census would go blind to the most load-bearing prompt in the estate while reporting a clean fleet.
  If you ever convert it, use `platform/validation.WalkSteps`.
  `TestFleetSurfaceSeesRefsNestedInsideASubWorkflow` pins it.
- **Do not "tidy" P6 into a first-person form** to match its siblings P1–P5. The sentence that filed
  the bug has no subject, and a first-person anchor measured **inert** on it. Its anchor is on the
  VERB deliberately: moving it earlier puts "have NOT been checked" outside `negatedClaimMatch`'s
  backwards window and silently breaks suppression that works today.
- **Do not "tidy" the `formatted` test fixture in `specclaims_test.go`.** It reproduces the real row's
  shape, including **no blank line before the "Would never say:" label**. My first fix tested the
  block's *first line*, my composed fixture had a blank line so it passed, and the live data still
  failed. A fixture composed to match the assumption exercises the assumption.
- **`cmd/brief-negation-check/` is NOT in council scope** (the `cmd/` widening was targeted at
  `cmd/config-key-audit/`). `dc9ccfda2` was refused client-side and **not forced** — that is the scope
  working. The original change was admitted only because it also carried `platform/` files.
- **Exporting the corpus:** write inside the pod and `kubectl cp` it out, then count **three ways**
  (pod, local, DB). A single `kubectl exec` stream silently dropped **302 of 2,585** rows on 08-27,
  and a per-domain loop hit the `kubectl exec -i` stdin-eater. Full recipe in the RUNBOOK §4.

---

# 5. RESIDUALS — forward items, none blocking closure

1. **A POISONED EVIDENCE REGISTER passes every layer we have.** Found by the council's compliance
   seat. `evidence_base` is deliberately excluded from the spec scan because it stores the banned
   phrases themselves — so a fabricated *source* or *fact* written into the register sails through
   both the writer-side gate and the new detector, because every layer treats the register as ground
   truth rather than something to be checked. No instance found. **If one turns up it wants its own
   bug file**, not a widening of this detector.
2. **A COMPLETENESS mandate in a spec is not covered** by the spec detector (it scans the practice
   family). The page gate refuses that shape on the output side at blocker, so it cannot be served —
   but the instruction would survive in the spec and refuse every rebuild until a human reads the
   refusal.
3. **P6 has no post-deploy consumer.** The practice family is read by the build gate and
   `cmd/claimscan` only, not by `check_unverified_claims`, so it fires on rebuild and never over the
   installed base. Wiring it in changes what a shared sweep REPORTS on every site — a register entry
   plus a council round, not a quiet edit.
4. **Two P6 false-positive shapes are pinned by a test that asserts they FIRE**, so they are on the
   record rather than hidden: a third-party subject ("The FCA checks firms against the handbook, rule
   by rule") and the bare-"nothing" disclosure, which `negationCueRe` cannot see. Fixing the second
   means adding cues to a guard **shared by every claims family** — a guarantee change, architecture
   scope.
5. **`fleetSurface` duplicates census logic that `cmd/config-key-audit` already houses**
   (`relaygaps.go`, `sharedoutputs.go`, `livedeclarations.go`). Raised by the council's `reuse_agent`
   seat; a real unification debt, recorded rather than argued away.
6. **⚠ AN OWNER DECISION IS OPEN.** The council's `compliance` seat **dissented** on P6 shipping at
   WARNING rather than blocker, on a finance site where the fabrication ran 24 days. The owner ruled
   warning on 2026-08-27 and the counter-argument holds (a compliance-services client could say it
   truthfully; and at blocker this layer would refuse the honest correcting disclosure "Nothing here
   has been checked against the FCA handbook, rule by rule", which the negation guard cannot see).
   **It is on the record so it can be revisited deliberately rather than settle by default.**

---

# 6. WHAT THIS LANE GOT WRONG — read this before trusting any figure in it

Seven errors were logged in `WRONG_CALLS.md` (2026-08-27 and 2026-08-31). They share a shape worth
knowing before you re-use this lane's numbers: **every one was a claim made in passing, to support a
point that was not itself about the claim** — a green baseline asserted without measuring, a pattern
count added up rather than derived (four documents said "twelve"; it is eleven), a fleet rate taken
from a table that archives its rows, an **absence** concluded from a result set truncated by
`head -14` (three of twelve rows visible, the refuting row sorted first) and reported to another lane
as derived from source, a threshold misread inside the retraction of that very error, a frequency
inferred from a code comment and used to argue another lane out of running a query — and the detector
false positive in §3, whose first fix was validated by a fixture I had composed to match my own
assumption.

The measured work — corpus calibrations, mutation proofs, dry runs, artefact verification — held up.
The prose around it did not, and the prose is what travels. **Re-derive any figure here before
repeating it**; every one has the command that produces it, in the RUNBOOK or beside it.

---

# 7. THE STANDING FIVE

- `PLAN_2026-08-27_planted_marker_as_claim.md` — the six decisions and their reasons, incl. why the
  repair went before the detector and why no guard sits on `WriteSiteSpecAction`.
- `RUNBOOK_planted_marker_as_claim.md` — the retraction sweep, the guarded spec strip, the framework
  dispatch, the corpus export that took three attempts, the council schema, the artefact proof.
- `NOTES_planted_marker_as_claim.md` — the technical log, including every misstep, newest at the
  bottom.
- `README_where_we_are.md` — the owner's plain-prose log. **Append; never rewrite.**
- `SUMMARY_2026-08-27_the_claim_rules_now_read_the_instruction.md` — the milestone read-out. A second
  summary is owed only when the five headings would genuinely differ; closing 414 is that moment.
