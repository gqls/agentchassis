# Consolidation instructions (stage 1, step C1/C2)

You are one of 19 parallel consolidator agents merging the raw concept extraction
into the final register. Every extraction agent already produced structured
concept blocks (`### Name` + category/status/what/sources/relations/verify-later
fields); those blocks have been mechanically split and bucketed by category into
a single **cluster input file** assigned to you. You did not do the extraction —
treat the blocks as raw material to merge, not to re-derive.

## Your job

1. **Read your assigned cluster input file in full.** It contains every raw
   concept block for 1-15 categories, each block preceded by a
   `<!-- SOURCE: UNN_slug.md -->` comment naming which extraction unit it came from.

2. **For each category you were assigned** (see your prompt for the exact
   category → register filename → ID prefix mapping):
   - Group raw blocks that describe **the same real-world concept** — this is
     the core task. Duplication is expected and heavy: the same mechanism was
     often independently extracted by 2-4 different units (e.g. a live-tree unit
     and an archive-tree unit covering overlapping material, or a cross-cutting
     FOCUS doc and a domain-specific unit both touching the same bug). Merge
     duplicates into ONE entry.
   - Recognise duplicates by what they describe, not by exact name match — two
     blocks titled differently ("Snapshot-shadowing defect" vs "version+1000
     snapshot bug") are the same concept if the mechanism/evidence is the same.
   - When merging: combine `sources` (keep the best 4-8, prefer ones with a
     specific file#section or dated evidence over vague ones), combine
     `relations` (dedupe), combine `verify-later` (dedupe), and reconcile
     `status-signal` — if sources disagree (e.g. one says "aspirational", a
     later-dated one says "deployed"), prefer the most recent/most specific
     dated evidence and say so in status-evidence. If genuinely still
     unresolved, keep the most cautious status and note the disagreement.
   - Keep the richest `what` description (2-4 sentences) rather than
     concatenating all versions — synthesize, don't just paste both.
   - It is fine and expected for many concepts to have NO duplicates — carry
     those through as single entries, don't force merges that aren't real.

3. **Assign each final (post-merge) concept a stable ID**: `<PREFIX>-NNN` using
   the prefix given for that category in your prompt, numbered from 001 in the
   order you write them (order doesn't matter — just don't reuse a number).

4. **Write one register file per category** to
   `/home/ant/projects/agentchassis/docs/agent_docs/docs026_concept_register/register/<register-filename>.md`
   using this format per entry:

   ```markdown
   ### <PREFIX>-NNN — <Concept name>
   - **status:** deployed | partial | aspirational | superseded | abandoned | unknown
   - **status-evidence:** <dated phrase/evidence justifying the signal>
   - **what:** <synthesized 2-4 sentence description>
   - **sources:** <merged list, best 4-8, format `path#section` or `path (unit UNN)`>
   - **relations:** <merged, deduped>
   - **verify-later:** <merged, deduped — code paths/DB tables/workflows to check in stage 2>
   ```

   Precede each file with a one-line header: `# Register — <category-slug>` and
   a count: `<N> concepts, consolidated from <M> raw extractions across units
   <list the UNN unit tags you saw in the SOURCE comments for this category>.`

5. **Merging categories.** If, having read the material, two of YOUR OWN assigned
   categories are clearly the same thing under different names (this does
   happen — the taxonomy was seeded loosely on purpose), you may merge them into
   a single register file. If you do: write only the surviving filename, note
   the merge explicitly in that file's header ("absorbed <other-slug>"), and
   say so clearly in your final report so the index build accounts for it.
   Do NOT invent a brand-new category outside your assigned list — if material
   doesn't fit anywhere you were given, put it in the closest assigned category
   and note the mismatch in that entry's `relations` field instead.

6. **Write an index fragment** to
   `/home/ant/projects/agentchassis/docs/agent_docs/docs026_concept_register/register/.index_fragments/<your-cluster-name>.md`
   — one markdown table row per final concept, across ALL your categories:
   `| <PREFIX>-NNN | <name> | <status> | <one-line summary, <120 chars> | <register-filename>.md |`
   No header row needed (I'll concatenate all fragments and add one header).

## What NOT to do

- Don't re-read the original `extractions/*.md` files — everything you need is
  in your cluster input file.
- Don't fabricate sources or evidence — only use what's in the raw blocks.
- Don't drop concepts because they seem minor — the register's job is
  completeness; a one-line abandoned idea is still worth an entry.
- Don't spend time perfecting prose — synthesis quality matters more than
  polish. This is a working register, not a publication.

## Report back (final message, not a file)

- Per category: raw block count in → final concept count out (shows dedup rate)
- Total categories written, any merges you made
- 5-10 most significant / highest-value concepts across your whole cluster
