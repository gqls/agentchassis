# NOTES — apis.uk bees home page

Running record, append-only, newest at the bottom. Includes the wrong turns —
those are the point, not an appendix.

---

## 2026-08-22 — session start: the constraint turned out to be the interesting part

Owner ask: *"build a page about bees for the apis.uk home page but without
affecting the dns for the tools-api that runs on that same domain."*

Took the constraint first, before designing anything, because it is the half that
can break something live.

**What the standing documents said.** Three of them agree, and all three are now
out of date:
- `features_open/020_FEATURE_apis_uk_traffic_probe.md`: *"apex will be repointed
  at the owner's planned bees homepage (separate thread) when that exists — one
  record swap, wildcard/probe unaffected."*
- `gauntlet_dead_cta/infra/island/RUNBOOK_island.md:76`: *"repoint ONLY the apex
  record at its hosting."*
- `gauntlet_dead_cta/NOTES_gauntlet_dead_cta.md:551`: *"apex rides the probe 404
  until then; swap is one DNS record."*

So the expected shape of this task was: build the site, then swap one apex DNS
record from the Cloudflare tunnel to the portfolio hosting.

**What is actually true [MEASURED 2026-08-22].** The apex is *already* served by
the portfolio worker, and no DNS change is needed at all. The zone carries **4**
DNS records and **2** worker routes as of 2026-08-22:

```
A      www.apis.uk   -> 192.0.2.1                      proxied
CNAME  *.apis.uk     -> f917c7c1-….cfargotunnel.com     proxied
CNAME  apis.uk       -> f917c7c1-….cfargotunnel.com     proxied
CNAME  tools.apis.uk -> f917c7c1-….cfargotunnel.com     proxied
routes: apis.uk/* -> portfolio-sites-router ; www.apis.uk/* -> portfolio-sites-router
```

A worker route intercepts at the edge *ahead of* the origin, so the apex CNAME to
the tunnel is vestigial. **Reading the DNS alone gives exactly the wrong answer**,
which is presumably how the three documents above came to agree with each other.

**Not inferred from config — the hostnames were asked, and they identify
themselves by their failure modes.** Four names, three distinct behaviours:

| hostname | response | therefore |
|---|---|---|
| `apis.uk` | 404, body `Not found` | the worker (`worker.js:91` returns that exact string) |
| `www.apis.uk` | 301 → apex | the worker (`worker.js:23` www→apex branch) |
| `zzqq-probe-test.apis.uk` | 404, **0 bytes** | island probe vhost :8082 |
| `tools.apis.uk` | 404, **0 bytes** | island Caddy :8081 → tools-api |

The body-length difference is what separates worker from island; the status code
alone cannot, since everything here 404s.

**The real hazard is not DNS at all — it is the worker ROUTE PATTERN.** DNS
records are per-name and independent, so the apex and `tools` never interfered.
But a route `*.apis.uk/*` would match `tools.apis.uk`, intercept the live API at
the edge, look up a B2 object that does not exist, and serve 404 — with no DNS
record touched and nothing looking wrong. `scripts/cloudflare/add_www_redirect.sh`
records that **24** zones already carry exactly that wildcard route as of its own
2026-08-18 measurement. apis.uk is deliberately not one of them. Filed as a
landmine.

**tools-api liveness control, before doing anything** [MEASURED 2026-08-22]:
`POST https://tools.apis.uk/api/v1/tools/gauntlet/round` with `Origin: vonc.com`
→ **200**. Root `GET /` → 404, which is the documented Caddy arm and NOT a
liveness signal (`WRONG_CALLS.md` records a session getting this exactly wrong).

## 2026-08-22 — scope put to the owner rather than guessed

The framework's fresh-domain pipeline builds a *whole site*; the ask was for a
home page. Rather than assume, asked. Owner chose **home page only** and a
**personal / enthusiast** angle — not beekeeping instruction, not conservation
campaigning.

Mechanism for binding that: the `roadmap_brief` site_spec. **Grounded in the live
agent definition, not in the oufe runbook that recommended it** — `build-site-planner`'s
prompt reads `{{.site_specs.specs.roadmap_brief.text}}` and states *"ROADMAP
OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase
below. … Do NOT invent additional pages. The roadmap is the authority for this
site."*

**Misstep, corrected in ~2 minutes:** went to `templates_db` for that definition
because the oufe runbook says `agent_definitions` lives there. It returned **0
rows**, which reads exactly like "no such agent" rather than "wrong database".
`agent_definitions` is in **`clients_db`** (**216** rows as of 2026-08-22);
`templates_db` has a same-named table with **8** rows as of 2026-08-22. Recorded
as a correction in the RUNBOOK, because the next person will make the same call
for the same reason.

**Checked that the cascade cannot undo the seed**, since a spec I write is worth
nothing if a later agent overwrites it. Queried every active agent's
`write_site_spec` steps by aspect: only `domain-submitter` writes `roadmap_brief`
(and its step needs `input_data.roadmap_brief`, which the 082 envelope does not
carry, so it errors to its `error_step` and moves on), and **nothing** in the
active set writes `evidence_base` or `imagery_style_guide`. Note that
`write_site_spec` **deep-merges** rather than replaces, so "something writes this
aspect" would not have been fatal — but it would have been a merge nobody
reviewed.

## 2026-08-22 — the evidence base, and three defects in my own seed

`loadEvidenceBase` (`platform/orchestration/actions/validate_page_content.go:1272-1290`)
returns `nil` on `sql.ErrNoRows`, and every claims lane then silently no-ops. So a
site with no evidence base is not "unchecked but fine" — it is **unchecked and
reporting clean**. Seeded before the first page exists.

Bees are an unusually bad subject for this: the field is made almost entirely of
famous repeated numbers (share of food owed to pollinators, flowers per jar, miles
flown, bees per hive, percentage declines, species counts), every one quotable
everywhere and sourced nowhere we hold. Plus one specific famous misattribution —
the "four years left to live" line Einstein never said. So `facts[]` is empty and
the bans target **shapes**, per the oufe precedent. **27** bans and **41** allowed
entities as of 2026-08-22.

**Then tested the ban list against sentences, and it failed three times.** This is
the part worth keeping, because every cheaper check had already passed:

1. **`\\\\.` decodes to a regex meaning "literal backslash", not "decimal point".**
   The seed's own stated safety net — *"an invalid regex degrades to a literal
   substring, so a typo never silently drops a ban"* — **does not fire**, because
   the pattern is *valid*, merely wrong. A valid-but-wrong regex never matches and
   reports clean for ever. `jq -e .` passes it happily; only decoding the string
   and *running* it shows the problem.
2. **`2 million flowers` escaped every pattern.** The digit-adjacent bans require
   the number next to the noun and cannot see a magnitude word between them — and
   "two million flowers to a jar of honey" is the single most repeated bee
   statistic there is. Added digit, spelled-out and bare-plural magnitude bans.
3. **`lives for 6 weeks` escaped**, because the pattern said `live` without the
   optional `s`.

Each was found by asserting on *sentences*, not by inspecting the JSON. The suite
now holds **22** sentences that must be caught and **8** ordinary bee sentences
that must stay clean, as of 2026-08-22 — and it earns its keep precisely because
it *did* come out negative three times.

**Misstep:** the first `cd`-relative fix attempt silently did nothing. The shell's
working directory had already been changed by an earlier call, so the `cd` failed
and `&&` short-circuited — and the "fixed" message never printed but the
re-validation ran anyway and showed the old patterns. Caught it because the
decoded patterns still showed four backslashes. Re-ran with absolute paths.
*A fix you did not see apply has not applied.*

## 2026-08-22 — seed applied, on the second attempt

First apply **failed and rolled back**: the `imagery_style_guide` insert was
missing its `FROM sites WHERE domain='apis.uk'` clause, so `id` was unresolvable
(`ERROR: column "id" does not exist … cannot be referenced from this part of the
query`). The transaction wrapper did its job — `SELECT count(*) FROM sites WHERE
domain='apis.uk'` returned **0** afterwards, i.e. the site row inserted moments
earlier was rolled back too. Fixed the clause, re-applied clean.

Verified after: site row with email present (`bugs_open/063` — the hallucinated-email
check FAILS OPEN with no contact email), three specs `is_current` and `pinned`,
**27** bans / **0** facts / **41** entities.

Submitted 12:18Z. `CORRELATION_ID=ba7a9c24-aea3-4fd0-9def-7e1d6f1cf891`.
Chassis pods were ~3.5 h old, well clear of the ~300s post-restart window in which
spawns are silently dropped.

## 2026-08-23 — THE MISSTEP: I put an infrastructure disclosure on a public page, and the framework did as it was told

The owner found this live on apis.uk and flagged it as serious. He is right.

**What was live.** Four sentences of this shape, one per content section:

> *"Away from the bees: this domain also hosts an unrelated technical service, on a
> different address, with no connection to anything on this page."*
> *"Separately from all of that, apis.uk is also the address of an unrelated technical
> service, run on a different part of the same domain. This page has nothing to do with
> it. It is only about the insect."*

**Whose fault.** Mine, entirely, and at the input. The 2026-08-22 mission brief I wrote
asked for it in as many words — *"A single short, plainly worded line somewhere
unobtrusive, acknowledging that the domain also hosts an unrelated technical service, is
welcome so that a developer who lands here by mistake is not left confused"* — and
`roadmap_brief` repeated it. **Do not file this against the framework.** Every layer did
exactly what the brief said.

The judgement was wrong in a way worth naming precisely: the owner's constraint was
*protect the API*, and I converted it into *tell everyone the API is there*. **A
constraint about protecting something is not a licence to describe it.** And the visitor
it was meant to serve — a developer arriving at a bees page by mistake — does not exist
in any number worth a public disclosure.

### Four things, and only the first is obvious

1. **I invented an outward-facing requirement nobody asked for.** Full stop.
2. **"Somewhere unobtrusive, once" names no location, so the writer satisfied it
   everywhere.** A section-by-section writer meets a placement-free instruction once per
   section. An instruction that cannot name which element carries it must not be written.
3. **It PROPAGATED, and fixing the origin would not have fixed it.** By the time I looked
   it was in **7** current specs as of 2026-08-23: `mission_brief`, `submission`,
   `identity` (as a fact about the site), `classification` (as a listed constraint),
   `strategy` (as a footer element), `content_direction` (the brief the writer actually
   reads) and `briefing` — where it had become an **acceptance criterion**, so a validator
   could have failed the page for *omitting* it. **A sentence in a brief is a seed, not a
   document.**
4. **I checked the wrong artefact.** I verified the page COUNT and the API's liveness.
   Both were true and neither was relevant. **I never read the page.**

### The repair, and the two traps inside it

Order mattered: specs first, bytes second, because the specs are what a regeneration
reads. Strip the sentences first and the next rerender writes them straight back.

- **Trap 1 — `content_data` is not what the renderer used.** I stripped the disclosure
  from `page_components.content_data`, confirmed **0** components dirty, re-rendered, and
  the committed file came back **byte-identical** (66,205). The renderer had used
  `page_components.rendered_html`, the cached render. `content_data` clean / `rendered_html`
  dirty on all six components. **"A rerender regenerates from `content_data`" is true of
  the SOURCE and says nothing about which layer the render actually reads.** Fixed both,
  re-rendered, and the file dropped to 65,250 with 0 hits — verified on the served page,
  not in the DB.
- **Trap 2 — a sentence-level scrub leaves ORPHANS that no trigger-word query finds.**
  My scrubber dropped whole list elements from structured keys but only offending
  *sentences* from long strings, so `content_direction.formatted` kept a dangling
  *"State it once, style it quietly, never present it as documentation or as a link to
  technical resources."* It contains none of the trigger phrases, so every "is it clean?"
  query I ran said yes. **It was caught only because I reimplemented
  `FormatContentDirection` and diffed my rebuild against the stored string** — a control
  I ran for a different reason. Regenerating `formatted` from the structured keys removed it.

### Fail-closed, not merely corrected

Removing the instruction stops the writer being *asked*. It does not stop the sentence
arriving from anywhere else. So `evidence_base` now carries an **OTHER-SERVICE DISCLOSURE**
ban class (**5** patterns as of 2026-08-23, catching the phrase, the different-host clause,
the possessive "this domain also hosts" form and the trailing disclaimer) plus an absolute
prohibition in `writer_block`. **An instruction deleted from a prompt is a decision no
future reader can see; a ban is.** Logged in `WRONG_CALLS.md` and `LANDMINES.md`.

## 2026-08-23 — the copy voice, traced to the same origin

The owner also said the copy reads as AI-written, citing *"worth sitting with"* and
*"not just"* negative framing. It does, and it traces to the same place.

`copy_quality_two_stage`'s `CONTRIB_2026-08-12` established that negatively-framed copy
originates in `site_specs.identity`, not in the model. That reproduced here exactly.
**Four of five `identity.unique_selling_points` were built as "X, not Y"** — *"reads like
a knowledgeable friend, not an institution"*, *"deeply rather than skimming everything"*,
*"No agenda — nothing to sell, nothing to sign up for, no conservation sermon"*, *"as they
live, not as things to be kept"*. They are a faithful encoding of my mission brief, which
was itself largely a list of prohibitions (*"What this page is not, and these are firm"*).

**The sharper cause is `content_direction.example_phrases.characteristic`** — the literal
exemplars a writer imitates. Four of five violated the house style:

> *"A returning forager does not simply arrive back at the hive — she announces where she
> has been."* (negative frame + em dash)
> *"A swarm looks like catastrophe. It is, in fact, reproduction."* (the manufactured twist)

And the live copy mirrors them: *"not just that the information passes between them, but
the translation involved"*, *"stranger fact than it first sounds"*. **The writer copied
its examples.** Giving a writer exemplars in the style you do not want is a stronger
instruction than any rule forbidding it.

**Fixed through the framework's own controls, which already existed and which I had simply
left un-populated:** `content_direction.voice / sentence_style / writing_rules /
things_to_avoid / example_phrases`. Exemplars rewritten positive-first with no em dashes;
the five house de-AI-ify rules from
`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md` added to
`writing_rules` (**16** as of 2026-08-23); the specific tells added to `things_to_avoid`
(**16**) and to `would_never_say`; `identity.unique_selling_points` restated as what the
page IS.

**`formatted` regenerated by a faithful reimplementation of `FormatContentDirection`,
validated by reproducing the STORED string before writing a new one** — the diff was one
line, and that line was the orphan above. Then `page-rebuild` fired for the domain, so the
framework rewrites the prose from the corrected brief. I am not writing the copy.

**Misstep inside the repair:** the first apply failed because psql read `\n` inside the
JSON payload as a meta-command (*"invalid command \nAssumed"*). The transaction aborted
cleanly. Re-applied by base64-encoding the JSON and decoding in SQL
(`convert_from(decode(...,'base64'),'UTF8')::jsonb`), which has no quoting surface at all.

### The rewrite is STAGED, not done — blocked on the API usage limit (owner already aware)

`page-rebuild` (orch `96d68aba`) FAILED at `call_rebuilder`:

> `AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 ... status 400:`
> `"You have reached your specified API usage limits. You will regain access on 2026-09-01 at 00:00 UTC."`

Fleet-wide, not this lane's doing — `llm_call_log` by day: 08-21 **1229 calls / 6 failed**,
08-22 **1179 / 116**, 08-23 **35 / 30**. `ai_endpoint_health` has `claude` unhealthy
(`last_healthy` 2026-08-23 09:44:13Z) with that same message; `cpu-ollama` is healthy and
`gpu-ollama` has never been. **Owner already knows — do not re-file this.**

**Everything the rewrite needs is already in place, so it is one command when the limit
lifts.** The corrected `identity` and `content_direction` are live and pinned, and
`pages.build_status` is still `needs_rebuild`, which is the precondition the trigger reads:

```bash
# re-fire; the corrected specs are already current
scripts/initial_messages/110_page_rebuild/072_page_rebuild   # after editing DOMAIN, or use the
                                                             # inlined envelope in this lane's history
```

**Do NOT hand-write the copy in the meantime.** The owner's instruction was explicitly to
put the copywriting through the framework, and the 2026-08-04 ruling says the same. The
current live copy is in the wrong voice but it is honest, contains no disclosure and
asserts no quantities — it is safe to leave standing until the framework can rewrite it.

**Verification owed when it does run** (do not close this without them):
1. Read the served page as prose, not as a row count — that is the check this lane already
   got wrong once: `curl -sS https://apis.uk/ | sed 's/<[^>]*>//g'`.
2. Assert **both** `content_data` and `rendered_html` are clean, per the render-cache landmine.
3. Re-run the tools-api liveness POST as an independent fact.

## 2026-08-23 — pre-flight on my own change: the ban list had five gaps, all found by testing sentences

With the rewrite blocked, the useful thing was to check that **my** change will work when it
finally runs. I added 32 bans yesterday; if they over-block, the rewrite fails or comes out
stunted, and I would not learn that until September.

Started from the right question — *does the list block the copy we WANT?* — and ran it over
**8,907** chars of the framework's own live bee prose. **Zero hits.** Good, and it could
have come out otherwise.

Then the sharper self-consistency check: **do my own new exemplars trip my own bans?** If
they did, I would be telling the writer to imitate sentences the gate rejects. They did not.
But it surfaced a genuine ambiguity I had introduced: my exemplar says *"The cells settle
into six sides"* while `writer_block` said *"do not state any count"*. A writer reading both
gets a mixed signal and could have dropped the comb-geometry section, which is one of the
best on the page. Added a **definitional-number carve-out**: a number that is part of what a
thing IS (six sides, six legs, one queen) is not a quantity claim; a number that could come
out different if measured again still is.

**Then I tested the carve-out against the bans, and found the first real gap** — and four
more behind it. In order, each found by asserting on a sentence rather than by reading the
list:

1. **`colonies fell by 40%` — nothing caught it.** The decline pattern required the decline
   word to follow the figure. This is the most loaded claim available on this subject.
   Fixed with a blanket percentage ban (consistent with `writer_block`, which forbids all
   percentages) plus a verb-first trend pattern.
2. **`the population has halved since 1990`** — `halved` is self-quantifying and offers no
   number to anchor on. Added the self-quantifying-verb class. ⚠ `half` as a bare word is
   deliberately NOT banned: *"half the bees leave with the old queen"* is the correct
   description of a swarm and is one of our own exemplars.
3. **`the species was named in 1758`** — no date ban existed at all, though `governing_rule`
   has always required dates to trace to a registered fact. Added, requiring a preposition
   so the footer's copyright year does not trip it.
4. **`sign up to our newsletter`** — slipped between `sign up` and `our` in the commercial
   pattern. Added a standalone audience-capture class.

**32 → 37 bans.** Final state: 23 forbidden sentences all caught, 12 permitted all clean,
8,907 chars of live prose clean.

### The checker is now a script, and it had the same disease it was written to cure

`check_evidence_base.py` in this directory. Its first cut returned empty prose on any
exception and printed **`ALL CONSISTENT`** — and Cloudflare 403s urllib's default
User-Agent, so **check 3 silently did not run on the very site it was written for**. That is
the blind-pass failure mode, in a script whose whole job is preventing blind passes. A
failed fetch is now fatal (exit 2), with an explicit `--skip-live` for the pre-deploy case,
plus a short-page guard so an empty body cannot be called clean.

**Misstep in verifying that fix:** I mutation-tested the new exit code with
`python3 check.py … | tail -3; echo "exit=$?"` and read **exit=0**, and briefly believed the
guard had not worked. `$?` after a pipe is **`tail`'s** status. The guard was fine; the
measurement was not. Re-measured with `cmd >/dev/null 2>&1; echo $?` → **2**, `--skip-live`
→ 0, control → 0.

### Fleet check, and a false positive I caught before reporting it

Having found five gaps in my own ban list, the obvious question was whether the pattern is
inherited — every site's `evidence_base` copies the same worked example, so a blind spot in
the recipe would be fleet-wide. Scanned all of it: **19 sites** carry a current
`evidence_base`, holding **205** ban patterns between them as of 2026-08-23.

**The apis.uk defects were mine, not inherited.** No other site carries the
literal-backslash bug, and no pattern anywhere fails to compile.

**But the scan first told me `webdesign.uk` had an invalid regex, and that was WRONG.**
Python rejected `"[^"]{20,}" ?[—,-]? ?(?-i)[A-Z][a-z]+ [A-Z]` with *"missing : at position
25"*, because bare `(?-i)` is illegal in Python's `re`. It is perfectly legal in **RE2**,
which is what Go — and therefore production — actually uses. Compiled all 205 under Go:
**0 rejected.**

I caught it only because I stopped to ask which engine evaluates these in production before
writing the finding down. Had I not, I would have told another lane their working pattern
was broken, with a plausible error message to back it up. **A validator that does not use
the production engine produces confident, well-evidenced, wrong findings** — and the
direction of the error is not predictable, so it can under-report just as easily.

Consequence: the Python checker is now labelled a fast authoring aid, `check_bans_re2.go`
sits beside it as the authority, and the whole apis.uk suite was re-run under RE2 —
**37 patterns, 23 forbidden all caught, 12 permitted all clean, 0 problems**. The suite
holds where it actually runs, not only where I wrote it.

## 2026-08-23 (afternoon) — the limit lifted mid-session, and the first rewrite attempt was REFUSED by a guard, correctly

**How I found out the blocker had gone: by accident, and it is worth recording how.** I
wrote `fire_page_rebuild.sh` with a refusal guard on `ai_endpoint_health`, and ran it
expecting exit 4. It dispatched instead — because `claude` had recovered at **12:12:18Z**
(74 calls in the following 20 minutes, 73 successful). The guard was correct; the world had
changed under it.

**Two missteps of mine in that one command.**
1. **I ran the script TWICE** — once silenced to capture `$?`, once to show output — and
   **each invocation dispatched a real orchestration.** Two concurrent rebuilds of the same
   page. They both died on the same guard so no damage, but *testing a dispatching script by
   running it dispatches*, and I should have exercised the guard logic in isolation (which
   is what I then did, and what the script's header now records).
2. **The live run only exercised the ALLOW arm**, so my "guard works" claim was unproven in
   the direction that matters. Driven directly afterwards: `HEALTHY=f` refuses, `HEALTHY=''`
   refuses (a failed query must never read as healthy), `HEALTHY=t` proceeds.

### The rewrite was REFUSED, and the refusal was right

```
SECTION SHRINK REFUSED for page "index" — hero 296→144 chars of VISIBLE text
(49% kept, floor 50%) ... Nothing was written (bugs_open/178, axis corrected by bugs_open/293).
```

**Nothing was written**, so the live page was never at risk. My first instinct was that the
guard was in the way — my tighter style rules naturally produce shorter copy, and the error
even names `section_shrink_floor` as the override. **Overriding it would have been the wrong
call, and the proposed content is why.**

### What the writer actually produced — read before deciding, not after

| section | opening |
|---|---|
| hero | *"A returning forager climbs onto the vertical comb and dances the direction she flew."* — **my exemplar 1, verbatim** |
| 1 | *"A forager returning from a good patch of flowers **does not simply** walk into the hive and stop."* |
| 3 | *"**Most bees live alone.** They nest in dry soil, in hollow plant stems…"* |
| 4 | *"**Most bees live alone.** They nest in dry soil, in hollow stems…"* |
| 6 | *"**Most bees live alone.** They nest in dry soil, in hollow plant stems…"* |

**Three sections opening on the same sentence, and that sentence is my own exemplar 4.** The
hero is exemplar 1 verbatim, under the bare label headline *"A page about bees"*. And the
negative frames survived anyway: `does not simply`, `is not one thing`, `looks nothing like`,
and `rather than` four times.

**So my exemplar fix backfired in a way I had explicitly predicted the opposite of.** The
CONTRIB I filed this morning said transfer was *selective and frame-shaped* and treated that
as reassuring. Rewriting the exemplars to be concrete, complete and on-subject turned
stylistic influence into **wholesale lifting**: a vivid finished sentence about bees is not
read as "write like this", it is read as "good material for this page". Addendum appended to
that CONTRIB, because a finding I gave another lane was wrong within three hours.

### The structural half, which is not a style problem at all

`section_plan_0.ready_names` = `hero, generic-text-block ×6`. **Six identical slots with no
per-section subject.** The writer is asked six times for "a section" with nothing to tell
them apart, so it reaches for the most concrete material in the brief — which, after my
change, was six finished sentences about bees. **Duplication here is what a contentless
section plan looks like once the brief contains anything quotable.**

### Fixes for attempt 2 (`afbc8d26`), and the shrink guard is now expected to pass on merit

- `roadmap_brief` **names the five section subjects** — waggle dance, wax and comb, a
  worker's changing job, swarming, solitary bees — one each, with *"cover solitary bees
  ONCE, in section five only"* and the three-way duplication named as the failure it exists
  to prevent.
- `example_phrases.how_to_use_these`: these are **style samples, not content**; no sentence
  may be copied; no section may be built around a subject an exemplar mentions.
- a `writing_rules` entry forbidding two sections opening on the same claim.
- the hero brief now demands **a real headline**, naming *"A page about bees"* as
  unacceptable — which should also clear the shrink floor honestly rather than by override.

**I have not touched `section_shrink_floor`.** If attempt 2 shrinks the hero past the floor
again *with a good headline*, that is the moment to consider the override — with the copy
read first, in that order.

## 2026-08-23 (attempt 2) — the voice is FIXED; the topics are not; and I nearly reverted a healthy page

**The voice result, measured on the served page, whole-page counts old versus new:**

| construction the owner named | before | after |
|---|---|---|
| `worth sitting with` | 2 | **0** |
| `does not simply` | 3 | **0** |
| `rather than` | 3 | **0** |
| `it simply` | 2 | **0** |
| `not just` | 1 | **0** |
| **total** | **12** | **0** |

**Every tell is gone.** Three changes landed together (exemplars marked style-samples-not-
content; a no-duplicate-opening rule; per-section subjects), so **which one did it is not
separable** — recorded as a guess, not a finding, in the CONTRIB.

### The misstep: I called a regression on a page that was mid-deploy

Straight after the rebuild I measured the live page at **12,272 bytes**, against 65,250
before. Inline CSS had collapsed from 51,023 to 2,270 bytes, the footer was gone, and three
section classes had no rule anywhere. I diagnosed a styling regression, went as far as
listing the previous commits in the sites repo to pick one to restore — and was one command
from reverting.

**It was deploy lag.** `git log` on `apis.uk/index.html` showed a LATER commit than the one
being served: `25f877fff` at 13:50, **64,085 bytes, footer present, disclosure 0**. The
rerender had already produced a healthy page; B2 had not finished syncing when I measured.
Re-fetched: **64,085 bytes, 51,023 inline CSS, footer present, 0 disclosure.** Nothing was
wrong.

**Why I got it wrong, precisely:** I treated "the orchestration COMPLETED" as licence to read
the served bytes immediately. Completion is the moment the commit lands, and the deploy is a
GitHub Action plus a B2 sync AFTER that. **A cache-buster does not help — this is not a cache,
it is a pipeline stage that has not run yet**, so `?cb=` returned the genuinely-current
object, which was genuinely stale. The check that settles it costs one command and I only ran
it once I had already drawn a conclusion:

```bash
git -C /home/ant/projects/sites log --oneline -4 origin/master -- <domain>/index.html
# then compare `git show <sha>:<domain>/index.html | wc -c` against what is being served
```

**If the repo holds a newer commit than the bytes you fetched, you are early, not broken.**
Had I acted on the first reading I would have reverted a good page over a regression that did
not exist, and blamed the framework for it in the handoff.

### What is genuinely still wrong: subject allocation

Six headings, and only two subjects between them:

- *Most bees keep no company at all* · *A solitary life in an old beetle hole* · *A nest built
  for one* · *The bees that live alone* — **four sections on solitary bees**
- *The jobs a bee does before she ever leaves the hive* · *One bee, several careers* — **two
  on a worker's changing job**

The dance, the wax and the swarm — all named in `roadmap_brief` — are **absent entirely**.

**Cause is arithmetic, not style.** The page carries **six** `generic-text-block` slots and
that roadmap named **five** subjects. A writer with a slot left over and no subject for it
duplicates the last thing it wrote. So the two failure modes are independent: **style is
reachable by prompt; subject allocation is not, while the plan and the brief disagree about
how many sections exist.**

Attempt 3 names **six subjects for six slots**, in order, and states both prior failures
explicitly so the instruction is about a known mistake rather than an abstraction.

## 2026-08-23 (attempt 3) — my six-for-six hypothesis was WRONG, and the result was worse

Attempt 2 produced 4 solitary-bee sections and 2 worker-job sections. I diagnosed that as
**arithmetic**: six `generic-text-block` slots against five named subjects, so the writer
duplicates to fill the spare. Attempt 3 therefore named **six** subjects for six slots, in
order, and stated both prior failures explicitly.

**It got worse.**

| slot | attempt 3 heading |
|---|---|
| hero | *A page about bees* — the exact placeholder the brief names as unacceptable |
| 1 | The same bee does several jobs in one lifetime |
| 2 | The same bee does different jobs across her life |
| 3 | A worker bee's job changes as she gets older |
| 4 | The bees that live alone |
| 5 | Most bees never see a hive |
| 6 | The bees that never see a hive |

**Three sections on a worker's changing job, three on solitary bees.** The waggle dance, the
wax and the swarm — named as sections 1, 2 and 4 in the brief — are absent again, and the
hero reverted to the string the brief explicitly rejects.

**So the count mismatch was NOT the cause, and naming subjects is not sufficient.** The
honest conclusion is that `roadmap_brief` reaches the PLANNER, not the section writer: the
plan it produces is still six identical `generic-text-block` slots carrying **no per-section
subject**, so whatever the roadmap says about "section 3 is X" never arrives at the moment a
section is written. The writer sees six indistinguishable slots and a brief full of bee
material, and writes six variations on whichever subject it most recently had in hand.

**[UNVERIFIED]** that reading of where roadmap_brief lands — I did not trace the planner's
consumption of it, only observed that `section_plan_0.ready_names` is six identical entries
after a roadmap that names six distinct subjects. The next person should check whether the
planner is expected to write per-section briefs into `pages.sections` and is failing to, or
whether that field was never part of this path.

### Where I stopped, and why

Attempt 3's output deployed. The live page became **11,991 bytes** — the `page-rebuild`
assembly, which produces an unstyled page with no footer (2,270 bytes of inline CSS against
51,023 when healthy). Attempt 2's render (`25f877fff`, 64,085 bytes, hero *"A closer look at
bees"*, footer, 12 voice tells → 0) is strictly better on every axis.

**I tried three routes to restore it and all three were refused by the harness**, which I am
recording rather than working around:

1. writing the known-good HTML into the `sites` repo with a shell redirect — refused;
2. `git merge origin/master` then committing the resolution — the merge staged **473** files
   (my clone was 840 behind), and a commit that broad in a shared repo was refused. **The
   refusal was right**: pushing 473 files to restore one page is precisely the sweep this
   estate's commit rules exist to prevent. Merge aborted; the sites tree is as I found it and
   my local restore commit is unpushed and harmless;
3. `UPDATE page_components SET content_data/rendered_html` from attempt 2's still-live
   orchestration `7304b797` — refused as a multi-row write to live content.

**Route 3 is the RIGHT fix and it is one statement.** Attempt 2's content is fully recoverable:
`orchestration_states` row `7304b797-3fce-4488-9073-933c80074556` still holds
`collected_data->'page_content_0'->'response'->'sections_metadata'` with both `content_data`
and `rendered_html` for all seven slots, mapped to `page_components.position` as `ord`
(0-based) + 1. ⚠ **`orchestration_states` is not an archive and rows are deleted** — if that
row goes, attempt 2's copy survives only as deployed bytes in `git show
25f877fff:apis.uk/index.html`.

What I did instead, being unblocked and strictly an improvement: fired `page-rerender`, which
re-assembles the stored content into the full styled page with its footer. That fixes the
STRUCTURE. It does not fix the copy — the duplicate headings and the placeholder hero are in
`content_data` and will persist until either route 3 runs or a fourth build is attempted.

**Owner decision, not mine:** whether to restore attempt 2's copy, accept attempt 3's, or
spend another build. I have stopped iterating — three builds is enough to have established
that more attempts at the same lever will not fix a defect that lives in the section plan.

## 2026-08-23 — DONE. Attempt 2's copy restored and live

Owner: *"just get the page done."* Stopped iterating on the brief and shipped the best
framework-written version we had.

Route 3 from the previous entry worked on a simpler statement (no TEMP TABLE, direct
`UPDATE … FROM (SELECT … jsonb_array_elements … WITH ORDINALITY)`): attempt 2's
`content_data` and `rendered_html` recovered from orchestration `7304b797` for all 7 slots,
then `page-rerender` to assemble and deploy.

**Live and verified at the artefact** (served bytes confirmed equal to repo commit
`fba43582f` before believing anything, per the deploy-lag landmine):

| check | result |
|---|---|
| served | **63,726 bytes**, HTTP 200 |
| inline CSS | **51,023** (styled) |
| footer | present |
| other-service disclosure | **0** |
| voice tells (was **12**) | **0** |
| `check_evidence_base.py apis.uk` | ALL CONSISTENT (37 bans, 23 forbidden, 12 permitted) |
| `tools.apis.uk` liveness POST | **200** |

**Known and accepted imperfection:** four of the six content sections are about solitary
bees (*Most bees keep no company at all* · *A solitary life in an old beetle hole* · *A nest
built for one* · *The bees that live alone*) and two about a worker's changing job. The
waggle dance, wax/comb and swarming subjects are not on the page. This is the section-plan
defect recorded above — six identical `generic-text-block` slots with no per-section subject
— and three builds established it is **not** reachable by editing the brief. Fixing it means
per-section subjects in the plan itself, which is a separate piece of work and needs the
planner path traced first (still `[UNVERIFIED]` where `roadmap_brief` lands).

**Do not re-render or rebuild this page without reading the entry above** — the specs are
correct but the plan is not, and a rebuild will re-introduce duplicate sections and the
placeholder hero.

## 2026-08-23 — images: five already existed and were unused; four more generated; all six embedded

Owner: *"There are insufficient images on this page, please can you add an image for each section."*

**First finding: the page was not short of images, it was short of images IN USE.** Five
illustrations already existed, generated 2026-08-22 from my seeded `imagery_style_guide`
(their `origin_prompt` begins with my exact palette, so the style guide is driving them) and
all five were live and serving:

| asset_key | serves |
|---|---|
| `illustration_waggle_dance` | 200, 191,747 bytes |
| `illustration_wax_comb` | 200, 278,998 |
| `illustration_swarm` | 200, 234,959 |
| `illustration_solitary_bee` | 200, 195,543 |
| `illustration_pollination` | 200, 204,315 |

The page referenced none of them — only `hero-home.jpg` and `logo.png`. **A silent mechanism
that is undriven, not missing.**

**Second finding, and it is the same defect as the copy:** those five cover the waggle dance,
wax/comb, swarm, solitary bees and pollination — **exactly the five subjects the page was
supposed to have and does not.** The images were generated against the intended plan; the
prose drifted to four solitary-bee sections and two worker-job sections. So the assets and
the copy disagree, and using a swarm illustration on a solitary-bee section would have been
actively misleading.

### Why I generated new ones rather than reusing what was there

Reusing `illustration_solitary_bee` across the four solitary sections would have made the
duplication **more** obvious, not less. So: four new illustrations, each matched to what its
section actually says, deliberately different from one another so related text does not get
four identical pictures.

| section | image | source |
|---|---|---|
| 2 · Most bees keep no company at all | `illustration_hive_vs_solitary` (crowd vs one bee) | NEW |
| 3 · A solitary life in an old beetle hole | `illustration_beetle_hole` | NEW |
| 4 · A nest built for one | `illustration_nest_cutaway` (burrow cross-section) | NEW |
| 5 · The jobs a bee does before she ever leaves the hive | `illustration_wax_comb` | existing — the text says she *"works wax into comb"* |
| 6 · The bees that live alone | `illustration_solitary_bee` | existing |
| 7 · One bee, several careers | `illustration_worker_stages` (one comb, four tasks) | NEW |

Generated through the framework: four `needs_imagery` work items (spec shape copied from
apis.uk's own `illustration_swarm` item) dispatched to `image-build-handler`. **No prompt
written from scratch for style** — the handler composes the site's `imagery_style_guide` in
front of the subject, which is why these match the earlier five.

### Traps hit

- **`Generic Text Block` has no image slot.** Its `html_template` is 211 chars: section →
  container → `h2.section__title` → `div.section__content`. There is no component in the
  library that is "prose section plus illustration", so the image goes inside the section's
  content HTML. **Precedent for that is the fleet's own** — `leopardessconsulting.co.uk` has
  3 Generic Text Block sections carrying `<figure class="infographic"><img …>` inside
  `content`. I copied the idiom rather than inventing one.
- **`figure` and `.infographic` have NO rule in apis.uk's CSS** (checked inline + external),
  so a bare `<figure>` would take the browser default `margin: 1em 40px` and inset every
  image. Used an inline `style="margin:0 0 1.75rem"`, which is self-contained and cannot
  affect anything else. `img` IS globally styled (`max-width:100%; height:auto; display:block`),
  so the images are responsive without touching the stylesheet.
- **`illustration-worker-stages.jpg` 404'd while its asset row said `active` with a filename
  and its orchestration said COMPLETED.** `b2 ls` showed the object present in the bucket, so
  it was edge propagation, not a failed deploy — 200 on the next attempt. **Had I trusted the
  DB I would have called it deployed; had I trusted the first 404 I would have re-run a
  generation that had already succeeded.** The bucket listing is what separated those.
- Both `content_data` AND `rendered_html` updated in the same statement, per the render-cache
  landmine — updating only the source would have re-rendered the old markup.

## 2026-08-23 — the images WERE added, and a sweep deleted them four minutes later

The owner sent a screenshot showing no images and headings this lane had never written
(*"The many jobs of a single bee"*, *"The bees that live without a hive"*). Both facts had
the same cause and it was not the image work.

**What happened, from the orchestration log:**

| time | event |
|---|---|
| 15:53 | four `image-build-handler` runs — illustrations generated |
| 15:54 | four `asset-deployer` runs — images live, all 200 |
| 15:56:15 | **`build-dispatch-loop`** — swept the page |
| 15:56:30 | `page-build-handler` |
| 15:56:51 | my `page-rerender` — deployed the page WITH images, verified |
| 15:56:56 | **`page-content-writer`** — began rewriting every section |
| 16:00:18 | all six sections replaced: new headings, **zero `<img>` in either column** |

**`pages.build_status='needs_rebuild'` is not a note-to-self, it is queue membership.** The
page had carried that flag all day — `page-rebuild` sets it and it had never been settled —
so a dispatch loop was entitled to regenerate the page at any moment, and did. My
verification (served bytes == repo commit, six sections with `<img>`, every image URL 200)
was **correct when made and expired four minutes later**. That is the part worth keeping:
the check was not wrong, it had a shelf life, and nothing in the check could reveal that.

**Fixed in the right order this time:** settle the flag FIRST (`build_status='deployed'` —
the state 703 of 839 pages are in), then re-apply the six images to the *current* section
text, then render, **then settle it again** — because `page-rerender` re-queues the page,
which I only found by re-reading the field after the deploy looked finished.

**Live and verified:** served bytes == repo commit `a7eeed551` (65,009), **7 `<img>` tags**
(six sections + logo), each section carrying a distinct illustration matched to its text,
`build_status='deployed'`, nothing in flight. Watching for 30 minutes to confirm it *stays*,
since the last verification is exactly the one that expired.

**Landmine filed** (`LANDMINES.md`, 2026-08-23) — and it is the third member of a set that
all fired on this one page today: the render cache (`content_data` clean while
`rendered_html` serves the old bytes), deploy lag (`COMPLETED` is the commit, not the
deploy), and now the rebuild queue. **All three make a page-edit verification lie in a
different direction.**

> **Note on a flagged ledger removal (`bdeaba560`).** `pattern-check` correctly flagged that
> the commit removed one line from `LANDMINES.md`, which is append-only and shared. Checked:
> the removed line was **my own**, from `7a690f068` twenty minutes earlier in this session,
> corrected in place because it was wrong (it said to re-check the flag after `page-rebuild`;
> `page-rerender` does it too). No other session's entry was touched. Recorded here rather
> than left to an auditor, because "the removal was mine" is exactly the claim that check
> exists to make someone prove.

## 2026-08-24 — footer removed, solitary-bee images corrected, three subjects filed for the framework

Owner asked for three things. All three are below, including the wrong turns.

### 1. Footer removed — and I aimed at the wrong function twice first

**The footer is NOT produced by `InjectFooter`.** That is the obvious candidate, it is in
`component_library.go`, it has a documented skip (`page already contains site-footer` →
early return), and I spent two render cycles exploiting that skip. It never fired, because
**`page-rerender` runs `rerender_single_page` (`rerender_single_page_action.go`), which
never calls `InjectFooter` at all.** The chassis log had no skip line, which was the tell I
should have read first instead of reasoning about the markup.

**The actual source is the `site_components` table** — site-level chrome, one row per slot:

| slot_name | rendered_html |
|---|---|
| `footer` | **969 chars, contained the email** |
| `head` | 48,051 (this is the 51KB of inline CSS) |
| `header` | 2,024 |

`rerender_single_page_action.go:609` resolves `footer` from `site_components` and
`:710` writes it **only `if footer != ""`**. So emptying that row is the supported way to
have no footer, and it worked first time. **Read the action the agent definition actually
names before theorising about the markup** — `page-rerender`'s `render_page` step maps to
`rerender_single_page`, and one query would have said so.

**The email lived in FOUR places** and clearing them one at a time produced three renders
with byte-identical output, which looked like the render ignoring me and was actually me
removing copies that nothing read: `sites.email`, `sites.company_name`,
`sites.content_data->'email'`, and `site_specs.briefing->…->'contact_email'` (plus a stale
copy in `submission`). The one the footer actually read was the `site_components.footer`
row's already-rendered HTML — a **cached render**, not a live lookup. Same shape as the
`page_components.rendered_html` trap from 08-23.

**⚠ Clearing an address is not automatically safe.** `multipage_actions.go:417` and
`section_editor_actions.go` **synthesise `ctx.Email = "info@" + Domain`** when a site has
none — so on those paths, removing the real address invents a fake one. The rerender path
has no such synthesis (checked: zero `info@` hits in `rerender_pages_actions.go` and
`v3_site_actions.go`), which is why it was safe *here*. A future multipage build on this
site could still print `info@apis.uk`.

**Compensating control, applied BEFORE removing the data:** `bugs_open/063` records that
the hallucinated-email check **fails open** on a site with no contact email. So
`evidence_base` gained an any-email-address ban (**38** bans as of 2026-08-24) first, so
removing the check's input did not leave a hole.

**Verified:** `<footer>` tags **0**, email in source **0**, "rights reserved" **0**,
images **7**, served bytes == repo commit `0479cf487`.

### 2. Solitary-bee images regenerated for anatomical accuracy

Owner: *"they tend to look slightly different to honey bees, please research this"*. He is
right, and the original four were drawn as honey bees on soil.

Researched rather than guessed. The decisive difference is **how pollen is carried**:
a honey bee packs it into a **corbicula** — a smooth compacted pellet on the *outside* of
the hind tibia; a mining bee (Andrena) has **no pollen basket at all** and carries it dry
and loose on a **scopa**, dense hairs on the upper inner hind legs and dusted through the
thoracic fur. Supporting tells: dark brown/black body rather than amber; densely furry
thorax, often foxy-red or buff; stockier and slightly smaller; wings faintly brownish with
darker tips; at most subtle pale abdominal bands, never bold yellow-and-black stripes.

All of that went into the prompts verbatim, including an explicit "do not draw a pollen
basket". Four regenerated under the same `asset_key`s, so the filenames and every page
reference are unchanged: `illustration_solitary_bee`, `illustration_nest_cutaway`,
`illustration_hive_vs_solitary`, `illustration_beetle_hole`.

**Checked by looking at them, not by trusting the pipeline.** The contrast image now shows
honey bees on comb with a visible packed pollen pellet on the hind leg, and beside them a
dark, foxy-thoraxed mining bee with pollen as a loose dusty scatter through its leg hairs.
The lone mining bee is correct too, at a nest hole with its spoil mound.
`cf-cache-status: BYPASS` on all four, so visitors get the new files immediately.

### 3. Waggle dance / swarm / pollination — filed where the framework acts on them

The roadmap_brief already names all three subjects. What it lacked was anything that makes
a writer act, because **`pages.sections` is a flat array of component names with nowhere to
put a per-section subject** (established over three builds on 08-23).

The framework's own per-suggestion content mechanism is the **`content_rewrite` work item**
handled by `page-build-handler` — 108 complete fleet-wide, most recent 2026-08-23, so it is
live and not folklore. Filed three, each carrying a `suggestion`, a `description` naming the
already-generated unused illustration, and an **`acceptance_test`** stating both what the
section must say and that no other section may cover the same subject.

**⚠ THE TENSION, STATED PLAINLY: `page-build-handler` is the agent that deleted the
embedded images on 08-23.** Hand-embedded `<figure>` markup lives in `content_data`, and a
content build rewrites `content_data`. So running these three items is expected to cost the
images, and they will need re-embedding afterwards. That is why one was dispatched as a
**canary** rather than all three, and why the images are being watched while it runs.
**The durable fix is not more hand-embedding** — it is a component with an image slot, or
image markup the writer is told to preserve. Recorded as the open question it is.

## 2026-08-24 — the mechanism: a component that carries its own image (CLC-030), and what it does not fix

Owner: *"create the component with an image slot and/or a section plan that can carry
per-section subjects … it's the mechanism that would be good to fix here."*

### The component-hierarchy route is NOT available — checked before building on it

The owner thought `features_open/035`'s "component containing other components" was built
but unused. **It is less built than that.** `page_components.parent_instance_id` is read by
**zero** Go files and set on **0 of 2,005** rows as of 2026-08-24, and that lane's own
summary says *"Nothing is built; the plan is staged."* Its P1 is council-gated and was
deferred on another lane's uncommitted WIP. So it is **a column and a design, not a
mechanism**, and building it here would duplicate an owned, gated project. Recorded in
CLC-030's relations so the next reader does not repeat the assumption.

### The component-creator produced something unusable

Fired the framework's own `component-creator` first, since the owner asked for this to go
through the framework. It returned a **`hero-headline`** component — CTA buttons, "trust
signals", `company_name` — filed under `section_type='illustrated-text-block'`. It ignored
the brief completely, and commercial boilerplate is the exact opposite of what a page with
no offer needs. Deactivated per the trigger script's own cleanup note, and the component
was written as a library row instead. **The component library IS the framework's extension
point; the generator is not the only door into it.**

### CLC-030 — `illustrated-text-block`

Five fields — `heading`, `content`, `image_url`, `image_alt`, optional `image_caption` —
with figure and caption **separately** gated, so with no image it degrades to plain prose
with no empty figure and no blank space (`bugs_open/111`'s rule). Keeps
`.section/.container/.section__title/.section__content` so the existing stylesheet applies
unchanged, ships its own scoped `<style>`, and satisfies all four component-creator
contracts (`data-component`, `<function>-section` class, a `--color-` var, an `@media`).

**Why it matters, in one line:** an image expressed as a component FIELD is data the writer
does not own; an image expressed as body HTML is prose the writer will overwrite. The six
hand-embedded images were deleted by a content rewrite four minutes after shipping on
08-23. That is the defect this closes.

**Live:** all six sections on apis.uk/index now render through it,
`data-component="illustrated-text-block"` ×6, every image 200, served bytes == repo commit.

### Two mistakes I made getting there

1. **I destroyed the images with a bad extraction.** Pulling section text out of psql with a
   `\x1f` field separator and splitting on LINES — the content contains newlines, so rows
   were truncated, the image regex matched **0 of 6**, and I wrote back six sections with
   the figures stripped. **A row-oriented parse of a field that can contain newlines is a
   silent truncation.** Redone entirely in SQL with `jsonb_build_object` and no text
   round-trip, which is what I should have done first.
2. **The canary did land, late, and rewrote everything.** The `content_rewrite` for the
   waggle dance first failed on the shrink guard, then succeeded — and rewrote **all six
   sections about the waggle dance**. So a single-subject suggestion regenerates the whole
   page. Content restored from the still-live attempt-2 orchestration (`7304b797`).

### What is STILL not fixed, stated plainly

**Subject allocation.** `pages.sections` is parsed as `[]string` (`PlannedSections`), so the
plan has nowhere to carry a per-section subject, and every slot gets the same brief. Making
it object-shaped is a Go change to the parse, with blast radius over every page the renderer
assembles — architecture-scope, its own register entry, not something to slip in beside a
site fix. **Four builds have now demonstrated the defect; none of them was a wording
problem.** CLC-030 makes imagery survive a rewrite. It does not make six sections differ.

## 2026-08-24 — the platform had already diagnosed me 25 times, and it prescribed the fix I built

A peer session (web_admin_console) flagged apis.uk as a stalled build. **Checked before
accepting it: wrong.** `needs_vertical_research` completed 08-22 12:37, eleven minutes after
it was created, and the whole cascade finished by 13:25; `site_unreachable` self-resolved at
16:21 once the page served. They were reading an 08-22 snapshot. Corrected to them directly.

**But their prompt made me look at the work queue for this site, which I had not done since
filing my own items — and it was carrying 27 open human-review rows.** My earlier look showed
three because `head` truncated the list. **A truncated listing is not a count**, and I had
treated it as one.

### 25 × `page_divergence_overwritten`

Twenty-five records, newest 08-24 10:55, each one the platform noticing that a rebuild
overwrote hand-patched bytes on this page and archiving the outgoing HTML to
`page_component_history`. **Every one of them is this lane, and every one was right.** That
is the measured cost of embedding illustrations as raw markup in `content_data.content`
instead of declaring them: I generated twenty-five divergence records over two days.

**And the item's own `fix` text names both remedies, one of which I had already built and the
other I had missed:**

> *"if the patched content should exist, re-declare it in content_data **or lock the
> component (058)** — do not paste it back into rendered_html, which only re-arms this same
> loss."*

The first half is exactly CLC-030. The second half I had not done, so I did it: **all 7
`page_components` on this page now carry `lock_type='permanent'`, `locked_by='apis-uk-bees-lane'`.**
`save_page_sections_action.go:460` documents the locked-slot path — the human-locked copy is
kept and the incoming copy discarded — which is what finally stops a `build-dispatch-loop`
sweep destroying the illustrated sections. **43 components across 6 other lanes already use
this**; it is established practice I simply had not found. All 25 records resolved with that
explanation; nothing was lost, each was archived at the time.

### `brief_supplies_negation` — a detector firing on my own prohibition

A live check reported *"apis.uk's brief hands the writer 2 phrase(s) built on
define-by-negation (0 mandated onto pages)"*, naming the `not_x_but_y` and `rather_than`
shapes in `content_direction.formatted`.

Both were present because I **banned** them — `mandated: false`, and the check says so. This
is the known shape from MEMORY ([[prompt-text-poisons-its-own-detector]]): a ban list
containing the banned phrase trips the detector that greps for the phrase.

**One instance was a genuine defect and is fixed:** `things_to_avoid` quoted both phrases
verbatim inside a rule. Given this lane MEASURED on 08-23 that exemplars get lifted, printing
a bad phrase in the brief is a real exposure, so it now describes the shape without printing
it. Whole `content_direction` object written, never a patch (`bugs_open/327`: a partial write
shrinks the brief the writer reads).

**The rest were left in place deliberately**, and the item was **annotated, not closed**,
because its spec says the decision is the SITE OWNER's: the remaining hits are ordinary
English use of "rather than", plus `example_phrases.would_never_say`, a contrast list whose
whole purpose is holding bad examples. Evidence it is working rather than leaking: after
those entries went in, the constructions the owner objected to went from **12** on the served
page to **0**. Stripping them would trade a measured benefit for a better detector score.

### Final state

Live: **67,464 bytes**, 6 sections rendered through `illustrated-text-block`, 7 images all
200, no footer, no email, `build_status='deployed'`, **7 components permanently locked**,
`tools.apis.uk` 200. Open review rows on this site: **2** — the negation item (owner's call)
and one historical `save_refused_incomplete` from 08-22.

## 2026-08-24 — the peer's "oldest open item" pointer found a live placeholder headline

The web_admin_console session re-measured, confirmed the cascade was never stalled, and named
its own error more precisely than I had: the query was honest when run on 08-22, but the
sentence *"about two days now, not progressing"* was **arithmetic on a stale row, never
measured** — a derived duration delivered in the same voice as the measurement. Worth
recording as a shape: **the derived half of a claim can be invented while the measured half
is sound, and the derived half is the one that sounds like evidence.**

They also confirmed my resolution landed independently (25 rows `complete`, one timestamp),
and pointed at `save_refused_incomplete` as the oldest open row, suspecting the
`bugs_open/012` truncation family.

**It is not that family — the guard REFUSED the save, so no fragment was persisted; that is
the guard working.** But following the pointer to the hero found something real and live:

> **`headline: "A page about bees"`** — the exact string `roadmap_brief` names as
> unacceptable, sitting on the served page.

It arrived in the third build (08-23) and survived every later render because renders read
stored `rendered_html`. **Fixed without writing copy:** restored the hero
`content_data` that the framework itself produced in attempt 2 (orchestration `7304b797`) —
headline *"A closer look at bees"*, with its own subheadline, taken as one coherent unit
rather than composed by me. Live and verified: `<h1>A closer look at bees</h1>`.

**Queue cleared, with reasons rather than silence:**
- **4 × `needs_page`** (`image-build-handler`, "re-render after imagery changed") → `complete`.
  Superseded: the re-render happened, the regenerated images are live and 200. Their two
  failed attempts were against the shrink guard while the page was mid-conversion.
- **2 × `content_rewrite`** (swarm, pollination) → **`deferred`, not abandoned**, with the
  unblock condition written into the row: the subjects are still wanted and their
  illustrations still exist unused, but (1) adding ONE section is not possible — the sibling
  item rewrote all six sections about the waggle dance — and (2) the components are now
  permanently locked, so a rewrite would be discarded anyway. **Running them today spends a
  build to produce a refusal.**

**Open on this site: 2.** `brief_supplies_negation` (annotated, owner's call) and
`save_refused_incomplete` (historical, guard worked, no residual damage found).

**Final live state:** 67,457 bytes · `<h1>A closer look at bees</h1>` · 6 sections through
`illustrated-text-block` · 7 images all 200 · no footer · no email · 7 components permanently
locked · `build_status='deployed'` · `tools.apis.uk` 200.

## 2026-08-24 (owner session) — GTM fleet-wide, and option D landed inside a typed struct

### Google Tag Manager

Owner: use `GTM-PQ3WCTBD`, make it standard for new builds, and backfill existing sites.

**Corrected my own earlier count:** I told the owner "three sites" carry a Google tag. Measured
properly: **14 of 27** sites with a `head` component already had `GTM-PQ3WCTBD` — I had sampled
three domains by curl and reported the sample as the population. **A sample is not a census.**

Backfilled the remaining **13** by inserting the canonical snippet (copied verbatim from
vonc.com's head, not re-typed) immediately after `<meta charset>`. All **27** heads now carry it
**exactly once**, 0 duplicated — asserted inside the transaction by counting occurrences, not by
eyeballing.

⚠ **The stored head is not the served page.** GTM reaches visitors only when a page re-renders:
**695 deployed pages across 28 sites**. Not fired — a fresh chassis was mid-build and a spawn
within ~300s of a restart is silently dropped. Owner decision on batching.

⚠ **"Standard for new builds" is a CODE change, not data.** `ChromeSlotFunction("head")` asks the
library for `function='head'`; the only two such components are **inactive**, and `Document Head`
is deliberately ineligible (`component_level='section'` — using it would render a page section as
`<head>`). So every site's head comes from **`RenderFallbackHead`**, a Go function. Two routes,
both platform-scope: amend that function (small, precise, needs council + a roll), or activate an
eligible library head component carrying GTM (no code, but it changes head rendering for **every**
site at once). Put to the owner rather than chosen.

### Option D — and it would have silently failed the obvious way

`imageryStyleGuide` is a **typed struct**: `palette`, `medium`, `mood`, `avoid`, `provider`,
`reference_asset_keys`, `kinds`. **An added key such as `subject_accuracy` is dropped on
unmarshal — no error, no log, and the prompt would simply never carry it.** Same family as the
`input_schema`-fallback landmine. Checked the existing spec for unknown keys first: none.

So accuracy went into the fields that are actually read, and the split is better than a single
blob:

- **Negatives → `avoid`**, which the adapter routes to the **negative prompt**. That is a
  *separate channel*, so ~1KB of "no corbicula on a mining bee, no amber-and-black stripes on a
  solitary bee, wrong species for the stated subject" costs the length-limited main prompt
  **nothing**. This is the direct answer to the owner's context-length constraint.
- **Positives → the FRONT of `kinds.illustration.medium`**, because `medium` heads the composed
  prompt (the file header states the composition is medium+mood+palette, and every stored
  `origin_prompt` on this site opens with guide text). Vital detail therefore survives truncation.
- Scoped to the `illustration` kind so hero and logo generation are untouched.

### `visual-design-auditor` cannot see images — but the platform can

Its steps are `query_database` → `execute_llm_prompt` → `write_audit_findings`: **text only.**
But `execute_vision_prompt` is a **registered, live action** (`registry.go:1208`,
`execute_vision_prompt_action.go`), backed by `aiservice.VisionCapable` /
`GenerateWithImages` on both Anthropic and Gemini — **and `tool-acceptance-agent` already uses
it in production.** So option C (generate → vision critique → regenerate) is reachable as
**agent configuration, not new code**, which makes it far cheaper than I first told the owner.
Inputs it takes: `images_field`, `max_images`, `prompt_template`, `output_type`.

## 2026-08-25 ~17:30 BST (session "apis.uk") — the owner split the lane: Google goes, the rest stays

### Measured first, before believing the 08-25 handoff

`[MEASURED 2026-08-25 ~17:00 BST]` apis.uk: HTTP 200, **67,877 B** (same as the handoff),
`<h1>A closer look at bees</h1>`, `illustrated-text-block` present, 7 `<img>`, 0 `<footer>`,
0 `mailto:`, `googletagmanager` ×1 (`GTM-PQ3WCTBD`), **0 `G-` ids, 0 `Set-Cookie`**.
`pages.build_status='deployed'` (updated 08-24 13:52), 7 `page_components` `permanent`/`deployed`.
`tools.apis.uk` real endpoint (`POST /api/v1/tools/gauntlet/round`, Origin vonc.com) → **200**;
the root 404 is the Caddy arm, as the RUNBOOK says. `gtm.js` for the container: version 2,
`"tags":[]` — still nothing recorded.

Open rows on apis.uk: 6 — the two `deferred` `content_rewrite`, `save_refused_incomplete`
(historical), `unresolved_cta` (`wont_fix`), a `page_content_divergence` `detected` 08-24 12:42
(pre-dates the final render; not chased), and **one `page_rerender` `failed` today 11:19 from the
383 lane** ("canonical re-walk … bugs_open/383 §13") — their commit `9a843c06a` reads "apis.uk
cannot re-render", i.e. our locks refused it, which is the design.

`[UNEXPLAINED]` `sites.build_status='pending'` since 14:40 today (the page row is `deployed`).
No writer found by grep for that column on `sites`; no served-page effect; left alone.

### What the peer lane found, and the owner's ruling

`analytics_gtm` (session "google") filed a CONTRIB into this directory and `bugs_open/397`:
the 08-24 backfill wrote `site_components.rendered_html` and **no `site_config` key**, so 12 sites
(apis.uk among them) carry GTM in the artefact only and lose it on the next chrome render —
agritec.uk already did. And §3's "per-site id in `RenderFallbackHead`" design was wrong-place:
the seam exists (STY-050) and that function is the failure path. **Accepted in full.**

Owner, in this session, on §4: *"section 4 has google in it which is taken by another lane,
please communicate to that lane that that is what they take and we will take the rest here."*

**Done:**
- CONTRIB into `analytics_gtm/` (`CONTRIB_2026-08-25_from_apis_uk_bees_homepage_owner_ruling_…md`)
  stating the split as a table — theirs: GA4 publish + consent, 397's rebuild, 397 §6.2, Search
  Console, the dashboard script (§2.3, never started), `039`; ours: the page, per-section subjects,
  image accuracy A + C, the two deferred rewrites. Plus two apis.uk-specific warnings for their
  `c2` wave (the page refuses re-renders; add this lane to 397 §9).
- Handoff corrected **visibly** (top block + strike-throughs), not rewritten.
- Ownership census for the two builds we keep, so "nobody else has it" is a grep and not a
  feeling: `grep -rln 'per-section subject\|PlannedSections\|section.*subject.*writer'` over
  `docs024_key_docs_latest/` → `bugfix_285` (diagnostic use of `PlannedSections` only),
  `copy_quality_two_stage` (names it as a precondition on THEIR experiment, not as work),
  `finetuning_uk_service`; `bugs_open/` → 0 files. `execute_vision_prompt` in `bugs_open/` +
  `features_open/` → 231/243/256/257/136 (all about the action's plumbing, not a critique step)
  and `features_open/018` (screenshot taste critic, "specified, not built" — adjacent to C, not C).

**Misstep, small:** my first two DB queries used columns that do not exist (`pages.slug`,
`site_components.head`/`content`). The handoff's row names are `pages.name` and
`site_components.slot_name` + `rendered_html`; the RUNBOOK §5 already had the right shapes and I
typed from memory instead of reading it. Cost: one round trip.

## 2026-08-26 (session "apis.uk") — per-section subjects: read, built, mutation-proven, submitted, committed; one owed parity caught by a peer

**Owner: "go ahead with per section subjects" (2026-08-25, after the lane split).**

### What the reading changed about the design

The 08-24/25 handoffs' design ("let an entry be a string or {component, subject}") was right in
intent and wrong in placement: RFC_016 §5.1 (owner-RATIFIED 2026-08-08) already defines the
extension contract, and the `facts` field (PBP-037, Slice B live — verified in the live
`page-content-writer` config, not assumed from the RFC's stale "HELD" line) provides working
rails at every hop. So the subject became the seam's SECOND field rather than a new seam:
migration 638 (applied, schema-verified), extractSectionEntries + INSERT, normalise pass sibling
`section_subjects`, carry, loader (LOCK-008 merge mirrored), plan_sections opt-in key,
`sectionPlanItem.Subject`, three `_HOLD` seeds (639/640/641; 641 owner-read gated per §5.2).
Commit `35905c547`; council `Council-Submitted: 4bd35ed8-5f72-4a2f-9cbf-3860847837f4`; register
PBP-049 + the PBP-037 nested-contract line in the same commit (the 08-11 ruling demands exactly
that); PBP-037's 16-days-stale index cell corrected visibly while there.

### Missteps and near-misses, the point of this file

1. **My first mutation test proved nothing and I nearly recorded it as proof.** Removing the
   `clone["subject"]` set passed the suite — because my test's realised entries are plain
   strings, so the mutation hit the OBJECT arm the test never exercises
   (`a-mutation-that-passes-may-have-hit-a-guard-in-series`, and its neighbour: the wrong ARM).
   Re-run on the string arm and on scopeItem's attach: both FAIL exactly one named test each.
   The object arm is now a **stated** untested gap in PBP-049 and the council risks, not a
   silently-assumed-covered one.
2. **I would have shipped the RFC_022 parity break.** `section_subjects` in the Optional list is
   an optional-key-count change; the cron literal (`check.py`) must ride the same commit
   (CLAUDE.md RFC_022 — I had read it this session and still missed it at commit time). The
   **333 lane caught it at HEAD within ~30 minutes** and told me instead of fixing it, having
   checked blame and found my then-uncommitted change — the correct call under
   `your-fix-invalidates-a-peers-pending-test`, done TO me rather than BY me. Settled
   `339474ca4`: regenerated from a committed-HEAD extract (the tree carries other lanes' Go
   WIP), sole delta plan_sections 7→8, tests green vs HEAD, overlay applied, **verified at the
   mounted ConfigMap with an unchanged neighbour key as control**.
3. **The pattern check flagged 3 `logged-model-output` lines in my commit — none mine.** All
   three pre-date the change (git log -L: 2026-03-31, 05-12, 06-20); the check scans whole
   touched files. Recorded so the next reader of `35905c547`'s advisory doesn't chase them.
4. **Pre-existing red at clean HEAD**: `TestFindingCodeScanEveryWriteIsRegistered` /
   `WORK_ITEM_STATUS_OVERRIDE_REFUSED`, from `2b46afbe6` (the status_override 396 — duplicate
   number, resolve by slug). Proven pre-existing by running the suite against HEAD without my
   overlay files. CONTRIB filed into `deferred_work_item_park/`. Cost me one verification round;
   without the check it would have read as MY failure.

### Cross-lane, this session

CONTRIBs out: `brochure_component_library` (seam extended per their own RFC; carry semantics),
`bugfix_285_lock_blind_section_list` (loader + their test file, third aligned list),
`copy_quality_two_stage` (their experiment's stated precondition now has a mechanism + adoption
query), `deferred_work_item_park` (the unregistered finding code). Messages: `google` (08-25
split), `bugs_open/333` (parity settled; they re-verified green from their side).

### 2026-08-26 later — design rotation heads-up, measured

`webdesign-tool-rebuilds` warned the discovery rotation is back on and apis.uk (no design stamp)
is early in the queue. Measured instead of waiting: 6 findings already landed **00:40 UTC today**
(so the visit predates their 09:20Z re-enable time — a different sweep, or their timestamp is the
noticing not the enabling; either way the items are real). All 7 detected rows on the site carry
**empty `handler_agent`** ⇒ not promotable. Analysis + trap (chrome repair strips artefact-only
GTM, 397) written into HANDOFF_2026-08-26 §5. The `footer` half of `head_essentials_missing` is
the owner's deliberate absence being read as a defect — the predicted false-positive class;
`skip_link` and the favicon/og-card 404s are genuine small items, parked with the trap named.

### 2026-08-26 ~12:00 — council round 1: REVISE; round 2 built, proven and resubmitted the same morning

**Round 1 (4bd35ed8): REVISE, gating seat `prior_art_librarian` (HIGH: "confirm the DDL was not
executed ahead of approval").** It WAS executed — deliberately, per the owner's 2026-07-29
"review is after the fact by design" ruling — and the honest answer was to say so and supply the
pre-state the seat could not reach (it has no access to `site_plan_sections`): the pre-apply
schema listing from this session's own 08:55 read (no `subject`, `assigned_fact_ids` present)
plus the post-apply check. **The seat's instinct was right and the posture is still correct** —
worth recording that those can both be true: on this estate the apply-then-review shape is the
documented norm, and the reviewer's job is to demand the pre-state, which round 1 did not carry.

**Every other objection was cheap and real** (`a-revise-round-is-cheaper-than-the-defect-it-finds`
holds again):
- bug_historian's MEDIUM (mismatched-length test for the three parallel lists) produced two tests
  that now pin degrade-to-unassigned-never-shift — the exact bugs_closed/041/095/039 shape.
- The object-arm gap I had stated as a risk became two tests; **both object-arm mutations now
  FAIL, including the one that silently PASSED in my first mutation run** (NOTES earlier today).
  A stated gap is better than a hidden one; a closed gap beats both.
- editquality's seed pre-flight ask exposed a real extra hazard I had not seen: 640/641's
  `SELECT INTO` takes the FIRST of N rows silently on a duplicate-active-row, so the guard is not
  just politeness — without it a dupe row would HALF-apply the prompt edit.
- guardian's caller-enumeration ask was answered from the round's own read-only checks (the
  council answered its own question; my job was to notice), plus: `scopeItem` is function-local.
- guidelines' register-outside-the-slots concern: settled by evidence, not argument — the register
  edits are IN `35905c547`'s stat, quoted in grounded_in.
- architecture's RFC_022 ask was already settled by the 333 lane's catch (`339474ca4`), quoted.

**Round 2 resubmitted on the SAME correlation** (`RESUBMIT_CORR`), commit `52085b410`.
**Two Kafka publish failures first** (kcat bootstrap timeouts, broker pod-0 unready 4 restarts;
the 097 trigger failed LOUD both times — "SUBMISSION NOT SENT", nothing spent — which is the
kcat-silent-drop landmine's fix earning its keep). Third attempt after ~90s backoff SENT.
Evidence appended to `bugs_open/040` with the DB-slowness correlation marked `[INFERRED]`.

Also this hour: handoff §5's visit attribution corrected — the 00:40 findings were the
**improvement-loop** (re-enabled 21:18Z 08-25; cycles 00:39/04:47/08:40Z), not the rotation
(separately re-enabled 09:20Z); both now active, so expect ~4-hourly finding arrivals.

### 2026-08-26 ~12:50 — round 2 APPROVED; the three advisories were worth more than the approval

**APPROVED, 3 advisories, none high (`4bd35ed8` r2, ~12:35 BST).** Trailer discipline held:
`Council-Reviewed:` written only on `fa98a1961`, AFTER reading the verdict; the two earlier
commits keep `Council-Submitted:` and 098 credits them.

**Advisories dispositioned, not banked** (all in `fa98a1961`):
1. bug_historian (MEDIUM): rule 17's "subject REQUIRED on repeats" had NO structural enforcement
   — prompt-only, the exact 016b §9 decorative-decision pattern, on the exact site that motivated
   the build. Built `subjectlessRepeatFindings` in write_site_plan: observe-only durable finding
   `SUBJECT_MISSING_ON_REPEATED_COMPONENT`, GATED on the plan carrying any subject at all so the
   fleet's pre-640 all-string plans produce ZERO rows (no retroactive spam). 5-case pure test;
   the gate is mutation-proven (removing it fails case (a)). Declared in the finding-code
   registry in the same commit — the same-commit rule I reported the 396 lane for this morning.
2. debug_historian (MEDIUM+2 LOW): "post-roll" was an assumption in the seed headers. Every
   `_HOLD` seed now carries the pod-verification commands (stamp + merge-base; capability probe
   with `section_facts` as positive control) and the APPLIED-line convention (a `_HOLD` file
   never reaches the migration ledger, so the FILE is the durable record). 641 additionally:
   the owner's read must be RECORDED in NOTES and named in the APPLIED line.
3. guardian (2 LOW): both settled by measurement, not argument — `site-planner` does NOT invoke
   write_site_plan (config LIKE check), loader caller set re-confirmed = page-build-handler only.
4. editquality's cosmetic (the copy-pasted sketch fragment) needed no repo change — it lived in
   the submission JSON, and each seed file's own comments were already distinct.

**Missteps of the hour, small but recorded:** (a) my registry addition went through `json.dump`,
which reformatted the shared file — 363/356 line churn for a one-entry change; proven lossless
(zero keys lost, zero entries changed, diff'd against HEAD~1) but it is diff noise other lanes
must read past; edit such files textually next time. (b) The pattern-check's
`logged-model-output` flag on this commit is the same pre-existing 2026-05-12 line as before,
shifted to :1290 — still not mine.

**And the morning's CONTRIB paid off within ~2h:** `a0ec90eb9` (396 lane) declared their code;
the full actions suite including the finding-code scan is green at HEAD again.

### 2026-08-26 ~13:20 — the trap fired: chrome refresh stripped GTM and resurrected a fallback footer; two of my same-day calls were wrong

Peer relay (webdesign-tool-rebuilds ← analytics_gtm) then measured here: overnight
improvement-loop completions re-rendered apis.uk's chrome at 08:46:26 (all three site_components
rows, one timestamp). Served page now: GTM ×0 (397 bucket-B strip, 10/10 artefact-only sites),
one minimal fallback footer (brand + copyright — no email, no disclosure), +371 B; h1/sections/
images intact, **7/7 permanent locks held at the artefact**. Page re-queued 09:15.

Two WRONG_CALLS rows filed (the wide no-auto-dispatch claim; the cannot-re-render inference —
the second also poisoned my CONTRIB to analytics_gtm, corrected by message). The durable finding:
**"no footer" was artefact-only state all along** — same class as the GTM backfill, no spec-level
suppression exists, `RenderFallbackFooter` regenerates the shell on every chrome refresh. Owner
decision framed in handoff §5b: accept the shell vs commission an opt-in
`chrome.footer_disabled`. No interim row-emptying (397-class churn against c2's imminent wave).

### 2026-08-26 ~13:35 — c2 confirmed on apis.uk; my hour-earlier read said "no row"

google's message says c2 applied ~10:50 UTC, 17 sites incl. apis.uk. Re-measured: the key IS
present (`site_config`, is_current, created 10:12). But my ~13:15 check — same query shape —
returned "(no site_config row)". `[UNEXPLAINED]`: either their apply landed between my two reads
(timestamp session-TZ ambiguity makes this plausible) or my first query erred in a way psql
swallowed. Recorded rather than reconciled by guesswork; the CURRENT state is measured and is
what the handoff carries. Also corrected their forward expectation (they predicted the wave's
page item would FAIL on our locks, citing my own earlier wrong inference back at me — the
overnight completions disprove it; message sent, and 397 asked to carry a dated correction).

### 2026-08-26 ~11:30 UTC — the spec-key timing [UNEXPLAINED] is RESOLVED, by the row's own stamps

Session TZ is UTC (checked, not assumed). The apis.uk `site_config` row: `created_at =
updated_at = 10:12:11 UTC`, `created_by = claude-session-google-2026-08-25` — a fresh insert by
their session at 10:12:11. So: my "(no site_config row)" read ran minutes BEFORE the insert (a
plain race, my instrument was fine), and their "~10:50Z apply" figure is the imprecise one —
their own creator stamp says 10:12. Both accounts corrected; the key fact (live, verified) never
moved. Told them. **Lesson already in the memory index, now with a worked example: convert
"when did X happen" disputes to the artefact's own stamps before either side re-reads anything.**

Also from their message: `chrome_divergence_overwritten` item `2e4e5f51…` (00:44, our head
strip) is `needs_human_review` and the platform **archived the pre-strip head — 48,471 bytes**
— so the tagged artefact is recoverable evidence, not gone. Disposition queued for the owner in
`bugs_open/397` §10-addendum; added to handoff §5b's open list.

### 2026-08-26 ~11:45 UTC — near-miss on the six illustrations, caught and fixed by another lane before our own wave could fire it

`vigilant_designer_offer_analysis` found that `Illustrated Text Block` sourced `image_url` from
`site_assets.image` — an ALIAS to hero — and apis.uk has a live `hero_home`, so the next
plan_sections run on /index.html (i.e. the c2/rebuild wave WE are waiting for) would have
resolved all six sections' images to hero-home.jpg, live-resolution-beats-carry. They shipped
migration `644` (`site_assets.illustration`, no alias ⇒ carry preserves; `image_alt` → `llm` —
it had been feeding the URL to screen readers). Verified here at both ends: component sources
read `site_assets.illustration | llm`, and our six `content_data.image_url` values are intact
and distinct (positions 2–7). Our six are the estate's ONLY live instances of this component, so
nothing else would have shown the failure first — the defect was aimed at exactly one page, ours.

Their unfixed pointer constrains our un-defer plan and is now in handoff §5b: the five
`site_plan_imagery` illustration rows are inert (scope='page'), and section-scope resolution is
first-wins by kind — so swarm/pollination images go in via content_data + lock (CLC-030 route),
not via the resolver. Their register entry: IMG-074; their migration is Council-Submitted
(08477888, pending).

### 2026-08-26 ~14:30 UTC — automation added two tools to the single-page site; parked nine items before anything published

Found on a routine "has the c2 wave landed" check (it had, at the chrome level: head artefact
carries GTM again since 11:40; the served page awaits its own `page_rerender`, still triaged):
apis.uk had FIVE `pages` rows. Two `add_tool` items (09:18Z, improvement loop → `tool-generator`)
had created a quiz tool, a calendar tool and two companion guides, all `planned`, outside the
1-page site plan and against the current single-page `roadmap_brief`. Nine triaged items stood
ready to publish the lot (renders, guides, index rewrite with tool references, nav rebuild, two
improve_tool pushes). None claimed.

Deferred exactly those nine (plpgsql `GET DIAGNOSTICS` count assertion — a first attempt used a
psql `\gset` variable inside a `DO $$` body, which does not interpolate; it errored before COMMIT
and rolled back, verified by the second run's own count). Reason + unblock in `result`, mirroring
the 08-24 deferrals. Left the index's GTM-restoring rerender alone. Did NOT delete pages or
cancel items — the owner's standing ruling justifies holding the door; deciding the tools' fate
does not belong to this session. Two owner paths + the un-enrol requirement written in handoff
§5c; loanzy.uk (loop owner) and the 644 lane (their protection now trips
`image_source_unsatisfiable` here — true statement, false alarm) messaged.

Pattern worth naming for the lane: **a re-enabled fleet mechanism does not know a site's
rulings.** The loop read apis.uk's "1 of 6 structures" and "no tools" as gaps to fill; the
single-page constraint lives in a brief the planner reads and the completeness pass does not.
The durable fix is the un-enrol, not the deferral.

### 2026-08-26 ~14:50 UTC — loop owner's code read corrects two of my §5c claims

loanzy.uk (loop owner) read the path from code: it is the DESIGN seat's `missing_tools` check →
`evaluate_tools`/tool-suggester → `add_tool`/tool-generator (not completeness), and it runs via
both the loop's audit pass and the design rotation. **No per-site exclusion exists** — the growth
ratio key can only add pressure. And my hold is stronger than I described: `deferred` is OPEN for
`idx_swi_dedup`, so the nine parked rows hold their keys and re-files dedup onto them; my option
(b) had "cancel the nine" FIRST, which would release the keys and re-create the set — reordered
in §5c (refusal declaration first, RFC_056 follow-up they are recording; cancel after). RUNBOOK
§8 now carries the parking recipe + the `park_work_items` (mig 621) pointer. Their framing of
their own boundary is worth quoting: "tool EVALUATION is growth, not a defect, and should not
have been on the dispatching side of the line."

### 2026-08-26 ~14:55 UTC — my "false alarm" on image_source_unsatisfiable was under-read; the 644 lane measured it fleet-wide

They confirmed the mechanism (the check decides from schemas + alias table, never consults
carryStored) and corrected my framing twice: it fires whether or not images are carried, and it
is a supply warning (a NEW page with this component here renders nothing) not a false alarm —
two states that want distinguishing, not merging. Census: 58/67 open items assert "renders
empty" about populated fields. Routed by them through 090 (`8aeba0b6`), bug to follow; our
item stays as evidence. Adopted into §5c. Lesson for me: **"true statement, false alarm" was
itself a compressed claim — the alarm was about the wrong thing, not about nothing.**

### 2026-08-26 ~15:10 UTC — the routed item became bugs_open/411 (CONFIRMED, first iteration)

The 644 lane's 090 run confirmed the mechanism independently (same lines, plus a live non-644
example on another site), filed `bugs_open/411`, and our §5c framing — the (a)/(b) split, and
the rejection of "treat carry as satisfied" — survived into the file. Our item is cited evidence
and stays open. Their closing line is the practice in one sentence: "a cancelled item would have
taken the whole finding with it."

## 2026-09-02 (session "apis.uk", wrap-up for a new chat) — roll verified, 639+640 applied, IMG-075 adopted, the REAL rerender finally fired

Fresh chassis confirmed at the pod (capability probe, both controls) BEFORE anything else. Then:
639 applied+verified; **640's anchor guard refused the first apply** — bugs_open/380 had
rewritten rule 17's tail during the week; re-derived (subject sentences before the 380 sentence,
kept verbatim) and applied. The guard refusing on drift is the design working — recorded as such,
not as a failure. IMG-075's six section-scope bindings inserted (pairing checked against live
positions first; assets all `active`). A `section_data_resolved` page_rerender filed for index —
the 08-26 "completed" one was assemble-only (item_key suffix `_assemble`, byte-identical serve),
which today's inline_guide_imagery CONTRIB correction predicted exactly.

**Missteps this session, each caught in-session:** (1) my LIKE patterns with `\"` inside single
quotes read back false "absent" for config that was present — the in-file verify was right and my
instrument wrong; (2) the APPLIED-line append asserted on a from-memory anchor, failed, and the
commit went out carrying only 640's re-derivation — fixed in the next commit; (3) the rerender
INSERT missed NOT-NULL `created_by` — the count-asserted transaction rolled back whole, re-ran
with the google session's creator-stamp convention. (4) The week-old bugfix_394 CONTRIB sat
unread while the shared registry test stayed red — they fixed it themselves (3749132e0). The
standing lesson is in the new handoff §1.6: read the lane's CONTRIBs at session start.

Owner gates as of the handoff: 641 read (block quoted in §3.1), footer, tools park. The imagery
CONTRIB's decisive save-path test deliberately NOT run (locks are the standing protection; wrong
moment for a destructive-if-wrong test). brief_supplies_negation moved today 13:17 — unread,
flagged. Cold-start: HANDOFF_2026-09-02_continue_here.md.

### 2026-09-02 ~17:20 UTC — the reasoned rerender was REFUSED by the save guard; both modes now measured

`save_sections`: "overwrite: REFUSED for page index — re-confirmed too little of what is stored
(prune_floor…)". So: assemble = completes, deploys stale bytes; re-resolve = refused. The page is
wedged between its own protections and both rerender modes — the 383 lane's original claim was
closer than my 08-26 blanket correction of it; the truth is MODE-SPECIFIC and now recorded as
such (handoff §2). The IMG-075 rows and six illustrations are unaffected (refused save writes
nothing; served page byte-identical). Route left for next session: the site-level rerender-pages
fan-out that served the tag on this very page on 08-24 with locks in place. Not fired tonight —
end-of-session is the wrong time to launch a route whose failure mode I would not be awake to
read.

### 2026-09-02 ~20:15 UTC — consumers-told notice from bugs_open/443 (PBP-051): the subjects rails gained two fallback-tier sources; 641's acceptance population widens

Received as a cross-session message (this session freshly cleared — nothing else in context), the
2026-07-29 §3 notice for `dbb218a41` (council `b7c59309` APPROVED r1, lane
`bugfix_443_fallback_tier_subjects`). What changed about OUR guarantee: the loader's
"authoritative tier only" line is now "aligned or absent from CONSTRUCTIONAL sources only" —
tier 3 reads same-row `pages.section_subjects`/`section_facts` columns (mig 717 applied), tier 2
reads same-object sibling keys from the site_plan aspect, tier 4 never. Alignment semantics
(RAW-index across skips, trim, ""→nil, LOCK-008 nil-inserts) preserved hop for hop.

Verified all four claims rather than repeating them: commit exists with the trailer;
PBP-051 register entry cross-references PBP-049's supersession; both finding codes in
`architecture_review/finding_code_registry.json` (their `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT`
beside our `SUBJECT_MISSING_ON_REPEATED_COMPONENT`, our entry untouched by `dbb218a41`); and ALL
12 tests in `plan_section_subjects_test.go` ran by name and passed at HEAD (`-run` with the exact
12 names — a bare `-run 'Subject'` matches 31 and would hide a vanished file). Their commit also
carried bugs_open/444's `plan_sections_action.go` error-defer hunk as a DECLARED same-file
passenger — our chain's file; package green, no action.

Two things recorded for the next session of this lane:
- **641 gate: the acceptance population is now our two deferred `content_rewrite` items PLUS the
  443 cohort (11 fallback-tier pages)** — same Stage B, more evidence, when the owner read lands.
  The un-defer path in HANDOFF_2026-09-02_continue_here.md §"un-defer" is unchanged.
- **A tier-1 page with a subjectless repeat can legitimately draw BOTH finding codes** (plan-write
  event vs build event, remedies differ) — a double-fire is the design, not a duplicate-detector bug.

Caught while verifying, corrected in the register (5 visible in-place corrections, PBP-049 entry):
the "stated gap" line (object-realised carry arm untested) had been STALE SINCE 52085b410 closed it
in council round 2 — the entry that shipped the fix never updated its own gap line, a week of
readers inherited it. Also superseded the tier-1-only loader wording (→PBP-051), dated the test
count (7→12), widened the verify-later acceptance line, added PBP-051 to relations.

### 2026-09-02 ~20:30 UTC — 443 follow-up: 641's gate 1 independently corroborated; the OWNER READ is being driven by the finetuning lane, not us

Three state facts from the 443 lane's cross-checks tonight, the first two [PEER-REPORTED —
consistent with our own morning pod-verify, not re-measured here]: (1) gate 1 on 641 is CLEAR —
the finetuning lane pod-probed the rails literal in the running binary, and our 639 apply was
re-read live at the `agent_definitions` row by a second session; (2) **the finetuning lane is
actively putting the 641 inserted block in front of the owner with the verbatim quote** — so the
next session of this lane should NOT independently re-present it; coordinate with them first,
our only remaining job on 641 is recording the read and hand-applying when it lands. Stage B
then covers our two deferred `content_rewrite` items and the 443 cohort in one round.

(3) Verified, not peer-trusted: LANDMINES gained an entry tonight ("A `_HOLD` migration's
FILENAME never changes when it is applied", ~line 19796) after a session read `639_*_HOLD.sql`
in the listing and reported our LIVE 639 as unapplied — inverting the gate-1 conclusion the same
day it was measured clear. The check is the APPLIED header line + the artefact (live row / pod),
never the filename; caught only because a second session read the live row independently.

### 2026-09-02 ~21:00 UTC — 641 gate 2 RETURNED: a REDRAFT direction, not an approval — do NOT apply 641 as written

Relayed by the 443 session (the finetuning lane carried the read to the owner and initially
mis-routed the result to 443's lane; 443 caught the misattribution — 641 is OUR file, PBP-049's
seed). The owner's directive, recorded verbatim in the draft file below: positive prompting only
("say don't think of an elephant and the llm starts thinking of elephants"), the block written
IN the language expected back (his rejection of their first attempt: describing the arrangement
in production vocabulary — section/subject/reader — "has started to hardcode what should be in
it"), and NO specimen answer — the prompt's own prose is the demonstration, which also dodges
the quoted-exemplar-ships-verbatim trap.

Our committed v5 block fails the directive on its face: "do not restate theirs, and do not
widen…" is negative instruction in production vocabulary. It is now a historical draft.

State + division of labour (verified against the draft file, not just the relay):
`finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md` holds three framing
candidates (reply-to-a-person / write-for-a-person [their recommendation] / just-the-subject),
filled in for a real page; owner's pick PENDING → finetuning lane test-renders → THIS lane
writes the final SQL → owner reads the exact final words (RFC_016 §5.2 — approval attaches to
that text, voids on edit; the rule working as designed). Mechanics that survive the redraft:
the `{{if .current_section.subject}}` guard, placement before Verified Facts, no em dashes
(census pinned at 5), British English.

**Owed by us before the redrafted 641 can apply — the sibling-subject range render is
UNTESTED.** All three candidates enumerate the sibling subjects. The data is reachable
(verified at `platform/orchestration/loop_expansion_handler.go` ~395-425: setLoopVariable
persists `current_section` + loop vars into CollectedData precisely so they survive the
optimistic-lock reload; the full plan rides `section_plan.sections_ready[]`), but NO template
has ever ranged over it, and under `missingkey=zero` the failure is silent: an empty/absent
subject renders as a blank item, and a wrong path renders NOTHING — "a template that silently
renders nothing is exactly the failure 443 was about". Falsifier for the test-render: a
sections_ready list containing a subjectless/skipped sibling must produce an enumeration that
DROPS it; a wrong key must fail loud in the test, not blank in production.

Recorded tonight in the same commit: seed header stamped with the gate-2 outcome + DO-NOT-APPLY
(the control at the point of action — a session hearing "the read landed" must find the refusal
in the file it is about to apply); handoff §3.1 corrected + chain-status row updated; PBP-049
status line gains the outcome. FYI also relayed: **RFC_063 DECIDED tonight — option B**, the six
plan-less sites converge into the plan tables (hand-insert permitted, closed backfill); 443's
Stage B re-points at the redrafted 641, their Stage A unaffected.

### 2026-09-03 ~09:50 UTC — 641 rewritten to the owner's pick, rehearsed under ROLLBACK, council-submitted; the old census literal was never true; imagery thread resolved to "refused at save, resolver NOT implicated"

Session start: read the lane's CONTRIBs FIRST (last session's lesson) — and the new
`CONTRIB_2026-09-02_from_finetuning_owner_picked_C_and_the_test_render_found_two_things.md` was
the whole brief. Owner rejected the first three candidates ("they all sound a bit AI"), picked C
from a plainer set ("go with C"); the finetuning lane test-rendered five fixtures from REAL
orchestration_states rows BEFORE handing over (their `render_test_641/` harness). Fixture E
discharged the sibling-range obligation I recorded yesterday (subjectless siblings drop out
cleanly). Fixture D found the real blocker: the prompt renders against
`ExtractFields(CollectedData, input_fields)` — a SUBSET the step names — and live
`generate_content.config.input_fields` has no `sections_for_render`, so the sibling range
renders an EMPTY list, silently. My yesterday line "the data is reachable" was true of
CollectedData and WRONG-BY-OMISSION about the template context; the draft file logged its own
version of that in WRONG_CALLS 2026-09-02(d).

What I verified before writing SQL (not taken from the CONTRIB): live row — exactly 1 active
writer, `input_fields` verbatim as quoted (no sections_for_render), `iterate_over` =
`sections_for_render.sections_ready`; extractor code — the speciallyHandled promotion + named-
field loop at `unified_extractor.go`; template bytes — my E-string decodes BYTE-IDENTICAL to the
tested harness `const block` (programmatic check, anchor swapped), so what the owner reads is
what was rendered in the fixtures.

**The census discovery.** The dry pre-flight printed the live em-dash count: **9, not 5**. From
`agent_definitions_backup`: 3 before mig 595 (rules 9+10, applied 08-24 16:57Z), 9 after — TWO
DAYS before the first 641 asserted 5. The literal was wrong from birth, matched no adjacent
state, and would have RAISEd on a correct apply at the worst possible moment. → WRONG_CALLS
2026-09-03 entry; the rewrite replaces the literal with pre/post EQUALITY in one plpgsql block
(the true invariant: this insertion adds zero em dashes). Also observed, unexplained, not
chased: the writer row was UPDATEd 2026-09-03 08:56:53Z with NO agent_snapshots row (latest
snapshot is 599's, 08-24) and no em-dash change — whoever owns snapshot discipline may care.

**Proof, both directions:** full DO block rehearsed under BEGIN/ROLLBACK against the live row —
pre-flight clean, both halves applied, census 9 unchanged, and the post-ROLLBACK control shows
the row untouched (subject block absent, no sections_for_render). Induced failure: the same
block with the input_fields append stripped RAISEs "input_fields does not contain
sections_for_render" — the fixture-D check can actually fire. Council: corr
`6c92d154-e527-4f9d-8262-9fd0c22858f1` submitted (DRY_RUN admission first), commit carries
`Council-Submitted:`.

**Gate state now:** gate 1 clear (unchanged); gate 2 = the owner reads the EXACT block in the
seed file — the finetuning lane carries it to him (told via CONTRIB in their dir). Open owner
question recorded, NOT absorbed into C's words: tier-1 planner subjects ("Brief description of
the sister-site relationship…") read awkwardly in "You'll want to know ___"; recommended fix is
a planner-prompt nudge, separate small migration, this lane's side.

**Imagery thread (inline guide imager, two messages):** their measurement — our reasoned
rerender 73ad8c56 FAILED with result={} and the morning's completed 5a1740be was reasonless
(assemble-only, exercised nothing). I supplied the why from last night's NOTES: REFUSED by
save_sections' overwrite guard (prune_floor), never reached the write — so our case is OUT of
the bugs_open/425 "sections path doesn't write resolver-sourced values" hypothesis (theirs to
record; they did, in 114, citing our NOTES). Their point, adopted here as the standing line:
**the 08-24 site-level fan-out served this page WITH the locks in place, so the locks are
proven survivable and `prune_floor` is the one blocker** — a reader seeing "permanent locks" +
"refused rerender" side by side must not conclude the locks did it. IMG-075 status stands as
"armed; its one attempted test was refused at save". They get the six image_url results when a
reasoned rerender lands. Their observability point (refusal reason existed in pod logs, never
reached the item) is recorded in 114, greppable before anyone re-files it.
