# 430 — forking a tool component drops `js_content`; one live page serves a 404 asset, and the naive fix would introduce a new bug class

**Filed:** 2026-09-02, from the "themes" session, while researching how JS snippets should
fit into a new theme-kit registry (`docs/agent_docs/docs024_key_docs_latest` theme system
plan). Not itself part of that build — filed standalone because it turned out to be live.

**090 diagnosis substitute** (per the 2026-07-31 owner ruling): this is local and
self-evidencing — a single INSERT's column list, read directly, plus first-hand DB and
production verification below (not inferred, not delegated). No 090 run.

**Credit**: found collaboratively with two peer sessions on this same tree —
`webdesign-tool-rebuilds` (platform seat) first flagged the code comment and ran an
initial census; `webdesign-tool-rebuild` (grind seat) corrected that census and found the
scoping-mismatch risk in §4 below, which is the reason this is NOT being fixed in this
filing. All DB queries and the production curl in §2 were run independently by the filing
session against live state, not taken on either peer's word — see §3 for where the two
peers' own counts needed correcting.

## 1. The defect

`platform/orchestration/actions/deploy_tool_action.go:332-361`, the fork-on-deploy INSERT
(`INSERT INTO content_components (...) SELECT ... FROM content_components WHERE id = $3`,
run when a library tool is first deployed to a site) omits `js_content` from both the
column list and the SELECT list. The code's own comment (lines 326-331) already names
this: *"js_content is also not copied (the known landmine — see 019 correction); harmless
while tools stay inline."* That qualifier is false today for at least one live case.

## 2. Live evidence (verified directly against the DB and production, 2026-09-02)

```sql
SELECT id, name, function, forked_from, created_from, source_agent_type, is_active,
       length(coalesce(js_content,'')) AS js_len,
       (html_template LIKE '%tools/assets/%.js%') AS refs_external_js
FROM content_components
WHERE function ILIKE '%provocation-heat-rater%' OR function ILIKE '%equity-release%'
ORDER BY function, forked_from NULLS FIRST;
```

| row | js_len | refs external asset | forked_from | linked to a live page? |
|---|---|---|---|---|
| `tool-provocation-heat-rater-vonc-com` (parent) | 2614 | yes | — | — |
| `tool-provocation-heat-rater-vonc-com-vonc-com` (fork, `2c7f7c67`) | **0** | yes | parent above | **yes** — `page_id decb69b3`, `/tools/provocation-heat-rater/index.html`, `vonc.com` |
| `tool-equity-release_pre_037` (parent) | 1950 | yes | — | — |
| `tool-equity-release_pre_037-mortgagecalculator-co-uk` (fork, `befacff0`) | **0** | yes | parent above | no `page_components` row — not deployed |

Production confirmation, run before either peer reply landed:
```
curl -o /dev/null -w '%{http_code}' https://vonc.com/tools/assets/tool-provocation-heat-rater.js
→ 404
curl -o /dev/null -w '%{http_code}' https://mortgagecalculator.co.uk/tools/assets/tool-equity-release.js
→ 200   (served — but from the FORK's own row, which has js_content='' — see §3 landmine)
```

**One confirmed live production defect**: `vonc.com/tools/provocation-heat-rater/` is a
served page whose interactive JS asset 404s, because the fork that serves it
(`2c7f7c67`) lost `js_content` on fork and `collectJSAssets`
(`rerender_single_page_action.go:338+`) publishes straight from that row's own column,
never the parent's.

## 3. A landmine inside this bug: a loose `LIKE '%<script%'` check misreads the failure

Both peer sessions' first-pass censuses, and the filing session's own first attempt,
used some form of `html_template LIKE '%<script%'` to ask "does this fork still have its
JS inline, so is it actually unaffected?" **That check cannot distinguish a real inline
`<script>` body from the `<script src="/tools/assets/{function}.js">` REFERENCE tag that
`separateInlineJS` itself splices into `html_template` at extraction time** — both start
with `<script`. A fork whose `js_content` is empty but whose `html_template` still carries
that spliced reference tag reads as "has inline script, not affected" on the loose check,
when it is exactly the broken case. This is why the two peer censuses disagreed with each
other (7 vs 1 "genuinely affected" rows) and why the filing session's own direct check
above used `html_template LIKE '%tools/assets/%.js%'` (does it reference the external
asset) plus `length(js_content)` (is the asset it references actually populated) instead
— those two signals together are what actually distinguishes broken from fine, not the
presence of any `<script` substring.

**Anyone re-running this census: don't reuse the loose check.** The corrected query, one
row per genuinely-broken fork fleet-wide, as of 2026-09-02:
```sql
SELECT f.id, f.name, f.function
FROM content_components f
JOIN content_components p ON p.id = f.forked_from
WHERE f.component_level = 'tool'
  AND coalesce(p.js_content,'') <> ''
  AND coalesce(f.js_content,'') = ''
  AND f.html_template LIKE '%tools/assets/%.js%';
```
This returns exactly the two rows in §2's table. Whether one is "live damage" still
requires the `page_components` join in §2 — a row matching this query with no linked page
is latent, not live.

## 4. Why this is NOT being fixed in this filing — the scoping mismatch

Credit: `webdesign-tool-rebuild` (grind seat), verified by the filing session.

The fork's `html_template` is not a raw copy of the parent's — it is `$6` in the INSERT,
i.e. `scopedToolHTML`, the parent's template passed through
`ConvertTemplateToInstanceScope`, which rewrites element ids (`id="x"` →
`id="c-<function>-x"`) so two instances of the same component don't collide
(`bugs_closed/283`). `js_content`, by contrast, is published **verbatim** by
`collectJSAssets` — only `StripToolDocHeader` is applied, no id-rewriting.

**A naive fix (add `js_content` to the fork's column list) would pair scoped HTML with
unscoped JS** — the exact id-mismatch failure class `bugs_closed/283` exists to prevent,
and one that has never occurred in production before, because no component today is both
scoped (i.e., has been forked) and separated (i.e., has non-empty `js_content`) at once.
This fix would create the first ones. If the JS references element ids by their original
(unscoped) form — plausible, not yet checked — every such fork would break in a new way
even after "fixing" the drop.

**This is a design question, not a known-good fix**, and it needs an answer before code
changes: either (a) apply the same instance-scoping transform to `js_content` that
`ConvertTemplateToInstanceScope` already applies to `html_template`, so both stay
consistent, or (b) confirm the specific components affected don't reference scoped ids in
their JS and a plain copy is safe for those. Do not ship a bare column-list fix without
resolving this.

**Not a new seam — same mechanism as an already-diagnosed defect.** Credit:
`webdesign-tool-rebuilds` (platform seat). `ConvertTemplateToInstanceScope`'s
instance-prefixing (`InstanceToken = "c-"+function`) is exactly what broke 110/112 of the
tool-health check's historical failures in `staged_component_build`'s own diagnosis
`91228c39` (corr `2b64e510`) — bare acceptance-criteria selectors there, verbatim
`js_content` here, same "anything bare can never match the rendered instance" mechanism.
**Route the scoping question to `staged_component_build` (active lane) rather than treat
it as fresh** — they already own the criteria-vs-renderer decision this bug's fix depends
on.

## 5. Scope note for anyone touching `deploy_tool_action.go`

`scripts/pattern-check.py`'s `check_unrepaired_component_write` deliberately does **not**
allow-list `deploy_tool_action.go` (LANDMINES.md ~3092) — that is a separate, unrelated,
open debt (can an LLM-authored string reach `page_components.rendered_html` unrepaired
through this file). If a future edit here trips that check, it is a true positive on that
separate issue; do not silence it by adding this file to `COMPONENT_WRITE_ALLOWED` as a
side effect of fixing this bug.

`platform/` is council-review scope — any code change here should go through the council
gate before or alongside committing, per CLAUDE.md.

## 6. Fix candidates, ranked

1. **Resolve the scoping question first** (§4), then extend the fork INSERT's column
   list — the actual code change is still small once the scoping approach is decided.
2. **Backfill**: the two existing broken forks (§2) are a separate, smaller follow-up —
   fixing the fork path going forward does not repair rows already forked. `befacff0`
   (latent) can wait; `2c7f7c67` (live, 404ing right now) is the one worth a manual
   backfill (copy the parent's current `js_content`, scoped correctly per whatever §4
   resolves to) independent of when the general fix lands.
3. Not recommended: leave as-is. `store_generated_component_action.go:242` separates
   inline `<script>` into `js_content` on every store, so the set of parents with
   separated JS — and therefore the set of forks that would lose it — grows by addition,
   not a fixed, bounded set. [MEASURED 2026-09-02, webdesign-tool-rebuild]: 7 of 261
   active tool components currently carry non-empty `js_content`; that number moves.

## 7. How to verify a fix

- Re-run §3's corrected query — a fixed fork path should never let it return a new row.
- For the backfill: confirm `vonc.com/tools/assets/tool-provocation-heat-rater.js` stops
  404ing, and that the served page's calculator actually functions (not just that the
  asset 200s — a scoped/unscoped id mismatch could serve 200 with broken behaviour).
- Before closing: confirm no component that is both forked (scoped html) and separated
  (non-empty `js_content`) still ships with unscoped JS.
