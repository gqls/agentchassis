# NOTES — bugs_open/084, asset reference resolution

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-05 — picking the bug, and proving nobody else had it

`who-owns.py` is a COMMIT reader, so it cannot see a session mid-fix. It returned
"OWNED or recently active" for essentially every candidate, which is not a usable
signal on a tree this busy. What discriminated was grepping the live session
transcripts:

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -name '*.jsonl' -mmin -180 -size +10k); do
  echo "=== $f ==="; tail -c 400000 "$f" | grep -oE 'bugs_open/[0-9]{3}' | sort -u | tr '\n' ' '; echo
done
```

38 live sessions, ~40 of the 45 open bugs claimed. Untouched by any of them:
**084, 085, 096, 107, 113, 114, 122, 146**.

- `096` self-reports **CLOSED 2026-07-28** in its own header and was simply never
  moved to `bugs_closed/`.
- `085` is fixed and verified live on both render paths; what remains is
  site-specific follow-up explicitly owned by `brochure_component_library`.
- `113`/`114`/`122` are brochure/dartsonline lane residue.
- `107` is recorded in its own file as owned by `vigilant_designer` Phase 4.

`084` is genuinely unowned: one commit, 2026-07-26, nothing since.

## 2026-08-05 — the bug is still valid, read from the code

`platform/orchestration/actions/discovery_checks/check_tool_acceptance.go:433-440`
is unchanged from the filing:

```go
case "asset_loads":
    if ch.Path == "" {
        skip(ch.ID, "asset_loads with empty path")
    } else if strings.Contains(html, ch.Path) {
        pass(ch.ID, "asset path referenced: "+ch.Path)
```

Presence in the HTML string, never a request. And nothing else fleet-wide fetches
a `<script src>`. The three discovery checks that DO make outbound HTTP are
`check_backend_unreachable.go`, `check_backend_entry_orphaned.go` and
`check_tool_acceptance.go` (which fetches the page itself, not its assets).

## 2026-08-05 — MISSTEP: my first measurement was a regex over raw HTML, and its only finding was a phantom

I ran this to size the damage:

```sql
SELECT DISTINCT s.domain, m[1]
FROM page_components pc ..., LATERAL regexp_matches(pc.rendered_html, '<script[^>]+src="([^"]+)"','g') m
WHERE p.deployed_at IS NOT NULL;
```

It returned `webdesign.co.uk | ...` — a `<script src="...">` with a literal
ellipsis — and `curl` confirmed **404**. I very nearly wrote that up as a live
production defect, LLM truncation shipped to a tool page.

It is not a script tag at all. Pulling the surrounding characters showed it is
**prose inside the tool's own JavaScript**:

```
// We want to keep anything that looks like <script src="..."> or <link rel="stylesheet">
```

A comment, inside a `<script>` block, describing a regex. No browser ever
requests it. **A regex over raw HTML cannot tell an element from a mention of an
element** — and the population it is most likely to mis-fire on is exactly the
population this bug is about: tool pages, whose whole content is JavaScript that
talks about HTML.

The check: parse the DOM. The re-measurement fetched all 541 deployed pages and
extracted `<script src>` / `<link rel=stylesheet href>` with `html.parser`, which
sees elements only. Logged in `WRONG_CALLS.md`.

## 2026-08-05 — the honest blast radius: the defect is real and the current population is ZERO

Re-measured properly (`scratchpad/realsweep.py`, `resolve.py`):

- 541 deployed pages fetched; 509 returned 200 (16 × `lendzy.co.uk` 522 origin
  down, 15 × `leopardessconsulting.co.uk`/`fundamentallyai.com` 404 — a different
  defect, not this one).
- **854 `<script src>` elements, 96 distinct referenced assets** (scripts +
  stylesheets, relative URLs resolved against their page).
- **96 of 96 return 200.** The single non-200 was `000` on
  `fundamentallyai.com/assets/css/styles.css` and returned 200 on retry —
  transient, not a finding.

And the negative control, because a status check is worthless if a miss returns
200 anyway (`bugs_open/132` says B2 sites serve a JSON blob for a miss — the
question is what STATUS rides with it):

```
https://webdesign.co.uk/tools/head-architect/definitely-not-here.js   404  application/json
https://robot-hands.com/tools/assets/definitely-not-here.js           404  application/json
https://vonc.com/tools/assets/definitely-not-here.js                  404  application/json
https://loanandmortgagecalculator.co.uk/assets/js/definitely-not-here.js 404 application/json
```

A miss is an honest 404, so a status check discriminates. It just has nothing to
find today. **State this plainly wherever the fix is described**: this is a
regression guard for a class that has bitten (`bugs_closed/041`), not a repair of
live damage.

## 2026-08-05 — the finding that changes the SCOPE: candidate 1 is already recorded as RFC-scope

084's fix candidate 1 is *"make `asset_loads` actually load"*. That has already
been argued and ruled on, in a council round I would not have found without
grepping the mechanism rather than the bug number.

`docs/agent_docs/docs024_key_docs_latest/experience_register/harvest/entries/CC-001_feed-driven-teaser-list.json:255`
(round 4, 2026-07-29):

> "My reason for deferring `feed_loads` described an implementation quirk as
> though it were an inherent limit; **the true reason is that fixing it would
> change what `asset_loads` means for 63 live documents, which the owner's
> 2026-07-29 ruling makes RFC-scope.**"

and `NOTES_experience_register.md:1064`:

> "63 `doc_plans` fences already use `asset_loads`, so making it fetch changes
> what the type MEANS for all of them — RFC-scope under the ruling made hours
> earlier."

The cost of that ruling is recorded too: a legitimate clause — "the feed is
actually fetched" — had to be **re-typed to Tier 4** and now runs nowhere,
because nothing drives a browser for the register yet
(`NOTES_experience_register.md:978`). The platform gave up a capability rather
than change the meaning of a shared type inside a bug patch. That is the correct
call and it binds me too.

> **The 63 figure has decayed and I am correcting it rather than repeating it.**
> Measured 2026-08-05 against the live DB:
>
> | population | plans | `asset_loads` occurrences |
> |---|---|---|
> | `doc_plans WHERE is_current` | **6** | **8** |
> | all plan versions (history) | 66 | 153 |
>
> So the *live* blast radius is 8 checks in 6 current plans, not 63 documents —
> the 63 counted every superseded version. This does not overturn the ruling:
> RFC-scope under the 2026-07-29 ruling is triggered by a change to what the
> shared mechanism GUARANTEES, not by a headcount. But a future RFC should argue
> from 8, not from 63.

## 2026-08-06 — LIVE and ENABLED on v1.0.1257, and the honest gap that remains

The roll landed. Verified on the running pod **before** touching config, because
an unregistered check name fails the whole discovery step:

```
asset_reference_404      13   MINE
agentchassis-discovery    1   MINE (the probe UA)
unresolvable_reference    1   MINE (the finding kind)
asset_reference_405       0   NEGATIVE control
image_url_404             7   pre-existing positive control
```

Both replicas on `v1.0.1257`, started 09:52Z; the exec ran at 10:00Z, so well
past the ~300s post-restart window in which a dispatch is silently dropped.

Then `agent_definitions`: `design-discovery-agent` 22 → 23 checks. The UPDATE ran
inside a transaction whose verify block is a `DO ... RAISE EXCEPTION`, not a
`SELECT` — **a verify block made of `SELECT`s cannot stop a `COMMIT`**, which is
a trap already recorded on this estate. Fixture updated in the same commit
(`42e117c5e`).

**Found while enabling, and not mine: the fixture was under-asserting by one.**
`literal_markdown` (bugs_open/184's lane) was live on `quality-discovery-agent`
and absent from `liveConfiguredChecks`. It resolves, so there was no production
risk — but a roster that drifts silently is the exact defect that file exists to
prevent. Added, and named as not-mine in the commit rather than slipped in.

> **THE GAP, STATED PLAINLY: enabled is not run.** The check is in the binary and
> in the config, and it has still **never executed**. `improvement-sweep` is
> `enabled=f` (IMP-016) and nothing else drives design discovery, so the only way
> to make it run today is `294_TRIGGER_improvement_loop_v1.sh` — which fires the
> FULL loop, `discovery → triage_findings → call_dispatch`, and dispatches real
> content fixers at a real customer site. That is an outward-facing action for
> the sake of a verification, so it is being asked about rather than done.
>
> This is the same class of claim this whole lane is about. "Live" is a fact
> about the binary and the config. "Works" is a fact about a run, and I do not
> have one. Anyone reading this file later should not upgrade the first into the
> second — and a future clean sweep is *still* not proof, because the population
> is zero and a silently broken check reports exactly that.
