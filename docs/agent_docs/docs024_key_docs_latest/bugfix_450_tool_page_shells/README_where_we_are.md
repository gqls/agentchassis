# Where we are — the tool pages that are not tools (bugs_open/450)

*(Plain-prose log for the owner. Append only, newest at the bottom.)*

---

**2026-09-03, morning — what this is, and why it is a separate piece of work from the seotools clean-up.**

When seotools went live it had seven pages at addresses like `/tools/robots-txt-tester/`. All seven
answer, all seven look finished, all seven carry the right headline — and none of them contains a
tool. They are articles *about* tools. The site whose entire product is tools shipped seven pages
promising tools that are not there.

You have already decided what to do about those seven: build the tools they promise. That is the
portfolio lane's job and it is in hand. This lane is the other half — making sure the next site
does not do the same thing, because nothing about seotools was special.

Here is the sequence, because the shape of it is the whole problem. The planner writes the site
plan and names the tool pages. At that moment the tools do not exist: they are invented later by a
separate process that visits one site every three hours or so, and when it arrives it invents its
own names — on seotools, not one of the seven names it chose matched a name the planner had used.
So the planner is naming pages that nothing will ever fill. Meanwhile the plan validator does spot
the problem and files a note saying "this page needs a proper build, not the generic one". Nobody
reads that note. It goes into a queue and sits there.

Then the ordinary machinery does exactly what it is built to do. Other pages link to
`/tools/robots-txt-tester/`. A checker notices the link points at a page that has not been built,
and files a repair: "build the target page". The builder picks it up, reads the plan, and the plan
says the page consists of a headline and a text block. So it writes a headline and a text block,
deploys it, and the link is no longer broken. Everything reports success. There is a guard meant to
stop exactly this — but it asks a question that does not apply here, so it waves the page through.

The blunt version: **a tool page is judged by whether it has a form on it, and nothing in the
system was asking that question.**

**What I am changing.** Two things, independent of each other.

The first is a rule the builders consult: *a page typed "tool" that contains no tool is not
available for generic building.* Note what that rule is not — it is not a flag someone sets and
someone else must remember to unset. It is a question asked fresh each time, so the moment the real
tool arrives the page stops being protected and everything proceeds normally. That mattered more
than it sounds: the obvious version of this fix was to mark the page somehow, and when I looked, I
found the field people would use for that has never once been changed by any code in this system,
in either direction. A mark nobody clears is a page that stays stuck for ever.

The second is upstream: stop the planner planning tool pages whose tools do not exist. That one is
a sibling of a gate another session shipped yesterday for a near-identical problem with listing
pages, and I am deliberately reusing its shape and giving it its own on/off switch, so if mine
misbehaves theirs is unaffected.

**One thing I checked before writing any of it,** because it could have broken your seotools
clean-up: when the tool pipeline builds a tool onto one of those seven pages, does my new rule get
in its way? No — the tool component is inserted before the pipeline asks for the surrounding
content, so by the time anything consults the rule, the page has a tool in it and the rule stands
down. Your wave is safe either side of this landing.

**Timing.** A chassis build is going out now; my work is not in it and will ride the next one after.
I will say plainly when it is live rather than implying it from the fact that it was committed —
those are different things here.
