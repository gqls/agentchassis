# SUMMARY — 2026-08-02 — the agent-state gather (bugs_open/172)

## What we're trying to do

Make the diagnosis loop tell the truth about its own evidence. When the platform
diagnoses itself, it assembles a dossier for the model that decides what went wrong.
One section of that dossier lists the agents the symptom named and what each has
recently been doing. If that section can drop an agent without saying so, the
deciding model reads the silence as "this agent wasn't involved" — so a cap that
under-reports is not untidiness, it is a wrong-answer mechanism aimed at the one
component whose job is to be right about causes.

## Where we've come from

This is the fourth pass over the same shape in the same corner of the system. A
2026-07-20 audit swept these files for silent caps and cleared them; `bugs_closed/164`
then found one the audit had missed; 164's own review asked whether a fourth existed,
and `172` was filed in answer — a count-based cap that truncates the list of agents
without a marker, and does it non-deterministically because the underlying query had
no sort. Each pass narrowed to the shape it happened to grep for, and each time the
next instance was found by someone else, later.

172 was filed as **latent**: it allows five agents and we had never seen more than
four named.

## What we've done

Confirmed the latency claim, then measured the section as it actually renders — and
that is where the job changed. The same gather gives all the named agents **one
shared allowance** of AI-call history, handed out by recency, so the busiest agent
takes the lot. Of the 72 retained dossiers carrying this section, **23 named more
than one agent and had history; in all 23, every line belonged to a single agent.**
Reproduced directly against the live table: ask about three agents and you get ten
rows, all from one, while another with 18,286 calls to its name renders nothing.

So the ticket's cap was asleep and its neighbour had been awake since at least
20 July. Both are fixed:

- each named agent gets its **own** allowance instead of sharing one;
- dropped agents are **named**, and the heading counts what it actually covered;
- an agent with no history is **stated**, so silence stops being ambiguous;
- the agent list is **sorted**, so two identical diagnoses can no longer examine
  different agents and both report success.

It went through the council gate and was **approved first time**. One of its four
advisory objections was a real catch and is now fixed: my new "no history" message
asserted more than the data supports, because `agent_type` was relabelled on
26 July — an agent can have history under a former name. It now states what is true
and names the boundary. A second objection asked why nobody audits for the *other*
instances of this shape; that audit found a fourth site, filed as **`bugs_open/181`**.

Two mistakes are written up rather than tidied away. A test I wrote to prove the
sorting fix passed happily when I deleted the fix — it was checking my own fixture
against itself, because a mock hands back rows in the order the test chose. And
twice while measuring for 181 I read a clean zero as a finding when the query had
examined nothing at all. Both are in `WRONG_CALLS.md` and in the new bug file's own
guard.

## Where we are now

Fixed, tested by induction (the type cap cannot fire in production, so it is driven
with the cap lowered), mutation-proved, council-approved, and committed —
`3761a04ca` and `c8031e284`, correlation `d47b826e-6fc6-42ad-a2ef-62b1f1ba0b88`.

**It is not yet live**, and `172` therefore stays OPEN. The bar here is *fixed AND
live*, not *fixed and committed*: the defect is reproducible until an image carries
the change. A build was held back deliberately — another session had a council round
mid-flight and a chassis roll would have killed it.

Every dossier written before this ships still carries the flaw, and they are retained
about a month, so a landmine entry now warns anyone reading one not to treat a
missing agent as an absent one.

## Where we're going

Three things, in order. Roll the chassis and verify at the running pod with both a
positive and a negative control — the runbook names the exact strings — then close
`172` to `bugs_closed/`. Pick up or route `181`, whose first question is not the fix
but why 276 retained dossiers contain no rendered code-search output at all. And, if
this shape appears a fifth time, take the `bug_historian` seat's advice and ask why
there is still no shared guard, rather than patching a fifth instance.
