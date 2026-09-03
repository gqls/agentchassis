# 338 — the voice gate's DENSITY rules are applied to a single sentence, where they are not measurements

**Filed** 2026-08-20 by the `meta_description_never_backfilled` lane. **Status: OPEN —
FIX COMMITTED 2026-09-02 (`425398a01`) AND SHIPPED ON `v1.0.1356` 2026-09-03 08:58Z.
⚠ CLOSURE IS BLOCKED ON AN OWNER DECISION, NOT ON WORK: §6's acceptance test can never
pass, and the reason is not the fix (§9).**

> **UPDATE 2026-09-02** (`bugsweep_2026_08_26` lane, continuing
> `docs024_key_docs_latest/bugsweep_2026_08_26/HANDOFF_2026-09-02_continue_here.md`).
> The Go change is written, tested and committed; it is inert until an image is rebuilt
> and rolled, so this file stays OPEN on the estate's fixed-AND-live bar. Council gate
> **SUBMITTED**, `106802fc-ad14-4beb-b622-147c3a0ab982` — verdict unread at time of writing.
> Mechanism registered as **CQ-035** in `docs026_concept_register/register/content-quality.md`.
>
> **Still biting when picked up**, re-measured before any code was written:
> `leopardessconsulting.co.uk` and `oufe.com` have exactly **1** blank active page each,
> and they remain precisely the only two of the **9** enabled gates that leave the
> thresholds unset. §3's table still holds. ⚠ But the blank page on leopardess is now
> `case-study-automated-intelligence-pipeline`, **not** the page §2 quotes — so the
> failure is RECURRING against new pages, not a single stuck row, and
> `orchestration_states` has since aged out (its refusal is no longer readable there).
>
> **Two corrections to this file are below, at §4 and §3.** Both are recorded rather
> than edited away.
>
> **⚖ COUNCIL VERDICT: APPROVED** (`106802fc-ad14-4beb-b622-147c3a0ab982`, 2026-09-02
> 18:25Z) — 9 approve, 2 advisory objections, none high-severity. The `architecture` seat
> recorded `ARCHITECTURE_SIGNAL: point_fix | DEFLECTIONS: 0` and confirmed the RFC_022
> shape (opt-in, unsafe side default, call sites enumerated rather than asserted).
>
> **⚠ AND THE APPROVED ROUND STILL FOUND A REAL HOLE, now filed as `bugs_open/442`.**
> The `bug_historian` seat objected [medium] that this fix corrects the false-positive
> TRIGGER but leaves the silent-loss MECHANISM intact: a refusal — right or wrong —
> returns a nil error, so the step COMPLETEs, the scheduled task stamps a clean run, and
> nothing downstream asserts on it. **Investigated and CONFIRMED**: there is no
> conditional on `save_result` anywhere in the live `meta-description-backfiller`
> workflow, and the one human-facing `result_message` names four refusal reasons while
> omitting all three copy-gate ones. So §2's observation that "nothing about it reads as
> a failure" was recorded here as *context* and is in fact **its own defect** — see 442.
>
> The other objection (`editquality` [medium]: the partial diff might orphan the `blocks`
> variable and fail to compile) is **answered by fact rather than by argument**: `blocks`
> is still passed to `checkBannedClaims`, the package builds, and
> `scripts/verify-head-builds.sh` reports OK at HEAD `425398a01`. A fair objection to
> raise from a partial diff — the sketch could not show it.

> **Resolve by SLUG** (`voice_gate_density_rules_on_a_single_sentence`) — bug numbers
> collide on this tree, and `git log` the FILE PATH, not the number.

---

## 1. What a reader needs first

A **voice gate** is a per-site set of copy rules held in `site_specs` (`aspect='voice'`,
key `voice_gate`). It carries **two different kinds of rule**, and the distinction is the
whole of this bug:

- **CONTENT rules** — `banned_phrases`, a list of regexes with reasons
  (*"harness — hype verb"*, *"trust — overused; say what is checked/verified instead"*).
  These are true of **any** string, however short.
- **DENSITY / DISTRIBUTION rules** — `mean_sentence_words`, `long_sentence_words`,
  `long_sentence_share`, `em_dash_per_1000_words`, `triads_per_page`,
  `expect_contractions`. These are **statistics computed across a corpus**. `ScanVoiceTells`
  computes them over a page's rendered blocks, which is what they were designed for.

`save_page_meta_description` (register **SEO-004**, `bugs_open/320`) applies the **whole**
gate — both kinds — to a single meta description, which is one sentence of ~15-25 words.

**Over one sentence the second kind degenerates.** "Mean sentence length" of one sentence
is just its length. "Long-sentence share" is 0 or 1. "Expects contractions" over 20 words
is noise. None of them is a measurement at that sample size.

## 2. The failure, observed in production

`[MEASURED 2026-08-20]` first scheduled run of the backfiller, on
`leopardessconsulting.co.uk`. The writer produced this, which is good ordinary copy well
inside a search result's display budget:

> "Read research-backed insights on AI adoption across healthcare, finance, hiring and
> data security, with findings on what builds and breaks confidence in these systems."

and the action refused it:

```
reason: voice_tell
detail: long_sentences: average sentence too long for the register
        ("mean sentence length 24.0 words")
```

24 words, against a default trip of **22** (`voicetells.go:179-180`,
`meanTrip := defaultF(g.cfg.MeanSentenceWords, 22)`).

The page stayed blank. **On an hourly schedule that is a permanent hourly LLM bill for a
page that never fills**, and nothing about it reads as a failure: the orchestration
COMPLETEs, the scheduled task stamps a clean run, and only `save_result.reason` says
otherwise.

## 3. The evidence that this is not just one lane's opinion

**`[MEASURED 2026-08-20]` of the 9 sites with an enabled voice gate, SEVEN have already
switched the length checks off by hand:**

| domain | `mean_sentence_words` | `long_sentence_words` | `long_sentence_share` |
|---|---|---|---|
| gamesdesign.co.uk | **100000** | **10000** | 1.1 |
| lendzy.co.uk | **100000** | **10000** | 1.1 |
| loancash.co.uk | **100000** | **10000** | 1.1 |
| noted.co.uk | **100000** | **10000** | 1.1 |
| relojistas.com | **100000** | **10000** | 1.1 |
| vetcomparison.uk | **100000** | **10000** | 1.1 |
| webdesign.uk | **100000** | **10000** | 1.1 |
| leopardessconsulting.co.uk | *(unset → default 22)* | *(unset → 25)* | *(unset)* |
| oufe.com | *(unset → default 22)* | *(unset → 25)* | *(unset)* |

Those are not thresholds. `100000` words and a share of `1.1` (i.e. >100%) are values
chosen so the check can never fire, while keeping the banned-phrase list.

**The estate has already voted on this, one site at a time, and nobody wrote down why.**
That is the part worth acting on: seven sites carry an undocumented workaround for a rule
that does not fit, and the two that did not carry it are the two this bug bites.

## 4. Where to fix it

`metaDescriptionFailsCopyGates` in
`platform/orchestration/actions/save_page_meta_description_action.go`.

It currently does:

```go
if findings := gate.ScanVoice(blocks, false); len(findings) > 0 {
    f := findings[0]
    return "voice_tell", fmt.Sprintf("%s: %s (%q)", f.Check, f.Reason, f.Matched)
}
```

`VoiceFinding.Check` already names which rule fired — `banned_phrase`,
`em_dash_density`, `triad_density`, `long_sentences`, `no_contractions`,
`flourish_ending`, `strawman` (`voicetells.go:85-93`). So the filter is available and
needs no new plumbing.

**The fix is to keep the content findings and drop the distribution ones** for a
single-value field. `banned_phrase` stays. `long_sentences`, `no_contractions`,
`em_dash_density` and `triad_density` are corpus statistics and should not gate one
sentence. (`flourish_ending` / `strawman` are judgement calls — decide deliberately
rather than by omission, and say which way in the code comment.)

⚠ **Do NOT fix this by raising a site's thresholds.** That is relaxing a checker to fit
the content, and worse, it disables the rule for the site's **pages** too — where it does
work and is the thing the gate was built for.

⚠ **Do NOT drop the em-dash rule as "a density rule".** The house style bans em dashes
outright; a *rate per 1000 words* is the wrong instrument over 20 words, but **any** em
dash in a meta description should still be refused. ~~If the density check is dropped, put
a flat "contains an em dash" test in its place rather than losing the rule.~~

> **CORRECTED 2026-09-02 — the REQUIREMENT above is right and the REMEDY was wrong; the
> implemented fix keeps `em_dash_density` unchanged.** Caught by doing the arithmetic
> before writing the replacement. The rule is `emDashes / totalWords * 1000`, a rate over
> **words**, so at 20 words one em dash scores **50.0** against a default trip of 3 — and a
> single em dash trips the default at any length below **333 words**, which every
> single-value field is. It *already* means "contains an em dash" here. Worse, a
> hand-rolled flat test would have **ignored site config and re-gated the seven sites**
> that set `em_dash_per_1000_words: 100000` to switch the rule off — a fact §3's table
> does not show, because that table omits the em-dash column. (Re-censused 2026-09-02: all
> seven set it.)
>
> **So the axis is not content-vs-density.** It is what the signal is a rate OF:
> a **rate over words** reduces correctly at n=1 and travels; a **count per page**
> (`triad_density` trip 4, `negation_density` trip 12) or a **share over sentences**
> (`long_sentences`) does not. `strawman` and `flourish_ending` are per-hit patterns and
> travel too — so the landmine's "filter to `banned_phrase`" advice would have dropped
> three working rules. `flourish_ending` is KEPT deliberately, as §4 asks: it anchors on
> the opening of the final sentence (`Ultimately,`, `In short,`), so at n=1 it is an
> ordinary pattern match on the only sentence there is.
>
> ⚠ **AND THIS SECTION'S OWN CHECK LIST WAS ALREADY STALE.** The list above omits
> **`negation_density`**, added by `bugs_open/305` — correct when written, wrong by
> birthday, and reading as current ever since. There are **8** check names as of
> 2026-09-02. That is why the fix does not hand-keep a list: `voiceCheckKinds` is
> exhaustive and `TestEveryVoiceCheckIsClassified` reads the `Check:` emission sites out
> of `voicetells.go`, failing on any unclassified check **or** any stale map entry.

## 9. ⚠ THE ROLL LANDED — AND §6's ACCEPTANCE TEST IS UNFALSIFIABLE (2026-09-03)

**The fix is shipped.** `agent-chassis` pods run `v1.0.1356`, started 2026-09-03 08:58:07Z.

⚠ **Neither sanctioned provenance route can confirm the commit, and both failed for
recorded reasons — do not read that as a doubt about the fix.**
1. The log stamp is out of range. LANDMINES' own precheck: pods started 08:58Z; at 09:16Z
   `kubectl logs <pod> | head -1` showed **09:10Z** and **09:16Z**, so the startup line had
   already rotated. This is that entry's time-limited case, not "unstamped".
   ⚠ Worse, `grep 'build provenance'` on this service now matches **LANDMINE TEXT ABOUT
   build provenance** — the chassis logs whole council/diagnosis payloads, so the recipe
   greps up its own documentation. Already a landmine; confirmed live today.
2. **The release was applied from an UNCOMMITTED overlay bump.** `git show HEAD:` on the
   chassis overlay says `newTag: v1.0.1353`; the working tree and the running pods say
   `v1.0.1356`. So the build point is **not in git history** and no ancestry check is
   available. ⚠ I first ran `git merge-base --is-ancestor` against an empty variable, and
   it "passed" its own control by ERRORING — a false green. Logged in `WRONG_CALLS.md`.

**AND THE ACCEPTANCE TEST CANNOT COME OUT EITHER WAY.** §6 says to watch the two blank
pages on `leopardessconsulting.co.uk` and `oufe.com` fill. They cannot:

| page | visible text | eligible components | passes the backfiller's `> 200` gate |
|---|---|---|---|
| `leopardess…/case-study-automated-intelligence-pipeline` | **0** | **0** | ✗ |
| `oufe.com/contact` | **124** | 2 | ✗ |

`load_pages_missing_meta` requires `page_visible_text_len(p.id) > 200`. **Both pages fail
it, so the backfiller never selects them, so the gate is never consulted, so the voice-gate
fix cannot change their state — ever.** A reader running §6 after the roll would see "still
blank" and reasonably conclude the fix did not work.

**Fleet-wide, `[MEASURED 2026-09-03]`, with a demand control:** all **37** remaining blank
active pages (11 sites) fail the gate — **zero** are selectable. The control that makes
that meaningful rather than a broken instrument: pages that DO have a description number
**1,164**, average **4,401** chars of visible text, and **1,137 (97.7%)** clear the gate.
The blanks average **8** characters. So the function works, and every remaining blank is a
near-empty page.

**Why §6 looked right when written.** On 2026-08-20 the blanks WERE gate refusals. Those
pages have since been filled — migration `501`'s ≤20-word instruction (§7) keeps candidates
under the trip. What is left is a different population, blank for a different reason. §6
was correct when written and is now measuring something that no longer exists.

### What could actually close this bug — an OWNER DECISION, not more work

⚠ **And note `501` makes the fix largely inert for THIS caller**: it asks the writer for
≤20 words, which clears the default trip of 22 anyway. The fix's value is (a) the seam, for
the short-field callers §5 expects, and (b) removing the trap for any site that sets a
lower `mean_sentence_words`. Neither produces a natural live exercise.

1. **Close on the code proof.** Unit tests induce both arms against the LIVE leopardess gate
   config, every corpus-only case is control-first, and four mutations were proven red. The
   image is shipped. *Recommended* — with the caveat stated in the closing note.
2. **Induce a live exercise deliberately** — a page with >200 chars of content and a
   >22-word candidate on a default-threshold site. Decisive, but it is a manufactured case,
   and manufacturing it means writing copy the house style would reject.
3. **Leave OPEN indefinitely.** Honest, but it will not resolve on its own: `501` prevents
   the natural case, so this waits for a caller that does not exist yet.

**A live exercise is decisive only if the written description has a mean sentence length
> 22 words (or any sentence > 25).** Below that the old gate would have accepted it too and
the observation proves nothing — the same "induce both arms" trap as §6, one level up.

## 5. Blast radius

**Small and bounded today**: `save_page_meta_description` is the only caller applying the gate
to a single-value field, and only **2 of 27** sites are affected.

> **CONFIRMED 2026-09-02 by enumeration rather than assertion**
> (`grep -rn 'ScanVoice(' --include=*.go | grep -v _test.go`): the other two production
> consumers are `check_voice_tells.go:214` (the page path) and **`cmd/voicescan/main.go:103`**,
> a CLI that scans whole HTML files — both corpora, both unchanged by the fix. ⚠ My first
> draft of that enumeration was wrong twice: it counted a **doc comment** at
> `save_page_meta_description_action.go:55` as a call site and **missed `cmd/voicescan`
> entirely**. Running the grep is what corrected it. The stale comment is repointed in the
> same commit.

**It grows the moment anyone reuses the gate on another short field** — a page title, a
nav label, an alt text — which is a reasonable thing to want, because the banned-phrase
half genuinely belongs there. That is why this is worth fixing at the seam rather than
per-caller.

## 6. How to verify a fix

```sql
-- the two affected sites, before and after
SELECT s.domain, p.name, length(COALESCE(p.meta_description,'')) AS len
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain IN ('leopardessconsulting.co.uk','oufe.com')
  AND p.status='active' AND COALESCE(p.meta_description,'')='';
```

**Induce both arms, or the test proves nothing:**
1. A candidate containing a banned phrase must still be **REFUSED** with
   `reason=voice_tell`, `detail` naming `banned_phrase`.
2. A clean 24-word candidate on `leopardessconsulting.co.uk` must now be **WRITTEN**.

A test that only exercises (2) is indistinguishable from one that removed the gate.

## 7. Interim mitigation already in place

Migration `501` asks the writer for **≤20 words**, which clears the default trip of 22.
That is a workaround and the migration says so in its own header. It is also better copy
on its own merits, which is why it was acceptable as an interim — but it does not make the
gate correct, and a site that sets `mean_sentence_words: 15` would break it again.

## 8. Provenance

Not run through `090`. **Substituting first-hand verification per the owner ruling of
2026-07-31, and stating the substitution rather than omitting it:** the failure was
observed in a real production run (the refusal text is quoted verbatim from
`orchestration_states.collected_data.save_result`), the deciding code path was read
(`voicetells.go:179-180`, `save_page_meta_description_action.go`), the threshold arithmetic
is arithmetic, and the 9-site config census is a direct query over `site_specs`. The claim
is narrow — one function applies a corpus statistic to a one-element sample — and does not
assert a cause outside the symptom.

Landmine written (`LANDMINES.md`, *"The voice gate's DENSITY rules are statistics over a
CORPUS…"*). Full trail: `bugs_open/320` §13 and
`docs024_key_docs_latest/meta_description_never_backfilled/`.
