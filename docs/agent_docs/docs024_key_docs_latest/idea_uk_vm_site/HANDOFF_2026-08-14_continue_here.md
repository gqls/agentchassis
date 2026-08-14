# HANDOFF 2026-08-14 — the fleet VOICE arc. Read this, then §X.53–§X.56.

**Supersedes `HANDOFF_2026-08-11_continue_here.md` as the cold-start file.** That file is
still correct and still the reference for **RFC_015** (decision records, the citation
gate, the rebuild door) — none of it changed, and §3 of it is still the list of things
that will mislead you about *that* mechanism. This file covers everything after 2026-08-12.

Cold-start order: **this file → RUNNING_NOTES §X.53–§X.56 → `HANDOFF_2026-08-11` §3–§5
→ `README_where_we_are.md`**.

---

## 1. What this arc is, in one paragraph

The owner read `idea.uk/report.html` and said the copy sounded AI-written, was
relentlessly negative, and that **"honest" was spattered all over the place** — banned
across all sites, with one blessed exception (the report hero). He also asked for **fewer
"riddles"** (his word: *"slightly obscure follow-on text that you have to think hard to
understand"*), and for the report to say what it is worth **relative to a single agent
call the reader could make themselves**. It turned into a fleet voice pass, because the
cause was not the writer.

## 2. THE FINDING — the copy was commissioned, and the owner had already ruled on it

**Every sentence the owner objected to traces to `site_specs.content_direction`**, written
by `domain-research-classifier` on 2026-06-21 and never revised. *"A thinking partner, not
a verdict machine"* was a **defined key term, verbatim**. *"That honesty is not a flaw…;
it is the point of it"* mirrors a writing rule reading *"the site's honesty about limits
is a feature, not a weakness"*. Changing the model would have changed the wording, not the
shape — the sibling lane's 08-11 ruling stands: **the model is not the lever.**

**And the owner had already ruled this once**, on 2026-07-18, on
`leopardessconsulting.co.uk`: `\bhonest(ly)?\b` in a `voice_gate`, reason *"overused; show
the honesty, do not label it"*. It was enforced there for 25 days and **never propagated**
— `check_voice_tells` is live and driven by `quality-discovery-agent`, opt-in with the
unsafe default OFF, and only **2 of 23 sites** had opted in. The mechanism worked; the
adoption was the defect. [[zero-adoption-means-read-the-mechanism]].

## 3. State — [VERIFIED] 2026-08-14, do not redo

| piece | state |
|---|---|
| `report.html` rewritten | LIVE. antithesis **4 → 0**, negation **37% → 16%**, "honest" **2 → 1** (the blessed hero clause) |
| voice specs swept | **40 strings across 16 specs**, superseded not edited; `formatted` regenerated |
| page copy | **124 `section_edit` items, all `complete`** across 12 sites |
| visible-text census | **53 → 18 pages**, 9 sites |
| narrow gate armed | **7 sites** (+2 that already had one = 9 live) |
| gate has FIRED | 5 items since arming — it works |

Verified at the **served page**, not item status: one item reported `complete` while
carrying *"workflow completed but its result could not be delivered"*.

## 4. FOUR CORRECTIONS TO MY OWN EARLIER CLAIMS — inherit these, not the originals

Written in the notes as they happened; repeated here because a fresh session will
otherwise carry the wrong numbers forward.

1. **"Banned phrases only" is NOT achievable, and the gate is wider than I said.**
   `strawman` and `flourish_ending` fire **unconditionally** — no config. Worse,
   `ParseVoiceGate` **appends `globalTellPhrases()`** (`voicetells.go:109`), a built-in
   list of **13** fleet-wide AI-tells: `unlock`, `leverage`, `seamless`, `delve`,
   `tapestry`, `dive into`, `cutting-edge`, `game-changing`, `best practices`,
   `whether you're`, `at the end of the day`, `it's worth noting`, `in today's <world>`.
   **My 9-case verification passed 9/9 and still missed this**, because none of my test
   strings happened to contain a global tell. A green suite proves only what it exercises.
2. **Class C is NOT a shared-mechanism risk.** I said "15 components and they are SHARED".
   Joining through `page_components`: **14 of 15 serve ONE site and ONE page**; only
   `contact-block` is shared (2 sites). The double-domain suffix is a cloning artefact.
   **I asserted blast radius from a component NAME instead of a join.**
3. **Most of class C is CODE COMMENTS, not copy** — a JS comment in `contact-block`, a CSS
   comment in `gauntlet-interface`, one naming the `honesty_rails` compliance mechanism.
   Only `funding-fit`'s *"Where is the idea, honestly?"* is reader-visible. **~1 real fix,
   not 18.**
4. **Every count before §X.56 was inflated by `<script>`/`<style>`.** Reader-visible is
   what matters. Reuse this predicate, not a raw `rendered_html` match:
   ```sql
   regexp_replace(regexp_replace(html,'(?is)<(script|style)[^>]*>.*?</\1>',' ','g'),'<[^>]+>',' ','g')
   ```

## 5. A REGRESSION ALREADY HAPPENED — and it names the next job

`idea.uk/tools` `tool-list` was **recreated 2026-08-14 16:44** carrying *"An honest steer
on which funding routes fit your stage"* — **the exact string fixed on 08-12**.

**Cause: `pages.meta_description`, which I never swept.** The `tool-list` / `guide-list`
components render it **verbatim** as the card blurb — the same double-duty structure the
mortgagecalculator lane found for `pages.title` (*"`pages.title` does TWO jobs"*). Fixing
`page_components.content_data` alone is therefore **not durable**: the next rebuild reads
the page row and puts it back.

**4 rows fleet-wide, and they are the top of the next session's list:**

| site | page | text |
|---|---|---|
| idea.uk | `tool-funding-fit` | "…**An honest steer** on which funding routes fit your stage…" |
| idea.uk | `index` | (matches; check before editing) |
| leopardessconsulting.co.uk | `use-cases` | "…each **honestly labelled** and grounded in a system…" |
| mortgagecalculator.co.uk | `guide-first-time-buyer` | "**An honest and comprehensive** guide for first-time buyers…" |

## 6. What is open, with real numbers

**18 pages / 22 components still reader-visible**, in three classes:

- **B — no framework path (8 components, 3 sites).** `content_data` is **NULL**; two also
  have **`component_id` NULL**, so there is no template to re-render from and no field for
  `apply_section_edit` to write. Real, visible copy — `finetuning.uk/our-position-on-ai`
  serves an `<h2>` reading **"Our Honest Position on AI"**, plus three CTAs (*"Just an
  honest chat"*). **Filed, not hand-rolled**, per CLAUDE.md: when the framework cannot do
  it, that is a bug to file. Same defect the leopardess lane already knows.
- **C — component templates (13 components).** Mostly comments (see §4.3). The one real
  fix is `funding-fit`'s visible question label.
- **meta_description (4).** §5. Do these first — they are cheap and they are the only ones
  that have already regressed.

**Arming the rest of the fleet — read the volume first.** Reader-visible pages that would
fire, `[MEASURED 2026-08-14]`: `honest` **18**, `whether you're` **11**, `leverage` **7**,
`unlock` **3**, `best practices` **1**, `seamless`/`delve`/`tapestry` **0**; plus
`strawman` (~41 at last count) and `flourish_ending`. So a fleet-wide arm files roughly
**60–80** items, not zero — because of the built-ins in §4.1, not because of the owner's
rule.

> ⚠ **The measurement trap I nearly shipped into this file:** my first count of
> `whether you're` was **83 pages**, from a loose `~* 'whether you'` that also matches
> "whether you want". The real regex `\ywhether you'?re\y` gives **11**. A 7.5× error,
> caught only by re-reading the pattern I was claiming to measure.

## 7. Traps — read before touching any of it

- **A global cleanup regex damages text far from the edit site.** Mine turned dartsonline's
  *"the same weight as **an 80%** barrel"* into *"as **a 80%** barrel"* — on a string that
  merely contained the target word elsewhere. Structural checks cannot catch it. **Use
  exact substring replacement** for copy edits, never global rules.
- **Assert that every rule FIRED, not only that nothing is left.** 50 hand-written phrases
  produced 49 replacements; the miss was `"See the\nhonest test above."` — a newline
  mid-phrase, invisible to substring matching. Without that assertion the site would have
  been reported clean.
- **Blunt deletion breaks meaning, and no shape check sees it.** *"the rules stay that's
  why"*, *"as as the inputs"*, *"An and comprehensive guide"*, and cookly's heading
  *"What we're honest about"* → *"What we're about"*, which **inverts** it.
- **Target `apply_section_edit` by `page_component_id`, never `slot_name`.** Two slots on
  `report.html` are both `generic-text-block`; a name-keyed edit is ambiguous.
- **RFC_015 will refuse you, and it is right to.** Four edits to D-004-protected guide copy
  were gated, naming the decision. Re-fire with **`acknowledges_decision`** (D-004 stands;
  a one-word change is made in knowledge of it) — **not** `supersedes_decision`.
- **Do not "fix" what the owner has accepted.** The sibling lane's recorded mistake: four
  sound card descriptions reported as defects purely for matching a ban-list. The filler
  list is a smell, not a crime.
- **Do not touch:** the ban regex itself or its reason line; `submission` /
  `mission_brief` (the record of what was *asked for* — webdesign.co.uk's literally says
  "TONE AND HONESTY"); `vertical_landscape` (findings about *other people's* sites);
  `briefing.honesty_rails` (a compliance mechanism, `[VERIFIED]` no Go reader);
  "dishonest" meaning *unfair*.

## 8. Recipes

```bash
# reader-visible census — the ONLY count to trust (raw rendered_html over-reports)
```
```sql
WITH v AS (SELECT s.domain, p.id AS page_id,
  regexp_replace(regexp_replace(pc.rendered_html,'(?is)<(script|style)[^>]*>.*?</\1>',' ','g'),'<[^>]+>',' ','g') AS t
  FROM sites s JOIN pages p ON p.site_id=s.id JOIN page_components pc ON pc.page_id=p.id
  WHERE pc.build_status IS DISTINCT FROM 'removed' AND s.domain NOT LIKE 'pool-%')
SELECT domain, count(DISTINCT page_id) FROM v WHERE t ~* '\yhonest' GROUP BY 1 ORDER BY 2 DESC;
```
```bash
# arm the narrow gate on a site whose copy is clean (refuses to clobber an existing gate)
python3 docs/agent_docs/docs024_key_docs_latest/fleet_copy_quality/arm_voice_gate.py <scratch> <domain>[,<domain>...] --dry-run

# verify a gate against the REAL parser, not by reading the JSON — the scratch module
# under <scratch>/gatecheck has a `replace` onto the repo and runs ParseVoiceGate+ScanVoice.
# ALWAYS include a control that must come out NOT opted in.
```

Config of record: `fleet_copy_quality/voice_gate_narrow_2026-08-12.json`.
Fleet CONTRIB (for the 11 other lanes whose sites were edited):
`fleet_copy_quality/CONTRIB_2026-08-12_the_honest_ban_and_the_voice_gate_nobody_opted_into.md`.

## 9. Older residuals from this lane, unchanged

From `HANDOFF_2026-08-11` §5, none touched by this arc: the first organic signed Stripe
webhook; tools-page card images and tool-page heroes; `derive_brand_head_assets`
(favicon/og-card are live 404s); **news at `/data/latest-news.json` — still 404, and
`content_sources` for idea.uk is still 0** (§X.53: the 08-04/05 dispatch ids resolve to
nothing in any table, and that is un-diagnosed); the empty-kind → SDXL image-routing hole;
and the ingress landmines (grey is no longer a safe rollback — `ufw allow 80,443` first).
