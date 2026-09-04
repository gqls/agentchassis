# TRIAL OUTPUTS 2026-09-04 — the Grok arms, verbatim, for the owner's own read

Same prompt as every arm in `EXPERIMENT_2026-08-31_model_trials.md`: `llm_call_log` id
`79257fb4-fcfa-4ff6-9923-dc4e7fcd2b6a` (finetuning.uk, `about-content` section of the APPROACH page,
built 2026-08-26 19:17Z on `claude-sonnet-5`). Rendered here as prose from the model's JSON so it can be
read aloud; the raw JSON is what was scored. Nothing here is edited. The Fable and Gemini arms of 08-31
were not stored verbatim, which is why they cannot be re-scored today — this file exists so that the
Grok arms can be.

## Production baseline — claude-sonnet-5, the shipped section (stored `response_text`)

**Section title:** How we think about deployment

**Highlights:**

- **Vendor neutral by design** — We pick whichever model handles a given job best, whether that is text, images, code, or reasoning, rather than locking you into one provider's tools.
- **The model is yours to keep** — When we fine-tune a model on your documents, the result belongs to you and can move to whichever host you choose later.
- **Human review where it matters** — You decide which decisions need a person to check them and which can run on their own, and every decision stays visible either way.
- **Data control built in, not bolted on** — Private deployment options, including models that run in-browser or on your own servers, keep your documents off third-party infrastructure.

### We start with the problem you already have
Most AI projects fail because someone falls in love with a technology before anyone has described the task properly. We do the opposite. Before we talk about models, agents, or automation pipelines, we want to know what the work actually looks like today: who does it, how long it takes, and where it breaks. That conversation usually points to one of a small number of well-understood mechanisms, rather than something bespoke and exotic, and knowing which one saves you paying for capability you do not need.


### No single model handles everything well
Different tasks suit different tools. A model that writes clean prose is not necessarily the one you want reasoning over a spreadsheet, and a model that is excellent at code is rarely the cheapest option for simple document lookup. We work across open-weight models such as Llama, Mistral, and Phi, alongside commercial options, and choose per task rather than committing your business to one vendor's roadmap. Where we fine-tune a model on your own documents, using efficient techniques such as **LoRA** (Low-Rank Adaptation, a way of teaching a model your specifics without retraining the whole thing) or its variant **DoRA**, the trained model belongs to you and can move to a different host if your circumstances change. That portability is not a footnote. It is what stops a two-year AI deployment turning into a two-year contract with a single supplier.


### Retrieval before invention
A lot of what looks like AI reasoning is actually retrieval done well. A **RAG** system (Retrieval-Augmented Generation, where the model searches your own emails, documents, or records before answering) lets a model give a specific, sourced answer instead of a plausible-sounding guess. We build these systems to point at your material first, because a model that has read your actual pricing sheet answers a pricing question more reliably than one working from general training data alone. You can read more about how this fits together on our [services page](/services.html).


### Data control is part of the design
If your documents include client records, financial detail, or anything covered by GDPR, where that data physically sits is not a minor technical detail. We offer private deployment options, including models that run on your own infrastructure or directly in the browser, so nothing needs to leave your side to get useful work out of it. You choose what leaves your systems and what stays, and that choice is made before deployment, not negotiated afterwards.


### Human judgement stays in the loop
Automation does not mean nobody is watching. Every pipeline and agent system we build includes configurable review gates, points where a person checks a decision before it goes any further. Some clients want a person to approve every quote before it goes out; others are happy to automate the routine cases and reserve review for anything unusual. Either is a reasonable choice, and it is yours to make, task by task, rather than something the system decides for you.


### Who this suits
This approach tends to work best for operations and finance teams who want a system they can explain to a regulator, a client, or their own board, rather than one they have to trust blindly. If that description fits your situation, our [how we work](/how-we-work.html) page sets out what the first few weeks of a project usually involve, and our [case studies](/case-studies.html) describe the kind of problems we have solved this way before.

## grok-4.6 #1 (reasoning)

`grok-4.6` · xAI response id `48fca3e0-25a5-978b-83b2-4a807ef946d4` · 315s wall · input 8905 (cached 512) · output 20418 (reasoning 18873) · xAI-reported cost $0.1396

**Section title:** The work comes first

**Highlights:**

- **Your data stays yours** — Models can run on your infrastructure, on a machine in the office, or in the browser, so company files can stay on your side.
- **The trained model belongs to you** — We train an open-weight model on your documents so it picks up your terms, your rules, and your tone. The finished weights are yours to keep and to move.
- **A person still signs off** — Review gates let you approve the steps that carry risk. The system takes the volume, and every decision stays visible.

Deployment, for us, is the job of taking a repetitive piece of work inside your business and giving it a system that can carry the routine part while people keep the judgement.

We begin with the process as it stands: who touches it, which documents it uses, where it fails, and what a good result looks like. After that we choose a model, a place to run it, and the points at which a person still has to approve the output. That order matters because a model aimed at a muddled process will still produce a muddle, and you will have paid for it.

The hour of the week you want back only counts if the work that fills it is regular enough to hand to software. We'll say so if it isn't.


### The shape of the system
A process that repeats, uses the same kinds of documents, and has a recognisable right answer is a candidate for automation. We map the steps, the data it needs, and the exceptions a person currently handles. From there we can say whether you need a pipeline that moves work along, a system that searches and reasons over your own files (retrieval-augmented generation, or **RAG**), or a model trained on your documents so it speaks in your terms.

The technical choice follows the job. Text, images, code, and reasoning each get whichever model does that job well. We work with open-weight models such as Llama, Mistral, and Phi, meaning the weights can live on a host you control and can move if you change supplier later.

If you want to see the kinds of work this usually covers, the [use cases](/use-cases.html) are the place to start. The steps from first call to a working system are set out in [how we work](/how-we-work.html). A companion note, [AI will not save a broken process](/blog/ai-wont-save-your-business-if-your-processes-are-broken.html), explains why the mapping comes first.


### Training a model on company documents
When the work has a house style, a set of product rules, or a way of writing the firm already uses, we train a specialised model on that material. The technique is called **fine-tuning**. We adapt a capable open model using only a small extra set of parameters (methods such as LoRA and DoRA), which keeps the job smaller and the result portable. The trained model knows your terms and your constraints.

A fine-tune with us starts at £99. Costs and scope sit on [what it costs to work with us](/pricing.html). For the mechanics of training and hosting, see [your own AI model](/your-own-model.html) and the [technical details](/technical-details.html).


### Where the model runs
Company data is both an asset and a compliance problem, so where the model runs is part of the design. For a UK firm that question sits next to GDPR. Finance and compliance leads usually want an answer before anyone pastes a client file into a public chat tool, and they are right to ask. We'll say which option keeps data on your side, and which option sends it out to a supplier.

If you want a structured way to think about that risk, use the [GDPR and AI data risk self-assessment](/tools/gdpr-ai-risk-assessment/index.html). A longer explanation of pasting work into public tools is in [this piece on ChatGPT and company data](/blog/chatgpt-has-your-data-does-that-matter.html).


### Review sits where the risk sits
An automation pipeline can take a document, extract what is needed, draft a response, and pass the work along. A **review gate** sits at the points that carry risk: a quote before it goes to a customer, a filing before it is submitted, a summary before it goes out.

You approve those steps. The system handles the volume, and every decision stays visible so you can see what the model produced and where a person changed it.

When the work is wider, we design networks of agents that can analyse information, talk to each other, and split a large job into parts. **Orchestration** is the layer that coordinates them. We also install and configure those frameworks so you can run your own agents once the system is live. Human review remains available at every stage unless you have chosen to switch a gate off. After launch, quality sweeps can keep a site, a data set, or a tool improving between human review cycles.

The [services list](/services.html) sets out automation pipelines, agent systems, RAG over company data, bulk collection, and private deployments as separate offerings, because most firms only need one of them to start. A short chooser, the [fine-tuning, RAG, or prompting guide](/tools/model-approach-selector.html), can help you see which shape fits before we talk.


### Built for admin-heavy UK firms
This work is built for operations managers and owners at UK firms of about 10 to 250 people, in professional services, logistics, facilities management, and other admin-heavy trades. You have probably tried ChatGPT. You already know it can draft an email. A system earns its place when it can take a real slice of operational work, sit inside the process you already run, and leave judgement with the people who hold it.

Finance and compliance leads in the same organisations usually need to see the data path before they will approve anything. That's a reasonable demand, and it belongs in the design conversation.

If that's your situation, pick one process and [start a conversation](/contact.html). We'll tell you when the work is a poor fit. For a wider view of how we think about the technology, read [our position on AI](/our-position-on-ai.html) or [AI for UK small business](/ai-for-uk-small-business.html).

## grok-4.6 #2 (reasoning)

`grok-4.6` · xAI response id `e3caac81-147f-9518-b4bf-5c518ae8752f` · 233s wall · input 8905 (cached 8832) · output 15317 (reasoning 14420) · xAI-reported cost $0.0965

**Section title:** The system follows the work

**Highlights:**

- **The job on the desk** — We begin with a process someone already runs, then match the system to that process.
- **Review stays visible** — You set the gates. People approve the steps that carry risk, and each decision can be seen.
- **Built to stay yours** — Models and pipelines can run on infrastructure you control, and what we build is yours to keep.

### Who we build for
UK operations teams come to us with work that is already happening: request screening, quoting, document reading, and the shuffling of facts between systems. The people we speak to usually run professional services, logistics, or facilities businesses, or something equally heavy on admin, with a team somewhere between ten people and a few hundred. You've probably tried ChatGPT. You want hours back in the week, fewer avoidable errors, and a straight answer on GDPR and data risk before anything is switched on.


### How a project takes shape
We name the job first. A person on your team is already doing it, often with a spreadsheet, an inbox, and a fair amount of memory. We look for the stretch where the work is repetitive enough to automate and important enough to watch.

An automation pipeline can take an inbound request, extract the useful facts, and prepare the next action, with a human review gate on anything you mark as needing a person.

Staff keep asking the same questions of the same files. A retrieval system lets the model search those files and answer from them. That pattern is called **retrieval-augmented generation** (RAG), and that's how internal emails, documents, financial records, and knowledge bases become usable by the model.

When the answers need to sound like your business, we train a capable open-weight model on your own documents. **Fine-tuning** is the name for that work. Methods such as LoRA and DoRA change a small part of the model so it picks up your terms and your way of working, and the trained model belongs to you. You can read more about [training a model on your documents](/your-own-model.html).

Some work is too branched for a single pipeline. In those cases we design a group of AI agents that can analyse information, pass it between them, and carry a larger task, with the review points you set. The steps from a first conversation to a system in production are on [how we work](/how-we-work.html).

Text, images, code, and reasoning each get the model that handles them well. Open-weight models such as Llama, Mistral, and Phi are ones you can host yourself, and they can be adapted to your domain and run on the host you choose.


### Where the work runs
Finance and compliance leads ask, quite reasonably, what leaves the building. **Private deployment** is the thing we design around: models that can run on your infrastructure, locally, or in the browser, so you decide what is sent anywhere else. A trained model you own can be explained to an insurer, an auditor, or a client who issues a questionnaire, because you can say where it lives and who can see the inputs.

Human oversight sits in the same design. You approve the steps that carry risk. The routine steps can run on their own, and a record remains of what the system did.

After launch, quality checks can keep sites, data, and tools improving between human reviews.


### What belongs in the system
A repetitive task with a clear input and a clear output is a good candidate for this kind of system. A process with no owner, or with rules that change every time someone asks, will give you a faster version of the same confusion.

If a process on your desk looks like a candidate, [start a conversation](/contact.html).

## grok-4.20-0309-non-reasoning #1

`grok-4.20-0309-non-reasoning` · xAI response id `c4c460a3-e64a-99f5-bef8-ff1cb4bc657a` · 5s wall · input 8167 (cached 128) · output 440 (reasoning 0) · xAI-reported cost $0.0112

**Highlights:** (empty — the model returned `[]`)

We build practical AI systems that work inside real UK businesses. Our approach starts with your actual problems rather than the latest model releases.

Most companies we speak to have already tried generic AI tools. They discovered that ChatGPT alone cannot reliably handle their documents, processes or data without making things up or exposing sensitive information. We take a different route.


### Start with the work, not the technology

We first look at where time is being lost or errors are creeping in. Only once we understand the task do we decide whether the right answer is a simple automation pipeline, a retrieval-augmented generation system (**RAG**), a custom-trained model (**fine-tuning** with LoRA or DoRA), or a network of specialised AI agents.

This order matters. It stops organisations from buying impressive technology that solves nothing useful.


### Three principles we follow on every project


- Your data stays under your control. We can deploy models privately, run them locally, or keep everything inside your own infrastructure.
- We use open-weight models wherever possible. The systems we build belong to you, not to a rented API.
- Human oversight remains part of the design. We build review gates so you decide what gets automated fully and what still needs checking.


### Who we work with

Our typical clients are UK SMEs with between ten and two hundred and fifty people, usually in professional services, logistics, facilities management or other admin-heavy sectors. They are curious about AI but rightly sceptical. They want concrete time savings and fewer errors, not science projects.

If you are tired of reading about what AI might do one day and would rather talk about what it can do for your business next month, we should talk.

[Start a conversation](/contact.html)

## grok-4.20-0309-non-reasoning #2

`grok-4.20-0309-non-reasoning` · xAI response id `d0087584-8600-999f-8398-d0b8508442d0` · 5s wall · input 8167 (cached 128) · output 489 (reasoning 0) · xAI-reported cost $0.0113

**Highlights:** (empty — the model returned `[]`)

Our approach to AI deployment is built on one simple idea: the technology should do the repetitive work so your team can focus on the parts that need human judgement.

We design and build systems that fit inside your existing processes rather than asking you to redesign how you work. That means starting with the tasks that are already well understood, clearly defined, and eating up the most time. From there we decide which combination of tools will actually solve the problem without creating new ones.


### Start with the problem, not the technology
Before we talk about models or architectures we look at your workflow. We identify where information moves between people, where decisions get repeated, and where errors tend to creep in. Only once that picture is clear do we choose the right approach, whether that is a straightforward automation pipeline, a retrieval-augmented generation system that answers questions from your own documents, or a network of specialised AI agents working together.


### We keep your data under your control
Your documents, emails, financial records and customer information never leave your environment unless you explicitly decide they should. We can fine-tune open-weight models using techniques such as LoRA so the resulting system belongs to you and runs on infrastructure you choose. Private deployments, local models and in-browser options are standard parts of how we work, not add-ons.


### Human oversight stays in the loop
Every system we build includes clear review gates. The AI handles volume and repetition; a person reviews the decisions that matter. This is not a limitation of the technology. It is how we make sure the output remains reliable and legally sound for UK businesses.

The same principle applies whether we are building [AI Automation Systems](/services.html), [Custom AI Model Training](/your-own-model.html), [RAG Systems for Company Data](/services.html), or [Intelligent Agent Systems](/services.html). We pick the technique that fits the job rather than forcing every problem into the same shape.

If you would like to see how this could work inside your business, [book a discovery call](/contact.html).

---

## The sonnet dose arms (arm 5), added 2026-09-04 13:30Z

Same prompt again. These are the control arms for the reasoning hypothesis: `claude-sonnet-5` with
thinking disabled, and with adaptive thinking at production's default effort. The `effort: "max"`
arms produced NO text at all (whole budget spent thinking, `stop_reason: max_tokens`), so there is
nothing to reproduce for them — that absence is the finding.

### sonnet-5, thinking DISABLED #1

**Section title:** How we decide what to build

**Highlights:**

- **No fixed vendor** — We work with open-weight models such as Llama, Mistral and Phi rather than tying a client to one vendor's platform. The model we train stays yours, portable to whichever host you choose later.
- **Human oversight by design** — Every automation pipeline we build can include a review gate, so a person checks the output that matters before it goes anywhere. What runs unattended is a choice you make, not a default we set.
- **Data stays where you put it** — Where a system touches your documents, emails or financial records, it runs on your infrastructure or a private deployment we configure for you. What leaves your systems is your decision.

We start from the problem in front of you, not from a model we happen to like. A business that wants its staff to stop answering the same question five times a day needs something different from a business drowning in supplier invoices, and the right AI system follows from that problem rather than from whatever is fashionable that quarter.


### Where this comes from
Most AI advice aimed at small and mid-sized businesses is written by people selling one platform, or written so generally it could apply to anyone. We built FineTuning because the useful middle ground, someone who will look at your actual workflow and tell you plainly what will and won't help, was hard to find. Our services page ([services.html](/services.html)) lists what we build, but the approach behind each one is the same: work out what the process actually needs, then pick the tool that fits it.


### What we actually do differently
- **Vendor-neutral by habit, not by slogan.** Text, images, code and reasoning each suit different models, so we choose per task rather than routing everything through one provider because that's the account we already have open.- **Fine-tuning that produces something you own.** Using techniques such as LoRA and DoRA, which adapt a large model efficiently rather than retraining it from scratch, we train a model on your own documents so it answers the way your best person would. The trained model belongs to you, not to us.- **Privacy treated as a real requirement.** Where the work involves emails, contracts or financial records, we can deploy privately, including models that run locally or in-browser, so your data does not need to leave your systems to be useful.- **Review gates instead of blind automation.** An automation pipeline (a sequence of automated steps that used to be manual work) can run unattended or pause for a person to check its output. Which one you choose depends on what's at stake, and that choice stays visible throughout.
### Who this suits
This approach tends to work best for operations managers and business owners who are curious about AI but tired of vague promises, and for finance or compliance leads who need a straight answer on data exposure before anything gets approved. If that's you, our position on data and vendor choice is set out further in [our-position-on-ai.html](/our-position-on-ai.html), and the practical side of a first engagement is covered in [how-we-work.html](/how-we-work.html).


### What we won't claim
Not every process benefits from automation, and not every business needs a custom-trained model when a well-configured off-the-shelf tool would do the job. We'd rather tell you that early than build something you didn't need. Where AI genuinely fits, we can show you a working system before you commit to the larger project, which is a better test of fit than any amount of description on a page like this one.

### sonnet-5, thinking DISABLED #2

**Section title:** How we think about deploying AI

**Highlights:**

- **Human oversight stays switched on** — Every automation we build has a point where a person can check the work before it goes anywhere, and you decide where that point sits.
- **No fixed vendor, no lock-in** — We pick from open-weight models like Llama, Mistral and Phi rather than tying you to one vendor, and the model we train for you belongs to you.
- **Your data stays where you put it** — Fine-tuning on your own documents keeps that data on your infrastructure. You choose what leaves, and what stays put.

### What we actually do
We build systems that take a specific piece of manual work out of a business and hand it to a model trained or configured for that job. That might be an **automation pipeline** that screens incoming requests, a **RAG system** (retrieval-augmented generation, where the AI searches your own documents before it answers) that lets staff query your knowledge base instead of hunting through folders, or a fine-tuned model that has learned your industry's language from your own paperwork. The mechanism changes depending on the problem. The starting point never does: we ask what is eating time or introducing errors, and we work backwards from there.


### Why we do not favour one model or vendor
Text, images, code and structured reasoning are handled better by different models, and that changes month to month as new ones are released. We work mainly with open-weight models, meaning the underlying model can be downloaded, adapted and hosted by anyone rather than locked behind one company's API. That matters practically: a model trained on your data using techniques like LoRA or DoRA (methods for adapting a large model efficiently, without retraining it from scratch) is portable to whichever host suits you, and it is yours to keep even if you later move away from us.


### Where data control fits in
Operations and compliance leads asking about AI deployment usually have one question underneath all the others: where does our data go? We design around that question rather than answering it after the fact. Private deployments, models that run locally or in-browser, and fine-tuning pipelines that never send your documents to a third party are all available, and which one is right depends on what you are handling and how sensitive it is. We would rather tell you a lighter-weight option is sufficient than sell you infrastructure you do not need.


### Human oversight, by design not by accident
Automation and agent systems (networks of AI models that can pass tasks between each other to solve something larger than one model could alone) can run with very little human involvement. That does not mean they should, and it does not mean ours do by default. Every pipeline we build has configurable review gates, so you decide what gets approved automatically and what a person checks first. That decision is visible and adjustable, not buried in the build.


### Who this suits
- Operations teams with a repetitive, rules-based task currently done by hand, where the volume justifies building something purpose-made.- Businesses holding sensitive client or financial data who need an AI system that does not send that data to a general-purpose chatbot.- Companies that have tried off-the-shelf AI tools and found them close but not quite right for how the business actually works.Not every business needs a custom model, and not every task is worth automating. Where a simpler tool or a process fix would do the job, we say so. You can read more about how a project actually runs from first call to delivery on [how we work](/how-we-work.html), see the range of what we build on [services](/services.html), or look at [case studies](/case-studies.html) for how this approach has played out on real projects.

### sonnet-5, ADAPTIVE thinking at default effort #1 (= production)

**Section title:** How we think about deployment

**Highlights:**

- **Problem first, technology second** — We start with the task costing you time and work out which kind of system actually solves it, rather than starting from a tool and looking for a use for it.
- **Whichever model does the job** — Text, images, code and reasoning each suit different models. We pick the one that fits the task instead of defaulting to one vendor.
- **Yours to keep** — Models trained on your documents and data kept on your infrastructure both belong to you, not to us, once the work is done.
- **A person still signs off** — Every automation and agent system we build can include a review gate, so you decide what runs on its own and what a person checks first.

### The problem before the technology
An AI deployment works when it begins with the task that is costing you time, not with a piece of technology looking for a reason to exist. We start by understanding what is actually slow or error-prone in your business, whether that is screening incoming requests, chasing quotes, or answering the same handful of questions every week. From there we work out which kind of system fits: an **automation pipeline** that removes a repetitive step from a process, a set of **AI agents** that split a large task between them and coordinate on the result, a model trained on your own documents, or sometimes a smaller fix that does not need AI at all. You can see the range of what this covers on our services page.


### Picking the model that suits the job
No single model is best at everything. Text, images, code and step-by-step reasoning tend to suit different models, so we choose per task rather than routing everything through one provider by default. Where a model needs to speak your industry's language specifically, we train it on your own documents using parameter-efficient techniques such as **LoRA** and **DoRA**, which adapt an existing open-weight model rather than building one from nothing. Open-weight models such as Llama, Mistral and Phi can be fine-tuned this way and moved to whichever host you choose, so the work you pay for does not lock you to one supplier. More on how this works in practice is on the trained-on-your-documents page.


### Keeping your data where it belongs
Company data, whether it is emails, financial records or client documents, is usually the most sensitive thing involved in an AI project, and it stays a genuine business concern rather than a checkbox to tick off. Where a project calls for it, we deploy models privately, including options that run locally or in-browser, so nothing leaves your systems unless you decide it should. A system that searches and reasons over your own internal data (a technique called **RAG**, or retrieval-augmented generation) can answer questions from your files without those files being sent anywhere else.


### A person still decides what matters
Automating a process does not mean removing the person who used to do it from the decision. Every pipeline and agent system we build can include configurable review points, so you set which decisions get approved automatically and which ones wait for a person, and every step the system takes stays visible. This matters more as a system scales, because the difference between a pipeline that handles ten requests a day and one that handles ten thousand is exactly the difference a review gate is there to manage.

- We name the business problem before naming any technology.- We choose the model or method that fits the task, not the one we happen to prefer.- Data stays under your control, with private deployment where that matters.- Review gates stay configurable, so oversight is a setting, not an afterthought.None of this suits every business, and we say so when it doesn't. If you want to talk through whether a project fits this shape, our how we work page sets out what the first conversation looks like.

### sonnet-5, ADAPTIVE thinking at default effort #2 (= production)

**Section title:** How we think about putting AI into a real business

**Highlights:**

- **Vendor neutral by design** — We work across open-weight models like Llama, Mistral and Phi, and choose whichever handles your task best rather than pushing one platform because it's the one we sell.
- **You keep the model** — When we fine-tune a model on your documents, the result belongs to you. You can move it to a different host at any point, because it was never tied to ours.
- **Human review stays in the loop** — Every automation pipeline we build includes configurable review gates, so you decide which decisions get approved by a person and which run without one.

Our approach starts with the work itself, not the technology. Before we talk about models or pipelines, we ask what task is taking up your team's time, where the errors creep in, and whether a machine can actually do that piece of it reliably. Most of the AI projects that go wrong start from the other end: someone picks a tool first, then hunts for a use for it. We do the opposite.


### Why we don't have a favourite model
Different tasks need different tools. A system that reads invoices, a system that writes marketing copy, and a system that reasons through a multi-step workflow are not the same problem, and treating them as one flattens what each is actually good at. So we work across open-weight models such as Llama, Mistral and Phi, and pick whichever suits the job in front of us. If a client already has a preferred provider, we work within it. If not, we recommend on merit rather than on what we happen to have on the shelf.


### Training on your own documents, not a generic dataset
Where a generic model gives generic answers, we can train one on your own material, using efficient techniques called **LoRA** and **DoRA** that adapt a model to your way of working without retraining it from scratch. The result answers the way your best person would, drawing on your documents, your language, and your edge cases. That trained model is yours. You can host it wherever you choose, move it later, or run it privately so your data never leaves your own infrastructure. You can read more about this on our page on [your own AI model, trained on your documents](/your-own-model.html).


### Keeping a person in the loop
Automation is not the same as removing oversight. Every pipeline we build, whether it screens requests, drafts quotes, or coordinates a network of AI agents working on a shared task, can include a review gate at the point that matters to you. You choose what gets approved automatically and what waits for a person to look at it first, and that choice is visible and adjustable rather than buried in a setting nobody remembers exists.


### Who this suits
This approach tends to fit businesses with a real, repeatable process buried under manual admin: operations teams processing the same request types daily, finance leads who need to see exactly how a figure was produced, or firms handling client data who need to know precisely where that data sits. It fits less well where the underlying process is broken rather than slow, because no amount of automation repairs a workflow that was never working in the first place.

- Start from the task, not the tool, and only bring in AI where it earns its place- Choose the model that suits the job, without loyalty to any one vendor- Train on your own material where a generic answer will not do- Build in review gates so a person stays accountable for what matters- Hand over ownership of anything trained specifically for youIf you want to see what this looks like against a live problem, our [services page](/services.html) sets out the individual pieces, and our [case studies](/case-studies.html) describe how they've come together on real projects.

