# HANDOFF — bugs_open/228, cold-start here (2026-08-09)

**State in one line: CLOSED IN SUBSTANCE. Framework fix committed, council-
APPROVED (4 rounds, 15 reviewers), LIVE fleet-wide. Component fix LIVE on
both real pages, shipped by a different session with a better implementation
than this lane's own staged one. Nothing outstanding on this bug.**

Cold-start reading order: this file → `bugs_open/228_…md` (read the WHOLE
file — it has two long contributions at the bottom, both load-bearing) →
`SUMMARY_2026-08-09_contact_block_transport.md` → `NOTES_…md` (newest at the
bottom) if you want the full blow-by-blow including all four council rounds.
Register entry: **LNK-031** (`docs026_concept_register/register/link-management.md`).

## What is done (do not redo)

| | |
|---|---|
| Framework fix | `85390ee33` — `RenderTemplateReportingMissing` (`component_library.go`) seeds `form_action` when a template references it, regardless of content-authoring; 3 new tests, mutation-checked by hand |
| Register | `LNK-031`, `link-management.md` — the `sanitiseFormAction`/`deliverableFormAction` mechanism had no entry before this, now documented with its landmine |
| Council | corr `46f87e4c-05fc-4a5c-bd6a-93a073b63253`, **APPROVED round 4** (15 reviewers, 2 abstained), after 3 REVISE rounds each catching something real (see NOTES for all four verdicts in full) |
| Deploy | `85390ee33` pod-verified live fleet-wide on `v1.0.1273` — checked by IMAGE across every currently-running Deployment-managed chassis-binary pod (not just the 2-pod `-l app=agent-chassis` label match — see the landmine below) |
| Component fix (shipped by `staged_component_build`, NOT this lane) | `contact-block` `html_template`/`js_content` and `contact-form` `js_content` all live; commits `2c62379d5`, `7d57b0342`, `b9c4d743d` |
| Live proof | Both real pages independently re-verified at the artefact by this lane after the other session's work: `robot-hands.com/contact.html` and `leopardessconsulting.co.uk/ai-readiness-quiz.html` serve a correct `mailto:` action; served JS asset (7,345 B) carries the new 5-branch logic, zero functional `setTimeout` |
| Tooling left behind, unexecuted but reusable | `apply_228_contact_block_fix.sh` (needle-guard + backup + auto-rollback + `RETURNING`-gated manual commit — a template for any future `content_components` mutation), `dispatch_228_rerenders.sh`, `verify_228_deployed_page.sh` (deploy-window-aware, cache-busted, single-fetch-per-check verifier — a template for any future post-rerender check). `js_content_after_228_fix.js` is superseded, kept as reference only |
| Registered landmines | `LANDMINES.md`: (1) `who-owns.py`'s VERDICT line false-positives on a bare digit-substring match in unrelated commits — read the "(none identified)" section, not the verdict line; (2) the pre-existing `-l app=agent-chassis` 2-pod undercount, re-confirmed the hard way this session |
| Missteps logged | `WRONG_CALLS.md` ×2 from this lane (a placeholder `Council-Submitted: pending` trailer the commit hook correctly rejected before any commit existed; citing a 2-pod grep as decisive despite the landmine already being in reach) — plus `staged_component_build`'s own entry for the duplication itself |

## Still open

**Nothing on this specific bug.** Three small, deliberately out-of-scope
residuals are named for whoever picks them up next, none urgent, none
blocking anything live:

1. Widen `check_contact_form_undeliverable.go`'s discovery-check scope to
   also match `data-component="contact-block"` (currently `contact-form`
   only) — now safe to do since neither component's JS unconditionally
   intercepts submit anymore, which was the reason it was excluded before.
2. Extend `contact-block`'s acceptance fence (`doc_plans`,
   `subject_type='component'`, `subject_key='contact-block'`, owned by
   `staged_component_build`) to assert the success path now requires a real
   destination, now that asserting it doesn't ratify a defect.
3. Architectural: a council reviewer (`bug_historian`) correctly noted that
   `RenderTemplateReportingMissing`'s generic "`<no value>` stripped to empty
   string, no error" behaviour is the same shape as the platform's
   worst-documented recurring bug class (Go template `missingkey=zero`).
   This fix narrows the exposure for one field; it doesn't close the generic
   one. Might be worth a register watch entry or a follow-on ticket if it
   bites again elsewhere.

## The original task, if you're continuing it rather than just verifying this bug

The instruction this lane was given was to find and fix **an** unowned bug
from `bugs_open/`, with full rigor (research, a Fable-drafted plan, council
review, docs, missteps logged). That instruction is satisfied — this was
that bug. If the intent is to keep going (pick up another unowned bug the
same way), start fresh: **`scripts/who-owns.py <N>` is not reliable
stand-alone** — its VERDICT line false-positives on bare digit-substring
matches (see the landmine above). Filter on the "(none identified)" section
under "likely OWNING workstream(s)" instead, and cross-check with
`git log --all --oneline --since="<a few hours ago>"` for the bug's SLUG, not
its bare number — a fresh, unowned-looking bug can be picked up by another
session between your check and your first commit, as happened here in
reverse (this lane owned 228 for hours; another session missed that and
duplicated the fix anyway, from a stale `who-owns.py` read taken at filing
time rather than re-run at fixing time).

## Landmines a fresh session must not step on

- **Never trust a 2-pod `-l app=agent-chassis` grep as "fleet-wide".**
  Enumerate by image: `kubectl -n ai-persona-system get pods -o
  jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.ownerReferences[0].kind}{"\t"}{.spec.containers[0].image}{"\n"}{end}'
  | grep agent-chassis`. Job-owned pods are per-work-item and age out on
  their own; Deployment/ReplicaSet-owned pods are the ones that must be on
  the current tag.
- **A single-service `kubectl apply`/deploy is not this session's to run** —
  precedent says it's blocked by a permission classifier; ask the owner to
  trigger the release, then pod-verify yourself.
- **`who-owns.py`'s printed VERDICT line is not the check** — read the
  "likely OWNING workstream(s)" section; `(none identified)` there is the
  real signal regardless of what the top-line verdict says.
- **A council submission's `file` field must be a real repo-relative path,
  no whitespace** — a descriptive label passes client-side validation and
  fails server-side as `complete_invalid` with no verdict row.
