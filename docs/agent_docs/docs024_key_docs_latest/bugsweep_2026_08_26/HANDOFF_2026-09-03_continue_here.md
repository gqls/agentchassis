# HANDOFF — bug sweep lane, 2026-09-03 (post-roll)

**Read this instead of `HANDOFF_2026-09-02_continue_here.md`.** That file is still correct
about mechanism and its §7 still lists the adjacent items; what changed is that the chassis
rolled, 338 shipped, and **338's acceptance test turned out to be unfalsifiable**. Where the
earlier file is now wrong I say so here rather than editing it.

**The lane is one owner decision away from closing.** There is no unfinished engineering.

---

## 1. STATE IN ONE TABLE

| item | state 2026-09-03 |
|---|---|
| **359** archived-still-serving | **CLOSED**, acceptance re-established, census instrument fixed (`4e26b1063`) |
| **404** reason vocabulary | **APPROVED r4** (`f2e4ac2a…`, 16:33Z 09-02, 3 advisory objections). Code committed. **Owed: read those 3 objections** |
| **407** nav declaration | **RULED, RUN, VERIFIED at the served page.** Drain was partial 09-02; re-sample if you care, do NOT re-dispatch |
| **338** voice gate on one value | **FIXED, APPROVED, SHIPPED on `v1.0.1356`.** ⚠ **OPEN only because its acceptance test cannot pass — see §3. This is the decision.** |
| **442** silent refusal *(new)* | **FILED, unowned.** Found by the council reviewing 338. Config-only partial fix is ~10 min |
| **320** meta_description backfill | Headline re-measured: **55.7% → 3.1%**. Residual changed KIND — see §4 |

---

## 2. THE ROLL — and why provenance is unavailable, which is NOT a doubt about the fix

`agent-chassis` runs **`v1.0.1356`**, pods started **2026-09-03 08:58:07Z**.

**Both sanctioned provenance routes failed, each for a reason already in LANDMINES.** Do not
spend time re-deriving this:

1. **The startup stamp had rotated inside 18 minutes.** Run the entry's own precheck first —
   `kubectl logs <pod> | head -1`. At 09:16Z the two pods' first lines were 09:16Z and
   09:10Z, so 08:58 was long gone. That is the *time-limited* case, not "unstamped".
2. ⚠ **`grep 'build provenance'` on this service matches LANDMINE TEXT ABOUT build
   provenance.** The chassis logs whole council/diagnosis payloads, so the recipe greps up
   its own documentation — my first attempt returned **1.9 MB**. Already a landmine;
   confirmed live today.
3. ⚠ **The release was applied from an UNCOMMITTED overlay bump.** `git show HEAD:` on the
   chassis overlay → `newTag: v1.0.1353`; the working tree and the running pods → `v1.0.1356`.
   `make release` bumps the makefile and overlays **in the working tree** and nobody
   committed it, so **the build point is not in git history and no ancestry check exists.**
   *(Tree `IMAGE_TAG` is already `v1.0.1357`, i.e. staged for the next build.)*

⚠ **Do not repeat my mistake here.** I ran `git merge-base --is-ancestor` with an empty
deploy-commit variable; both calls printed usage errors, exited non-zero, and my
must-be-absent control "passed" **because the command failed** — output read `ANCESTOR: no`
+ `CONTROL OK`, which is a confident false negative. **A control keyed on a non-zero exit
cannot discriminate: every failure mode of the command is also non-zero.** Full row in
`WRONG_CALLS.md`.

---

## 3. ⚠ THE DECISION — 338's acceptance test can never pass, and the reason is not the fix

**§6 of `bugs_open/338` says: watch the two blank pages fill. They cannot.**

`load_pages_missing_meta` (the backfiller's own selection query) requires
`page_visible_text_len(p.id) > 200`:

| page | visible text | eligible components | passes `> 200` |
|---|---|---|---|
| `leopardess…/case-study-automated-intelligence-pipeline` | **0** | **0** | ✗ |
| `oufe.com/contact` | **124** | 2 | ✗ |

The backfiller never selects them → the voice gate is never consulted → **the fix cannot
move them in either direction, ever.** A session running §6 would see "still blank" and
reasonably conclude the fix failed. That is the trap this section exists to prevent.

**Fleet-wide `[MEASURED 2026-09-03]`, with a demand control:** all **37** remaining blank
active pages (11 sites) fail the gate — **zero** selectable, averaging **8** characters of
visible text. The control that makes that meaningful: the **1,164** pages that DO have a
description average **4,401** characters, and **1,137 (97.7%)** clear the gate. The
instrument works; every remaining blank is a near-empty page.

**Why §6 was right when written.** On 08-20 those blanks WERE gate refusals. Migration `501`
(§7, the interim mitigation) asks the writer for ≤20 words, which clears the trip of 22, and
that population has since been filled. What remains is a *different* population, blank for
an unrelated reason. **A measurement can go stale by having its SUBJECT REPLACED, not only
by drifting.**

⚠ **And `501` makes the fix largely inert for this caller anyway.** ≤20 words clears the
default trip, so the natural exercise will not occur. The fix's real value is (a) the seam,
for the short-field callers 338 §5 and CQ-035 both expect, and (b) removing the trap for any
site that sets a lower `mean_sentence_words`. Both real; neither observable today.

### The three options, costed

1. **Close on the code proof.** Both arms induced against the LIVE leopardess gate config,
   every corpus-only case control-first, four mutations proven red, council APPROVED, image
   shipped. **Recommended** — state the caveat in the closing note rather than pretending to
   a live exercise.
2. **Induce a live exercise deliberately.** Needs a page with >200 chars of content on a
   default-threshold site AND a candidate whose **mean sentence length > 22 words** (or any
   sentence > 25). Decisive — but manufactured, and manufacturing it means writing copy the
   house style would reject. ⚠ Below that threshold the old gate would have accepted it too
   and the observation proves nothing: the "induce both arms" trap, one level up.
3. **Leave OPEN indefinitely.** Honest, but it will not resolve on its own — `501` prevents
   the natural case, so this waits on a caller that does not exist yet.

---

## 4. KNOCK-ON — 320's residual changed KIND, and that is also a decision

`bugs_open/320`'s headline was **407 of 731 (55.7%)** on 2026-08-19. Today it is
**37 of 1,201 (3.1%)**, and all 37 are the near-empty pages above. Updated in place with
dated figures.

**So what is left of 320 is a COVERAGE FLOOR, not a writing failure** — and *should a
near-empty page carry a meta description at all?* is an owner question, not a backlog. If
the answer is "no", 320's residual is zero and it can close too. If "yes", it needs a
different mechanism, because the backfiller structurally cannot see them.

---

## 5. WHAT IS ACTUALLY LEFT ON THIS LANE

**Blocking closure — decisions only (§3, §4).** No engineering.

**Owed, small:**
- **Read 404's three r4 objections** (`editquality`, `bug_historian`, `debug_historian`, all
  advisory). Both approved rounds this week still found real defects; that is the argument
  for reading them.
- **`bugs_open/442`** is unowned. Its candidate 3 — fix the `result_message` that names four
  refusal reasons and omits all three copy-gate ones — is config-only, live immediately, no
  roll, ~10 minutes, and removes an actively misleading surface.

**Adjacent, verified, none of them this lane's:**
- `platform/livespec` RED at committed HEAD (405 lane, `ffa1707b3`) — **unchanged, 8 days**.
- `TestNoHandSpelledTombstonePredicate` RED at committed HEAD on
  `check_unrendered_page_imagery.go:156/197/202/207` (`a87746b77`, 114/IMG-077 lane). A
  second instance of the same class.
- `_RELOCK` unclassified migration suffix — still WARNed by the 097 trigger.
- `WII-035` duplicate row id in the concept index (lines 415/416).
- ⚠ **The chassis overlay tag is uncommitted** (§2.3). Whoever owns releases should decide
  whether `make release` ought to commit its bumps — today the fleet's running tag exists
  nowhere in git.

---

## 6. FIRST COMMANDS FOR WHOEVER PICKS THIS UP

```bash
cd /home/ant/projects/agentchassis
git log --oneline -25

# 1. Do NOT run 338 §6 expecting a signal — read §3 above first. If you want the
#    numbers yourself, this is the query that shows why it cannot fire:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT CASE WHEN COALESCE(meta_description,'')='' THEN 'BLANK' ELSE 'has description' END AS grp,
       count(*), round(avg(page_visible_text_len(id))) AS avg_len,
       count(*) FILTER (WHERE page_visible_text_len(id) > 200) AS over_200
FROM pages WHERE status='active' GROUP BY 1;"
# expect: BLANK 37 / avg 8 / over_200 0   vs   has description 1164 / avg 4401 / over_200 1137

# 2. 442's cheap half — the result_message omitting voice_tell / banned_claim /
#    voice_gate_unreadable. Config-only, no roll.

# 3. 404's r4 objections:
#    diagnosis_artifacts WHERE correlation_id='f2e4ac2a-2bfc-4c82-ac99-d5fd7616edef'
#      AND kind='council_report' ORDER BY created_at DESC LIMIT 1
```

**Do NOT** "fix" 338 by raising a site's thresholds (it disables the rule for that site's
PAGES too), and **do NOT** follow 338 §4's own em-dash remedy — it is corrected in the file,
and following it would re-gate the seven sites that switched the rule off.

---

## 7. WHAT SHIPPED THIS LANE, FOR THE RECORD

`425398a01` the fix · `d8009cea8` docs + two corrections to the bug file's own remedy ·
`6413ad1f4` CQ-035 index row · `4c538efda` bug 442 + verdict · `8ad1c2869` notes ·
`8f1567f5a` closing the 09-02 handoff · `e41f6b332` post-roll findings.
Council `106802fc-ad14-4beb-b622-147c3a0ab982` **APPROVED**. Mechanism registered **CQ-035**.

⚠ **This tree swept my edits into other sessions' commits TWICE**, both on the fleet-wide
append-only ledgers (`LANDMINES.md`, `WRONG_CALLS.md`). Nothing was lost, but it reads as a
failed edit — **verify with `git show HEAD:<file>`, not `git log -- <file>`.**
