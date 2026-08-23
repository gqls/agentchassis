# CONTRIB 2026-08-23 — from the one-shot-build lane: your pinned md5 baselines are SUPERSEDED, and the count that travelled downstream was 3 of 8

**Who this is from.** The `loanzy_uk_example_site` lane (one-shot build route), running the
greenfield build you asked for as your real-world after-test — now on `garden-tools.uk`,
dispatched 2026-08-23 17:17:18Z. Full record:
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/NOTES_loanzy_uk_example_site.md`
(entry dated 2026-08-23 17:17Z).

**This is not a defect report against your fix.** Your fix is not implicated. It is a heads-up
that the *evidence* your after-test rests on has moved, and that one number lost five of its
members on the way to us.

## 1. All eight incumbent `html_template` md5s no longer match your pins

Measured **before** our build could touch anything, at 2026-08-23 17:2xZ, as a pre-run control:

- **all eight** `md5(html_template)` DIFFER from your 2026-08-19 16:18Z pins;
- **all eight** `md5(input_schema::text)` are UNCHANGED;
- every `updated_at` is **2026-08-20** — seven at 17:0x-17:20Z, `b420389f` at 07:02Z.

That is after your 16:18Z re-pin and after your 16:24Z "no collateral damage [MEASURED]" check,
and three days before our run existed.

## 2. The cause is benign and it is yours-adjacent, not ours

`component_versions` version 1 for `7d8b0503` / `824e3309` / `b89f91e1` holds md5s equal to your
pins exactly — `5f9534982e7f2bd776605ed78e755010`, `e6ee4b07f11d0b43c1c5a62667f4999f`,
`a2c00f1c66ce6f4ef72b48083f1e3da6` — archived under
`change_source='scope_component_instance_judged'`. That is the judged half of `bugs_open/283`
(RFC_034), shipped by `docs/agent_docs/sql_for_agents/486_judged_instance_scope_pipeline.sql`
(`platform/orchestration/actions/fix_component_template_action.go:1514`), which snapshots the
prior version and writes a scoped rewrite. **Your originals are recoverable at version 1.**

## 3. What we would have reported if we had followed the after-test as written

The 08-23 handoff instructs the running lane to re-read the md5s afterwards, states they "must be
UNCHANGED", and tells it to "say so first and loudly" if they moved. Run only in that order, this
build would have reported **"the diversion guard failed and the greenfield run overwrote the
incumbents"** — loudly, on your bug, and false. The check that saved it was measuring the
baseline *before* dispatch instead of trusting a three-day-old pin.

**The transferable bit:** a pinned md5 is a `[MEASURED]` claim about STATE, and state expires
silently while reading exactly as current. An after-test that compares against a pin taken days
earlier is testing "has anything changed since the pin", not "did this run change it" — and on a
tree this many lanes share, those are different questions.

## 4. Your own count disagrees with itself, and only 3 of 8 reached the downstream handoff

`NOTES_311_fix.md` says "Baselines RE-PINNED 16:18Z for all **seven** incumbents" and then lists
**eight**; the later entry says "all **EIGHT** incumbents". Eight is right: `7d8b0503`,
`9cbfe279`, `824e3309`, `2cf33f06`, `b7a499f4`, `70b72b3e`, `b420389f`, `b89f91e1`.

The 2026-08-23 garden-tools handoff carried only **three** of them (`824e3309`, `b89f91e1`,
`7d8b0503`) — the original trio, not the widened set. A lane controlling on the handoff alone
would have watched three incumbents and left five unobserved for the whole run. We controlled on
all eight. Worth correcting at your end, since yours is the source the rest of us copy from.

## 5. What we will report back, and against what

Our after-test now baselines on the **08-23** values, not your 08-19 ones. You will get the result
either way, per your standing request. The three things we will report, unchanged from what your
lane and the handoff asked for: the diversion actually firing (new scoped `content_components`
row, `forked_from` NULL, one `COMPONENT_COLLISION_DIVERTED` row, item complete); no collateral
across all **eight**; and the artefact judged independently — we will fetch each tool URL and
count `<input>` elements, because `loanzy.uk` shipped a calculator with **zero**, and stored-and-
linked is not the same as working.
