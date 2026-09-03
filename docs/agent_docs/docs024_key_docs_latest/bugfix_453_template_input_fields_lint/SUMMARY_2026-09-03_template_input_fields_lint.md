# SUMMARY — 2026-09-03 — the prompt-template lint

## What we're trying to do

Stop a particular kind of silent failure in how our agents build their prompts.

An agent writes its prompt from a template — text with holes in it, like
`Company: {{.company_name}}` — and the holes are filled from data the step was told to
collect, in a list called `input_fields`. If the template asks for something that is not on
that list, nothing breaks. The hole fills with nothing. The prompt goes to the model looking
entirely reasonable, one paragraph or one list shorter than intended, and there is no error,
no warning, and nothing in any report. Whoever wrote the template believes it is working.

That has now bitten us four separate times, and every single catch came from somebody
happening to write a test that rendered the thing and looked at the output. We wanted a
checker that asks the question mechanically, of every template we have.

## Where we've come from

The class was written up as `bugs_open/453` on 3 September by the lane that hit it most
recently, and it asked for this checker as its first fix candidate. A second lane then added
an important correction: there is more than one way for a hole to end up empty, and the
checker as described would only catch one of them.

Both lanes explicitly said they were not going to build it. That is where we picked it up.

## What we've done

Built it, ran it against everything we have, and shipped it.

It reads every live agent, finds every step that renders a prompt, and works out — from the
code that actually does the filling, not from a description of it — which names that template
can and cannot resolve. It parses the templates the same way the real system does, which
matters more than it sounds: a crude version misreads the inside of a repeating block and
would have reported dozens of problems that are not problems, which is exactly how a checker
gets ignored and switched off.

On its first run: 202 agents, 1,474 steps, 139 templates, nothing it failed to read. One
genuine "this can never work". Sixteen it honestly cannot decide because they depend on data
that only exists mid-run, and nineteen of the opposite problem — a step collecting something
no template reads, which costs a search through the whole data tree on every run for nothing.
Only the first kind fails the check. The other two are reported and block nothing.

The one genuine finding is on our main page-writing agent, and it is the same step the bug was
filed about. Somebody fixed that step by hand last week for a different missing name. This one
was sitting next to it and nobody saw it. That is the argument for the checker in a sentence.

## Where we are now

Shipped, registered, submitted for review, and honest about its limits.

The one finding is real but **causes no damage today**, and we know that because we measured
it rather than assumed it. The template has a whole "Research Findings" section that can never
fill. There is a research step wired up to feed it — so our first read was that we were paying
for research and throwing it away. We weren't. The research agent has not run since 18 January —
I first wrote "never, not once" here, which was wrong and read off a table that only keeps two
days; corrected the same day against a five-month log of every model call we make, where it
appears zero times. The switch that would turn it on is carried by none of our 554 component
definitions. Two things are dead and each one hides the other.
Nothing is broken now, and the first person to turn the research on will pay for it and get
silence, with the prompt still telling the model to cite sources it never received.

We also found, and wrote down, that the one piece of the runtime that does look at this pair
looks at the wrong end of the name — so it reports a success as a failure. Nothing in the fleet
currently triggers it, so it is a trap recorded for the next person rather than a bug to fix.

We deliberately did not change the page-writing agent. Making its research block work is a real
change to the busiest thing we run, it is not urgent on the evidence, and it should be somebody's
decision rather than a side effect of building a checker.

## Where we're going

Three things are open, and none is ours to close alone.

The checker is run by hand. It is one step from running nightly — the plumbing for that is
already in it — but this estate's own history says a check nobody schedules quietly stops
mattering. That needs doing.

The second way a hole ends up empty — where the name is right but the piece of data behind it
is missing — cannot be caught by anything that reads configuration, because the configuration
is correct. The other lane measured that one firing on about two thirds of our page-writer
prompts. The fix lives at the opposite end of the system and belongs to whoever takes it.

And somebody needs to decide whether the research block on the page writer should be wired up
or deleted. Right now it is neither.
