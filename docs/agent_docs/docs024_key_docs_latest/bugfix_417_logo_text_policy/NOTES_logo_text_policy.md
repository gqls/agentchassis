# NOTES — 417 logo text policy (append-only, newest at the bottom)

## 2026-08-31 — session open

Asked to take 417 and 420, checking first whether either had a live thread. `who-owns.py` said
both were OWNED and ACTIVE — but it reads COMMITS, so it lags a session mid-fix. Messaged the
three candidate owners directly rather than trusting it. All three replied within minutes:
loanzy formally handed me 417's residual, the delivery lane handed me 420's class fix, and
boxingonline confirmed nothing of theirs was in flight on either. **Asking was worth more than
the ownership tool**, and the tool's own docs say so.

## First measurement, before planning anything

669 and 670 both applied, verified at the row (not at a tracker — there is no
`schema_migrations_agents` table, and `schema_migrations` has no `version` column; two wasted
round trips). Verbatim licences remaining: 0. **But the current-logo-prompt total had moved 27 →
28, and one wordmark row carried no 670 override.**

That row is boxingonline.com's, created **12:36:55Z — 41 seconds after 669 applied at
12:36:14Z**. And its wording is *"no text **other than** the wordmark itself"* where the
exemplar said *"**outside**"*. Two independent class points fell out of one row:

1. a migration that fixes a prompt SOURCE cannot bound prompts already in flight;
2. the model REWORDS the licence, so no literal match bounds the class — which cuts against the
   contributing lane's own §3 warning, since counting the licence by literal is still a literal.

## The misstep I made in my own brief, and what caught it

I briefed the planning agent that `bugs_closed/028` "proved the provider DISCARDS negative
clauses", and used it as a hard constraint. **That is 028's PRE-fix state.** Its fix,
`foldNegativeIntoPrompt`, is live. The agent went and read the banana adapter log for the exact
failing generation:

```
2026-08-31T12:55:50.145Z banana_provider "Banana: folded NegativePrompt into positive prompt
as a prohibition clause"  kind=logo  negative_prompt="people, faces, text, signature, watermark…"
prompt_len_before=232  prompt_len_after=407
```

`prompt_len_before=232` matches the boxingonline plan prompt exactly, and the timestamp sits
between the plan row (12:36:55Z) and the asset (12:56:10Z). **So the model RECEIVED "text" as a
prohibition and lettered "BOXING NEWS" anyway.**

The fix I would have shipped is unchanged. The *reason* changed completely, and that matters:
under my wrong premise, "the negative channel is broken" invites the cheap alternative of
repairing the negative channel. Under the true one, **a folded negative loses to a positive
licence in the same prompt**, so no amount of negative-channel work could have helped. I would
have shipped correct code with a rationale a reviewer could rightly have rejected. Logged in
`WRONG_CALLS.md` as *a closed bug's finding expires when its own fix ships*.

Second-order consequence worth keeping: the clause standing in `default_brand_prompt.go` was
itself negative-framed, so **it was weaker than its own comment claimed, for as long as it stood.**

## What the concept census found that the literal census could not

~8 of the 28 current-plan logo prompts name their exact wordmark **on purpose** (cv1
'CareerPrep', idea.uk, oufe, relojistas, robot-hands, webdesign.uk, lendzy, loanzy) — and **four
of them never use the word "wordmark" at all**, so 670's arm (b) never touched them. That is the
paraphrase point proven twice over, and it turned the opt-in field from polish into a
prerequisite: an unconditional text-free clause would have wrecked those on regeneration.

## It had already fired a second time, and only a human eye found it

The boxingonline session downloaded the served PNG and **looked at it**. It reads "BOXING NEWS"
on a site called Boxing Online. No query would have found that: `origin_prompt` looked ordinary,
the asset row said `status='active'`, the page served 200. **The defect is silent until someone
opens the image** — which is the acceptance gap this bug file names, and the reason candidate 2
keeps being tempting.

They also found a *different* defect in the same file: it is a **two-panel design comp**, not a
logo — the mark on navy left, the mark plus lettering on grey right. My guard cannot catch that;
it is an output-acceptance failure, not an input-licence one. Routed to its own bug file rather
than folded in, because the fix layer is different (store time, not prompt time).

## The liveness probe I ran, and threw away

Before claiming the guard's coverage was total, I checked whether the two kindless legacy parents
still dispatch. The `orchestration_states` probe returned zero — **and returned zero for a
known-live control too**, so it proved nothing. Recorded as `[UNVERIFIED]` with a LANDMINES entry
and a post-roll census disconfirmation, rather than as "those parents are dead". A probe whose
control also fails is not evidence.

## Council submission — three schema errors, all caught free

`DRY_RUN=1` caught `operation: "create"` (must be `add`), `risks` as an array (must be a string),
and then my own over-correction: I "consistently" turned `grounded_in` into a string too, and it
must be an array. **I applied one field's error message to a different field.** The validator was
standing right there and would have told me for free if I had re-run after each single change
instead of batching a fix with a guess. Logged in `WRONG_CALLS.md`.

## Verification against HEAD — and why the first read was misleading

`verify-head-builds.sh --with … --test` printed **`FAILED`** and **exited 0**. Three failures,
one a build failure, all in `test/`. None of them touched my files — but that is an inference,
not a measurement. Ran the bare-HEAD control: **it fails in ~23 places**, and every failure in
my set appears in the control's set. So the honest claim is "my changes introduce no new
failure", not "HEAD is green" — it is emphatically not green.

## 2026-08-31, later — the opt-in got a live consumer within the hour, and it exposed a near-miss

The loanzy lane set `constraints.wordmark_text = "farmerinsurance"` on farmerinsurance.uk's
current plan row (`b6680524`), giving the owner's named exception a durable home rather than a
flag-to-owner note. I verified it first-hand rather than taking the report:

```
domain            | farmerinsurance.uk
constraints       | {"wordmark_text": "farmerinsurance"}
company_name      | (empty)
logo_text         | (empty)
has_670_override  | t
locked_at         | (null)
```

**`company_name` and `logo_text` are BOTH EMPTY.** My validator grounds the requested wordmark
against company_name / logo_text / **domain stem**. Had I written it against the two naming
columns only — which is the obvious reading of "check it is this site's own name" — **the
owner's own explicit instruction would have been REJECTED and silently degraded to a text-free
mark**, on the one site where he had specifically asked for lettering. The domain stem is not a
convenience; on this row it is the only thing that makes the exception expressible at all.

It grounds because `domainStem("farmerinsurance.uk")` → `"farmerinsurance"` and the ask
normalises to the same token. Pinned as a named test case with the reasoning inline, and the
mutation run: **deleting the domain-stem source alone breaks it.**

Two things worth keeping from this:

- The fable plan had flagged the shape from the other direction — cv1's `'CareerPrep'` grounds in
  no column and no domain stem, so it would degrade. I read that as an edge case about cv1. It
  was actually telling me the identity columns are unreliable in general (18 of 39 sites carry no
  identity spec, per `default_brand_prompt.go`'s own census), and farmer is the same gap landing
  on the case that mattered most.
- **It gives disconfirmation D a real subject.** The post-roll census needs to prove the guard is
  not OVER-applied — that a deliberate worded mark still comes back worded. Farmer is now that
  test, with a live row, rather than a hypothetical I would have had to construct.

One scope note for honesty: my council submission states `constraints.wordmark_text` had **0
occurrences in the repo** as the RFC_022 third condition. That claim was scoped to the repo
(`platform/ internal/ pkg/ cmd/`) and **remains true** — this is a DATA row, not code, and the
field is still opt-in with the unsafe side OFF. The scope judgement is unchanged. But the plain
reading of "zero live consumers" is now stale by an hour, and the reviewer should hear it from me
rather than discover it.

## 2026-09-02 (afternoon) — second roll, and the census that finally has subjects

Chassis rolled again 15:39/15:53Z. Re-probed the binary: both fixes survived, and I now have TWO
removed-string controls rather than one — `resolveLogoIntent` (deleted in round 3) and
`email_was_intake_value` (deleted by 420) are both absent, so the probe proves the REVISION and not
merely that some version of the code is there. That pairing is worth keeping as the house method.

**Disconfirmation A is 2 for 2.** advertise.co.uk generated a logo at 14:48Z and its `origin_prompt`
carries the clause. boxingonline's does too. So the guard has now reached two independent sites.

**Getting there took unpicking both timestamp columns, and the second one caught me.** I already
knew `created_at` lies (the upsert keeps the original date). Today I nearly made the mirror error:
FIVE logo rows show `updated_at` of 2026-09-02 — relojistas, homegarden, idea.uk, agritec,
webdesign.co.uk — none of which carries the clause. Read naively that is five post-guard
generations the guard MISSED, i.e. disconfirmation A failing badly.

It is not. `updated_at` is bumped by ANY write, and something else is touching these rows (`433`,
the empty-`mime_type` sweep, is the obvious candidate). **Neither timestamp column is a
regeneration signal.** The instrument that settled it is `site_work_items`, which is
insert-per-dispatch: exactly ONE logo item has completed since the guard besides boxingonline —
advertise.co.uk — and three more (`websitepromotion`, `seotools`, `designblog`) are `triaged` and
pending.

I nearly filed "the guard missed five generations" on a column that cannot support the claim. The
general shape: **when a timestamp disagrees with a content check, the content check wins** — a
column records that a row was written, never what wrote it or why.

**Those three queued items are the fence trigger's subjects**, which is the useful outcome: the
decision I recorded as "not on n=1" now has a scheduled way to become decidable rather than sitting
as a permanent maybe.

⚠ **advertise.co.uk's PNG has not been looked at.** It is disconfirmation C's second data point and
the cheapest outstanding action in the lane. The census cannot see lettering; only a person can.

---

## 2026-09-02 (later session) — disconfirmation C closed at 3 for 3, and the census got a THIRD instrument

### The two eye-checks the handoff owed

**advertise.co.uk — CLEAN.** Zero lettering, single composition, a broadcast/signal mark
(concentric arcs from a mast). No invented brand name. Bytes pulled from the bucket through the
adapter pod; recipe now in the RUNBOOK.

**boxingonline.com — CLEAN, re-checked independently, at BOTH artefacts.** Fist-in-a-square mark,
no lettering, single composition. First from `boxingonline.ugg2.com` (the served copy, 400×218 PNG)
— then, because the handoff records that the delivery lane shipped an *interim* solid-ground logo,
also from the asset row's own source object (`images/demo_client/20260902/bf55d4ed…`, 1408×768
JPEG). **Same mark.** So the served file is a derivative of the guard's own generation, not the
interim replacement, and the eye-check is against the thing 417 actually produced. *Checking the
served copy alone would not have established that* — worth doing whenever a lane has shipped an
interim artefact for the same asset key.

**designblog.co.uk — CLEAN.** Generated 17:03:23Z, mid-session, off the queue. White geometric
star mark, no lettering, one composition. Carries the clause.

**seotools.co.uk — CLEAN.** Generated 17:10:10Z, also off the queue. Compass/target mark, no
lettering, single composition.

**So: disconfirmation A is 4 for 4, C is 4 for 4, and 421's two-panel shape did not recur on any
of them.** Only `websitepromotion.co.uk` is left of the three, and its generation was REFUSED at
17:15Z by 424's matte guard (`border_keyed=0`) — so its item is back to `triaged` and it will
retry. **A 424 refusal is not a 417 signal**: nothing about the wordmark policy is implicated, the
image never reached storage. Do not score it either way.

### The census instrument the lane was missing: the STORAGE KEY's date directory

The handoff says neither `created_at` nor `updated_at` is a regeneration signal, and points at the
work-item trail instead. **The work-item trail is also incomplete** — boxingonline's 10:40Z
regeneration has **no `needs_imagery` row at all** (checked every item type on that site since
2026-09-01: 22 `page_rerender`, an `owner_critique`, a `chrome_divergence_overwritten` — no imagery
item). A census keyed on the trail would have reported ONE regeneration today, not two.

The instrument that caught both, and then caught the third:
```sql
substring(a.storage_path from 'images/[^/]+/([0-9]{8})/') AS key_date
```
**It is sound by construction, not by luck** — `dynamic_adapter.go:717` builds every key as
`images/<client>/<YYYYMMDD>/<fresh uuid>.png`, so a regeneration can never re-use an old key:
```
relojistas.com     updated 09-02  key 20260729  ← updated_at bumped, NOT regenerated
homegarden.uk      updated 09-02  key 20260825  ← ditto
idea.uk            updated 09-02  key 20260621  ← ditto
advertise.co.uk    updated 09-02  key 20260902  ← REAL
boxingonline.com   updated 09-02  key 20260902  ← REAL (created 08-31 — proof the key is fresh)
```
Eight of ten logo rows touched today were not regenerations. Blind spot to state: operator-supplied
rows (`amend-asset.sh`) have no dated key — gaswholesalers.com returns NULL.

### The check I nearly skipped, and it would have been a WRONG_CALL

I measured 12 of 12 logo source objects to be JPEG and was about to tell the 424 lane that their
code comment ("banana has returned PNG in every observed case") was refuted. **All 12 samples
predate their 15:39Z roll.** Checked the roll time before writing — the samples could not speak to
the matted path. The claim survived only because the adapter's own post-roll log line independently
says `"source_format":"jpeg"`. *Measuring 12 things does not make the sample the right population.*

### The one I nearly got wrong in the other direction, and it matters more

The first live matte run produced a ground ~74 Euclidean units off the requested `#FF00FF`. The
424 lane's constants carry a `[UNMEASURED]` note explicitly asking the next session to "tune from
the first real output and date the change" — so I had exactly the number they asked for, and
drafted it as the tuning figure.

**It is contaminated and I nearly shipped it as tuning advice.** That run's prompt still carried
the magenta negative-prompt contradiction the council caught: `b2322a203` was committed at **17:25**
and the chassis pods (`v1.0.1354`) started **15:39/15:53Z**, so at 17:03 the running build still
had `logoBackgroundNegatives = {..., "magenta", "#ff00ff"}` while the clause told the model to
paint the whole ground magenta. The ~74 units are most plausibly that contradiction, not model
drift. **A number can be correctly measured, correctly dated, and still answer a question nobody
asked.** Rewrote the CONTRIB to lead with "treat §4 as contaminated".

### What actually matters in that run, and it is not the number

`border_keyed=1` — a perfect pass from the fail-closed guard — on an artefact with **0.0% fully
transparent pixels** (alpha extrema (57,255); of 4,348 border pixels, 0 keyed out). `BorderKeyed`
counts flood *membership* (`dist <= outer`), while transparency needs `dist <= inner`. So the guard
scores 1.000 on a ground it left 98% opaque. **That is structural and survives any threshold or
prompt fix.** Contributed to 424, not filed as a new bug — their lane, their fix, shipped today.

### Serving: I wrote a rule here that was WRONG, four hours before disproving it

What I wrote at ~17:30: *"advertise.co.uk 404s every path and serves a stranger's Drupal install;
only sites with `publish_project` set are served by the `*.ugg2.com` worker (2 of them fleet-wide
today)."*

> **CORRECTED 2026-09-02 19:45 (same session).** The second half is **false and was false when I
> wrote it.** `publish_target`/`publish_project` govern mirroring to a **second** hostname
> (`publisher.go`: *"copies the tree under a second hostname prefix … served by the existing
> `*.ugg2.com` worker"*). A site's OWN domain serves whenever its DNS points at the worker —
> nothing to do with that column. **Measured 19:45: websitepromotion.co.uk, designblog.co.uk,
> seotools.co.uk, advertise.co.uk and gamedesign.uk all have `publish_target` EMPTY and all serve
> 200 with a 404 invented-path control.**
>
> The first half was true at 17:00 and stale by 19:45 — advertise.co.uk now serves our own site,
> header and all. **The domain was repointed mid-session.**
>
> **How I got it wrong:** I read `publish_target` on 5 rows, saw 2 populated, saw the one domain I
> had actually curled return 404, and turned that into a serving rule. **One failing probe plus a
> plausible-looking column is not a mechanism** — and `publisher.go`'s own doc comment, which I had
> already read for the bucket name, says what the column is for. The cost: it sent me to the bucket
> for bytes I could have curled, and I wrote the wrong rule into the RUNBOOK and into the owner's
> README before disproving it. Logged in `WRONG_CALLS.md`.

Still true and still worth keeping: a `Host:` override against the worker is 403'd by Cloudflare,
and a parked domain 200s every path — so **always run the invented-path control**.

---

## 2026-09-03 (session 3) — the clean test happened, and it was mostly other lanes' runs

**State on arrival.** The 09-03 handoff's ⭐ item was "trigger ONE logo regeneration and look at it",
with a warning to coordinate because the 424 lane owned those sites. On checking, that had already
started: 424 reset three logo work items at 09:23:49Z and `seotools.co.uk` regenerated at 09:30Z —
the first artefact ever produced with both fixes live. So the test needed watching and reading, not
triggering.

**seotools, read at the served bytes (09:45Z).** 200, 26,975 B, RGBA, 404 invented-path control.
Storage key under `20260903/` — the sound instrument, since `updated_at` has already lied once on
this fleet. Eye-check: a magnifying glass over a woven lattice. **No lettering, single composition,
no invented brand.** Its prompt carried the wordmark licence AND the text-free override, so this was
the adjudication case, and the override won.

**gamedesign, read at 12:28Z.** Completed 11:41:09Z on **attempt 3**. md5 changed
(`b4f0ed1091f9` → `01076df06e90`), 298,938 B → 76,830 B, key `20260903/`. Eye-check: an abstract
maze in terracotta/tan. **No lettering, single composition.** Contrast median 2.39:1, max 6.46:1 —
legible.

**designblog FAILED, all 3 attempts, and it is NOT evidence for 417.** Error, verbatim:
`border_keyed=0.000, want >= 0.95 — refusing to store`. A genuine guard refusal, so **no artefact
exists to eye-check**. Recorded in 417 explicitly because counting a refusal as a "clean generation"
would pad the fence trigger's evidence base with runs that never tested the prompt.
⚠ Confirmed it was a refusal rather than a run killed by the 12:06Z chassis roll — the refusal is
timestamped 11:36:58Z, *before* the roll, and carries the guard's own statistic. A killed run leaves
no such line while still burning an attempt, and from the item status alone the two look identical.

### A hypothesis I formed from the image and the measurement REFUTED

Looking at gamedesign's maze, the white regions *inside* the mark looked opaque, and I reasoned that
`BorderKeyed` measures only the outermost ring, so enclosed ground would survive the guard **by
design** — a real gap, and one worth telling the 424 lane about.

**Measured it before saying so, and it is false.** Near-white opaque pixels: **gamedesign 0,
seotools 0**, websitepromotion 41 (0.05%). The white I was looking at is transparency showing the
page through. The matte reached the interior regions too.

The cheap check took one command and would have cost the 424 lane a wasted investigation had I sent
it as a finding. **The image is not the alpha channel** — what looks opaque against a white page is
indistinguishable from what is transparent against a white page, which is precisely why this needed
a measurement and not a look.

### Despill fringe, since 424 asked for readings on a good result

Magenta-ish opaque pixels as a fraction of the image: **gamedesign 0.01%, seotools 0.05%**, versus
**websitepromotion 0.62%** (pre-fix). Substantially better on both post-fix artefacts.

### The exposure figure, and why a literal is legitimate here

~~`origin_prompt` census, 5 sites: **5 of 5 carry BOTH the wordmark licence and the text-free
override** `[MEASURED 2026-09-03]`. So the override is load-bearing on every logo generation, not
occasionally — there is no population where the licence is simply absent.~~

> **CORRECTED 2026-09-03, ~1 hour later, same session — this was WRONG and it flattered the
> conclusion.** The override clause itself ends *"…presupposes a **wordmark** or any text"*, so every
> prompt carrying the override matches `%wordmark%` **because of the override**. I was counting the
> prohibition as if it were the licence, and the census could not have come out any other way.
> Re-measured with the override sentence stripped first: **0 of 5** carry a licence outside it, and
> the true figure is **1 of 5** — `designblog`, which says *"abstract letterform or typographic
> symbol"* and never uses the word "wordmark" at all. So the licence survived 669/670 by being
> **reworded**, which is this bug's own central finding happening to my own detector.
> **What caught it:** the 424 lane messaged to say it was resetting designblog, which sent me to read
> that prompt in full. Not a check I ran — a peer's unrelated message.
> Full correction, and what it does to the rate: `bugs_open/417`, the struck exposure section.
This is a literal match, which 417's own finding says is a floor. Legitimate for a **measurement**
(and only a measurement): being a floor, the true figure can only be higher, which is the direction
that keeps the conclusion safe.
⚠ **Re-ran this census against the post-regeneration rows.** The morning's reading of gamedesign
described a prompt that no longer existed — the UPSERT replaces `origin_prompt` along with the
artefact. Dated evidence about a row that has since been rewritten is not evidence about the row.

### websitepromotion regeneration — attempt_count reset to 0, deliberately

Owner approved regenerating it now the matting fix is live. Used the **work-item reset**, not a
hand-built `orchestrate` publish, which avoids the missing-ORCHESTRATION-headers trap that cost this
lane ~50 minutes on 09-02.

Followed 424's reset shape with **one deliberate difference: `attempt_count` reset to 0.** 424 kept
theirs because those were genuine retries of a run that had failed; this item had **succeeded** (it
produced the pale logo), so this is a fresh owner-authorised request and the ladder should start at
zero. The evidence says the margin matters: seotools needed 2 of 3 attempts, gamedesign needed 3 of
3, and designblog exhausted all 3 and stored nothing. Leaving the counter at 2 would have given this
one attempt against a guard that refuses roughly as often as it passes.

Bounded downside, stated before firing: if every attempt is refused, nothing is stored and the
existing logo keeps serving. The failure mode is "no change", not "worse".

### Fired into a settled cluster, not a rolling one

A fresh chassis (`v1.0.1358`) rolled at 12:06:47Z / 12:07:16Z. A spawn dispatched within ~300s of a
chassis restart is **silently dropped**, so the reset was held until 12:30Z. Re-verified both fixes
against the **new** adapter stamp `d0252fd4d` rather than carrying `7bf1ff674` forward —
`b2322a203`, `fcbe6071c` and `6440ec968` all ancestors, with HEAD as a negative control (correctly
not an ancestor). Ancestry, not dates: `kubectl` prints UTC and `git log --date=format:` prints
`+01:00` with no visible offset.

### The census, third attempt — and the lesson is the TABLE, not the regex

Having corrected "5 of 5" to "1 of 5" above, I then checked the fleet and **"1 of 5" was wrong too**
— a five-site sample generalised, where those five happened to be the most recently regenerated
sites. Fleet-wide `assets.origin_prompt`: **25 of 30 carry a licence**, 19 of them with no override
at all (they predate it).

Then a third error surfaced in my own re-measurement: I widened the regex to
`(wordmark|letterform|typograph|monogram|lettering)` and got 28 of 33 — **because the override text
itself says *"no lettering, words, letters, numerals or typography of any kind"***. So `typograph`
and `lettering` match the prohibition. **The same trap, three times in one hour, in three different
disguises.**

**What finally settled it was changing TABLE, not tightening the pattern:**
- `assets.origin_prompt` = what a past generation received. Right for "did the override arrive?".
  Wrong for exposure — it is a historical record including artefacts that can never be regenerated
  under their old prompt.
- `site_plan_imagery.prompt` + `sp.is_current` = what the next generation will be composed from.
  **The RUNBOOK's own census table, which I should have used from the start.**

**And it comes with a free self-check I had been missing: `plan_contains_override = 0` of 33.** The
override is appended at composition time, so it is absent from plan text — meaning a licence hit
there *cannot* be the prohibition quoting itself. One query, and it is the disconfirmation control
all three earlier attempts lacked.

**Forward-looking exposure `[MEASURED 2026-09-03]`: 13 of 33 current logo specs** carry a
licence-shaped term. Excluded `lettering` deliberately — including it inflates 13 → 28 by matching
prohibitions.

### websitepromotion regenerated — and it got WORSE, which is 462 reproducing live

Landed 13:08:44Z on **attempt 2**, fresh `20260903/` key, guard PASSED, artefact stored and served.

| | before | after |
|---|---|---|
| transparent | 84.3% | **93.4%** |
| median contrast | 1.43:1 | **1.01:1** |
| max contrast | 2.55:1 | 20.87:1 |

**Every transparency signal improved; visibility went down.** It is a chevron drawn in white and
outlined in magenta: **85.4% of the mark is near-white** (invisible on a white header, median 1.01
is white-on-white), and **of the 669 pixels actually visible, 420 (62.8%) are magenta**, 244 of them
near-exact `#FF00FF`. The key colour the matte exists to remove is the majority of the visible logo.

Mechanism: the model painted a **light** mark on the magenta ground. Matting removed the ground
correctly — hence 93.4% and a passing `BorderKeyed` — but a white/magenta boundary despills to
magenta, and with the interior also white that fringe is all that has contrast.

⚠ **This retracts a figure I sent the 424 lane this morning.** I gave them despill at 0.01% /
0.05% and called it closed. Both samples had **dark** strokes, where a thin fringe is cosmetic. The
magenta fraction here barely moved (0.62% → 0.48%) — what changed is that nothing else is visible.
**Despill severity depends on mark lightness and my two samples were both dark.** Told them, and
asked them not to close the item on my numbers.

⚠ **The previous artefact is GONE** — the UPSERT mints a new key and there is no rollback. The old
bytes survive only because I fetched them first (md5 `c80adffc9b23`, 41,062 B). Any decision to go
back depends on that local copy.

⚠ **It also constrains 462's fix candidate 1:** a contrast check must run **after** matting and
measure against **the header**. Pre-matte it would see high-contrast white-on-magenta and pass.
