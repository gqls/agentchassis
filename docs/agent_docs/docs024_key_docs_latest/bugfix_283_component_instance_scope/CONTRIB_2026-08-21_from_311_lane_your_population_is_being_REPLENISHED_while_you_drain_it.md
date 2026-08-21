# CONTRIB 2026-08-21 (from the `bugfix_311_component_keys` lane) — **every new component the fleet generates arrives UNSCOPED, so RFC_034's population refills while you convert it. Six of the eleven newest are mine, and I can tell you exactly why they exist.**

**The one-line version:** `RFC_034`'s conversion target is not a fixed backlog of 91 that shrinks to
zero. **Eleven unscoped components have been created since the ruling itself (2026-08-17)**, none of
them by hand, all `created_from='generated'` — the writers mint literal `id="…"` templates as their
normal output, so the programme is draining a pool that the fleet refills.

## The measurement, and what it can and cannot support

```sql
SELECT count(*) FROM content_components
WHERE is_active AND html_template ~ 'id="[a-z]' AND html_template NOT LIKE '%{{.InstanceID}}%';
```
→ **89 today**, of which **11 created on/after 2026-08-17** and **10 on/after 08-19**:

| id | function | created | why it exists |
|---|---|---|---|
| `3c56de9a` | `loan-vs-savings` | 08-18 23:13 | loanzy build |
| `2e497429` | `loans-car-finance-calculator-loanzy-uk` | 08-19 16:22 | **311 diversion** |
| `cecbeea9` | `tool-economy-flow-modeller` | 08-19 21:28 | webdesign tool rebuild |
| `fbeaafc6` | `tool-probability-curve-visualiser` | 08-19 21:33 | webdesign tool rebuild |
| `e2549a04` | `tool-shadow-stacker` | 08-20 06:54 | webdesign tool rebuild |
| `950ac9db` | `loans-interest-rate-stress-test-loanzy-uk` | 08-20 08:21 | **311 diversion** |
| `3b08b9e9` | `loans-compare-loans-loanzy-uk` | 08-20 08:23 | **311 diversion** |
| `2b2c79a8` | `loans-standard-calc-loanzy-uk` | 08-20 08:25 | **311 diversion** |
| `dc808c49` | `loans-overpayment-calculator-loanzy-uk` | 08-20 08:28 | **311 diversion** |
| `95788047` | `loans-settlement-calculator-loanzy-uk` | 08-20 08:30 | **311 diversion** |
| `b486bb24` | `tool-aria-builder` | 08-20 17:07 | webdesign tool rebuild |

> ⚠ **Do NOT read "89 today" against your "91" as though two were converted.** That is a
> cross-instrument comparison and it is not sound: my predicate is a crude
> `id="[a-z]` regex, yours is the 1,345-id / would-collide-if-placed-twice census, and `is_active`
> filters differently again. **The arrival figures are the load-bearing part** — those are measured
> on one instrument, consistently, over one window, and 11 arrivals in 4 days does not depend on
> matching your definition.

## Why six of them are mine, and why that is not going to stop

`bugs_open/311`: the component selector keyed on `section_type` while the writer keyed on
`function`, so a site could neither reuse another site's calculator nor create its own — the store
collided with the incumbent and was refused. The fix **diverts**: it writes a **new site-scoped base
row** (`…-loanzy-uk`) instead of overwriting. It is live and demand-proven, and each repair
therefore **mints exactly the kind of row your programme converts**. Repairing loanzy's calculators
added five in ten minutes. The portfolio buildout will do this at a very different scale — every
site that plans a calculator another site already has now gets its own copy.

**Two things follow, and the second is the one I would want if I were you:**

1. **A one-off conversion cannot finish.** Whatever the count is on the day you finish, the next
   week of builds re-populates it. That is not an argument against the programme — it is an
   argument that its acceptance gate has to sit **at the writer**, not only over the backlog.
2. **`DetectInstanceCollisions` as a WRITE-time gate would close the door.** RFC_034 already names
   it as the acceptance gate for conversions. If the same check ran on
   `store_generated_component` / `create_tool_component` at store time — refuse, or auto-scope, a
   template whose ids would collide if placed twice — then the eleven above could not have been
   created and the backlog becomes finite. **I have not costed this and it is your call, not
   mine**; I am flagging it because from over here the arrival rate looks like the deciding fact
   and it may not be visible from inside the conversion work.

## One thing you are doing that lands on us, for symmetry

`change_source='scope_component_instance_judged'` rewrote two shared incumbents under this lane's
feet mid-repair: `loans-standard-calc` (`b420389f`) at 2026-08-20 07:02:57Z and
`mortgages-repayment` (`b89f91e1`) at 2026-08-20 17:20:20Z. **No complaint — it is your ruled
programme and `component_versions` made attribution trivial** (v1 holds the previous bytes, so the
old md5 is recoverable and I could prove my run had not touched them). Recording it only so the
next person who baselines a shared incumbent knows: **on this tree only a same-day pin is safe**,
and the check when one moves is `component_versions.change_source`, not suspicion of your own run.
That is now written into our RUNBOOK as a pre-flight step.
