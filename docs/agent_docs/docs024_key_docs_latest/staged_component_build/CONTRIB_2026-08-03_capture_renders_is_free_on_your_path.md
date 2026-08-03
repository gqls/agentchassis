# CONTRIB 2026-08-03 — your component acceptance path can photograph a passing page, and the code is already there

**From:** the brochure component library lane (`brochure_component_library/`), which owns
TL-035.
**To:** whoever picks up `staged_component_build` next.
**Why you are getting this:** the owner ruling of 2026-07-29 §3 — a shared mechanism's
other consumers must be **told**, not merely measured. You are the other consumer. This
asks nothing of you and changes nothing you own; it tells you a switch exists.

---

## 1. The one-line version

Your `request_component_browser_run` step already supports `capture_renders` — the
feature ships in the helper both acceptance actions share, so **adding one config key
turns it on, with no code change and no rebuild.** Your step config currently omits the
key, so it defaults to `false` and your runs photograph nothing when they pass.

## 2. How I found it, so you can judge the claim rather than take it

I was re-checking TL-035 (my lane's "photograph a page that PASSES, not only one that
fails") and my own re-check query showed a **render-less acceptance note newer than my
armed one** — yours, `teaser-reveal-panel`, 2026-08-02 21:53. For about ten minutes it
looked like my feature had broken. It had not. Both lanes file into the same
`doc_notes` category with the **same `created_by='tool-acceptance-agent'`**, so the note
cannot tell the two callers apart; the discriminator is the action on the orchestration
row. Your run `cee46f41-1ccb-49aa-a980-4914b4c43088`:

| | my path | your path |
|---|---|---|
| action | `request_browser_run` | `request_component_browser_run` |
| `capture_renders` in step config | `true` (seed 292) | **key absent** |
| response keys | …+ renders | `run_id, results, skipped, summary` |

**Your negative control is not implicated and was not the cause.** `neg_control` did
exactly what you built it to do — refused a `bad_page_id` where the component is not
placed, `__step_error` populated, `neg_control_confirmed_red`. I mention it only so you
know I read it and am not reporting it as a fault.

## 3. Why it is a config-only change

`platform/orchestration/actions/tool_acceptance_actions.go`:

- `RequestBrowserRunAction` → `return dispatchBrowserRun(...)` at **`:184`**
- `RequestComponentBrowserRunAction` → `return dispatchBrowserRun(...)` at **`:390`**
- `dispatchBrowserRun` reads the flag at **`:220`** —
  `datahelpers.GetBoolField(config, "capture_renders", false)` — and puts it in the
  adapter envelope at **`:268`**

Both callers build the **identical** `run_checks` envelope from the same helper. So the
adapter cannot tell your run from mine except by what the config asked for.

## 4. What to add, when you want it

In your `request_run` step's `config` (currently `profiles`, `error_step`,
`domain_field`, `page_id_field`, `site_id_field`, `criteria_field`, `function_field`):

```json
"capture_renders": true
```

Mine went in as a numbered seed rather than a bare `UPDATE`, and I would suggest the
same — a DB-only write leaves a key in `default_config` with no provenance at all.
`sql_for_agents/292_acceptance_runs_photograph_a_page_that_passes.sql` is the worked
example, including the guard that asserts a **neighbour** key so a surgical `jsonb_set`
can be told apart from a write that flattened the whole `config` object.

**A wrinkle specific to you:** your 21:53 run had **no `agent_definitions` row at all**
— `component-acceptance-probe` returns 0 rows fleet-wide, including snapshots and
soft-deleted. It ran from an **inline `workflow_plan`**. So if you are still dispatching
by hand, the key goes in the plan you dispatch; there is nothing seeded yet to patch. If
and when you register the agent, that is the natural moment to include it.

## 5. What I have NOT proved, stated plainly

**I have not run your path with the flag on.** "Adding the key yields renders on the
component path" is **[INFERRED]** from the shared helper — identical envelope, identical
`run_checks` action, one code path from `:390` — not observed at an artefact. It is a
strong inference and it is still an inference, and you should not record it as proven
until a run files a note with a `Rendered:` line on it. I did not fire a run on your
path because it is your lane and your dispatch, not mine to spend.

Two limits carried over from my side, both real and both worth knowing before you rely
on the output:

- **Nobody looks at the renders yet.** They land as `s3://` URIs inside a technical
  note — no page, no digest, nothing that puts an image in front of a person. That is
  the open owner call on my lane. Turning the flag on gets you the photograph; it does
  not get you a reader.
- **The PNGs have never been fetched.** The bucket is private and returns **401 for a
  key that does not exist exactly as for one that does** (proven with a deliberate
  nonsense key), so object existence rests on code, not observation. Marked
  `[UNFETCHED]` in my lane's notes too.

## 6. Pointers

- My lane's evidence: `brochure_component_library/EVIDENCE_2026-07-31_TL-035_capture_renders.md`
  (adapter half) and `EVIDENCE_2026-07-31b_TL-035_caller_half.md` (caller half, incl. the
  one council objection that changed the code).
- The trap that sent me here is filed fleet-wide in `LANDMINES.md` — *"Two acceptance
  callers file notes under ONE category with the SAME `created_by`"*. It names your
  action as the second caller, which is the honest framing: **the gap was in my arming,
  not in your build.**
- Register: TL-035 (`docs026_concept_register/register/tool-lifecycle.md`).
