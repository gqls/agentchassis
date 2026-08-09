# HANDOFF — 2026-08-09, fresh chat starts here

**Supersedes `HANDOFF_2026-08-08_continue_here.md`** for state; that file's §3 (the line)
and §6 (before you dispatch) are unchanged and still binding — read them, they are not
repeated here.

This session did two things: continued the D10 contract backlog (batch 6/6b), and then —
at the owner's instruction — **enabled the contact forms end to end.** The second half
collided with another lane and that is the first thing to read.

---

## 0. READ THIS FIRST — bug 228 is SHARED, and I overstepped

`bugs_open/228` (the fake contact form) is owned by **`bugfix_228_contact_block_transport`**,
which had a plan, two council rounds and a committed Go fix before I touched it. I filed
the bug, checked ownership at FILING time, and did not re-check before FIXING. The full
account, what is live now, and the one decision left for them are in a signed contribution
at the bottom of the bug file — **`bugs_open/228`, section "CONTRIBUTION 2026-08-09"**.
Logged in `WRONG_CALLS.md`.

**Do not run `bugfix_228_contact_block_transport/apply_228_contact_block_fix.sh`.** It will
abort on its needle guard; the abort is correct, the change is already applied.

**The open decision (theirs, not ours):** two JS implementations now exist for
`contact-block` — theirs (2,232 B, prepared, unapplied) and mine (7,325 B, applied, live,
proven on five branches). The prover is subject-agnostic:
`scripts/prove_contact_delivery.go <live-url> <candidate.js> <cb|cf>`. Run theirs through
it and pick. Forward-only either way.

---

## 1. Contact forms — what is LIVE, measured at the artefact

**`contact-block` — the bug. FIXED and delivering on both served pages.**

| page | state |
|---|---|
| robot-hands.com/contact.html | `mailto:robot-hands@contactforsales.com`, driven as a visitor, PASS |
| leopardessconsulting.co.uk/ai-readiness-quiz.html | `mailto:leopardess@contactforsales.com`, driven as a visitor, PASS |

Driven in a real browser against the **served** page: fill → submit → assert the mailto the
browser is actually sent to. It carries the name, the message and the reply address, is
addressed to that site's own configured inbox, the status never says "sent", and the typed
text is preserved. The `setTimeout` fake send is gone from both served assets (positive and
negative controls both run).

**`contact-form` — 13 pages. 12 live, 1 owed.**
Gained `#cf-status`, a script ref and real delivery JS. Live and verified on
ai-agent-orchestration, dartsonline, finetuning, fundamentallyai, gaswholesalers,
leopardess, oufe, robot-hands, vetcomparison, vonc (+ leopardess/contact.html canary).
**OWED: `idea.uk/contact.html`** — re-rendered and committed to `gqls/vm-sites` at 10:01Z,
served object still dated 05 Aug. Different repo and host from the other twelve; that
deploy path is the thing to chase, not the render.

**Blast radius, measured not assumed:** on the canary, a whole-page diff before/after was
**17 lines, every one of them the intended change** (form id, status div, script ref,
status CSS). Nothing else on the page moved.

### How it actually delivers, and the honest limit

A `mailto:` built in JS with explicit `subject=`/`body=`, from the address the platform
already holds in `sites.email`, via the platform's own `sanitiseFormAction`. **The visitor's
mail client is in the loop** — we hand off and say so ("Opening your email app…"), we never
claim receipt. That is this estate's decided mechanism (owner, 2026-07-17; `contact-form`'s
13 pages) and it needs no new infrastructure.

**A true server-side receipt is possible and is NOT built.** `tools.apis.uk` (the island
`tools-api`) already accepts cross-origin POSTs from these domains, and `platform/mailer`
(register PUB-003) is built, council-approved and has **zero importers**, with contact forms
named in its own docstring as the third queued consumer. Both new scripts route on the
action's scheme and already handle an `https` destination correctly (POST, report the
server's actual status, keep the text on failure) — **so switching is a config change, not a
code change.** What is missing is SMTP credentials (none exist in the cluster) and an island
deploy, which is owner-gated. Costed in the bug file.

---

## 2. Two platform traps this work found — both silent, both encoded in a script

1. **`page-rerender` has two paths and the wrong one looks like success.** Without
   `input_data.spec.reason` ∈ (`image_landed`|`section_data_resolved`|`cta_links_stale`) it
   assembles from each section's STORED `rendered_html`, so a **template** change never
   appears — while still republishing `/tools/assets/*.js` from `js_content`. You get the
   new script against the old markup, `COMPLETED`, and a green asset check. Measured here:
   the contact-block asset went 2,100 → 7,345 bytes while the form tag was untouched.
2. **`page_name` must be at `input_data.spec.page_name`** — the exact path, read out of the
   live `save_sections` config, not guessed. Anywhere else and `save_sections` returns
   `{"skipped":true,"success":true,"sections_saved":0,"reason":"no page name"}`: sections
   re-rendered and discarded, reported as success. Same family as `bugs_open/095`.

**Both are encoded in `scripts/RERENDER_page.sh <site_id> <domain> <page_id> [reason]`**,
which takes a reason and looks the page name up for you. Use it; do not hand-roll a kcat
dispatch. (And `kubectl run -i` inside a `while read` loop **eats the loop's stdin** — the
first rollout attempt dispatched exactly one of ten. Use an array, or `< /dev/null`.)

---

## 3. D10 contract backlog — state

**51 subjects proven end-to-end: 49 sections + 2 tools**, all S6-green in-cluster with the
negative control confirmed red.

- Batch 6 (08-08 eve): `news-listing`, `latest-news`, `case-studies-grid`, `contact-block`,
  `blog-listing` — 8 checks each, 8/8 mutants, S6 12/12.
- Batch 6b (08-09): `game-list` (7/7, S6 11/11), `ai-readiness-quiz` (9/9, S6 13/13).

**The rule the interactive pile forced, and the four worked shapes, are in
`HANDOFF_2026-08-08_continue_here.md` §3. Read it before authoring another interactive
fence.** In one line: *a fence must carry one check a static render cannot satisfy, or it
certifies a dead panel.*

⚠ **`contact-block`'s fence is now understated.** It deliberately asserted the VALIDATION
path only, so as not to ratify the fake success. That reason has gone — the component now
delivers. The right addition is a check that the success state is downstream of a
destination. Not done; it is the natural next edit to that PLAN.

### Remaining work

- **~10 interactive sections**, all single-placement: `tool-ai-vendor-trust-checklist`,
  `tool-archetype-taster-quiz`, `adoption-tracker-listing`,
  `tool-gripper-cycle-time-estimator`, `audience-check-form`, `model-directory-listing`,
  `protocol-tracker-listing`, `report-request-form`, `tool-ai-agent-roi-estimator`.
  ~30–45 min each. `gauntlet-interface` is **lane-owned — coordinate.**
  **Check first whether the subject is interactive at all** — `length(js_content) > 0` was
  WRONG for `game-list` (its script binds selectors absent from its own template, and no
  page loads it).
- **~10 ready tools** — re-run `CHECK_naming_contract.sh` + census first.
- **8 chrome-blocked sections** — fences authored, baselines cannot go green until each
  site's `hero.jpg`/logo 404 is fixed. Still the highest-value small repair on the board.
- **Drift rows (six)**: ported-page ×58 on lmc/loancash, `featured-content`, `pricing`,
  `leopardessconsulting.co.uk/blog.html`, `contact-block` on `finetuning.uk/case-studies.html`.

---

## 4. Standing defect list for the owner

1. **`bugs_open/228` — SUBSTANTIALLY FIXED** (see §1). Remaining: idea.uk's deploy, the
   fleet roll for `85390ee33`, and the JS choice (§0).
2. gaswholesalers.com: every page 404s `/assets/images/logo.png`. **4+ days.**
3. The `hero.jpg` 404 family (`bugs_closed/128`, measured 07-31, **still serving**) — ≥7
   sites, incl. vetcomparison.uk which 128's own list missed.
4. `finetuning.uk/index.html` 404s five `case-studies-grid` card images.
5. `article-body` ships no `pre`/`code` overflow CSS.
6. Broken tool pages: tool-gas-unit-converter, tool-ab-test-calculator (idea.uk serves raw
   `{{.placeholders}}`), tool-equity-release (active row, 404 URL).

---

## 5. Instruments committed this session (all under `staged_component_build/scripts/`)

| script | what it does |
|---|---|
| `gen_component_plan_sql.py <manifest>` | persist a batch of component PLANs; DO/RAISE length asserts |
| `prove_contact_delivery.go <url> <js> <cb\|cf>` | drive a contact component through all five destination branches in a real browser |
| `probe_mailto_form_encoding.go` | measure what a `mailto:` FORM hands the transport (GET destroys `?subject=`; POST hands it a body a mailto cannot carry) |
| `apply_contact_form_delivery.py [--apply]` | the exact-string, length-asserted component update |
| `RERENDER_page.sh <site> <domain> <page> [reason]` | single-page rerender with the two traps in §2 encoded |
| `contact_block.js`, `contact_form.js` | the delivery scripts as applied |

Nothing lives only in the scratchpad. It gets wiped between sessions; that is why the
persist generator had to be rewritten five times before this one.
