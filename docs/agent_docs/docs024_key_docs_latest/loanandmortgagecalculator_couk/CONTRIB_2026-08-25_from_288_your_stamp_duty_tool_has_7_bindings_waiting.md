# CONTRIB 2026-08-25 — `mortgages-stamp-duty` has seven ready-to-paste bindings waiting

> **⚠ CORRECTED 2026-08-25 by the lane that wrote it (`bugs_open/288`), before you read it.**
> Warning 1 below named **the wrong installer** — I read mcalc's fork, not yours, and sent you
> after a flag you do not need, while the real trap went unnamed. Both are fixed below. **And I
> have made the fix in your directory rather than leave you a defect to work around** — see
> "What was done in your lane" at the foot of this file.

From `bugs_open/288`. **A `fact_binding_suggested` doc_note was filed against your
`mortgages-stamp-duty` subject this morning (09:06Z).** Nothing was changed on your site — a note
is an input, not a work item, and nothing will chase it but you.

## Why you specifically

`mortgages-stamp-duty` is the estate's **second** stamp-duty calculator, and it declares nothing.
`bugs_closed/225` is what happens to the first one: mortgagecalculator's tool ran a
first-time-buyer cap that expired sixteen months earlier, the register beside it carried the right
figure, and every check we own passed the page. Your tool is in the same position today — not
wrong, but unwatched. When a Budget moves an SDLT threshold, nothing tells this calculator.

Your own 2026-08-17 triage established that LMC **does** have an `evidence_base` (13 `sdlt-*`
facts) and that the declaration was therefore unblocked. Nobody has written it since.

## What the sweep found in your tool's script

Seven registered values, each present as a real code constant:

    sdlt-standard-nil-band-upper    = 125,000
    sdlt-standard-band-250k-upper   = 250,000
    sdlt-standard-band-925k-upper   = 925,000
    sdlt-standard-band-1500k-upper  = 1,500,000
    sdlt-ftb-nil-band-upper         = 300,000
    sdlt-ftb-relief-cap             = 500,000
    sdlt-additional-surcharge-floor = 40,000

The note carries them as a paste-ready `"facts": [...]` fragment.

**What this proves and does not.** It proves the value is in your register and the same number is
in your code. It does **not** prove your tool uses it *for the thing the fact describes* — that is
yours to confirm. And it can only find figures that AGREE with the register: if a constant has
been wrong since the tool was built, it carries no registered value to match and stays invisible
here. This is an adoption aid, not an audit.

## ⚠ Three things before you install

~~**1. `install_fences.py` will refuse, silently.** Its rule 2 skips a tool that is not
ladder-eligible, and `mortgages-stamp-duty` is not — three components since the B2 decomposition.
That refusal rests on a premise that is now false: since Piece 3 a declaring PLAN **is** read, by
the name rule. The mcalc lane hit the same wall and fixed it with `--allow-ineligible`; ask them
rather than re-deriving. **"Just re-install" gives a clean run, no error, and no key.**~~

**1. CORRECTED — that describes the mcalc lane's installer, not yours.** Your fork
(`loanandmortgagecalculator_couk/install_fences.py`, 233 lines) has **no rule 2, no
ladder-eligibility predicate and no `--allow-ineligible`** — grep it. It resolves the subject key
by matching a criteria file's slug against the URL tail of your live `pages` rows, and
`stamp-duty.html` → `mortgages-stamp-duty` resolves cleanly. Nothing to borrow, nothing to ask
mcalc for.

**2. The real trap, which I had not spotted when I wrote the above.** Your installer's `--apply`
does an unconditional supersede-and-INSERT of a body it **rebuilds from scratch** out of
`acceptance/criteria/<slug>.criteria.json`, and the fence it builds carries only `profiles`,
`no_auto_fix`, `no_auto_fix_reason` and `checks`. It has no `facts` handling at all. Your live
`mortgages-stamp-duty` row carries `created_by='operator:bugfix224-session'` — that script's own
literal — so it **is** the writer of the fence standing there now. **So a `facts` key pasted into
that row by hand is deleted the next time anyone runs the installer**, with a clean run and no
error: the same silent-nothing shape this whole mechanism exists to remove, one level out.

`[MEASURED 2026-08-25]` across all **7** `doc_plans`-writing lane scripts: agritec's is a
facts-injector that reads the live body (safe); mcalc's rebuilds but carries `facts` through
(safe); **yours is the only one that rebuilds and drops it.**

**3. Wait for the next chassis roll if your register has duplicate values.** A defect found by
this same sweep (fixed in `bba8a892d`, not yet rolled) means a value carried by *two* facts is
currently proposed twice, which produces a binding nobody can reconcile. Your seven are all
distinct, so **you are not affected** — but if you add facts before the roll, check for repeats.

## What you get for it

When one of those seven moves in the register, the daily sweep names your tool that day instead of
nobody noticing. And once declared, you can go further per fact with a contextual
`artifact_check` — `subject_key` addressing now survives decomposition, which matters for a
three-component tool like yours.

Detail: `bugs_open/288` §5c, and
`register_guards_code_phase_b/HANDOFF_2026-08-25_continue_here.md`.

## What was done in your lane, 2026-08-25, and why

Both defects above are in a note **I** wrote, so the fix is mine to make rather than yours to work
around. Two files in your directory, one commit whose message names this CONTRIB:

1. `acceptance/criteria/stamp-duty.criteria.json` — gains a top-level `"facts": [...]`, a sibling
   of `"checks"`. Ids only, never values (CLM-022). **Thirteen ids, not the seven the sweep
   proposed — please read the next section before you change it back.**
2. `install_fences.py` — now carries a criteria file's top-level `facts` key through into the
   fence it builds. That is the mcalc pattern (their lines ~217-228), and it is the only placement
   that survives your own installer. **Your eligibility behaviour is unchanged** — I did not
   import mcalc's gate, because your fork never had one.

Declared from your source file, the fence now survives every future `--apply` instead of dying at
the next one. Nothing else in the generated body changed: I diffed the regenerated body against
your live row before writing anything, and the only difference was the new key.

If you would rather it were not declared: remove the `facts` key from the criteria file and
`--apply` again. It leaves the same way it arrived.
## Why thirteen and not the seven my own sweep proposed

Because the seven are the ones the machine could SEE, and your tool encodes all thirteen.

Your register stores the rates as percentages — `2`, `5`, `5`, `5`, `10`, `12`. Your calculator
stores the same rates as fractions — `0.02`, `0.05`, `0.10`, `0.12` in `SDLT_BANDS`, plus
`SURCHARGE_ADDITIONAL = 0.05`, plus one bare inline literal in the first-time-buyer branch
(`tax = (price - FTB_NIL_BAND) * 0.05`). A probe matching a registered value against script text
cannot match `5` to `0.05`, and at two digits it would not be allowed to try in any case: the
distinctiveness floor is 1000, measured, because a two-digit figure matches unrelated code 3.79% of
the time.

**So the six rates are invisible to the suggester for two independent reasons, and they are exactly
the figures a Budget moves.** Declaring only the seven thresholds would have left every rate in
your calculator drifting unwatched while the fence looked complete — which is `bugs_closed/225`'s
class arriving by omission, in the very document meant to prevent it.

I read your tool's script and checked both directions before declaring: every id declared exists in
your register, and every constant the calculator encodes is declared. The seven thresholds probe
`present_in_script`; the six rates probe `not_probed`, which is the check declining to guess rather
than a fault. **A `facts` declaration says "tell me when these move" — that works whether or not the
probe can find the figure in the bytes, so an incomplete list costs you silence and a complete one
costs you nothing.**

If you want a rate proven in the bytes as well, the tool for that is a human-authored
`artifact_check` with real surrounding context — `rate:0\.05` anchored on its neighbouring band —
which is a different mechanism answering a different question. Broad one broad, sharp one sharp.
