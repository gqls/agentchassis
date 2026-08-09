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

**`contact-form` — 13 pages. ALL 13 LIVE.**
Gained `#cf-status`, a script ref and real delivery JS. Live and verified on
ai-agent-orchestration, dartsonline, finetuning, fundamentallyai, gaswholesalers,
leopardess, oufe, robot-hands, vetcomparison, vonc (+ leopardess/contact.html canary).
idea.uk lagged the other twelve by ~5 minutes because it deploys to `gqls/vm-sites` (a
different repo and host); it landed at 10:06Z and is verified. **Nothing is owed here** —
the earlier draft of this handoff said idea.uk was outstanding, which was true when written
and stopped being true five minutes later. Re-check before quoting it.

**Blast radius, measured not assumed:** on the canary, a whole-page diff before/after was
**17 lines, every one of them the intended change** (form id, status div, script ref,
status CSS). Nothing else on the page moved.

### How it actually delivers, and the honest limit

A `mailto:` built in JS with explicit `subject=`/`body=`, from the address the platform
already holds in `sites.email`, via the platform's own `sanitiseFormAction`. **The visitor's
mail client is in the loop** — we hand off and say so ("Opening your email app…"), we never
claim receipt. That is this estate's decided mechanism (owner, 2026-07-17; `contact-form`'s
13 pages) and it needs no new infrastructure.

**Update 13:15Z — the class fix (`85390ee33`) is live on v1.0.1274 and proven operating;
the hand-seeded special case has been removed and both pages still deliver. See the bug
file's UPDATE section.**

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

**56 subjects proven end-to-end: 54 sections + 2 tools**, all S6-green in-cluster with the
negative control confirmed red.

- Batch 6 (08-08 eve): `news-listing`, `latest-news`, `case-studies-grid`, `contact-block`,
  `blog-listing` — 8 checks each, 8/8 mutants, S6 12/12.
- Batch 6b (08-09): `game-list` (7/7, S6 11/11), `ai-readiness-quiz` (9/9, S6 13/13).
- Batch 7 (08-09 eve, plan-first — Fable planned read-only, Opus implemented):
  `tool-ai-vendor-trust-checklist` (8/8, S6 12/12 — its `#vtc-c1` 20x20 check is
  `bugs_closed/157`'s own reproducer, now a permanent regression check),
  `tool-gripper-cycle-time-estimator` (8/8, S6 11/11), `tool-archetype-taster-quiz`
  (8/8, S6 12/12), `report-request-form` (STATIC by design — its only JS observable
  fires a real POST into `idea_uk_vm_site`'s operator funnel; PLAN forbids driving it),
  `model-directory-listing` (STATIC by design — re-hydration is idempotent; serve_local
  required even for the static fence). NOTES 2026-08-09 evening entry has the full table.

**The rule the interactive pile forced, and the four worked shapes, are in
`HANDOFF_2026-08-08_continue_here.md` §3. Read it before authoring another interactive
fence.** In one line: *a fence must carry one check a static render cannot satisfy, or it
certifies a dead panel.*

✅ **`contact-block`'s fence has been strengthened (2026-08-09 afternoon) — DONE, not owed.**
It had deliberately asserted the validation path only, so as not to ratify the fake success;
that reason went when the component started delivering. It now carries
`form-has-a-real-destination` (`form.cb-form[action]:not([action=""])`) — scheme-agnostic so
it survives a future receipt endpoint, and mutation-proven against **both** states the bug
passed through (no action at all; `action=""`). 9 checks, 10/10 mutants, S6 `f3cd89a2` 13/13.

### Remaining work

- **Interactive sections: 4 left, each gated, none plain.**
  - `tool-ai-agent-roi-estimator` — fence+mutants COMMITTED (7/7 desktop), NOT persisted:
    the trial found a REAL mobile overflow (`h3.roi-inputs-title` fixed 297.9px inside the
    tool). Unblocks when the CSS is fixed; do NOT gate the check to desktop to dodge it.
  - `audience-check-form` — deferred on two gates: the prover needs a POST-stub sibling of
    serve_local, and the S6 run fires a real POST into idea.uk's free-taster funnel —
    coordinate with `idea_uk_vm_site` FIRST.
  - `adoption-tracker-listing` + `protocol-tracker-listing` — ALL FOUR tracker feeds 404
    (measured; model-directory's serve fine). CONTRIB filed into `model_directory_pipeline/`;
    fence after they publish or de-script the templates.
  - `gauntlet-interface` — **lane-owned, coordinate.**
  **Qualify before budgeting**: a subject is interactive only if the JS binds selectors in
  its own template AND a served page loads the script AND the effect is observable and safe
  to drive — `length(js_content)>0` fails three different ways (game-list, trackers,
  model-directory).
- **~10 ready tools** — re-run `CHECK_naming_contract.sh` + census first.
- **8 chrome-blocked sections** — fences authored, baselines cannot go green until each
  site's `hero.jpg`/logo 404 is fixed. Still the highest-value small repair on the board.
- **Drift rows (six)**: ported-page ×58 on lmc/loancash, `featured-content`, `pricing`,
  `leopardessconsulting.co.uk/blog.html`, `contact-block` on `finetuning.uk/case-studies.html`.

---

## 4. Standing defect list for the owner

1. **`bugs_open/228` — FIXED, LIVE on all 15 served pages, and the CLASS FIX IS PROVEN.**
   `85390ee33` rolled on **v1.0.1274** (13:xx Z); pod-grepped on both replicas with controls
   (new marker 0 → 1 across the roll). Proven to be *operating*, not merely present: the
   hand-seeded `content_data.form_action` special case was REMOVED from both placements and
   they still render the site's mailto — a check that would have produced `action=""` had
   the class fix not been reaching that path, which is exactly what happened on v1.0.1270.
   No contact-block placement carries a hand-seeded key any more, and the fence edit in §3
   is done. **The only thing still open on 228 is the JS choice (§0), which is the owning
   lane's call.** Nothing blocks a visitor.
2. gaswholesalers.com: every page 404s `/assets/images/logo.png`. **4+ days.**
3. The `hero.jpg` 404 family (`bugs_closed/128`, measured 07-31, **still serving**) — ≥7
   sites, incl. vetcomparison.uk which 128's own list missed.
4. `finetuning.uk/index.html` 404s five `case-studies-grid` card images.
5. `article-body` ships no `pre`/`code` overflow CSS.
6. Broken tool pages: tool-gas-unit-converter, tool-ab-test-calculator (idea.uk serves raw
   `{{.placeholders}}`), tool-equity-release (active row, 404 URL).
7. **NEW (batch 7):** `tool-ai-agent-roi-estimator` scrolls sideways on mobile —
   `h3.roi-inputs-title` carries a fixed 297.9px width inside the tool
   (leopardessconsulting.co.uk/tools/ai-agent-roi-estimator.html). One-line CSS fix;
   its proven fence is committed and waits on it.
8. **NEW (batch 7):** all four tracker feeds 404 on ai-agent-orchestration.com
   (`adoption-tracker[-full].json`, `protocol-tracker[-full].json`) — the two tracker
   pages' client refresh has never once worked. CONTRIB filed with
   `model_directory_pipeline` (their publish trigger is the likely one-dispatch fix).

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
