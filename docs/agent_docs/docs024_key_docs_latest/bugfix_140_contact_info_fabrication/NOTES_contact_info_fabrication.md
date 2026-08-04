# NOTES — bugfix 140, the contact-info component fabricates business facts

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-02 — picking the bug, and how I decided it was unowned

`scripts/who-owns.py` is COMMIT-based and therefore lagging: it reported "OWNED
or recently active" for every recent bug, because *filing* the bug is itself a
commit that touches the file. On 175/174/173/172/171/168/165/163 it produced a
verdict indistinguishable from real ownership. Useless as a discriminator here.

What worked: grep the **live session transcripts** for `bugs_open/NNN`.

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
files=$(find . -name '*.jsonl' -newermt '2026-08-01T22:00:00' | tr '\n' ' ')
grep -ohE 'bugs_open/[0-9]{3}' $files | sort | uniq -c | sort -rn
```

26 sessions were live. **140 appeared 8 times across all of them, and 0 times in
any session's recent context** — against 518 for `165`, 428 for `138`, 330 for
`119`. Unowned by a wide margin.

> **MISSTEP, caught before it cost anything.** My first pass used a
> `tail -c 900000` per file inside a bash loop over 26 files × 60 bug numbers.
> It timed out at 120s having produced nothing, and `tail` panicked on several
> files. A single `grep -oh` over all files at once took seconds. The lesson is
> not about performance — it is that **I nearly concluded "no signal" from a
> command that had failed**, which is the "0 matches AND no error" trap wearing
> different clothes.

> **SECOND MISSTEP, and this one changed my read.** The whole-file counts
> disagreed with the recent-context counts, and I initially read the whole-file
> table as authoritative. It is not: a long session's *history* contains the bug
> INDEX (every session that ever listed `bugs_open/` matches every number), so
> whole-file counts measure "has this session ever seen the directory listing",
> not "is this session working it". `bugs_open/072` scoring 290 across sessions
> that have nothing to do with it is the tell — it is in the SessionStart
> landmine hook output, injected into every session in the fleet. **Recent
> context is the signal; whole-file is noise with a plausible shape.**

## Re-verifying the bug before working it — it had GROWN

Filed 07-29 against 6 live uses. Today: **8**. `dartsonline.com` and
`vetcomparison.uk` are new, and vetcomparison (rendered 07-31, two days ago) is
the first site to serve the fabricated PHONE as well as the hours:

```
$ curl -fsS 'https://vetcomparison.uk/contact.html?cb=…' | grep -oE 'tel:[^"]*|Monday[^<]*'
tel:+1234567890
Monday – Friday, 9am – 6pm
```

Site seven and site eight arrived exactly as the filing predicted. That settled
the fix-the-mechanism-not-the-rows question before I had to argue it.

## The finding that reframed the whole fix

I went in expecting to argue a POLICY change — "should a component invent a
plausible default or render nothing?" — which needs an owner, because it is a
judgement call about eight other lanes' sites.

It is not a judgement call. The component's **own `input_schema` already
declares the answer**:

```json
"hours":   { "source": "site_specs.identity.hours",   "on_missing": "skip_field" },
"phone":   { "source": "site_specs.identity.phone",   "on_missing": "skip_field" },
"section_title": { "source": "llm", "fallback": "Contact Us", "on_missing": "use_fallback" }
```

`skip_field` for the facts, an explicit `fallback` string for the label. **The
schema language already draws the exact distinction the bug is about, and only
the template ignored it.** So the change is "make the template obey its own
published contract", which needs no owner ruling at all.

Corroboration that made it unarguable: `SELECT count(*) FILTER (WHERE
content_data ? 'hours') FROM page_components` → **0 of 1089**. The Hours card has
never rendered a real datum anywhere in the platform's history. Every instance
of it ever served was fabricated.

## A second defect in the same row, worth more than the first

Template vars: `title, intro, email, phone, phone_display, hours`.
Schema fields: `email, hours, phone, address, intro_text, section_title`.

`{{if .title}}` and `{{if .intro}}` were **permanently false**. And all 8 sites
DO supply `section_title` and `intro_text` with real values — "Get in Touch",
"Contact Darts Online", "Reach us directly", "Contact VetComparison.uk". So **8
real headings and every intro paragraph were being silently discarded** in
favour of a hardcoded "Contact Information".

The quality pipeline caught this on **2026-05-18** and nothing consumed it:
`quality_score 80, schema_template_synced false, quality_issues ["template var
{{.intro}} has no schema entry"]`. Detection wired to nothing — the `083`/`115`
family, recorded here rather than re-litigated.

## Why the detector named for this defect never fired

`check_placeholder_contact` exists, is registered, is configured on
`quality-discovery-agent`, and raises items titled *"Fabricated contact info on
page X"*. Its patterns are the generic conventions: 555, `@example.com`, `123
Main St`, Lorem ipsum, `+1 (000)`, John Doe.

**Not one of them is a literal our own component library shipped.** Census over
every unlocked `page_components` row:

| | rows matched |
|---|---|
| all nine existing patterns COMBINED | **1** (`user@example.com`, a `ported-page`) |
| fabricated hours — no pattern exists | **8** |
| `+1234567890` — no pattern exists | **1** |

> **What I did NOT get to claim, and wanted to.** The obvious headline is "the
> detector ran and was blind". I checked instead of asserting:
> `quality-discovery-agent` *did* sweep four of the eight affected sites
> (ai-agent-orchestration 05-01, gaswholesalers 04-25, finetuning 04-24,
> leopardessconsulting 04-24) and raised `generic_theme` items, so those runs
> were real and other checks fired. But **every one of the eight current renders
> POSTDATES its site's sweep**, so no sweep has ever seen these artefacts. The
> blindness is certain from the code and the census; its having fired-and-missed
> in production is NOT established. Marked UNVERIFIED in the PLAN and left out of
> the commit message's claims.

Whether that agent is adequately driven is `bugs_open/149` Group B/C, and that
lane is actively worked — not touched here. A blind detector stays blind after
149 revives it, which is why the pattern gap was worth closing on its own.

## The unsound "improvement" I nearly made

The roster-of-literals design looks primitive, and the roster-free replacement is
obvious: **flag any rendered contact fact absent from the component's own
`content_data`**. Exact, no list to maintain, catches any future invented default.

It is **unsound**, and I found out by reading `RenderContext` before writing it:

```go
Email string `json:"email"`
Phone string `json:"phone"`
```

`contextToInterfaceMap` derives the scalar half of the template contract from
those json tags, so a component can legitimately render a phone that its
`content_data` does not hold — the value arrived from site identity.
**`idea.uk` is exactly that shape**, and the join would have flagged it as
fabricated. That is the over-fire the `component_write_guard.go` header warns
about: a guard that refuses good work gets switched off, and then it protects
nothing. Recorded in the check's own header so the next person does not
"improve" it into that test.

## The lint, and the two calibration failures it had before it was fit to ship

`scripts/check_placeholder_fallbacks.py`, house style of `check_cta_gates.py`.

**False positive #1, caught by reading the template instead of believing my own
tool.** First run reported `about-commercial-block` renders `fundamentallyai.com`
"when the site supplies no datum". Looked damning — one site's domain as every
site's default. It is not: the template is

```
{{if .built_by_url}}<a href="{{.built_by_url}}">fundamentallyai.com</a>{{else}}fundamentallyai.com{{end}}
```

a builder attribution rendered as a link or as plain text. **The `{{else}}` is
not substituting for a missing datum; it is one constant rendered two ways.**
Fixed generally rather than by exclusion: a fallback whose text also appears in
the branch it replaces is skipped.

**False negative #1, caught only because I ran controls.** After narrowing for
the above, I ran six synthetic controls. Control (b) — `Weekdays 8am to 5pm`, a
plain fabrication — **did not fire**, because my hours pattern was anchored on a
day-of-week name. Widened to catch a bare time RANGE as well. Without that
control I would have shipped a check that was inert against the most natural way
to write invented hours.

This is the "narrowing past a false positive can make a rule inert" trap, and it
was only visible because all four controls ran: a genuine fabrication must fire,
a label must not, a constant-rendered-twice must not, an honest non-claim
("Contact us for pricing") must not.

Final state — controls a/b/c fire, d/e/f silent; live library **5 findings
pre-fix, clean across 172 active components post-fix**.

## Applying the migration — the runner trap is real

`./scripts/migration/run-migrations.sh` dry run listed **18 pending files
belonging to other threads**. `--apply` takes every pending file in order; there
is no single-file mode. So: applied by hand with
`psql -v ON_ERROR_STOP=1 -f -`, then `--record-only` with a note saying exactly
that. Numbers 288 and 289 appeared in the directory from other sessions between
my writing 287 and applying it — the tree moves in minutes, so re-check the
number immediately before applying, not when you choose it.

## What is NOT done, and why the bug stays open

The template can no longer fabricate. **The 8 stored `page_components` rows still
serve the fabrication** until each page rerenders — a stored render is not
regenerated by fixing the template. So the defect is still reproducible on the
live fleet, which is the `/bugs_closed/README.md` bar, and 140 stays OPEN.
Deliberately not patching the rows per-instance: the template used to
re-fabricate on any rerender of an unlocked row, so per-instance patching was the
wrong tool (140's own candidate 3) — now that the source is fixed, the rows
self-correct on next rerender.

## 2026-08-02, later — COUNCIL: APPROVED at round 1, and the objections answered with checks

`40de12b0-36fa-4c06-82b4-995dc9098593` → **approved, "7 advisory objection(s) —
none high-severity"**, 4 seats abstained, 12 seats ran. Dispositions below; the
checkable ones were checked rather than argued.

### CHECKED — guardian, medium: "are the 8 `page_components` rows locked?"

The plan's own risk #2 says the rows self-correct on rerender. The guardian
pointed out that a **locked** row will not, so some sites could keep serving the
fabrication indefinitely while the source looks fixed. **I had not checked this.**

```sql
SELECT s.domain, pc.locked_at IS NOT NULL AS locked, p.build_status FROM page_components pc …
```

**0 of 8 locked.** All eight will self-correct. One incidental finding worth
carrying: `leopardessconsulting.co.uk` is `build_status = needs_rebuild` (the
other seven are `deployed`), so its correction depends on that rebuild happening,
not merely on a rerender.

### CHECKED — prior_art_librarian, medium ×2: two unverified existence claims

Fair objection: I justified two design choices by asserting that other code
already establishes the pattern, and attached no evidence. Both now verified.

1. *"tests already read their own migrations"* — `doc_subjects_common_test.go:74`
   and `write_experience_pattern_test.go:372` both build
   `filepath.Join("..","..","..","docs","agent_docs","sql_for_agents")`, `ReadDir`
   it and `ReadFile` the matching migration. **Confirmed.** Nuance I should state
   rather than let stand: they read the migration FILE and scan its text; neither
   extracts a `$mig$`-quoted body specifically, which is my addition. The
   precedent is "a test may read its migration", not "a test parses `$mig$`".
2. *"`check_list_empty_states.py` is a DB-reading advisory lint with the same exit
   codes"* — `:50-51` shells `kubectl exec … psql -U clients_user -d clients_db`,
   and its header documents "Exit 0 = clean". **Confirmed.** This is the one I had
   genuinely inferred rather than read — I took it from `check_cta_gates.py`'s
   header naming it as a sibling, and never opened it. The objection was right to
   flag it even though the claim turned out true.

### CORRECTED — bug_historian, medium: "a second live instance, identified and not fixed"

The seat objected that my submission names `about-commercial-block`'s hardcoded
`fundamentallyai.com` as a second live instance of the same fabrication shape,
demonstrates the new lint would catch it, and then does not fix it or file a work
item for it.

**The submission was stale on that point and the objection inherits the error.**
I wrote that sentence before running the lint; when I ran it, that finding turned
out to be the script's **first false positive** and I fixed the classifier rather
than the component. It is

```
{{if .built_by_url}}<a href="{{.built_by_url}}">fundamentallyai.com</a>{{else}}fundamentallyai.com{{end}}
```

— a builder attribution rendered as a link or as plain text, the same words
either way. Nothing is fabricated, and **there is no second live instance to
fix**; the lint reports clean across all 172 active components. No work item is
owed. (Logged in `WRONG_CALLS.md` — I nearly reported it as a real finding.)

The general lesson stands and is the seat's real point: a submission is a
snapshot, and mine went stale between writing and running. Say "expected to
catch X" and then report what it actually caught.

### ACCEPTED, owed at the next roll — debug_historian, medium: no pod-grep

Correct. The `check_placeholder_contact` patterns are Go and therefore **inert
until a chassis roll**, and I proposed no running-binary check. Unit tests do not
substitute for confirming the deployed chassis carries them. Added to the bug
file's closure criteria with a positive AND negative control, per the standing
fleet landmine that a roll is not evidence your fix shipped.

### RECORDED, not actioned here — reuse_agent + architecture, medium: the renderer should enforce `on_missing`

The strongest objection of the round, and both seats reached it independently.
This is the **second** time a template has been caught disobeying its own
`input_schema` contract (`bugs_open/111`'s footer, now `contact-info`), and both
times the fix was a hand-rewritten template. Nothing sits between
`input_schema.on_missing` and `executeGoTemplate` that enforces the contract, so
a third fact-bearing component will fabricate again unless the renderer gates
fields itself.

The architecture seat's own verdict is `point_fix` — *"no new namespace, wire
shape, or shared runtime contract … does not meet the needs_rfc trigger test on
its own terms"* — while recording `ARCHITECTURE_SIGNAL: insufficient` and naming
the RFC candidate: **a render-time guard driven by `input_schema.on_missing`,
applied once at the `executeGoTemplate`/`RenderContext` layer instead of
per-component.** Filed as `RFC_CANDIDATE` in this directory rather than silently
agreed with — routing it is the required response to a scope observation, not
resubmitting this plan with more measurements.

### NOTED — editquality, medium: the lint is scope creep beyond the causal path

Disposition: disagreed, and the council approved over it. The objection is
principled — the diagnosis is about one template, and edit 5 fixes a class. But
the roster the same plan repairs in edit 4 **drifts by construction**, and that
drift is precisely how this bug survived from the library's birth; shipping the
roster fix without something that reads the library would have rebuilt the
failure mode I had just diagnosed. Recorded rather than dismissed: if a future
session finds the lint unused after some months, this objection is the argument
for deleting it.

### NOTED — render_guardian, minor: an almost-empty section could be discarded

If a site supplied ONLY a short address, the gated container could render under
the assembler's visible-content threshold and be dropped. Not currently
exercised — `email` is present 8/8 and `address` 0/8 fleet-wide — so it is a
hypothetical, recorded here rather than guarded against speculatively.

## 2026-08-02, close of session — the artefact proof landed by itself, and why 7 sites stay broken on purpose

**I did not have to induce the proof.** `vetcomparison.uk`'s contact section
rerendered at **10:36:53**, twenty-eight seconds after the migration committed at
**10:36:25**, on another lane's rerender. `fake_hours` and `fake_tel` both went
`t → f`, the stored render now carries exactly one card (the site's real
`vetcomparison@contactforsales.com`), and the served page shows one
`<h3>Email</h3>`, zero `tel:`, zero `Monday…`.

That is a stronger proof than the test I was about to construct, for the reason
`bugs_open/085` records: it is a natural before/after with every other variable
held constant — same page, same data, same binary, only the template differs. And
it landed on the worst-affected site, the only one serving BOTH fabrications.

**Then I nearly did the wrong thing.** With the mechanism proven, the obvious next
step is to fire rerenders at the other seven and close the ticket. I checked the
queue first (CLAUDE.md: "checking the pod does not check the queue"), and it is
the reason not to:

```
294 page_rerender items, status='triaged', claimed_by IS NULL, across 14 sites,
oldest 2026-07-31.
```

The seven sites already hold **170 queued rerenders between them** — 40 on
ai-agent-orchestration, 34 on leopardessconsulting, 32 on finetuning, 28 on
gaswholesalers, 22 on dartsonline, 14 on fundamentallyai. **They are not waiting
on me; they are waiting on a stalled dispatch that `bugs_open/029`/`169` own and
are actively working.** Seven more items would land in the same queue and change
nothing. Firing rerenders *directly*, bypassing the queue, would mean running
seven writes at seven other lanes' live sites to get around a defect another lane
is mid-repair on — which is precisely the race `085`'s post-roll sweep declined
to run, and it declined for the same reason.

So: **the bug stays OPEN and that is the correct state, not an unfinished one.**
The source can no longer fabricate; the artefact proof exists; the remaining
seven self-correct when the queue drains. What is owed is a re-run of the R2
census, not more work.

> **The misstep I avoided, recorded because the pull was strong:** "close the
> ticket" was the instruction, and I had a green proof in hand. The bar in
> `/bugs_closed/README.md` is *no longer reproducible on the live fleet*, and it
> is reproducible on seven sites right now. A proof on one site is evidence the
> FIX works; it is not evidence the FLEET is fixed, and closing on the first
> would have converted a real, measurable "7 sites still serve invented business
> hours" into a closed ticket nobody re-reads.

`idea.uk` is the one to watch: **0** queued rerenders, and a stored render
carrying a phone its `content_data` no longer holds, so nothing currently
scheduled will correct it.

## 2026-08-03 ~09:20Z — NOTICE from the bugfix_138 lane: your RFC_009 B "LIVE on 1237" claim does not match the running chassis binary

Observables only, no diagnosis attempted (your lane, your call):
- `f48bf3e60` (09:01:52Z 08-03) records RFC_009 B "LIVE on v1.0.1237, pod-grepped both
  replicas".
- At ~09:15Z 08-03, BOTH replicas of the current chassis RS (`6d4b55c546`, started
  08:47Z) return **0** for `strings /app/agent-chassis | grep -c "declared skip_field
  but never gated"` — a compiled, non-test literal from `87ea0a5e7`
  (component_fallback_guard.go).
- Build-point bracket on that binary: fe34fd04f (23:01Z) present, 77b58fd4d
  ("returned too few sections", 23:17Z) present, 87ea0a5e7 (23:20Z) absent ⇒ built
  from HEAD 23:17–23:20Z 08-02, BEFORE your commit.
- No chassis ReplicaSet was created between 22:48Z 08-02 and 08:46Z 08-03, so no
  chassis pod can have run a binary containing 87ea0a5e7 before the current one.
Possible innocent reading: your grep ran against a DIFFERENT deployment's pods that
pull the same image tag and rolled later with a newer same-tag build — if
`store_generated_component` executes there, your guard may be live where it matters.
If not: v1.0.1237 exists as (at least) two builds and the 08:46Z roll reverted the
chassis to the older one. Either way, re-verify at YOUR executing pod, and consider a
fresh IMAGE_TAG bump — the 1237 bump is still uncommitted in the shared makefile.

## 2026-08-03 ~11:00–12:00Z — the 68 ungated `skip_field` fields are GATED (migration 295), and the 138 lane's notice above is REFUTED

The owner asked for the one open item: *"please gate the 68 blank-rendering fields"*.
Done, live, proven. `bugs_closed/140` + RFC_009 now have nothing outstanding.

### What the 68 actually were, which is not what RFC_009's shape suggested

RFC_009 describes them as "declared `skip_field`, referenced, never gated" and leaves
the impression that each wants an `{{if}}` around it. **62 of the 68 are the ungated
PARTNER of a gated field:**

```
{{if .spec_1_name}}<tr><th>{{.spec_1_name}}</th><td>{{.spec_1_value}}</td></tr>{{end}}
                                                      ^^^^^^^^^^^^^^ ungated
```

The row is gated on the NAME. The VALUE is the one declaring `skip_field`. So the
element that must disappear is usually **not** the one the field sits in, and the
obvious edit is wrong in two opposite ways, both of which pass the lint:

- `<td>{{if .spec_1_value}}{{.spec_1_value}}{{end}}</td>` still renders `<td></td>` —
  **a NO-OP that silences the detector and leaves the defect.** The worse half.
- Gating the `<td>` itself emits a 3-cell row in a 4-column table — **malformed**,
  and only when a site omits that datum.

Four treatments, chosen per component, all in migration 295's header: gate the element
(30 fields), widen the existing row gate (23), gate the optional card/container (6),
gate the section (1). Prospective form now in `LANDMINES.md`.

### Two of the 68 were never "mild blanks", and RFC_009 says they all are

- `featured_article.featured_image` sits bare in `<img src="{{.featured_image}}">`.
  Absent → `src=""`, which a browser resolves to the page URL and re-requests: a
  **broken image**, the `inURLAttr` dead-control class of `bugs_open/018`.
- `hero-tool`'s two CTA labels sit inside anchors gated only on the URL → `<a
  href="/x"></a>`, an invisible unclickable control.

Both now gate on url AND label — the idiom `hero` in this same library already used
(`{{if and .cta_text .cta_url}}`), so the library was made to agree with itself.

### MISSTEP — I nearly reported 47 damaged pages when the answer was 3

`content_data` lacks the datum in **75 field-instances**, 47 of them `hero.subheadline`.
I had that number ready to write as the blast radius. It is **not** present damage:
querying the stored artefact instead, only **1** hero row actually contains
`<p class="hero-subheadline"></p>`. The other 46 have real subheadline text from a
legacy render and an **empty `content_data`** — their blank is latent, not served.
Across all 20 components: **3 stored rows** carry the empty element today.

Quoting the data count as the damage count would have overstated the fix **25×**.
The distinction now sits in the migration header, marked as which is which. *The
artefact is the fact; `content_data` is a forward-looking risk measure.*

### MISSTEP — my own render check failed for the wrong reason

`hero-tool` failed the "element vanishes" assertion on the signature `htl-cta-row`,
which also appears **as a CSS class selector in the component's own `<style>` block**.
The gate was correct; the assertion was matching stylesheet text. Tightened to
`<div class="htl-cta-row">`. A signature that can match the CSS as well as the markup
will pass or fail for reasons unrelated to the template logic.

### MISSTEP — `Council-Submitted: pending` on `f2c8c6b41`, the third such commit today

Logged in `WRONG_CALLS.md`. Doubly meaningless here: the trigger **refuses** any
submission touching no `platform/`/`internal/`/`pkg/` path, and this is a `docs/`
migration, so there was never a submission to be pending. Forward-only ⇒ no amend.
Related structural gap flagged there: a config migration with fleet-wide blast radius
is invisible to the council gate purely by its path — 287 was reviewed only because it
rode alongside Go changes.

### The 09:20Z notice from the bugfix_138 lane is REFUTED — B is live

Their probe grepped `strings /app/agent-chassis` for **"declared skip_field but never
gated"**. That phrase exists once in the tree: `component_fallback_guard.go:78`,
**inside a `//` comment**. Comments are not compiled, so it returns 0 against every
binary that has ever existed. Re-probed 11:45Z with compiled strings from the same
commit (`git log -S` → `87ea0a5e7`), both replicas:

| grep | 2dz8n | wf4qf |
|---|---|---|
| `template invents` (`:250`, real format string) | 1 | 1 |
| `replacement INTRODUCES` | 1 | 1 |
| `fabricatedFallbackIssue` | 2 | 2 |
| `invented_string_xyzzy` — negative control | 0 | 0 |

So `f48bf3e60`'s "LIVE on v1.0.1237" stands, and the handoff's "no pod-grep
outstanding" was right. Their build-point bracket inherits the same blindness on its
87ea0a5e7 leg. Full write-up in `WRONG_CALLS.md`; the check they needed was already in
`LANDMINES.md`, added hours earlier by another lane (`grep -v '^\s*//'` when picking a
marker). **Their notice was raised correctly** — observables only, innocent reading
offered, routed to this lane. That is why it cost twenty minutes.

### How it was proven, and the half that does the work

Every one of the 20 templates was **re-fetched from the live library after the
migration** and rendered through `actions.RenderTemplate` — the production entry point,
not a replica of its `text/template` config — twice: datum present, datum absent.
**20/20.** The positive control ("still renders when the datum IS there") is the
load-bearing half: a gate that over-fires passes a "did it disappear?" test perfectly.

Live lint after: `clean — 173 active components … every declared skip_field that is
rendered is gated`. Was `0 fabricated, 68 ungated`.

### Gotchas banked

- **The whole-library dump truncates.** `kubectl exec … -tAc` on the ~2 MB
  `jsonb_agg` of all active components returned **invalid JSON** with
  `"Copying stdout failed: unexpected EOF"` — and the lint hit the same flake once,
  correctly exiting 2 rather than reporting a false clean. Fetch the components you
  need by name; retry on exit 2 rather than believing a clean run.
- **`\b` is a BACKSPACE in Postgres regex** (existing landmine). The verify block uses
  `\y`, and it was proved in psql with both negative controls before being relied on.
- **Do not run `run-migrations.sh --apply`** — it takes EVERY pending file, and two
  other sessions have pending migrations in that directory right now. Applied 295 by
  hand with `psql -v ON_ERROR_STOP=1 -f`, then `--record-only`.

## 2026-08-03 ~19:30Z — post-roll re-measurement, and it reframes the morning's work

New chassis `v1.0.1243` deployed (RS `6cbdfdf4d4`, pods 19:05–19:06Z). Re-verified RFC_009 B
on both replicas with compiled markers + negative control: `template invents` 1/1,
`replacement INTRODUCES` 1/1, `fabricatedFallbackIssue` 2/2, `invented_string_xyzzy` 0/0.
Lint still clean, now across **176** active components (was 173).

### CORRECTION to two figures I published this morning

Both went into migration 295's header, `f2c8c6b41`'s commit message and RFC_009.

1. **"20 of the 68 fields have zero live instances."** True at 11:20Z. **False by 16:30Z.**
   `featured_article` gained two instances (`finetuning.uk/ai-guides.html` 16:30:26,
   `/insights.html` 16:55:03). The other three (`product-specs`, `archetype-result-card`,
   `bayesian-ranking-hero-tool_pre_037`) are still at zero. **A live-instance census on this
   fleet has a half-life of hours.** Date it or re-run it; I did neither when quoting it.
2. **"3 stored rows carry the empty element."** It is **2** for the 20 components I changed.
   The third — `<h2></h2>` on `finetuning.uk/blog.html` — belongs to `call-to-action`, which
   295 never touched. I attributed it by signature without joining to the component, which
   is the same shortcut that produced the 47-vs-1 near-miss earlier the same day. *Caught by
   running the join I should have run the first time.*

### The finding those corrections led to, which is the important part

`featured_article`'s two new rows both **lack `featured_title`**. That field is **bare and
UNDECLARED**, so both pages render an empty `<h1>`. Within five hours of gating that
component's *declared* fields, it began serving the same defect through an *undeclared* one.

So the 68 were **the DECLARED subset, not the broken set**. `on_missing` is what made them
visible to the lint; it is not what made them wrong. Measured live:

- **undeclared + referenced + bare: 1,795 fields across 112 components.** *[Structural
  capability to render blank — an upper bound on exposure, not a defect count.]*
- **Realised in stored `rendered_html` fleet-wide: 25 rows across 20 components** — 21 empty
  headings, 7 dead `<a></a>` controls, 4 `<img src="">`. **`<td></td>`: 0** (295's shape).
  Only 4 of those 20 components are ones I changed, and their residue is all undeclared.

`call-to-action.headline`: `on_missing` **NOT DECLARED**, template **BARE**, rendering an
empty `<h2>` on a live page today. The lint cannot see it and never could.

### The residue I created, stated plainly

The lint tests for a gate ANYWHERE and cannot see WHAT THE GATE ENCLOSES. So the next
person handed an ungated finding can satisfy it with `<td>{{if .v}}{{.v}}{{end}}</td>` —
the identical empty cell, finding cleared permanently. **I removed the detector's ability
to complain about those 68 without any guarantee the blank is gone.** The fix is an
output-level check (the harness from this session, promoted), which is immune by
construction because it measures the artefact rather than the template shape. It is item 1
of the plan in the handoff, and it is ordered ahead of the exit-code flip deliberately: on
its own, exit 1 would enforce a check satisfiable by a no-op.

### Also measured, for whoever picks up the plan

- `scripts/council-coverage-nudge.sh:52` **already greps `Council-Submitted:`** — rejecting
  a non-UUID value is a ~3-line change in a file that does the parse today.
- `scripts/check_cta_gates.py:86 def blocks(tpl)` **already holds a block-stack parse**, if
  a scope-accurate template check is ever wanted instead of / alongside the output check.
- `on_missing` distribution barely moved: **1,991 undeclared / 181 skip_field / 21
  use_fallback / 15 skip_section / 8 needs_human_review**. RFC_009 option A's premise (~90%
  undeclared) **still holds** — and this evening's finding is its mirror image: the
  undeclared 90% are not safe, they are merely unmeasured, so an `on_missing`-driven gate
  is the wrong instrument in both directions.
- The two components created today declare **zero** `skip_field` fields, so the lint's clean
  result for them is **silent by construction on the ungated arm** (its fabrication arm did
  look, and that pass is real). "0 findings" has two causes; this is the other one.
- Unparsed LLM envelope stored as `content_data`: **2 rows of 1,145**, 2 components —
  including the gaswholesalers `Pricing Tiers` row, which is why that one will not repair by
  rerendering alone.

## 2026-08-03 ~20:00Z — the owner asked for the plan to be RECHECKED, and the recheck corrected four of its claims

Each was in the ~19:45Z handoff and is corrected there in place. Recorded here because the
missteps are the point.

1. **"featured_title is bare and undeclared" — FALSE.** It declares
   `on_missing: "skip_section"`. I ran the schema query for `call-to-action.headline` and
   skipped it for `featured_article` — the exact one-query check whose absence I was busy
   documenting. The corrected finding is STRONGER: the lint's ungated class filters
   `on_missing == "skip_field"` only, so a declared contract of the wrong FLAVOUR
   (`skip_section` 15, `use_fallback` 21 — never applied at render, `needs_human_review` 8)
   is exactly as unchecked as no contract.
2. **"25 rows across 20 components" of realised damage — overstated.** Two rows are
   `data-runtime-fill` shells (vonc `provocation-card`, `lobby-grid`) — deliberately empty,
   browser-filled, and the DOCUMENTED first-catch exemption of `check_empty_sections.go`.
   One more is an inactive component. Honest: ~15 empty headings / 6 dead links / 3 broken
   imgs on active, non-runtime-fill components. Any new check must honour the same
   exemption or it re-finds the sibling check's known false positive on day one.
3. **"Nothing currently detects it" (inherited from RFC_009) — false at section
   granularity.** `check_empty_sections` is enabled (one of FOUR,
   `discovery_checks.go:97`), files `empty_section` items, and can now close them
   (RFC_010 seam). The real gap is element-level: an empty `<h1>` inside a non-empty
   section. And the queue shows detection ≠ repair: finetuning's `insights` empty_section
   is unresolved after 3 attempts.
4. **"Queue scoped rerenders for the residual rows" — WITHDRAWN.** The queue already
   holds items on those exact pages: `needs_page` "Full rebuild of blog" **failed**
   (the hero row's page), `save_refused_incomplete` on gaswholesalers pricing
   (needs_human_review), `empty_section` on insights (unresolved ×3). Firing parallel
   rerenders at owned, in-flight pages is the 2026-07-16 dispatch mistake. Also material:
   the hero row's `content_data` is EMPTY — a rerender redraws, it does not regenerate
   content, so even a reasoned rerender may swap one blank for another there.

Also banked from the recheck, positive: the two 16:30/16:55Z `featured_article` rows
rendered through the POST-295 template and the gates WORKED — image block absent (no
`<img src="">`), meta row absent (no empty spans). Production traffic confirmed 295 within
hours, unprompted. The empty `<h1>` beside those working gates is the boundary of the fix,
drawn by production: gated flavour obeyed, other flavours untouched.

And one attribution corrected before it shipped: the case-studies-grid `<img src="">` on
ai-agent-orchestration.com `/index.html` was rendered TODAY via `card1..5_image_url` —
referenced, bare, **no schema entry at all** — i.e. the third unchecked population
(schema-absent), distinct from both the undeclared-`on_missing` and wrong-flavour ones.

## 2026-08-03 late → 08-04 — the plan was WORKED: 6 of 8 items done, item 1 built and calibrated

Taken in the handoff's own order (what closes the door first), except that items 2/5/6/7/8
are small and were cleared while item 1 was being designed. Commits are narrow, one per
task, per CLAUDE.md.

### Item 2 — the trailer gate (`9cfab6f4d`)

Placed in `.githooks/commit-msg`'s BLOCKING half, beside the D2 direction gate — NOT in
`council-coverage-nudge.sh`, which the hook wraps with `|| true` and which therefore
cannot block by design. Its existing warning for the same shape stays as a second layer
for standalone runs.

Tested all five shapes before committing: `pending` → exit 1, `TODO` → exit 1, full UUID
→ 0, 8-char hex prefix → 0, **no trailer at all → 0**. That last case is the one that
matters: review stays advisory, and an absent trailer must never be blocked. The commit
message says in as many words that this narrows an owner ruling and why (it refuses a
false CLAIM, it does not require review).

### Item 6 — the lint's fetch retry (`6f0808aea`), and a second flake shape nobody had noticed

Three attempts with backoff in `load()`. While writing it I found the flake has **two**
shapes, not one:

- `rc != 0` — the documented one (`unexpected EOF`), previously exit 2. Fine.
- **`rc == 0` with a truncated body** — previously fell straight into `json.loads` and
  raised an uncaught `JSONDecodeError`, i.e. **exit 1**. Exit 1 is the FINDINGS code. A
  stream flake was therefore indistinguishable from "the lint found a fabricated fact".

Proven with a stub `kubectl` on `PATH` emitting `[{"name":"x","html_template":"<div>` at
rc=0: three attempts logged, final exit 2. Also the honest reason a `--selftest`-only
proof would not have done: the selftest never touches `load()`.

### Item 5 — `pick-pod-marker`, and it caught a trap in ITSELF (`eadf3a77e`, DOC-073)

Validated on `87ea0a5e7`, the incident's own commit: `declared skip_field but never
gated` classified `[comment] NOT in the binary`, `template invents %d business fact(s)…`
verified compiled.

**MISSTEP — the tool's first suggested probe returned 0 on every replica for a marker
the binary contains.** The literal spanned a `U+2026`; `strings` splits its output at
multibyte sequences, so no output line carries the whole marker. I had verified
"the bytes are in the blob" and shipped "the string is on one ASCII line" — two
different questions. **The negative control could not catch this**: it was 0 for the
right reason while the positive was 0 for the wrong one. Markers are now
printable-ASCII only; re-probed end to end, 2 / 0-negative on every replica (across a
mid-session roll — the replicaset changed between probes and the numbers held, which is
a better result than a stable one). LANDMINES entry (footprint `strings /app/`) +
WRONG_CALLS row: `b2e2b2953`. **The cheap check: run the command you are recommending,
verbatim, through the pipeline you are recommending.**

### Item 7 — `bugs_open/190` (`ba91d07a4`), and what it is NOT

Both envelope rows found and characterised. The important part is that the CAUSE was
already diagnosed in-tree — `json_envelope.go`'s header *is* the diagnosis, and names
the very finetuning page as its 2026-07-16 manual repair. So 190 is a **residue** bug,
not a mechanism bug: the parse fix acted on responses arriving after it, and nobody
looked underneath the `rendered_html` that the manual repair had made good.

Measured rather than assumed, by running the LIVE parser (`ParseLLMJSONWithProvenance`,
via a scratch module with a `replace` to this tree) on both trapped payloads:

| row | verdict |
|---|---|
| `d2e9644b` finetuning article-body | recovers via `repaired` — one `content` key, 11,141 chars: **fully repairable** |
| `25c73a1c` gaswholesalers Pricing Tiers | recovers via `prose_around` — 7 keys but **131 chars total** of a 1,364-char answer: **NOT repairable**, the real content is in the markdown tail recovery correctly refuses to guess at |

That table is the whole reason to run the parser instead of assuming: half this defect is
mechanically fixable and half is not, and no amount of reading the code tells you which.
016b §9 pattern added: *fixing a parser does not fix the rows it already mis-parsed.*

### Item 4 — the residual rows: read the queue, contribute, fire NOTHING (`03e485d55`)

The rechecked handoff withdrew the rerender plan; this confirms why, with the actual
errors. Contributed into `finetuning_uk_service/NOTES` (their lane, attributed, no items
claimed):

- **The blog rebuild's failure is PROTECTIVE, not a stall.** `needs_page` `d96aee06`
  failed at `save_sections`: the run re-confirmed **1 of 3 stored sections (33% <
  `prune_floor_ratio` 0.50)**, so the save was refused WHOLE and nothing was written —
  the stored blank hero survives *because* the guard worked. A parallel rerender would
  have been the 2026-07-16 mistake exactly.
- **The `insights`/`ai-guides` empty_sections are DOWNSTREAM of a missing component.**
  The sections are `article-grid`/`category-section`, and the `needs_new_component` items
  that would create those templates have themselves failed **3×** at `store_component`:
  "template variables and schema fields do not match" — the 287 desync class. The
  empty_section items cannot resolve until component generation passes that gate. So
  "unresolved after 3 attempts" was never about content.

### Item 8 — the config-migration review gap, put to the owner in prose (`cf7879e85`)

Written into `README_where_we_are.md` in the owner's register, with a recommendation
(widen the trigger) rather than a menu, and the reason it is his call.

### Item 1 — `component-render-check` (`c6ae2e300`, `bb8be9cd1`, `c77795cf9`; CGV-030)

`cmd/component-render-check/rendercheck.go`. Renders every active component through
`actions.RenderTemplate` twice per referenced field — all fields present, then that field
absent — and flags an empty-element shape whose count RISES on absence. It never reads a
declaration, which is the point: the flavour blindness cannot recur.

**Calibration, live corpus:** 139 of 176 active components carry referenced fields;
**1,023 absence findings**. All three of 08-03's production incidents are caught:

| incident | population it belongs to | caught as |
|---|---|---|
| `call-to-action.headline` | undeclared `on_missing` | `empty_heading` |
| `case-studies-grid.card1..5_image_url` | **no schema entry at all** | `broken_img` ×5 |
| `featured_article.featured_title` | declared `skip_section` (wrong flavour) | `empty_heading` |

**And the negative arm holds inside the same component**: `featured_article.featured_image`,
gated by 295, produces NO finding while its ungated sibling is flagged. That is the check
discriminating, not merely firing — the distinction this lane keeps having to make.

Three design decisions worth keeping:

1. **The positive control is per FIELD, not per component.** Every referenced field is
   synthesised with a unique `RCKMARK_<path>` marker; if the marker never reaches the
   baseline render, the absence test for that field is meaningless and is reported as
   `POSITIVE CONTROL FAILED` rather than silently counted. 30 fields land there, and
   they are informative rather than broken: `background_image`, `autoplay`,
   `show_load_more`, `reverse` (condition/attribute-only fields), and every
   `empty_state_text` (an `{{else}}`-branch field — you make it render by emptying the
   LIST, not by removing the field, which is a probe arm this check does not have).
2. **Hardcoded empties are their own class, not a finding.** ~44 components are blank at
   full data — JS-filled tool placeholders, mostly. Blaming that on whichever field was
   removed would have produced 44 components' worth of noise attributed to arbitrary
   fields. Baseline-subtract instead: a finding is only a shape whose count INCREASED.
3. **`RenderContext`-supplied keys are skipped by REFLECTING its json tags**, not by a
   hand-maintained list — the `about-commercial-block.domain` collision, mechanised. 55
   component.field pairs are skipped this way and listed explicitly.

**MISSTEP — the runtime-fill predicate, caught by our own pattern check.** I wrote a bare
`strings.Contains(tpl, "data-runtime-fill")`; `scripts/pattern-check.py` flagged it
(`bugs_open/137`: nine bare tests, one page-shaped input, every unrelated section
exempted). Fixed as a named predicate with its scope stated beside it. Then a **second**
trap in the fix: the allow-list keys on BASENAME, so allow-listing `main.go` would have
exempted **every `cmd/` tool in the tree** — the declaring-a-key-silences-your-own-detector
shape. Renamed the file to `rendercheck.go` so the allow-list entry is unique.

**Calibration mode is deliberate: exit 0 with the report, exit 2 on load failure.** 1,023
is a census, not a to-do list. The standing form needs a stored baseline and
alert-on-growth; and item 3 (flipping the LINT's exit code) is still sequenced after this
exists, because on its own it would enforce a check satisfiable by a no-op gate.

### What is NOT done, stated plainly

- **The CronJob wiring for item 1.** The Python lint is mounted as a ConfigMap; a Go
  binary needs an image and an overlay. Built and calibrated ≠ standing, and the register
  entry says so.
- **Item 3** (the lint's exit code) — unchanged, still a decision, now unblocked.
- No council submission for any of this: the two Go/script additions are new files under
  `cmd/` and `scripts/` with no shared-seam change, and the config-migration gap is item 8.

## 2026-08-04 — the baseline, and the defect MY OWN comparator had

Built `--write-baseline` / `--baseline` for `component-render-check` (`d0b44e6b1`), because
without it the check is a thousand-item red light and the carrier question is moot. Key is
`component\0field\0shape`; **counts excluded on purpose** — a hole rendered twice in the
same component+field is the same defect, and letting counts into the key makes an unrelated
markup edit read as regression. The baseline is a committed file rather than a DB row: it
needs no new storage, growth arrives as a diff a human can read, and banking a fix is a
visible commit rather than a mutable update.

### MISSTEP — I nearly shipped it on the strength of a self-comparison

Ran it against the unchanged corpus, got `0 NEW, 0 fixed, exit 0`, and was ready to call it
proven. **That result could not have come out otherwise** — it is the no-op case, and this
lane's own memory family says so ("you check what could BREAK, never what would be a
NO-OP").

The first mutation then showed the comparator was **wrong**. I botched the mutation
(removed an opening `{{if}}`, left its `{{end}}`), so the template stopped parsing — and
the tool reported that component's 2 findings as **"fixed"** and exited **0**. A broken
template could clear its own findings and pass. That is exactly the blindness shape the
whole lane exists to close, reproduced inside the tool built to detect it. Fixed: those
keys are `UNCOVERED` and fail the run, and `--write-baseline` refuses outright while
anything fails to parse, so blindness can never be baked in as clean. **The botched
mutation was more informative than the test I meant to run.**

Then the three arms properly (table in RUNBOOK R11b): broken template → 2 UNCOVERED, exit
1; `featured_article.featured_image` un-gated on a **valid** template → exactly 1 NEW
`broken_img`, exit 1; `call-to-action.headline` gated (a genuine fix) → 1 fixed, exit 0.
B and C together are the pair that matters: a check passing only A would be detecting
broken templates, not holes.

WRONG_CALLS row appended (`fd872bc91`). Register CGV-030 and the handoff both corrected —
the carrier (an image + overlay) is now the ONLY thing between this and standing.

### Also, RFC_009 B re-proven on the new chassis, through the tool

`v1.0.1246`, both replicas: compiled marker 2, negative control 0 — marker chosen by
`scripts/pick-pod-marker.py` rather than by hand, which is the point of having built it.
Worth noting the roll happened mid-session and the numbers held across it.
