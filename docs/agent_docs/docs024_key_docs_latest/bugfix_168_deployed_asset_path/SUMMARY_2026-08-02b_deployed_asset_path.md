# SUMMARY — 2026-08-02b — live, approved, closed

Second summary of the day, and a new file rather than an edit of the first: the earlier one
said *"not live, still open, round 3 with the council"*, and all three of those changed. The
first is the record of what we believed at that point; this is where it landed.

Five parts: what we're trying to do · where we've come from · what we've done · where we are
now · where we're going.

---

## What we're trying to do

Make the platform agree with itself about where a generated image lives. The code that
**commits** the file into a site's repository and the code that **points a page at it** have
to produce the same path. When they don't, you get a file nobody references or a reference to
a missing file — and each half looks correct on its own.

## Where we've come from

Three lanes had already met this defect and each contained it locally: `142` declared a map of
the two filenames that can't be derived and left a tripwire test naming its own remedy; `128`
nearly shipped a check that would have reported a broken favicon and social card on **every
site in the fleet**, caught itself, and patched its own call site; and a council seat objected
that the residual had been deferred as "its own item" and never filed — which is how
`bugs_open/168` came to exist at all.

## What we've done

**Fixed the class rather than the case.** `storage.DeployedAssetPath` is now the single
derivation, called by the writer *and* all six readers. Before, the rule existed twice — once
in the deploying code, once in the lookup — held in step by a comment claiming they matched.
They did match. Nothing *made* them.

**Corrected the bug's own diagnosis before acting on it.** Its stated mechanism was too broad
and its fix candidate 2 would have *created* the drift it claimed to remove. Struck through on
the ticket with the reason, rather than quietly ignored.

**Took three council rounds, and each one earned its cost.** Round 1: one real code defect, and
three objections that were really *"you did the right thing and didn't show me"*. Round 2:
**gated at high, and I was wrong** — my change made a live social-card overwrite reachable via
eleven already-queued work items, which I had twice denied with measurements attached. Round 3:
**approved**, with two advisories that were still worth acting on — a "convention" I had
asserted without ever grepping for it (it existed nowhere but my own code), and a silent
fall-through that *my own round-1 fix* had introduced while closing a different one.

**Nine guards proven by deliberately breaking the code.** Four of those nine exist because a
reviewer disagreed with me and I went and checked instead of arguing.

## Where we are now

**Live on `v1.0.1229`, verified on both replicas, and closed.**

The verification used a **negative** control — a string the change *removed*, expected to read
zero — because a positive control only proves the grep works, not that your binary shipped.
Both replicas: negative 0, both new markers present. All 24 brand-head artefacts across 12
sites serve 200 after the roll. `bugs_open/168` → `bugs_closed/168`.

Two things worth saying plainly about how it got live. **Another session's build shipped it** —
builds come from committed HEAD, so the work rode out on someone else's tag bump. That is the
mechanism working as intended, and it is also the concrete proof of why a thread cannot "hold a
change back for review" here: the first commit went live *before* the council finished. And the
round-2 finding was caught with **no exposure window purely by luck** — the risky change and its
guard happened to still be unrolled together. Had the first commit rolled a few hours earlier,
that would have been a live incident rather than a review comment.

Two wrong calls of mine are in the fleet-wide log with the cheap check for each: measuring two
populations that could not answer the question, and a query in which **my own council
submissions became the evidence** I was measuring.

## Where we're going

Three things, none of them code this lane should write alone:

1. **An owner call:** the eleven stale queued items now *refuse* rather than clobber. Whether
   they should instead be re-pointed at re-derivation — the repair they actually want — is a
   data decision, and the code no longer depends on the answer.
2. **`bugs_open/179` finding A:** `deploy_path` still overrides the single derivation and is
   invisible to every reader. Measured empty across three populations *including* the standing
   queue that round 2 taught me to include — but two seats pressed it at medium, and measuring
   a hole empty is not closing it.
3. **`RFC_009`'s wider question:** the platform reconstructs an artefact's identity from its
   metadata instead of reading what the writer recorded. That is the shared root of `152`,
   `155` and `179`, and it wants designing *with* those lanes rather than inside a path helper.
