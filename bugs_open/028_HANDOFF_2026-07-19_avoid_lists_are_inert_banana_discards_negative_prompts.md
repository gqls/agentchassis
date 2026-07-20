# HANDOFF — every `avoid` list in the fleet is inert: Banana discards the negative prompt, and all declared kinds route to Banana

**Filed:** 2026-07-19, from the imagery workstream, after 9 generated tool heroes on
gamesdesign.co.uk violated their own `avoid` list in exactly the ways it forbids.
**Severity:** Medium-high. Nothing crashes; the imagery style guide's entire negative
half has simply had no effect since the Banana migration, and at least one documented
"hard-won fact" was attributed to it.
**Status:** OPEN — **FIX APPLIED 2026-07-20, INERT UNTIL AN IMAGE ROLL.** Stays OPEN
because the bar for `/bugs_closed/` is fixed AND live: this is Go, so the defect is
still reproducible in prod until a chassis image ships. Fix + evidence in §7 below.

> **CONFIRMED 2026-07-20 (bugfix-028 thread), beyond the original filing.** The
> mechanism was re-verified from code independently, and then proven end-to-end
> against the live DB, which the original filing had not done:
> - **All 11** gamesdesign.co.uk `content_hero` rows generated 2026-07-19 carry
>   `origin_model = 'banana/gemini-3-pro-image-preview'`. Every sampled image was
>   served by the discarding provider — so this is no longer an inference from the
>   routing table.
> - `assets.origin_prompt` for `content_hero_tool_xp_curve_designer` contains the
>   medium, mood and palette and **not one term** of the site's 240-char avoid list.
>   Assembled, shipped over Kafka, gone.
> - `avoid` has exactly **three** references in the whole Go tree
>   (`imagery_style_guide.go:128` emptiness check, `:194`, `:196`), and its only
>   consumer path terminates in that Debug log. There is no second path.

---

## 1. The mechanism, verified in code

`internal/adapters/imagegenerator/banana/provider.go:18`:

> `// NegativePrompt on provider.Request is ignored here (Gemini has no`
> `// negative-prompt concept). Provider logs at debug level if one is`
> `// provided; callers shouldn't rely on it being honoured.`

and at line 105 it does exactly that — logs at **Debug** and drops it:

```go
if req.NegativePrompt != "" {
    p.logger.Debug("NegativePrompt provided but Banana ignores it", ...)
}
```

Meanwhile `avoid` has exactly one destination. `generate_image_actions.go:333`:

```go
if avoid := styleGuide.avoidForKind(kind); avoid != "" {
    negativePrompt += ", " + avoid          // ← the ONLY use of avoid
}
```

`directionForKind` composes the POSITIVE prompt from medium/mood/palette only. So for
any Banana-routed kind, `avoid` is **fully inert** — it is assembled, logged, shipped
over Kafka, and thrown away.

## 2. The blast radius is "all imagery"

`internal/adapters/imagegenerator/routing.go:59` — **every declared kind is on Banana**:

```go
var kindProviderRouting = map[string]string{
    "icon": banana, "logo": banana, "illustration": banana, "infographic": banana,
    "sprite_sheet": banana, "content_hero": banana, "hero": banana,
}
```

SDXL — the provider that *does* honour negative prompts — is now reached only by an
empty/legacy kind or an explicit per-site `provider:"stability"`. So:

- every `avoid` in every site's `imagery_style_guide`, and
- every `NegativePrompt` in `kindDefaults` (`generate_image_actions.go:51`, incl. the
  logo entry described in-code as "biggest expected win is logo getting any negative
  prompt at all"),

have no effect on essentially all imagery the platform now generates.

> ~~**Caveat on `hero`:** its Banana routing is `bugs_open/011` R1, **fixed in code but
> not yet live** at time of filing. So `hero` may still be reaching SDXL in prod, where
> its negative prompt does work.~~ Everything else has been Banana-routed and live since
> v1.0.1135 or earlier.
>
> > **CAVEAT RESOLVED 2026-07-20 (011 R1 thread) — `hero` IS now live on Banana, so this
> > bug's blast radius is complete.** Both services rolled to **v1.0.1139** (pods started
> > 07:35); verified against the running binaries, not the tag:
> > `strings /app/image-generator-adapter | grep -c "UNROUTED KIND"` → 1 and
> > `… grep -c "routed_kinds"` → 1 on `image-generator-adapter-764d758d5c-lmp5j`;
> > `strings /app/agent-chassis | grep -c "site provider preference applied"` → 1 on
> > `agent-chassis-645674b498-rndg9`. (Log-message strings, not `case` values — the
> > Docker build does not retain the latter, which reads exactly like a stale deploy.)
> >
> > So `hero` — **84 of 155 planned images, the fleet's largest kind** — has joined the
> > inert-`avoid` set, including its `kindDefaults["hero"].NegativePrompt` of
> > *"text, watermark, signature, low quality, blurry, distorted"*, which genuinely did
> > work on the SDXL path. **011 R1 did not cause this defect; it extended it to the
> > largest kind**, which raises the value of fix candidate 1 rather than changing its
> > shape. Note the interaction cuts both ways: that negative prompt existed largely to
> > suppress SDXL's *garbled* text, and Banana renders text legibly — so the risk profile
> > shifted rather than simply worsening.
> >
> > **NOT yet observed end-to-end:** zero assets have been generated since the roll
> > (`SELECT … FROM assets WHERE created_at > '2026-07-20 07:35'` → 0 rows), consistent
> > with the owner's tool-imagery HOLD. The routing is verified *in the binary*; no hero
> > has actually been generated through it. First real generation is the observation that
> > would close that gap — and per §6 it is also the cheapest test of candidate 1.
> >
> > One more thing 011 R1 arms, flagged for whoever takes candidate 1: the
> > `maxImageryDirectionInPrompt = 200` cap (`/bugs_open/027` §4b, already cited above)
> > carries an in-code note that it is sized for *"the only generation backend (Stability
> > hosted SDXL)"* and its 77-token CLIP wall, listing Banana at *"~1000+ char effective;
> > cap could be raised significantly"* — explicitly deferred *"until provider routing
> > lands"*. **Provider routing has now landed**, so that deferral has come due: the cap
> > is now calibrated for a provider no declared kind uses.

## 3. The live evidence that prompted this

Nine `content_hero` images generated on gamesdesign.co.uk, 2026-07-19, whose
`kinds.content_hero.avoid` explicitly lists *"text, lettering, words, numerals,
labels, captions … white background, pale background, bright full-bleed colour field"*:

| # | asset | violation |
|---|---|---|
| 1 | `drop_rate_simulator` | renders the numerals **"100 100.100"** |
| 3 | `ehp_calculator` | pale near-white diagonal bands |
| 7 | `progression_architect` | large pale region |
| 9 | `xp_curve_designer` | sits on a **near-white ground** |

**4 of 9 violate the list, in its own terms.** The other 5 comply — which is what an
ignored constraint looks like: compliance by luck, not by instruction.

## 4. A documented "hard-won fact" is very likely a misattribution

This is the expensive part, because it is written down in three places and has been
repeated as a lesson.

`HANDOFF_imagery_best_in_class.md` (D14 findings), the imagery memory, and RUNBOOK
**A6.5** all record:

> "**Style drift in the ground colour is fixed via the style guide's `avoid`, not its
> `medium`** — 'deep charcoal ground' in `medium` did not stop a white background;
> adding 'white background, pale background, light background' to `avoid` did."

Those were `content_hero` generations — **Banana-routed**, therefore with `avoid`
discarded. The `avoid` edit cannot have caused the improvement. What was observed was a
**re-roll that happened to come out darker**, and the change made alongside it took the
credit. The 4-of-9 white/pale grounds above are that supposed fix failing on its first
real test at n=9.

I am stating the code fact as **verified** and the misattribution as **strongly
implied** — I have not re-run the original D14 generations with and without the `avoid`
edit, which is the experiment that would settle it beyond doubt.

> **STRENGTHENED 2026-07-20 (bugfix-028 thread).** The premise is now proven rather than
> inferred: every one of the 11 gamesdesign `content_hero` rows carries
> `origin_model = 'banana/…'`, and the stored `origin_prompt` for one of them contains
> no avoid term at all. So on the D14 generations the `avoid` edit provably reached
> nothing. The misattribution is as close to settled as it gets without re-running D14.
>
> ~~**The three documents still say it, and this thread has not corrected them** —
> `HANDOFF_imagery_best_in_class.md` (D14 findings), the imagery memory, and RUNBOOK
> **A6.5**. That is deliberate scope, not an oversight: they belong to the imagery
> workstream.~~
>
> > **CORRECTED 2026-07-20, SAME DAY, BY THE AUTHOR OF THE LINE ABOVE — and the way
> > it was caught is the point.** That claim was false, and I made it without opening
> > any of the three files. All three were **already corrected before I started**:
> > RUNBOOK A6.5 carries a dated `> **CORRECTED 2026-07-19**` block written by the
> > filing thread; the imagery memory carries its own correction; and
> > `HANDOFF_imagery_best_in_class.md` had already been reframed to *"Unverified and
> > next: whether the Banana path sends `avoid` as a negative prompt at all"* — a
> > flagged open question, not an assertion.
> >
> > **What caught it:** the council gate's `editquality` seat filed it as a *missing*
> > item — "the plan does not correct the three artifacts that actually assert the
> > false causal claim" — which sent me to read the files for the first time. It was
> > right that the submission asserted it; it inherited my error.
> >
> > So this bug's own fix repeated this bug's own failure mode: **I asserted the state
> > of three documents from a filing written a day earlier, rather than checking, while
> > writing up a defect that exists because someone asserted a mechanism rather than
> > checking.** Confidence was no signal in either case. Recorded here rather than
> > quietly deleted because a correction with the catching mechanism named is worth
> > more than a clean file.
>
> The durable lesson is not merely "this fact was wrong" but **how it was made** — a
> config edit and a re-roll happened together, the output improved, and the edit took
> the credit. Nothing in the loop was capable of noticing that the edited field was
> never read. After an image roll the `avoid` edit genuinely *will* do something, which
> is precisely when a stale "we already know how this works" note is most expensive.

## 5. Fix candidates

1. **Fold `avoid` into the POSITIVE prompt for providers with no negative-prompt
   concept.** Gemini responds to plain instruction ("no text or lettering, no white or
   pale background"). The positive prompt is already how `content_hero` gets its "no
   text or lettering in the image" clause — which is *also* frequently ignored, so
   phrasing matters and wants testing, not assuming. Mind
   `maxImageryDirectionInPrompt = 200` (`/bugs_open/027` §4b) — appending avoid terms
   to a capped string will silently truncate something else.
2. **Make the drop loud.** `provider.go`'s Debug line is invisible in practice. A
   provider that discards a caller's constraint should say so at Warn, or the
   capability should be declared on the provider interface so the action layer can
   choose a strategy instead of shipping a field into a void.
3. **Do not "fix" this by routing back to SDXL** — SDXL was abandoned for good reasons
   (no `ReferenceImageURIs`, weak style adherence, illegible text). Losing brand
   anchoring to regain a negative prompt is a bad trade.

## 6. How to verify a fix

> **CORRECTED 2026-07-20 — the second bullet was wrong for the fix that was actually
> applied, and following it would produce a false negative.** It said to read
> `assets.origin_prompt` and expect the avoid terms there. `origin_prompt` is written
> by the **action layer** (from the workflow's `origin_prompt_field`; see
> `sql_for_agents/107_image_build_handler.sql`), and the applied fix folds the
> constraint **downstream of that**, inside the Banana provider. So `origin_prompt`
> will *never* show the avoid terms, fix or no fix — and a thread checking there would
> conclude the list is still inert. That is the same shape of mistake this bug is made
> of, which is exactly why it is corrected here rather than quietly dropped.
> Why the fix went downstream anyway, and what was traded away, is in §7.

- Generate one `content_hero` on gamesdesign.co.uk (its guide's `avoid` names white
  grounds and numerals) and check the produced image for both.
- Confirm the constraint reached the model **from the adapter log**, not the DB. The
  provider logs the fold at Info:
  ```
  kubectl logs -n ai-persona-system <image-generator-adapter-pod> \
    | grep "folded NegativePrompt into positive prompt"
  ```
  The line carries `negative_prompt` (400-char preview — wide enough for real ~350-char
  avoid lists), `prompt_len_before` and `prompt_len_after`. A generation with a
  non-empty avoid list and **no such line** means the fix is not in the running image.
- Check the image itself, not the log line: the log proves the terms were *sent*, which
  is all this fix claims. It does **not** prove Gemini obeyed them — see §7's honest
  caveat.
- n=1 proves nothing here. Both defects in this file were only visible across a set;
  use 5+ and count violations rather than eyeballing one.
- **Do not verify by grepping the pod binary for `avoid` or `NegativePrompt`** — both
  symbols were present throughout the defect. Grep for `foldNegativeIntoPrompt`.

## 7. THE FIX AS APPLIED (2026-07-20, bugfix-028 thread)

**Shape:** candidate (1) from §5 — fold `avoid` into the positive prompt — but placed
in the **Banana provider**, not the action layer, plus candidate (2)'s "make it loud".
Candidate (3)'s warning was respected: nothing was routed back to SDXL.

`banana/provider.go` now translates `NegativePrompt` into a trailing prohibition clause
(`foldNegativeIntoPrompt`) instead of dropping it, and logs the fold at **Info**. The
`provider.Request.NegativePrompt` interface contract was tightened too: it used to tell
implementers that providers without negative-prompt support "log and ignore" — Banana
was *following the contract correctly*, so fixing only the call site would have left the
next provider author the same licence. 7 unit tests added (the package had none).

### Why the provider and not the action layer — the decision to re-examine

The action layer is the tempting place, because `assets.origin_prompt` would then show
the terms. Rejected for three reasons:

1. **Folding negation into the positive prompt is right for Gemini and WRONG for SDXL.**
   SDXL's CLIP text encoder cannot represent negation and tends to render what you asked
   it to omit. So the action layer would need to know which provider will serve the
   request — and that answer lives only in `routing.go`.
2. **A second hand-maintained provider list in a second package is the documented drift
   class here.** `routing.go`'s own header exists because a hand-maintained kind switch
   shipped two separate defects (`content_hero`, then `hero`/bugs_open/011). Queued
   diagnosis item `5db192c5` is already filed against exactly this shape: "the per-kind
   accessors ... each carry their own hand-maintained kind lists."
3. **It would feed `/bugs_open/027` §4b.** The action layer caps its composed direction
   at `maxImageryDirectionInPrompt = 200` with the palette composed **last**, so
   appending avoid terms up there silently pushes the brand colours off the end. By the
   time a prompt reaches the provider the cap is already applied, so nothing the fold
   adds can evict anything. §5 candidate (1) flagged this hazard; placing the fold
   downstream removes it rather than working around it.

One edit in the provider also covers the whole class at once — the style guide's
`avoid`, every `kindDefaults.NegativePrompt` (including the logo entry §2 names), and
the `input_data.constraints` folding — for every kind and caller.

### What was traded away, stated plainly

`origin_prompt` no longer witnesses the constraint (see the §6 correction). Mitigations:
the Info log with a 400-char preview, an in-code comment at the log site saying why
`origin_prompt` is the wrong place to look, and the §6 correction itself.

A `FinalPrompt` field on `provider.Result` to carry the sent prompt back for storage was
considered and **rejected as unconsumed surface**: `origin_prompt` is populated from
workflow config (`origin_prompt_field`) at four-plus sites per
`107_image_build_handler.sql`, so actually recording it is a multi-workflow config change
that would also change what `origin_prompt` means for every historical row. That is a
separate, deliberate piece of work — worth doing if a thread ever needs the
prompt-as-sent for more than this bug.

### What this fix does NOT claim

**That the constraint is now obeyed.** A prohibition in the positive prompt is a softer
instrument than SDXL's true negative conditioning, and the evidence that it is only
partly honoured is inside this very bug: `xp_curve_designer` had `near-black #121212`
in its **positive** prompt and still came back on a near-white ground. This restores the
constraint to having *any effect at all*. The wording ("the image must not contain or
use: …" — phrased for objects *and* styles, since avoid lists mix them) is a first
attempt that wants tuning from observation. **Count violations across 5+ generations.**

Expect fleet-wide visible change on the next image roll: every Banana generation now
carries a prohibition clause it did not carry before. Sites whose avoid lists are wrong
or self-contradictory will now *show* that, where the list was previously inert and
harmless. Some may look worse before they look better — that is a true signal, not a
regression.

## 7b. Council gate verdict — REVISE (`d35844da-f533-42da-b096-4f82cc2839bc`)

Round 1 was **void, not a verdict** — a harness defect, filed as `/bugs_open/036`.
Round 2 returned **REVISE**: 7 approve, 3 object (editquality, bug_historian,
guardian), 5 abstained on relevance. **No `Council-Reviewed:` trailer is claimed on
the fix commit** (`32f2d51e2`) — that trailer is earned by APPROVED only.

The core Banana fix was endorsed by every seat that looked at it, including the
placement argument (`constitution`: *"a genuine translation rather than a workaround,
and the rejected alternative is explicitly stated with three concrete reasons"*;
`reuse_agent`: *"the rejected alternative is the one that would actually have produced
a second, competing implementation"*). What drew objections was **edit 2 — the
interface-contract change — being comment-only.**

**Resolved by checking, in response to the objections:**

| Challenge | Seat | Outcome |
|---|---|---|
| *"SDXL uses this directly" is asserted, not verified — despite the author's own stated methodology* | bug_historian | **Verified, claim holds.** `stability/provider.go:185–201` reads `req.NegativePrompt`, falls back to the kind default, and appends it as a weighted `api.TextPrompt`. It is true negative conditioning. The seat was right that I had taken it from the old comment rather than reading the code. |
| Are Banana and Stability the only implementers? | guardian | **Yes, exactly two** — `var _ provider.Provider` appears only in `banana/provider.go:84` and `stability/provider.go:139`. No third provider is at risk. |
| Is any code or test coupled to the old "log and ignore" wording? | guardian | **No** — but the grep found the *same licence phrase* one field down, on `ReferenceImageURIs`. Now rewritten too. That discard is legitimate **because it is loud** (Warn in `stability/provider.go` *and* again in `dynamic_adapter.go`), which is precisely the standard the new wording sets: the difference was never the phrasing, it was Warn-in-two-layers versus Debug-in-one. |
| Was an existing sentence-boundary/append helper missed? | reuse_agent | **Checked.** `endsWithSentenceBoundary` does exist in `generate_image_actions.go:1161` — but in `platform/orchestration/actions`, and importing that into a provider adapter inverts the dependency direction. Six lines of pure duplication is the right call; recorded so the next reader knows it was a decision. |
| The three documents asserting the false lesson are not corrected | editquality | **The objection was right, my premise was wrong** — all three were already corrected before I started. See the correction block in §4; this seat is what caught it. |

**OPEN — deliberately not resolved here, because it is an owner call.** Two seats
(editquality, bug_historian, both medium) object that a comment cannot stop the next
provider repeating the bug, and bug_historian says so explicitly: *"documentation-as-guard
has already failed once on this exact field ... a human should decide whether the
interface-contract half needs an enforced fail-loud mechanism."* Both agree it does not
block the Banana fix. Options, none taken:

1. **Conformance test** iterating registered providers, asserting each either consumes
   `NegativePrompt` or fails loudly. Needs the API client injectable per provider —
   `banana.New` constructs `p.client` internally today, so this is a constructor change.
2. **Capability on the interface** (e.g. `HonoursNegativePrompt() bool`), letting the
   action layer choose a strategy. A real signature change to a shared interface, and it
   re-opens the placement question this fix closed.
3. **Accept the comment**, on the grounds that with only two implementers and a
   now-explicit contract, the marginal defect risk is small.

## 8. Related

- `/bugs_open/027` §4b — the palette-truncation defect, found in the same session. Both
  are the same shape: **a structured, brand-approved instruction is silently discarded
  between the style guide and the model**, and the output looks deliberate.
- `/bugs_open/011` — the routing work that put every kind on Banana. This bug is its
  unnoticed consequence, not a defect in it.
