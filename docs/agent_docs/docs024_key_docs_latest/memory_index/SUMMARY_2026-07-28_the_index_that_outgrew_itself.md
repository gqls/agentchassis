# SUMMARY 2026-07-28 — the index that outgrew itself

*Written to be read aloud. First summary in this series; the design record it draws on is
`~/.claude/projects/-home-ant-projects-agentchassis/memory/memory-index-how-it-works.md`.*

---

## What we're trying to do

Every session that opens this repo gets a memory index loaded into it automatically,
before anyone asks for anything. It is meant to be the map: what work is live, what has
already been learned the hard way, and where to go for the detail. The whole value of it
is that it arrives **unbidden** — nobody has to know to go and look.

That is also its whole problem. Anything loaded into every session costs something in
every session, and there is a hard ceiling past which it stops loading at all. So the
index has to earn its size continuously, and it is written by thirty-odd concurrent
sessions who each have an excellent local reason to add one more line.

## Where we've come from

The index started as one file holding everything and grew until it strained. In late July
it was split once already, on **lifecycle**: bug cases that were fixed *and* live moved to
an on-demand archive, and everything else — including all workstreams, "even when done" —
stayed in the always-loaded file. That bought some room.

It did not hold, because the growth is not a one-off. It is the ordinary output of a busy
fleet: every closed bug wants a line, every new lane wants a line, every hard-won lesson
wants a line, and all of them are right. The file was hand-compacted **five times** in a
fortnight. Each compaction worked and each was overtaken within days.

By the morning of the 28th the tooling had begun nagging for a much smaller file, and the
only obvious lever was to move the durable practices — the ~38 lines of "here is how this
system bites you" — out to an on-demand file. The owner looked at that and refused it, in
writing, with the reason attached: **those lines are auto-loaded precisely because the
lessons most worth heeding are the ones nobody thinks to look up.** He accepted a larger
file instead and said to revisit when bugs actually closed, not when the nag fired.

That ruling turned out to be the most useful thing in the file, and not for the reason it
was written.

## What we've done

By the evening of the 28th the index had reached **24.0KB against a 24.4KB hard read
limit** — the point past which it does not load at all and every session loses its map.
That is a different and worse failure than the nag, and it was about four hundred bytes
away. Concurrent sessions had added roughly 2.4KB in a single evening.

So the question came back, and this time it was answered by measuring first rather than by
reaching for the obvious lever. The two candidate halves turned out to be **almost exactly
the same size** — practices 9.2KB, workstreams 8.7KB. Size could not decide it. What
decided it was asking a different question: *which of these survives being behind a link?*

- A **workstream** is looked up deliberately. You already know which lane you are on, and
  every one of those lines already ended in a pointer to its own cold-start document. Put
  it behind a link and almost nothing is lost, because the lookup was always intentional.
- A **practice** is not looked up at all. It has to arrive uninvited or it does not arrive.
  Behind a link it quietly becomes "read it if you happen to feel like it".

So we moved the workstreams — forty lanes — and left every practice line exactly where it
was. The owner's morning ruling was not overturned; it was the thing that made the right
answer visible, because it had already articulated why one half was load-bearing in a way
the other was not.

The split was checked rather than eyeballed: 104 entries in, 64 plus 40 out, **none lost,
none duplicated, none invented**, verified by set comparison against a copy taken before
the change. That check matters more than it sounds — a split that silently drops a line
looks exactly like a tidy one.

The index went from 24.0KB to **16.2KB**, under target, with the practices intact.

## Where we are now

The always-loaded file is 16.2KB and holds three things: the durable practices (8.9KB,
55%), the open-bug entries (4.2KB, 26%), and a small set of volatile status banners plus
the header (3.1KB). The workstreams sit in a companion file that loads on demand, behind
one pointer line. A third file still holds the closed bug cases from the July split.

The rule that came out of this is more general than the one it replaced. The old rule was
about lifecycle — open things stay, closed things go. The new one is about **retrieval
mode: move what is retrieved on purpose; keep what has to arrive unbidden.** It happens to
agree with the lifecycle rule for closed bugs, which is a good sign, and it disagrees with
it about workstreams, which is what we acted on.

**What is honest to say about the position:** we have bought room, not solved the problem.
At the rate observed this week — a couple of kilobytes in an evening — 16.2KB is days, not
weeks. And the easy half has now been spent: there is no third block of eight kilobytes
sitting there waiting to be moved somewhere cheaper. The next squeeze will have to come
out of the practices themselves, which is exactly the material we have twice now decided
is worth protecting.

## Where we're going

The next move is compression rather than relocation, and the measurement points at it
clearly: the practice lines have a **median length of 200 characters** and run as long as
427. Capping them at roughly 120 — the imperative and the pointer, with the reasoning left
in the topic file each line already links to — would recover about **4.4KB without moving
anything out of the auto-loaded file.** That keeps the property we have twice chosen to
protect, and it is the last large lever that does.

Beyond that the real question is whether to keep hand-maintaining a shared file at all.
Six compactions in a fortnight is not a maintenance history, it is a design signal: thirty
sessions are editing one document with no schema, no size budget, and no way to say no at
the moment a line is added. The structural answer is to **generate** the index from the
topic files — each one declaring its own single-line summary and whether it belongs in the
always-loaded tier — with a budget enforced when it is built rather than discovered when
it breaks. That is real work and it is not urgent today, but it is the thing that would
stop us having this conversation a seventh time.

The immediate decision for the owner is only the first of those: whether to spend the
4.4KB compression now, or wait until the ceiling is close again. Waiting is defensible.
Doing it while the file is calm is easier than doing it in a hurry at 24KB.
