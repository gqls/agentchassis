# HANDOFF — bugfix 417/420/462 lane — 2026-09-03b, continue here

**Supersedes `HANDOFF_2026-09-03_continue_here.md`** (kept — its §5 traps and §6 owner decisions are
unchanged and still live). That file's ⭐ item was "trigger ONE logo regeneration and look at it".
**Done, four times over. 417 looks fixed; a different problem is now the urgent one.**

**Bug files — resolve by SLUG, all three numbers are ambiguous:**
- `bugs_open/417_HANDOFF_2026-08-31_planner_logo_exemplar_licenses_a_wordmark_it_never_names_so_the_image_model_invents_a_brand.md`
- `bugs_open/420_HANDOFF_2026-08-31_order_intake_publishes_the_billing_email_as_the_sites_public_contact_and_registers_it_as_a_renderable_claim.md` ⚠ the *other* 420 is the negation gate's prose walker
- **`bugs_open/462_HANDOFF_2026-09-03_a_logo_can_be_perfectly_rendered_correctly_deployed_and_illegible_and_no_check_can_see_it.md` — NEW, filed today at the owner's instruction. This lane owns it.**

**Working docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_417_logo_text_policy/`
Read `SUMMARY_2026-09-03_logo_text_policy.md` if you want the plain-prose version first.

---

## 1. FLEET STATE `[MEASURED 2026-09-03 13:10Z]`

A fresh chassis rolled at **12:06:47Z / 12:07:16Z (`v1.0.1358`)**; adapter stamp
**`d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`**. All three relevant fixes re-verified as ancestors of
that stamp — `b2322a203` (417's prompt fix), `fcbe6071c` (424's guard fix), `6440ec968` (the matte) —
with HEAD as a negative control (correctly not an ancestor). **Re-verify against the CURRENT stamp,
not this one; another roll has been announced since.** Use ancestry, never dates.

### All four logo generations today, read at the served bytes with 404 controls — FINAL

| site | outcome | key date | carried a licence? | lettering? | legible? |
|---|---|---|---|---|---|
| seotools.co.uk | landed, attempt 2 | `20260903/` | no | **none** | max 7.64:1 ✅ |
| gamedesign.uk | landed, attempt 3 | `20260903/` | no | **none** | max 6.46:1 ✅ |
| websitepromotion.co.uk | landed, attempt 2 | `20260903/` | no | **none** | **median 1.01:1 ❌ WORSE than before — `bugs_open/462` §6** |
| **designblog.co.uk** | **landed 14:30:30Z, attempt 3 of round 2** | `20260903/` | **YES — `letterform`** | **none** | **min 5.83:1 ✅ best of the day** |

**All four are settled and all four are text-free.** The `licence?` column is the one that matters
for 417: only designblog's prompt actually put the override in an argument, so it is the only row
that is evidence about *adjudication* rather than about the model's default behaviour.

**The two extremes are the same code on the same day**, which is the cleanest single piece of
evidence that 462 is per-generation variance and not a systemic defect: median **1.01:1** against
**11.05:1**, hours apart.

---

## 2. WHAT IS LEFT, ORDERED

1. **~~THE DECISION ON 462~~ — RULED 2026-09-03: candidate 2, report it afterwards.** A post-hoc
   finding, NOT a fail-closed refusal at store time. Recorded in `bugs_open/462` §7 with what the
   ruling accepts: an illegible logo still ships and is repaired after detection. **Do not re-open
   the ranking** — §3 ranked candidate 1 first and the owner ruled against that ordering knowingly.
   **What is left is 462 §7a, and it is a genuine choice**: does the measurement live in the render
   audit (browser, sees the real rendered backdrop) or in a standalone check over stored assets +
   the theme token (cheaper, sweeps all 30 existing logos, but trusts the declaration over the
   render)? **Answer refined by the 424 lane and it is now two pieces of work, not a choice: ship
   the standalone sweep NOW for the "how widespread is this" answer, and plan the render-audit
   version as the one that STAYS correct.** The deciding objection is staleness, not coverage — a
   theme token read at check time is a snapshot, and colour churn is documented here
   (`generic_theme` landmine; `bugs_open/396` is a design run rewriting the theme row byte-for-byte),
   so the cheap check drifts into **false passes**, the one direction this bug is already about.
   Write the sweep knowing it is temporary: share its threshold, and **record the theme value it
   measured against in the finding**, so a later reader can tell "passed against a palette that no
   longer exists" from "passed". **Routing is also unsettled** —
   `css-patch-agent` repaints a CSS class and cannot fix a pale PNG.
   ⚠ Binding constraint from §6: **measure after matting, against the header, never the keyed
   ground.** A pre-matte check sees high-contrast white-on-magenta and passes happily.
   ⚠ Build it against **websitepromotion**, the motivating case: median 1.01:1, 85.4% near-white,
   live. A detector that does not flag it is not working.
2. **~~websitepromotion's logo~~ — RULED 2026-09-03: do NOT restore.** It keeps the
   white-and-magenta mark. **So the estate is knowingly serving one illegible logo, by decision** —
   this is not an outstanding repair and should not be re-raised as one. The two `PRESERVED_*.png`
   files in this directory are now **evidence, not a rollback option**; keep them, they are the only
   copy of the pre-regeneration artefact and the pair is what makes §6's mechanism legible.
3. ~~**Wait for `designblog.co.uk`.**~~ **LANDED 2026-09-03 14:30:30Z on attempt 3 — CLEAN, and it
   is the decisive run.** It carried BOTH the override and the `letterform` licence (verified on the
   post-regeneration row), and produced an abstract open-cube mark with an arrow: **zero lettering,
   one composition, no invented brand** — and the best artefact of the day on every other axis
   (0 near-white opaque px, **min contrast 5.83:1**, median 11.05:1, fringe 0.022%). Full write-up
   in `bugs_open/417`, "THE DECISIVE RUN". **The adjudication case is now n=1 and it passed — but
   n=1 bounds nothing, so the trigger does NOT close.** The next evidence comes from the other 12
   licence-carrying sites when they regenerate, not from sites carrying no licence.
4. **The fence decision (417)** — still deliberately not taken, and the grounds have MOVED. See §3.
5. **417, 420 and 462 all stay OPEN.** 417 on the fence residual and item 3; 420 on its §C residual.

---

## 3. THE FENCE — the evidence base is weaker than this morning's handoff claimed

**8 clean generations, 0 lettered** → bounds the lettering rate at only **≈31% (95%, rule of
three)**. A weak bound, and the failure mode is silent.

**But the sample is worse than weak — it is systematically the wrong sample.** Only `designblog`'s
plan still carries a permitting phrase, and it has never produced an artefact. The other eight show
the model not painting text **when nothing asked it to**, which is a far weaker claim than "the
override beats a licence". **The honest sample size on the adjudication case is n = 0.**

**Recommendation: fence stays UNBUILT, trigger stays OPEN** — not because the evidence reassures, but
because *the evidence on the case that matters does not exist yet*. Item 3 above produces it.

**Forward-looking exposure `[MEASURED 2026-09-03]`: 13 of 33 current logo specs** carry a
licence-shaped term — robot-hands, relojistas, vetcomparison, dartsonline, oufe, lendzy, webdesign,
noted, loanzy, cv1, farmerinsurance, boxingonline, designblog.

### ⚠ READ THIS BEFORE RE-RUNNING THAT CENSUS — I got it wrong three times in one hour

1. `origin_prompt ILIKE '%wordmark%'` → **5 of 5. FALSE.** The override clause itself ends
   *"presupposes a **wordmark** or any text"*, so it matched the prohibition. Unfalsifiable.
2. Stripped that one sentence → **"1 of 5". ALSO FALSE** — five recently-regenerated sites
   generalised to a fleet. Fleet-wide `origin_prompt` is 25 of 30.
3. Widened the regex to include `lettering`/`typograph` → **28 of 33. FALSE AGAIN** — the override
   also says *"no lettering, words, letters, numerals or **typography** of any kind"*.

**What fixed it was the TABLE, not the regex.** `assets.origin_prompt` answers "did the override
reach *this* artefact?" and is a historical record (19 of its 25 licences predate the override).
**`site_plan_imagery.prompt` + `sp.is_current` is what the NEXT generation composes from** — the
RUNBOOK's own census table. **Its free self-check: `plan_contains_override = 0` of 33**, because the
override is appended at composition time — so a licence hit there cannot be the prohibition quoting
itself. That is the disconfirmation control all three earlier attempts lacked. Exclude `lettering`.

---

## 4. WHAT THIS SESSION DID

- **417 verification completed at the artefact.** Three fresh marks eye-checked, zero lettering,
  single composition each, 421's two-panel shape absent. Two carried licence + override together.
- **`bugs_open/462` FILED** (owner-approved) + the transferable pattern in `016b` §9. Its argument
  is sharper than expected: the exclusion that leaves this uncovered is *correct* — `over_image`
  contrast is skipped because *"the adapter's own header calls the backdrop unknown"* — and the logo
  case **inverts every term** of that reason. A gap between two sound decisions, which is why it
  produces silence rather than a symptom.
- **websitepromotion regenerated** (owner-approved) and **got worse** — 462 §6 has the full numbers.
- **A LANDMINE**, verifier-confirmed `STILL_VALID`: a roll-killed run and a fail-closed refusal are
  the same four fields; read `error` for the guard's own statistic and compare against pod
  `startTime`, in one timezone.
- **Two CONTRIBs to the 424 lane** (they folded the first into their NOTES with credit): gamedesign's
  landing, designblog's exhaustion making their retry-ladder decision concrete, the threshold finding
  (every refusal exactly `0.000`, nothing near `0.95`, so attempts are the lever not the constant),
  and the light-mark/key-colour interaction.
- **Three corrections logged**, two of them retractions of things I had already told other people.

## 5. MY OWN CLAIMS THAT TURNED OUT FALSE — check these before trusting anything above

- **The census, three times** (§3). `WRONG_CALLS.md` has the row; the marker `[MEASURED]` was applied
  correctly every time and prevented nothing, because the measurement could not come out otherwise.
- **Despill "0.01% / 0.05%, not worth chasing" — RETRACTED.** Both samples had **dark** strokes,
  where a thin magenta fringe is cosmetic. On websitepromotion's white mark the same fringe is
  **63% of everything visible**. Severity depends on mark lightness; my two samples were both dark.
  The 424 lane has been told not to close their item on my numbers.
- **An enclosed-ground gap in 424's guard — REFUTED before I sent it.** gamedesign's interior white
  looked opaque and `BorderKeyed` only inspects the border ring, so I reasoned enclosed ground would
  survive by design. Zero near-white opaque pixels: it is transparency over a white page. **Over
  white, opaque and transparent are visually identical.**
- **"Committing will make `who-owns.py` see 462's ownership" — FALSE**, checked after committing. It
  still says `(none identified)` and files this lane under *"probably just a cross-reference"*.
  462 §5 is the ownership record; do not read the tool as saying 462 is free to route work at.

## 6. STILL THE OWNER'S — unchanged, see the 09-02 handoff §5

1. ~~**RFC_058, the identity model** — lane recommends B~~ **RULED 2026-09-03: OPTION C** (ordering
   party, operating party, published contact, **subject as first-class**, not derived as this lane
   had recommended). **Plus two owner additions — read the RFC's ruling section, because the second
   one changes the SHAPE, not just the branch:**
   - a **fifth identity**, the selling party (the platform sold through more than one front),
     **named and explicitly deferred by the owner**;
   - **more than one contact per identity — NOT deferred.** This kills "one column per identity"
     outright and pushes the consent state down onto the *contact*, adding a third distinction no
     field on `sites` can hold: *"has contacts, none published"* vs *"has no contacts"* vs *"we
     asked and the answer is none"*.
   ⚠ **The two additions point the same way and a future session must not decouple them:** a
   relation is what addition 2 forces, and a relation is also what makes addition 1 cheap to defer
   (a fifth identity becomes a ROW, not a migration). Shortcutting addition 2 back into columns
   silently converts the owner's deferral into future migration work.
   **Unblocked now; what is still owed is §5.4's reader census** (writers refreshed today, still 4;
   the 14 readers remain dated 2026-08-31 and every one must learn which identity it reads).
2. **The 420 §C residual** — does the narrow ruling extend to *derived* contacts? 28 specs carry one.
   **Re-framed by the 2026-09-03 RFC_058 ruling and arguably subsumed by it:** a derived contact is
   just a contact row, so the live question becomes **what CONSENT STATE a classifier may write on
   one**. ⚠ Do not let that be answered in two states — under Option C the honest answer for a
   derived contact is likely a third (*recorded, not published, never asked*), and a two-state
   answer designs the fill-only-if-empty inversion back in at row level, which is the very defect
   §C exists to record.
   **Timing agreed with `site_delivery_and_editor` 2026-09-03: they will carry it when the owner
   next has a DELIVERY decision in front of him** — the question is abstract standalone and concrete
   when "who may we email, on whose say-so" is already live — and hand the answer back verbatim.
   **It stays on THIS lane's record**; they are not taking ownership, and the answer lands in
   `bugs_open/420` and RFC_058.
3. **Ordering cannot reopen until the intake chat asks the contact question** — box-side.
4. **`bugs_open/421` still has no owner.**
5. ~~**462's fix shape** and **whether to restore websitepromotion's previous logo**~~ — **BOTH
   RULED 2026-09-03**: report-afterwards, and do not restore. See §2 items 1–2 and `bugs_open/462`
   §7. **NEW and open:** 462 §7a — where the measurement lives (render audit vs standalone check),
   and which handler a logo-legibility finding routes to.

## 7. IF YOU READ ONE THING

**Every real defect this session found was found by opening the artefact, and every one had a green
status beside it.** A logo passed a fail-closed guard, gained 9 points of transparency, and became
*less* visible — 85% of it white on a white header, and 63% of what you can see is the temporary
background colour that exists to be deleted. Meanwhile the number I trusted most, a dated
`[MEASURED]` census, was wrong three times in an hour because I kept searching for a word that the
rule I was measuring contains. **The census proves the instruction arrived. Only looking proves it
worked — and if your measurement could not have come out otherwise, you have not measured anything.**
