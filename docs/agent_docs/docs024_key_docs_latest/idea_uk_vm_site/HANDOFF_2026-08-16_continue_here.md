# HANDOFF 2026-08-16 — the head surfaces are clean, the BODY is regrowing. Read this, then §X.57–§X.58.

> ## ⛔ SUPERSEDED as the cold-start file — read `HANDOFF_2026-08-18_continue_here.md` first (2026-08-18)
> Still the reference for the honesty-arc history, §4 (the gate's head blind spot — still
> true) and the §7 traps. The §5 work list is done or ruled dead; the §6 residuals are
> re-verified in the 08-18 file (favicon/og-card now LIVE at the served page).

> # ⛔ SUPERSEDED IN PART — OWNER RULING 2026-08-17 CLOSED THE HONESTY ARC
> **§1 and §5 below are HISTORY, not instructions.** The owner ruled: *"we have dealt
> with the honesty problem enough. It doesn't need any more sweeps. just stop the overuse
> probably by a single stop word in the content writer agent."*
> **Done — migration `454`** adds rule 19 to `page-content-writer`'s `prompt_template`
> (config only, live immediately; ROLLBACK companion exists). **So: do NOT run the `090`
> on the regrowth, do NOT arm the gate more widely for this word, and do NOT re-fix
> `funding-fit`.** The 30 pages already carrying it STAY. The ruling also implicitly
> settles that this is a DENSITY problem, not zero-tolerance.
> ⚠ **If you are here to "clean the word out of the writer prompt", STOP** — four uses in
> that prompt are anti-fabrication rules about the model's own truthfulness and deleting
> them trades style for invented content. See §X.59.
> **Everything else in this file still stands**, in particular §4 (the gate cannot see
> titles or meta descriptions) and §6–§7.

**Supersedes `HANDOFF_2026-08-14_continue_here.md` as the cold-start file.** That file is
still correct on everything it describes and its §7 traps are still live — but its §5
job is DONE (banner added there) and its §6 numbers are now WRONG in a way that matters:
they went the wrong direction. `HANDOFF_2026-08-11` remains the reference for **RFC_015**
(decision records, citation gate, rebuild door); none of that changed.

Cold-start order: **this file → RUNNING_NOTES §X.57–§X.58 → `HANDOFF_2026-08-14` §7
(traps) → `HANDOFF_2026-08-11` §3 → `README_where_we_are.md`**.

---

## 1. The one thing to know

The fleet "honest" ban was implemented on 2026-08-12 by cleaning **instances** — 16 site
specs, 124 page sections, 12 sites. That worked, and it did not hold. **`[MEASURED
2026-08-16]` the reader-visible count has gone 53 → 18 → back up to 30 pages across 11
sites, and 23 of the 37 matching components were CREATED AFTER the sweep**, the newest
dated the same day I measured. It is ordinary LLM prose — `hero`, `generic-text-block`,
`call-to-action`, `faq`, `article-body` — on sites this lane never touched, not cloned
tool templates.

**So the open question is no longer "what did we miss", it is "what keeps writing it".**
Do not re-run the sweep first: it will produce a good number and you will be back here in
four days.

## 2. State — [VERIFIED] 2026-08-16, do not redo

| piece | state |
|---|---|
| **head surfaces** (`pages.title`, `pages.meta_description`, current `site_plan_pages`) | **0 fleet-wide**, denominators 686 / 246. Six fixed 08-15, all verified at the served page |
| **report hero blessed clause** | **RESTORED and VERIFIED AT THE SERVED PAGE** (owner ruling 08-16). `curl https://idea.uk/report.html` returns **exactly 1** occurrence of the word, and it is the blessed clause; item `complete`, 0 retries |
| **D-005** | FILED and enforceable — covers `report`/`hero`, guard asserts the served page contains `honest assessment`, tagged both `decision` and `decision-record` |
| voice gate | armed on 9 sites; **blind to titles and meta descriptions** (see §4) |
| body copy | **REGROWING — 30 pages, 11 sites.** Diagnosis DROPPED by owner ruling 2026-08-17; the 30 stay |
| **the word, going forward** | **migration `454`** — rule 19 in `page-content-writer`'s `prompt_template`. `[VERIFIED 2026-08-17]` at runtime: 3 writer calls after the apply, **3 of 3** carrying the rule in `llm_call_log.prompt_rendered`. Snapshot holds the pre-change config, ROLLBACK companion exists |

## 3. OWNER RULINGS 2026-08-16 — two, and the second is easy to miss

> *"restore it. you're is fine there."*

1. **The blessed clause goes back** and is now protected by D-005. Done.
2. **`whether you're` STAYS in that hero sentence.** This is a ruling **against a built-in
   tell phrase**. `ParseVoiceGate` appends 13 fleet-wide AI-tells that **nobody chose for
   this estate** (`voicetells.go:109`) and `whether you're` is one of them. **They are not
   owner rules.** If a `voice_tells` item is ever raised against that clause, close it —
   do not reword owner copy to satisfy a built-in. Ask before treating any of the other 12
   as a defect.

## 4. THE STRUCTURAL FINDING — two enforcement mechanisms, one shared blind spot

**`check_voice_tells` cannot see `pages.title` or `pages.meta_description`.** Read the
query, do not infer it: `ScanVoiceTells` (`check_voice_tells.go:171-177`) selects
`pc.rendered_html FROM page_components` and takes only `p.name`, `p.page_type`, `p.url`
off the page row. Those columns are not in `page_components` at all — the head is
assembled per render (`rerender_single_page_action.go:617-620` splices the title into the
site-level head; `:625`/`:1017-1028` fills a page-scoped blank with the description).

So **arming the gate does not protect the head**, and §X.56's census could never count it.
Proof it is not theoretical: `leopardessconsulting.co.uk`, whose own gate has banned the
word since 2026-07-18, served *"each honestly labelled"* in its meta description for 28
days. `validate_page_content` / `banned_claims` has the **same** gap at build time, so
agreement between them is not corroboration. Landmine filed and synced.

**RFC_015 has the same shape of gap:** the citation gate seams are component writes. A
rebuild can change a page's title or meta description citing nothing, and nothing refuses
it. That is how D-005's clause could be deleted with both edits reporting `complete`.

## 5. WHAT TO DO NEXT — in this order

1. ~~Confirm the hero at the artefact.~~ **DONE before handoff** — served count is 1 and
   it is the blessed clause. Re-check only if D-005's guard ever files a
   `decision_regression`, which is now the mechanism that will tell you.
2. **Diagnose the regrowth — DO NOT re-sweep.** Three candidate causes, and the second is
   one query away and untested by anyone:
   - **(a)** the sweep covered 16 specs of 23 sites; new spec rows still carry the word —
     `webdesign.co.uk/offer_ordering` (08-15) and `webdesign.uk/evidence_base` (08-14) are
     both current and both match, i.e. written after the sweep.
   - ~~**(b) `domain-research-classifier` may still WRITE the language.**~~
     **TESTED AND REFUTED 2026-08-16 — do not spend a session on it, and do NOT "clean"
     these prompts.** Both `domain-research-classifier` and `page-content-writer` DO match
     `honest` in their live `default_config`. **Every match is an instruction about the
     AGENT'S OWN truthfulness, not a directive to write the word into copy:** *"where
     research is thin, say so honestly in the confidence fields rather than fabricating
     detail"*, *"be strict about the mission but honest about evidence"*, *"if a field has
     no honest value, give it an empty string"*, *"it is ALWAYS better to be honest and
     general than specific and fabricated"*. Those are **anti-fabrication rules** —
     stripping them to satisfy a copy ban would trade a style preference for invented
     content. ⚠ **This is why a grep on a prompt is not evidence about behaviour: the
     polarity is invisible to the match.** A prompt saying *"never write 'honest'"* would
     hit identically.
   - **(c) ⭐ NOW THE LEADING CANDIDATE — a generic model habit**, in which case only an
     armed gate holds it, and the gate is on 9 of 23 sites and **files** rather than
     blocks. Sampling the new copy is what points here: the usage is ordinary English
     spread across unrelated writers and sites — *"an honest read of your own credit
     file"* (loancalculator), *"honest, critical feedback"* (idea.uk), *"the more honest
     way to see which loan costs less"* (loancalculator), *"an honest readiness tier"*
     (fundamentallyai), *"the failure modes named honestly"* (leopardess). No shared spec
     term, no shared component — just the word turning up wherever prose is written.
     **Note several are TOOL components regenerated after the sweep**, incl. idea.uk's
     `funding-fit` (*"1. Where is the idea, honestly?"*), which the 08-14 handoff listed
     as class C's one real fix — it has been REWRITTEN since and still carries it, so
     fixing that component once was never going to hold either.
   ⚠ **Do not count `voice`-aspect spec matches as failures** — those are the ban regexes
   themselves and must contain the word.
3. **This is a cross-cutting structural claim, so file a `090` before asserting a cause**
   (the 2026-07-31 ruling). "New copy carrying a banned word appears fleet-wide four days
   after the ban" is exactly the class.
4. Only then decide on remediation, and prefer a fix to the generator over another sweep.

## 6. Still open from earlier, unchanged

- **Class B** — 8 components, 3 sites, `content_data` NULL (2 also `component_id` NULL);
  real visible copy incl. `finetuning.uk/our-position-on-ai`'s `<h2>`. ⚠ **The 08-14
  handoff says this was "filed" — I could not find a `bugs_open/` case for it.** Either
  file one or stop describing it as filed. **Its TITLE was fixable and is now fixed** —
  the class-B filing over-reached from the body to the whole page, so re-check the other
  seven for the same over-reach.
- **Class C** — one genuinely reader-visible fix left: `funding-fit`'s *"1. Where is the
  idea, honestly?"* inside the tool's own markup. The rest are code comments; leave them.
- **Arming the gate on the remaining sites** — remember §4: whatever you set, it will not
  protect the head.
- **Older residuals** (from `HANDOFF_2026-08-11` §5), re-verified 2026-08-17:
  - **favicon / og-card 404 — CAUSE FOUND 2026-08-17, contributed into `bugs_open/131`
    (og-card slug).** Not a generator bug: the generator is never reached. 9 of the last 10
    `needs_brand_head_assets` items carry **no `spec.mode`**, so they fall through
    `asset-deployer`'s conditional chain to `deploy_image_asset`, which refuses correctly
    ("re-derive it (mode=brand_head) instead") — **and the item completes anyway** with
    `skipped:true`. 21 such items read `complete` fleet-wide. Also hits webdesign.co.uk,
    webdesign.uk, cookly.uk (**other lanes' — same one-liner, their call to run**). An
    idea.uk item with the correct mode was filed 2026-08-17
    (`created_by='claude-ideauk-brandhead-20260817'`); **verify at the artefact, both must
    be 200:** `/assets/images/favicon.png` and `/assets/images/og-card.png`.
    ⚠ **Still `triaged` at handoff, and that is QUEUE POSITION, not failure.** The
    dispatcher serialises per site (`NOT EXISTS (… status='claimed' … same site)`) and
    orders `created_at ASC`, so a newly filed item goes to the BACK. Another session
    started a whole-site reassemble at 15:41 on 2026-08-17 — **35 items ahead of this
    one**, and idea.uk completed 1 item in the 3 hours after. Give it hours, not minutes.
    If it is still `triaged` a day later, check for a stuck `claimed` row on the site
    (that one blocks every other item) rather than assuming the item is malformed —
    `attempt_count` stays 0 while it waits, which is the tell that it has never been tried.
    ⚠ Read the URL the PAGE references — `/favicon.ico` 404s and proves nothing, nothing
    links it. `logo.png` (the derivation INPUT) is 200, which is what makes this "the
    deriver never ran" rather than "no source".
  - news at `/data/latest-news.json` still **404**, `content_sources` for idea.uk still
    **0** (fleet 49) — untouched, and the 08-04/05 dispatch mystery (§X.53) is still
    un-diagnosed.
  - first organic signed Stripe webhook; tools-page card images and tool-page heroes; the
    empty-kind → SDXL image-routing hole; ingress landmines (`ufw allow 80,443` FIRST,
    grey second).

## 7. Traps this session hit or nearly hit

- **A queue depth is not a prediction.** 147 items ahead and zero completions on 3 of 4
  sites read as stalled; everything delivered ~45 minutes later. Both times.
- **`ORDER BY created_at ASC LIMIT 1` plus month-old items at the head is NOT proof of
  head-of-line blocking.** One query refuted it — `build-dispatch-loop` had just claimed a
  *newer* item on the very site holding the July rows. I was one step from filing it.
- **The `build provenance` log recipe is INOPERATIVE on agent-chassis** and now
  self-poisoning: chassis log history is ~90 seconds, and the logs carry LANDMINE TEXT
  about build provenance, so the grep matches the corpus and returns confident nonsense.
  Use the binary probe with a present-control and an absent-control, or skip it.
- **Filing a decision guard BEFORE the change is live** files an immediate, true, useless
  `decision_regression` against your own page. D-005's script refuses to insert until the
  phrase is really in the stored assembly.
- **Build a replacement string server-side with `replace()`** rather than retyping copy
  that contains `—`, `£` or an apostrophe. Assert the untouched half too: D-005's restore
  aborts if `whether you` went missing.

## 8. Recipes

```bash
# the head census -- BOTH layers, denominators visible so a zero cannot be blind
```
```sql
SELECT 'pages.title' AS surface, count(*) FILTER (WHERE p.title ~* '\yhonest') AS hits, count(*) AS scanned
FROM sites s JOIN pages p ON p.site_id=s.id WHERE s.domain NOT LIKE 'pool-%'
UNION ALL SELECT 'pages.meta_description', count(*) FILTER (WHERE p.meta_description ~* '\yhonest'), count(*)
FROM sites s JOIN pages p ON p.site_id=s.id WHERE s.domain NOT LIKE 'pool-%'
UNION ALL SELECT 'site_plan_pages (CURRENT)',
       count(*) FILTER (WHERE spp.title ~* '\yhonest' OR spp.meta_description ~* '\yhonest'), count(*)
FROM sites s JOIN site_plans sp ON sp.site_id=s.id AND sp.is_current
     JOIN site_plan_pages spp ON spp.plan_id=sp.id WHERE s.domain NOT LIKE 'pool-%';
```
```sql
-- is the body regrowing? the vintage split is the question, not the total
WITH v AS (SELECT s.domain, pc.created_at,
  regexp_replace(regexp_replace(pc.rendered_html,'(?is)<(script|style)[^>]*>.*?</\1>',' ','g'),'<[^>]+>',' ','g') AS t
  FROM sites s JOIN pages p ON p.site_id=s.id JOIN page_components pc ON pc.page_id=p.id
  WHERE pc.build_status IS DISTINCT FROM 'removed' AND s.domain NOT LIKE 'pool-%')
SELECT CASE WHEN created_at > '2026-08-12' THEN 'created AFTER the sweep' ELSE 'pre-sweep' END,
       count(*), count(DISTINCT domain) FROM v WHERE t ~* '\yhonest' GROUP BY 1;
```

Applied SQL of record: `idea_uk_vm_site/sql/2026-08-15_fix_head_title_meta.sql`,
`2026-08-15_dispatch_head_rerenders.sql`. Fleet CONTRIB:
`fleet_copy_quality/CONTRIB_2026-08-15_the_voice_gate_cannot_see_titles_or_meta_descriptions.md`.
