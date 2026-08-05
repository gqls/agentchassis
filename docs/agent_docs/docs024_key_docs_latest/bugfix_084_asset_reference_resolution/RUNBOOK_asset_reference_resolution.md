# RUNBOOK — bugs_open/084, asset reference resolution

Every command here was hard to get right once. The gotcha is attached to the
command, not kept in a scrollback.

## Is anyone else on this bug?

`scripts/who-owns.py <n>` reads COMMITS, so it is blind to a session mid-fix and
on this tree it returns "OWNED or recently active" for almost everything. What
actually discriminates is the live transcripts:

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -name '*.jsonl' -mmin -180 -size +10k); do
  echo "=== $f ($(stat -c %y "$f" | cut -c1-16)) ==="
  tail -c 400000 "$f" | grep -oE 'bugs_open/[0-9]{3}' | sort -u | tr '\n' ' '; echo
done
```

**Gotcha:** `tail -c` on a file being appended to by another process panics
(`uutils` coreutils, `Invalid argument`) — that is the file growing under you,
not a bad command. Skip those and re-run; the sessions you cannot read are the
ones actively writing, so treat their bugs as claimed.

**Gotcha 2:** one session in the sweep mentioned ~40 of the 45 open bugs. It was
`ls bugs_open/` output, not ownership. Discount any session whose hit list is
almost the whole directory.

## Measure the real population — do NOT regex rendered_html

The wrong way, and it produced a phantom finding (see NOTES):

```sql
-- WRONG: matches a <script src> mentioned inside a JS comment
LATERAL regexp_matches(pc.rendered_html, '<script[^>]+src="([^"]+)"','g')
```

The right way — fetch the served pages and parse the DOM. Scripts live in
`scratchpad/realsweep.py` + `resolve.py`; the shape is:

```python
from html.parser import HTMLParser      # sees ELEMENTS; a mention is unreachable
# <script src>  and  <link rel~=stylesheet href>
# resolve each with urllib.parse.urljoin(page_url, src) — never build the URL
```

**Gotcha:** `curl` writing `000` is a *connection* error, not a status. Re-probe
before tallying — `fundamentallyai.com/assets/css/styles.css` gave `000` once and
`200` on retry, and counting it would have been a fabricated finding.

Page list to feed it:

```sql
SELECT s.domain, p.url, p.rebuild_policy
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.deployed_at IS NOT NULL AND COALESCE(s.domain,'') <> ''
ORDER BY 1,2;
```

## The negative control, which the measurement is worthless without

A status check only discriminates if a MISS is not a 200. On this fleet it is an
honest 404 with an `application/json` B2 error body:

```bash
for u in https://webdesign.co.uk/tools/head-architect/definitely-not-here.js \
         https://robot-hands.com/tools/assets/definitely-not-here.js \
         https://vonc.com/tools/assets/definitely-not-here.js; do
  curl -s -o /dev/null -w "%{http_code} %{content_type} $u\n" --max-time 20 "$u"
done
```

## Tests, and proving the guards are load-bearing

```bash
go test ./platform/orchestration/actions/discovery_checks/ -run AssetReference404 -v
go test ./platform/orchestration/actions/discovery_checks/          # the coverage guards
go test ./platform/orchestration/actions/ -run DiscoveryCheck       # the registration fixture
gofmt -l platform/orchestration/actions/discovery_checks/
python3 scripts/pattern-check.py
```

**`gofmt -l` lists files that are NOT formatted** — an empty line for your file is
the pass. Two files in that directory (`check_image_url_404.go`,
`check_misdirected_cta_test.go`) were already un-gofmt'd on arrival and are
another session's; do not sweep them into your commit.

The mutation harness (run it again if you change a guard — a guard no test can be
made to fail against is decoration):

```bash
F=platform/orchestration/actions/discovery_checks/check_asset_reference_404.go
cp "$F" /tmp/orig.go
# ...apply one mutation, run the -run AssetReference404 suite, expect a NAMED failure...
cp /tmp/orig.go "$F"; diff -q /tmp/orig.go "$F"   # prove you restored it
```

Recorded results, 2026-08-05 — each mutation caught by a named test:

| mutation | caught by |
|---|---|
| delete the 404-confirmation second probe | `CandidateNotFoundNotReproducedIsDiscarded` (+2) |
| treat any 4xx/5xx as a finding | `InconclusiveStatusesFileNothing` |
| dedup key on basename only | `SameBasenameDifferentDirectoriesAreTwoItems` (+1) |
| resolve relative refs at the site root | `RelativeReferenceResolvesAgainstThePage` (+2) |
| widen the selector to `<img>` | `ImgAndIconAreNotOurs` |

## Enablement — the ordering is not negotiable

An unregistered check name **fails the step** (`bugs_open/149` B4, chassis
v1.0.1211). So:

1. **Commit, then build.** `make build-<service>` archives committed `HEAD`; an
   uncommitted change is silently left out and you get an image missing your work.
2. **Bump `IMAGE_TAG`** (makefile ~line 16) — a same-tag rebuild ships the node's
   stale cached binary.
3. **Pod-grep the running pod, with both controls.** A roll is not evidence your
   fix shipped; the image may predate your commit.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec "$POD" -- sh -c '
  strings /app/agent-chassis | grep -c asset_reference_404;      # POSITIVE: expect > 0
  strings /app/agent-chassis | grep -c agentchassis-discovery;   # POSITIVE: the UA this adds
  strings /app/agent-chassis | grep -c asset_reference_405;      # NEGATIVE control: expect 0
  strings /app/agent-chassis | grep -c image_url_404'            # pre-existing: expect > 0
```

Do all four in the SAME exec. The negative control is what proves you tested the
pipeline rather than your own spelling.

4. **Only then** add the name to config — and **update the fixture in the same
   commit**, because `liveConfiguredChecks` in
   `platform/orchestration/actions/discovery_checks_registration_test.go` pins
   what the live agents are configured with. Adding the check to
   `design-discovery-agent` without adding it there leaves the fixture asserting a
   stale roster.

```sql
-- HOLD until the pod-grep above passes. design-discovery-agent, beside image_url_404.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,run_checks,config,checks}',
         (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
           || '["asset_reference_404"]'::jsonb)
 WHERE type = 'design-discovery-agent'
   AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
   AND NOT (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks' @> '"asset_reference_404"');
```

**Gotcha:** `workflow.steps` is an OBJECT keyed by step name, not an array.
`jsonb_array_elements` errors with *"cannot extract elements from an object"* and
`jsonb_path_query('$.workflow.steps[*]')` returns nothing at all — a silent zero,
which is worse. Enumerate with `jsonb_each`:

```sql
SELECT ad.type, s.key, s.value->'config'->'checks'
FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND s.value->>'action' = 'run_discovery_checks';
```

**Gotcha:** `pages` has no `slug` column — the served path is `pages.url`.

## Proving it bites in production, with no live positive available

There is no broken reference on the fleet, so the check cannot be shown to work by
waiting. Induce one, on a page you own, and revert:

1. Pick a throwaway page and add a `<script src="/does-not-exist-084.js">` to its
   rendered component.
2. Run the discovery sweep for that site; assert exactly one
   `asset_reference_404` item whose `item_key` carries the resolved URL.
3. Remove the reference, re-render, re-run; assert the item retracts through
   `CheckResult.Resolved` — **and note the stated gap**: retraction fires on a
   still-referenced URL that now returns 200, so if you DELETE the reference
   instead of fixing the file, the item stays open by design.

```sql
SELECT item_key, status, severity, spec->>'http_status', spec->>'surface'
FROM site_work_items
WHERE item_type = 'asset_reference_404' ORDER BY created_at DESC LIMIT 10;
```
