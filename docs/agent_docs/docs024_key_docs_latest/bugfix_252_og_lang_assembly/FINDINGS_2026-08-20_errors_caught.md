# FINDINGS — errors and defects caught while fixing bug 252 (og/lang slug), 2026-08-20

Written at the owner's request: *"note down the errors you caught so they can be addressed."*

Everything here was found **incidentally**, while doing something else. None of it is the bug I was
asked to fix. Each entry says what it is, how certain I am, whether anything is already filed, and
what I recommend. **Nothing below is fixed by this lane** unless it says so.

Ordered by what I'd address first.

---

## A. Live defects nobody has filed

### A1. `webdesign.co.uk` serves 117 pages with no `<head>` element at all
**Certainty: MEASURED.** Its head component (`webdesign.co.uk Document Head`,
`14cf6193-c8f0-4640-9cf1-f8b5347e6885`, its own `function` `webdesign-couk-head`) is a bare fragment:
no `<head>` open tag, no `</head>` close tag. The served page begins
`<!DOCTYPE html><html lang="en"><meta charset="utf-8">` — browsers imply the element, so it renders,
but nothing in the platform can address it.

**Why it matters beyond tidiness:** every helper that splices into a head keys on `</head>`
(`injectCanonicalLink`, `injectPageJSONLD`, `injectRobotsNoindex`, `injectBrandHeadTags`, my
`spliceOpenGraph`). Each one has its own private fallback for the missing marker, and they do not all
agree — `injectBrandHeadTags` **returns the head untouched** when it cannot find `</head>`, so this
site silently opts out of the whole brand block. **117 assembled pages, the largest site in the
fleet**, and it is invisible because every injector "succeeds".

**Recommend:** a bug file. The fix is probably to give that component a real `<head>` wrapper, but it
is a hand-authored chrome component on a live site, so it wants its own canary. **Not urgent, but it
will keep silently excluding that site from every future head feature** — it already excludes it from
the lang half of bug 252 and from the brand head block.

### A2. `head-seo-standard` has three template branches that can never fire
**Certainty: MEASURED.** The template references `{{.og_title}}`, `{{.og_description}}`,
`{{.og_image}}` and `{{.site_name}}` — **none of which is declared in its `input_schema`**. The live
field list is `accent_color, background_color, canonical_url, description, font_url,
gtm_container_id, primary_color, secondary_color, structured_data, text_color, theme_css, title`. So
`{{if .og_title}}…{{else}}{{.title}}{{end}}` has always taken the else branch, and `{{.title}}` is
empty at a site-level render — **which is the direct cause of the blank `og:title` on 4 sites.**
`{{if .canonical_url}}` is declared but nothing anywhere sets it, so that branch has never fired
either.

**Already handled, partly:** migration 507 removes the two branches that were actively emitting a
blank. The inert ones are left alone deliberately.

**Recommend:** the general problem is worth a check, not a patch — **a template referencing an
undeclared key fails silently and forever**, and this component had four. A cheap audit
(extract `{{.key}}` references per template, diff against `input_schema` keys) would find the rest
fleet-wide. There is already a landmine about the sibling failure (a *scalar* schema entry being
silently skipped); this is the same family from the other end.

### A3. `spliceMetaDescription`'s legacy fallback can corrupt any blank tag in the head
**Certainty: MEASURED, with a reproduction.** When the exact blank
`<meta name="description" content="">` is absent, it fills **the first `content=""` it finds
anywhere in the head**, whatever tag owns it. With blank `og:` placeholders present I reproduced it
writing the page description into `<meta property="og:image" content="…">`.

**This lane narrowed it but did not remove it.** Ordering the calls so my splice runs last means the
fallback can only reach a tag I am about to rewrite — but that only protects `og:title`,
`og:description` and `og:url`. **Any other blank tag in any head is still exposed**, and the existing
test (`inject_canonical_link_test.go:133`) pins the behaviour as intended.

**Recommend:** a bug file to make the fallback target `name="description"` specifically, or drop it.
It exists to preserve pre-2026-08 behaviour on heads shaped differently; a census would probably show
zero such heads, in which case it can just go.

---

## B. Tooling defects — our instruments giving misleading answers

### B1. The `090` diagnosis loop cannot see served bytes, and says so as "UNVERIFIABLE"
**Certainty: MEASURED on run `af31ec22`.** It confirmed the mechanism from source, then could not
reach a single real page, for two structural reasons: it queries **`pages.rendered_head`, which is
vestigial and returns 0 rows fleet-wide**, and every `site_components.rendered_html` row it fetched
came back **truncated before the `</head>` tail** — precisely where the evidence sits.

**Why this needs addressing rather than just documenting:** the output is indistinguishable from "your
claim is doubtful". A session that files a `090` about anything a visitor can see gets a courteous
non-answer and may weaken a correct claim, or spend another run that cannot succeed either.

**Filed:** LANDMINES entry (footprint `090_TRIGGER…`, `pages.rendered_head`,
`site_components.rendered_html`). **Recommend beyond that:** two cheap fixes — stop the loop reaching
for the three vestigial `pages` columns (they have no data and one caller, itself a known bug), and
either raise the truncation limit for `rendered_html` or fetch the tail. A third option is to give the
loop an HTTP fetch capability, which is the real gap.

### B2. `kubectl logs -l app=agent-chassis | grep 'build provenance'` matches other lanes' payloads
**Certainty: OBSERVED today.** The chassis logs whole council/diagnosis payloads, and those payloads
contain the *text* "build provenance" (they quote the landmine about it). My provenance grep returned
2.4MB of another lane's landmine corpus. This is already a documented landmine; I mention it only
because **it fired on the exact command CLAUDE.md recommends**, so the recommended command may want
an anchor (e.g. match the startup line's own prefix).

**Recommend:** tighten the recipe in CLAUDE.md to an anchored match.

### B4. `rerender-chrome` reports COMPLETED having skipped its own start step
**Certainty: MEASURED on corr `83d980c2` (2026-08-20).** Dispatched with the envelope its own seed
header documents. Returned `complete | COMPLETED` in seconds. The stored head was **untouched**
(`updated_at` still 2026-08-18), and `collected_data` contains **no `site_components_result` and no
`render_site_components` key** — only `complete`, whose `result` is a verbatim echo of the request.
No `__step_error`, no `error`.

The live definition inspects as correct: `start_step: render_site_components`, config
`{"slots":["header","footer","head"],"force_rerender":true}`, `next_step: complete`, `output_field:
site_components_result`. And `force_rerender` does propagate correctly in the Go
(`render_site_components_action.go:78` → `:311` → the `if !force` exit), so the flag is not the
problem — **the step never ran at all.** No locks on the slots either.

**Why it matters:** this agent exists to be the cheap lever — re-render chrome without the per-page
fan-out — and it is the thing a session reaches for when the scheduled pipe is too slow. It currently
**succeeds without doing anything**, which is the platform's own "a `complete` status is not a
repaired artefact" failure in its purest form.

**Not chased by this lane** — it belongs to the 226/351 lane and my propagation path (the
`stale_chrome` fingerprint) does not need it.
**Recommend:** a bug file against seed `351`. Until then, anyone using it must diff
`site_components.updated_at` before and after rather than trusting the COMPLETED.

### B5. New sites default to `en` and nothing will ever notice
**Certainty: INFERRED from mechanism, but the mechanism is now measured.** Migration 508 sets
`locale.lang` per site as a one-shot. `indoorplanters.co.uk` was created the day I authored it and was
caught only because I had written a fail-closed guard. **The next site created gets no `locale.lang`
at all**, falls to the Go default `en`, and there is no check anywhere that would report it.

**Recommend:** either give the site-creation path a `locale.lang` default, or add a small daily check
("real sites with no `site_config.locale.lang`"). The second is cheaper and also catches sites created
by paths nobody remembers. Flagged rather than built because "what language is this site?" is a
product decision and a silent default is exactly what bug 252 was about.

### B3. Migration numbers have no allocator, and collisions are now routine
**Certainty: MEASURED.** `497` and `498` were each already doubled by two lanes before I started. I
read the directory, saw max `501`, wrote two files as `502`/`503` — and by the time I committed,
**`502`, `503`, `504`, `505` and `506` had all been taken** by three other lanes. I renumbered to
`507`/`508`.

**Why it matters:** the runner takes every pending file in a directory, and two files sharing a number
have no defined order. Today the damage is confusion; the day two same-numbered migrations touch the
same row it is a real conflict.

**Recommend:** the cheapest real fix is to stop using a shared integer — e.g. allocate from a
`migrations_reserved` table, or name files by timestamp. Failing that, a pre-commit check that
refuses a duplicate number would catch it at the only moment anyone is looking.

---

## C. Coverage gaps — checks whose premises this fix falsifies

### C1. Two live checks document "the shared `<head>` cannot carry a per-page value" as settled fact
**Certainty: READ IN SOURCE.** `verify_site.py:71` (`OG_PER_PAGE`) exempts `og:url` from validation as
an "accepted loss", and `check_site_structural_validity.go` (~`:55`, `:1029`) excludes `og:url`
entirely with a long rationale resting on the same premise.

**Both go false the moment bug 252 ships, and neither fails loudly** — they just keep passing. This is
the exact shape that produces "a PASS from a blind check outlives the blindness".

**Recommend:** retire both in the same lane as the rollout. Listed as a close-out item on
`bugs_open/252`; flagging here because close-out items get dropped when a lane ends.

### C2. The wholesale idempotency guard is the generic mechanism, and it is untouched
**Certainty: READ IN SOURCE, and the council's `bug_historian` seat raised it independently.**
`injectBrandHeadTags` skips its entire block if the head contains `rel="icon"` OR `og:image`. That is
why `webdesign.co.uk` gets no brand tags at all, and why the 4 duplicated-tag sites duplicate (the
guard cannot see a blank `og:title`).

**The important framing, which I did not have until the council said it:** fixing `og:url` fixed a
*symptom*. **Any future per-page-varying tag added to that block reproduces bug 252 exactly.**

**Filed:** `bugs_open/322` item 4. **Recommend:** treat 322 item 4 as the higher-value half of that
file, not a tidy-up.

### C3. Four head-producer fixes have now landed on one producer only
**Certainty: COUNTED, and the council's `architecture` seat objected on it.** `injectPageJSONLD`
(2026-07-28), `injectCanonicalLink` (2026-08-02), `injectRobotsNoindex` (2026-08-09) and now this one
all live on `assemblePage`; `AssemblePageAction` (3 active agent types) gets none of them. The
convergence question has been open and architecture-scope since 2026-07-29 with no owner decision.

**Consequence, concretely:** a page rebuilt through the other producer now gets **no `og:url` at all**
rather than a wrong one — better, but different, and the blast radius is unmeasured because nobody has
established which live pages are built by which producer.

**Recommend:** an owner decision on SEO-003. I have written a binding escalation threshold into
SEO-005 at the council's direction — **a fifth instance raises an RFC rather than taking a fifth
patch** — but that is a tripwire, not a decision.

---

## D. My own errors (full accounts in `WRONG_CALLS.md`)

Listed for completeness; all are corrected and none shipped.

1. **I had my own ordering constraint backwards, wrote it into the code as settled fact, and my test
   could not fail.** Swapping the order to check the test would fail — it passed. The discriminating
   fixture then showed my chosen order wrote the page description into `og:image`. *Caught by
   mutation-testing the test, not by review.*
2. **A `lang="en"` grep silently missed the two emitters that mattered**, because Go escapes the
   quotes, and I briefly concluded the code was already fixed. *A pattern containing a quote cannot
   match a literal that escapes it, and the failure is directional — it returns the decorative hits.*
3. **A mutation reported PASS because my `sed` had silently failed to apply it** — a false pass,
   identical in appearance to a non-discriminating test.
4. **I asserted a control count (13) from memory in a migration guard**; the measured value was 14.
   Caught before shipping only because I ran the query.
5. **I guessed two column names** (`site_components.html` / `.component_type`) instead of reading the
   schema first, as CLAUDE.md requires.
6. **I planned around a stale tree state.** The plan was built around "the package does not compile
   and both my target files are dirty"; by the time I executed, that session had committed and both
   were clean.

---

## E. Not an error, but the thing most likely to be misread

**The chassis build deployed today (`v1.0.1319`, cut 10:18Z) does NOT contain this fix**, which was
committed at 14:03Z. Proven at the binary on 2026-08-20 14:35Z, with both controls:
`spliceOpenGraph` **absent**, positive control `injectCanonicalLink` PRESENT, fabricated negative
control absent.

**So both migrations stay HELD and no canary is possible yet.** A fresh build is not evidence a fix
shipped — it is evidence a build happened.
