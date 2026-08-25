# HANDOFF — ai-agent-orchestration.com. START HERE. Written 2026-08-25 ~12:15Z.

**Supersedes `HANDOFF_2026-08-22_continue_here.md`.** That file's state table is still accurate for
contrast, carousels and images — all three remain done and were re-verified today. What it does not
contain is the incident below, which this lane caused.

> ## ✅ The original ask is COMPLETE and intact. ⚠ One incident, caused by this lane, now fixed at source.
>
> | thing | state (verified 2026-08-25) |
> |---|---|
> | **contrast** | **0 firm failures** on index / about / pricing / services. Unchanged |
> | **carousels** | Live on index + enterprise-reference-deployment. Opt-in; other 2 sites still OFF |
> | **images** | 10/10 card images serving 200 |
> | **`NNN+` on model-directory** | ⚠ **CAUSED BY MY MIGRATION `557`.** Fixed at source by `611` (theirs) + `613` (mine). Clears from the live page at the next 6-hourly publish |
> | **"3 pages are 404"** (08-24 CONTRIB) | **REFUTED.** All serve 200 at `/<name>.html`; the filer curled the extensionless form, which 404s for every page by hosting design |
>
> **Nothing is blocked and nothing is urgent.**

---

## 1. The incident, because it is the useful part

`557` (mine, 08-22) told the writer: *"take the live value from the `aao-agent-definitions` fact.
Phrase it as **"NNN+ AI agents"**."* Two errors compounding:

1. **`NNN` is a stand-in with no substitution machinery behind it.**
2. **The writer is never given the facts list** — its prompt on the unscoped path contains
   `writer_block` and not the values.

So it printed what it was shown. **137** instructed calls since 08-22, **14** copied `NNN`
verbatim, **0** wrote the value; zero `NNN` in any writer response before 08-22. One reached the
public.

⚠ **I had verified `557` at the artefact and still missed it, because I verified the thing I
CHANGED and never asked what the writer would DO with the replacement.** The check is one query:

```sql
SELECT prompt FROM llm_call_log WHERE agent_type='page-content-writer' ORDER BY created_at DESC LIMIT 1;
```

⚠ **The placeholder lives in `content_data`, not just `rendered_html`** — so a `template_changed`
rerender does NOT clear it. Only a copy regeneration does. Reaching for this lane's usual
propagation route would have looked like a fix and changed nothing.

**Transferable, and it is not "avoid placeholders":** prose written into a prompt is **input, not
documentation**. A human resolves "NNN" or "take the value from X" against context they already
have; a generative reader has only the bytes in the prompt, and its failure mode is to **render a
dangling pointer** rather than notice it dangles. WRONG_CALLS 2026-08-25.

## 2. What is fixed, and who owns which half

- **`611` (the `bugs_open/387` session, landed 11:20Z today)** — rewrote `writer_block` to name
  **rounded floors** ("more than 150 active agent definitions") instead of stand-ins, and bans
  letter stand-ins outright. It preserves every ban and the NOT-TRACKED list and carries `557`'s
  own history note forward. **It is a better block than mine — read it before touching this row.**
- **`613` (mine, today)** — the `writer_line` field, which `611` deliberately left and they flagged
  back. ⚠ **They reported two defects; censusing all seven writer_lines found FIVE**, in two
  classes: three frozen dates beside a live `{value}` (`aao-orchestrations` was not reported), and
  two lines that would publish figures `writer_block` explicitly forbids. **Acting on the report's
  stated scope would have left three of five in place.**

## 3. ⚠ `writer_block_managed` is STILL NOT SAFE — do not flip it

`composeWriterBlock` (`refresh_evidence_base_action.go:996`) builds the block from `writer_line`s
and `allowed_entities` and **nothing else**, so flipping the flag deletes both NEVER-write bans and
the whole NOT TRACKED / NEVER STATE list. `613` removed one precondition; it is not the last. The
`387` session has proposed the missing `writer_block_guidance` carry to the **`bugs_open/288`**
lane, which owns that file. **When it lands, flipping managed on is this lane's call** and it
retires `611`'s interim block.

## 4. Verified-today facts a fresh session should not re-derive

- All six checked pages serve **200** at `/<name>.html`. The extensionless form 404s **by design**
  (`scripts/cloudflare/worker.js:40-44`) — do not file that as a bug; it has been filed and refuted
  once already.
- `NNN` appears on **model-directory only** — 0 on adoption-tracker, protocol-tracker, index,
  about, pricing.
- The agent-count fact reads **200** today (175 on 07-26 → 196 on 08-19 → 199 on 08-22 → 200).
  **This is why no count is written in prose anywhere on this site any more.**
- Migrations `469`, `557`, `559`, `560`, `611`, `613` all in force; carousel flag on 2 placements;
  10 image URLs bound.

## 5. Open, none blocking

1. **Confirm `NNN` has left the live page.** `model-directory-publish` runs every 6h (last 06:26Z).
   `curl -s https://ai-agent-orchestration.com/model-directory.html | grep -c NNN` must be **0**.
   If it is not 0 several hours after 611/613, the regeneration is not picking up the new block and
   that is a real finding.
2. **The 17 parked `contrast_failure` items** — still `deferred`, untouched since 08-11, because
   the site's render audit has not run here since 08-10. Not evidence of a live defect. Now that
   all four pages measure 0 this site is a clean disconfirmable test for `bugs_open/296`.
3. **`composeWriterBlock` negative-guidance carry** — tracked by `bugs_open/288`, not this lane.
4. **Two `overImage` contrast findings** — approximate by the adapter's own admission and excluded
   from every figure this lane has quoted. A different, fuzzier problem.
5. **`bugs_open/364`** (clock times read as business claims) — fix is LIVE since `v1.0.1332`.

## 6. Practice notes earned this week

- **Ask the CAPABILITY, not the commit.** `service_binary_capabilities` (one row per pod carrying
  `git_commit`, 15-min heartbeat) then `git merge-base --is-ancestor`. Grepping the binary for your
  own commit returns ABSENT for a binary that certainly contains it — see
  `platform/buildcapability/buildcapability.go`; two lanes have been burned.
- **A peer lane's CONTRIB is another doc.** Verify at the artefact, and **census the class rather
  than the instances they listed** — twice this week the reported scope was smaller than the defect.
- **Hard-wrapped docs make `grep -F` report false absences.** Unwrap
  (`tr '\n' ' ' | tr -s ' '`) or, better, `git diff --numstat` — a byte-compare cannot be fooled by
  a line break or a stray `**`.
- **Your uncommitted work is not safe on this tree.** My WRONG_CALLS entry was swept into another
  session's commit (`001211abf`) before I committed it. Nothing was lost; verified by byte-compare.
