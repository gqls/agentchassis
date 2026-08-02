# PLAN 2026-08-02 — `bugs_open/140`, the contact-info component fabricates business facts

**Lane taken** 2026-08-02. Bug filed 2026-07-29 by the oufe workstream and left
**unowned** — `scripts/who-owns.py` names no workstream, and a grep of all 26
live session transcripts (modified since 2026-08-01 22:00) finds **zero**
mentions of `bugs_open/140`. Nobody is on it.

---

## What the bug is

The single shared `content_components` row `function='contact-info'`
(`0bd72302-e9bf-4dc0-a615-41a9c919bf17`, 2,573 chars) renders Email / Phone /
Hours cards **unconditionally**, and each one substitutes an invented value when
the site's datum is absent:

```
{{if .phone}}…{{else}}+1234567890{{end}}                            ← tel: href
{{if .phone_display}}…{{else if .phone}}…{{else}}+1 (234) 567-890{{end}}
{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}
{{if .email}}…{{else}}info@example.com{{end}}
```

## Re-verified 2026-08-02 — still valid, and it has GROWN

The filing measured 6 live uses. Today there are **8**:

| domain | has `phone` | has `hours` | renders fake hours | renders `+1234567890` |
|---|---|---|---|---|
| ai-agent-orchestration.com | t | f | **t** | f |
| **dartsonline.com** | t | f | **t** | f |
| finetuning.uk | t | f | **t** | f |
| fundamentallyai.com | t | f | **t** | f |
| gaswholesalers.com | t | f | **t** | f |
| idea.uk | f | f | **t** | f |
| leopardessconsulting.co.uk | t | f | **t** | f |
| **vetcomparison.uk** | f | f | **t** | **t** |

`dartsonline.com` and `vetcomparison.uk` are new since the filing —
vetcomparison rendered **2026-07-31**, two days ago, and is the first site to
render the fabricated **phone** as well as the hours. Served proof, on the wire
today:

```
$ curl -fsS 'https://vetcomparison.uk/contact.html?cb=…' | grep -oE 'tel:[^"]*|Monday[^<]*'
tel:+1234567890
Monday – Friday, 9am – 6pm
```

**Site seven and site eight arrived exactly as the filing predicted.** That is
the argument for fixing the mechanism rather than the eight rows.

## The finding that reframes the fix — the template contradicts its own schema

The component's `input_schema` **already declares the correct behaviour**:

```json
"hours":   { "source": "site_specs.identity.hours", "on_missing": "skip_field" },
"phone":   { "source": "site_specs.identity.phone", "on_missing": "skip_field" },
"address": { "source": "site_specs.identity.address", "on_missing": "skip_field" },
"section_title": { "source": "llm", "fallback": "Contact Us", "on_missing": "use_fallback" }
```

`on_missing: skip_field` is the contract. The template does the opposite: it
substitutes a fabricated value. **So this is not a policy change requiring the
owner to choose between two defensible behaviours — it is making the template
obey the contract it already publishes.** The schema also shows the library has
a first-class way to express a *legitimate* default (`"fallback": "Contact Us"`
on `section_title`, a label) — so the distinction between a label default and a
fabricated fact is one the schema language already draws, and only the template
ignores it.

Corroborating: **`hours` is supplied by 0 of 1,089 `page_components`
fleet-wide.** The Hours card has never once rendered a real datum, anywhere, in
the platform's history. Every instance of it ever served was fabricated.

## Second defect in the same row — template/schema desync, detected in May, acted on by nobody

Template variables: `title`, `intro`, `email`, `phone`, `phone_display`, `hours`.
Schema fields: `email`, `hours`, `phone`, `address`, `intro_text`, `section_title`.

So `{{if .title}}` and `{{if .intro}}` are **permanently false** — the keys are
never populated, which is why `title`/`intro` are absent from `content_data` on
all 8 sites and every site renders the hardcoded "Contact Information". And
`address`, declared and sourced, is rendered nowhere.

The quality pipeline already caught this and nothing consumed it:

```
quality_score 80 | schema_template_synced f | quality_checked_at 2026-05-18
quality_issues  ["template var {{.intro}} has no schema entry"]
```

Detection wired to nothing — the `bugs_open/083` / `115` family. Recorded, not
re-litigated here.

## Why the detector that is named for this defect did not catch it

`check_placeholder_contact` (`discovery_checks/check_placeholder_contact.go`)
exists, is registered, is configured on `quality-discovery-agent`, and raises
work items titled *"Fabricated contact info on page %s"*. Its pattern set is a
roster of generic placeholder conventions — `555-…`, `(555)`, `@example.com`,
`123 Main St`, `Lorem ipsum`, `+1 (000)`, `John Doe`.

**It does not contain a single one of the literals our own component library
ships.** Measured over every unlocked `page_components` row fleet-wide:

| pattern class | rows matched today |
|---|---|
| existing patterns (all nine, combined) | **1** (`user@example.com`, a `ported-page` on webdesign.co.uk) |
| fabricated hours — no pattern exists | **8** |
| `+1234567890` / `+1 (234) 567-890` — no pattern exists | **1** |

So the detector is blind to 8 of the 9 live fabrications, and the 8 it misses
come from the platform's own library.

> **[UNVERIFIED — and deliberately not claimed]** I can NOT say the detector
> "ran and was blind" on these pages. `quality-discovery-agent` did sweep four of
> the eight affected sites (ai-agent-orchestration 05-01, gaswholesalers 04-25,
> finetuning.uk 04-24, leopardessconsulting 04-24) and raised `generic_theme`
> items, so the runs were real and other checks fired. But **every one of those
> eight contact renders postdates its site's sweep**, so no sweep has ever seen
> the current artefact. The blindness is certain *from the code and the pattern
> census*; its having fired-and-missed in production is not established, and I am
> not asserting it. `placeholder_contact` has raised **0 items, all-history**.

Whether that agent is adequately driven is `bugs_open/149` Group B/C's question,
and that lane is **actively worked** — I am not touching it. A blind detector
stays blind after 149 revives it, which is why the pattern gap is worth closing
on its own.

## The fix, ordered by what closes the door

1. **Make the template obey its schema** (DB config, live immediately). Each
   contact card renders only when its datum exists; the fabricated literals are
   deleted outright. Align template variables to the declared schema field names
   so the dead `title`/`intro` branches become live `section_title`/`intro_text`,
   and render the declared-but-orphaned `address`. Gate the container so an
   all-absent component renders nothing rather than an empty shell — the rule
   `bugs_open/111` already established for the footer
   (`RenderFallbackFooter`, `d4731109d`), never applied to the section component.
   **This makes the bad state unrepresentable: a datum nobody supplied cannot
   render.**

2. **Teach the runtime detector the dummies our own library ships** (Go →
   chassis build). `check_placeholder_contact` gains the fabricated-hours and
   dummy-phone classes, so a reintroduction is caught in the rendered artefact.
   Honest about its reach: inert until the agent that hosts it is driven (149).

3. **A standing lint that derives from the library, so the roster cannot drift**
   (`scripts/check_placeholder_fallbacks.py`). Built in the house style of
   `scripts/check_cta_gates.py` (`bugs_closed/023`, the closest prior art — same
   shape: hard-coded fallbacks in shared component templates, repaired by
   migration and kept repaired by a standing lint). It parses every ACTIVE
   template, finds `{{else}}` literal branches, and separates a **label** default
   ("Read more", "Get Started" — legitimate) from a **fact** default (contact
   detail, price, address, domain — a fabrication). It has a finding to prove it
   works before I write it: `about-commercial-block` hard-codes
   **`fundamentallyai.com`** as an `{{else}}` fallback in a *shared* component.

Rejected: a migration stamping real hours into 8 rows (only if the owner states
them; does not close the door for site nine), and per-site stored-instance
patches (the template re-fabricates on any rerender of an unlocked row).

## Blast radius, measured before submitting — not left for a reviewer

Gating each card on its datum, against today's live data:

- **email** — present on 8 of 8. Every site keeps its Email card. The
  `info@example.com` fallback is currently dead on all 8 (the one live
  `@example.com` row fleet-wide belongs to a different component,
  `ported-page` on webdesign.co.uk).
- **phone** — present on 6 of 8. Those six keep the card. Two lose it:
  **vetcomparison.uk**, which today publishes `+1234567890` (pure fabrication),
  and **idea.uk**, whose stored render shows a real number its `content_data` no
  longer holds (the `117` stored-artefact drift family) — on its next rerender
  under today's template it would publish `+1234567890` too, so gating prevents a
  *future* fabrication there rather than removing a present truth.
- **hours** — present on 0 of 8. All eight lose the Hours card. **That is the
  correction**: nobody stated those hours and no evidence register holds them.

Checked and NOT a fabrication, so deliberately untouched: `+44 (0) 7934 524 911`
appears on six sites, which looks alarming until traced — it comes from
`sites.content_data.phone`, it is the owner's own number, and these are the
owner's own portfolio sites. Real datum, correctly propagated. Not mine to
"fix".

## Consumers to tell (owner ruling 2026-07-29 §3 — measuring is not telling)

The 8 sites belong to other lanes. The change to their guarantee is: *a
contact card whose datum you never supplied will stop rendering, so a contact
page may carry fewer cards after its next rerender.* Named in the council
submission and recorded in the bug file; the two whose visible output changes
first are vetcomparison.uk and idea.uk.

## Diagnosis-loop position (owner ruling 2026-07-31)

Not filed to `090`, and stating why rather than omitting it: I am **not filing a
new cross-cutting root cause** — I am fixing an existing bug whose cause is
local and self-evidencing. Every claim above is first-hand: the template and
schema read from the live row, the fabrication reproduced on the wire with
`curl`, the fleet counts measured by query, the detector's blindness read from
its source and confirmed by a pattern census. The one claim I could not verify
first-hand (did a sweep ever see these renders) is marked UNVERIFIED above and
not asserted.

## Council

Platform code changes (`platform/…`) → council gate before/alongside the commit,
per CLAUDE.md. The template change is shared config across eight other-lane
sites, which the filing itself says needs a council run or the owner's nod.
