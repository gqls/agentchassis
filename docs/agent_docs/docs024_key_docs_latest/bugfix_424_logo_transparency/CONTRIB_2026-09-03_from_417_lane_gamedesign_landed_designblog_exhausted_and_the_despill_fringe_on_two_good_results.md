# CONTRIB from the 417 lane — 2026-09-03, ~12:35Z

Your handoff's "Live status" section says the three reset runs were still landing and asks the next
reader to check rather than trust what was printed. I checked, for my own lane's reasons (417 needs
to eye-check each new artefact for lettering), so here is everything I read. **All of it is at the
served bytes with a 404 invented-path control, not at the DB status.**

Nothing here needs anything from me — it is yours to fold in or discard.

---

## 1. `gamedesign.uk` — SUCCESS, and it took the full ladder

Completed **11:41:09Z on attempt 3 of 3**. Verified genuinely new by the storage key's date
directory (`20260903/`), not by `updated_at`.

| | before (pre-fix) | after |
|---|---|---|
| md5 (first 12) | `b4f0ed1091f9` | `01076df06e90` |
| bytes | 298,938 | **76,830** |
| dimensions / mode | 400×400 | 400×400 RGBA |
| fully transparent | **0.0%** | **62.0%** |

Magic bytes `89504e47`, colour type RGBA. **So two of your three resets landed** (seotools,
gamedesign).

## 2. `designblog.co.uk` — FAILED, all three attempts, and your decision #3 is now concrete

Item error, verbatim:

> `step call_imagery_gen failed: … bugs_open/424: model did not honour the key-colour instruction
> (border_keyed=0.000, want >= 0.95) — refusing to store; source_format=jpeg`

**This is your guard working correctly** — three times it looked at the output, judged the ground
had not been keyed, and stored nothing. The site still serves its pre-fix logo and the item is now
terminal.

Your handoff item #3 asks whether logo generation needs a longer leash if a site exhausts
`max_attempts=3`. **It has now happened, to a real site.** The distribution across the four runs I
can see is the useful part, and it is not reassuring:

| site | attempts to land a good result |
|---|---|
| seotools.co.uk | 2 of 3 |
| gamedesign.uk | **3 of 3** |
| websitepromotion.co.uk (in flight, my lane's) | attempt 1 refused, 2 pending |
| designblog.co.uk | **exhausted, nothing stored** |

⚠ **I confirmed this was a refusal and not a run killed by the 12:06:47Z chassis roll.** The two are
byte-for-byte identical in `status` / `attempt_count` / `completed_at`, and the roll landed the same
morning. The refusal is timestamped **11:36:58Z, before the pod start**, and carries your guard's own
statistic, which a killed run cannot fabricate. Worth knowing because reading it the other way would
have put a false failure against your fix. Written up as a landmine (`LANDMINES.md`, footprint
`site_work_items.error`).

## 3. Your open item #4 — the despill fringe on a genuinely good result

Magenta-ish opaque pixels as a fraction of the whole image (`r>150, b>150, g < both−30`):

| artefact | magenta | note |
|---|---|---|
| `gamedesign.uk` (post-fix) | **0.01%** | 8 pixels |
| `seotools.co.uk` (post-fix) | **0.05%** | 42 pixels |
| `websitepromotion.co.uk` (pre-fix) | **0.62%** | 542 pixels |

**An order of magnitude better on both post-fix artefacts.** On gamedesign it is 8 pixels in a
160,000-pixel image, which I would call diagnosed-and-not-worth-fixing unless you disagree.

## 4. ⚠ A hypothesis I formed about your guard and then REFUTED — do not chase it

Looking at gamedesign's maze, the white regions **inside** the mark looked opaque to me. I reasoned
that `BorderKeyed` measures the outermost ring only, so an enclosed region of ground would survive
the matte and pass the guard **by design** — a real structural gap, and I was about to send it to you
as one.

**I measured it first and it is false.** Near-white opaque pixels: **gamedesign 0, seotools 0**,
websitepromotion 41 (0.05%). Those interior areas are transparent; they look white because the page
behind them is white.

Recording it because the wrong version would have cost you an investigation, and because the general
form is worth having: **over a white page, opaque white and fully transparent are visually
identical**, so a look can generate a confident structural claim about your matte that the alpha
channel refutes in one command.

```bash
python3 -c "
from PIL import Image
px=list(Image.open('logo.png').convert('RGBA').getdata())
op=[p for p in px if p[3]>200]
print('near-white OPAQUE:', sum(1 for p in op if min(p[:3])>240), 'of', len(op))"
```

## 5. Not your bug, but adjacent, and now filed — `bugs_open/462`

Your lane scoped out mark legibility and handed it to mine; the owner has since approved filing it.
**`bugs_open/462`** — a logo can be perfectly rendered, correctly deployed and illegible, and no
check in the estate can see it.

The reason it matters to *you* is one line in its fix candidate 1: the natural remedy is a
fail-closed contrast statistic **at store time, beside your `BorderKeyed` guard, on the same retry
ladder**. Given §2 above, adding a second fail-closed statistic to a ladder that already exhausts is
a real interaction, not a hypothetical. **462 sharpens your decision #3 rather than competing with
it**, and I have said so in the bug rather than proposing anything unilaterally.

Also relevant to your threshold-constants item: these runs are the first real population for
`minBorderKeyed=0.95`, and **every refusal is `0.000` — not merely below threshold, but exactly
zero.** Read first-hand from `site_work_items.error` for two of them (designblog's terminal attempt,
websitepromotion's attempt 1); your own handoff reports the same `0.000` for seotools' and the other
first attempts.
⚠ Stated narrowly on purpose: the `error` column holds only the **last** attempt's message, so
earlier attempts on the same item are not readable there and I am not claiming to have seen them.

The split is bimodal with nothing anywhere near 0.95, which suggests **the constant is not currently
the binding choice** — the model either keys the ground or does not key it at all. Worth recording
before anyone tunes it, because a threshold that nothing lands near cannot be validated by the runs
that pass it, and lowering it would not have saved a single one of these refusals.
