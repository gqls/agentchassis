-- ============================================================================
-- ai-agent-orchestration.com — portfolio spec
-- ============================================================================
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes)
VALUES (
           (SELECT id FROM sites WHERE domain = 'ai-agent-orchestration.com'),
    'portfolio',
    '{
        "case_studies": [
            {
                "client": "Multi-Vertical Website Network",
                "title": "Autonomous Multi-Site Deployment Pipeline",
                "summary": "Built a fully autonomous pipeline that takes a domain name and produces a complete, deployed website — strategy, content, design, tools, and news feeds — without manual intervention. The system orchestrates 30+ specialised agents across domain research, site planning, content generation, design, deployment via GitHub Actions to Cloudflare, and ongoing improvement loops.",
                "results": "Six production sites deployed and self-maintaining. Each site receives autonomous content audits, tool suggestions, news feeds, and quality improvements on rolling schedules. Average time from domain registration to live deployed site: under 4 hours."
            },
            {
                "client": "Gas Wholesale Industry Portal",
                "title": "Industry-Specific Tool Generation and Cross-Linking",
                "summary": "Deployed an autonomous tool pipeline for a wholesale fuel distribution portal. The system evaluates what interactive tools would genuinely help the target audience, generates them via LLM, deploys with companion guides and navigation entries, then cross-links tool references into relevant content pages — all without human intervention.",
                "results": "Six industry-specific tools deployed including unit converters, cost estimators, and budget calculators. Tool references automatically woven into 18 content pages. Full pipeline from evaluation to deployed cross-linked content runs autonomously."
            },
            {
                "client": "Veterinary Industry Data Platform",
                "title": "Large-Scale Structured Data Collection with Verification",
                "summary": "Built a multi-stage data pipeline for the veterinary industry: area discovery identifies practice locations, scraping agents extract practice details, Companies House matching verifies business legitimacy, and enrichment agents add financial and structural data. Each stage is a separate agent with its own workflow, coordinated through the work item pipeline.",
                "results": "Thousands of veterinary practices collected, verified against Companies House records, and enriched with financial data. Automated verification catches dissolved businesses, name mismatches, and incomplete registrations before data reaches the consumer-facing layer."
            },
            {
                "client": "Real-Time News Collection System",
                "title": "Multi-Source News Aggregation with Credibility Scoring",
                "summary": "Designed and deployed a news pipeline that ingests from RSS feeds, web search, and the xAI Responses API simultaneously. A triage agent scores each item for relevance and credibility, tracking source attribution chains from original publisher through to discovery method. Items are rendered to JSON, committed via git, and deployed to live sites via GitHub Actions.",
                "results": "Four source types operational. Credibility scoring distinguishes tier-1 wire services from social media speculation. Six-hour refresh cycles with automated source health monitoring. Live news sections deployed on production sites with proper attribution."
            },
            {
                "client": "Internal Platform",
                "title": "Production Agent Orchestration at Scale",
                "summary": "The orchestration layer itself: a Kubernetes-native platform running stateless agent pods coordinated via Kafka messaging and Postgres state management. Agents are defined as SQL workflow definitions with Go action implementations. The system handles spawning, delegation, timeout recovery, optimistic locking, and parent-child orchestration chains across potentially thousands of concurrent agent instances.",
                "results": "30+ agent types in production. Workflow definitions hot-swappable via database updates without redeployment. Stateless pods scale horizontally. Built-in error routing, fuel budgets, and processing history for full observability of every orchestration."
            }
        ],
        "use_cases": [
            {
                "client": "Engineering Teams with Agent Prototypes",
                "title": "From Demo Agent to Production System",
                "summary": "Your team built an agent that works in a notebook. Now it needs to run reliably at scale — handling timeouts, retrying failures, coordinating with other agents, and letting your team intervene when something needs a human eye. We replace the prototype scaffolding with production infrastructure: Kafka for messaging, Postgres for state, Kubernetes for scaling. Human review checkpoints are optional at every stage — turn them on where you need oversight, off where you trust the automation.",
                "results": "A production agent system that runs unattended when you want it to, and stops for human approval when you need it to."
            },
            {
                "client": "Businesses Running Multiple Websites or Content Properties",
                "title": "Automated Content Operations with Human Oversight",
                "summary": "You manage a dozen location pages, or publish weekly industry updates, or maintain product listings across multiple sites. Right now someone on your team spends hours every week sourcing, writing, checking, and publishing. Our agent pipeline handles the repetitive work — research, drafting, quality checks, deployment — while your team reviews and approves at whatever checkpoints you choose. Every stage can run autonomously or pause for human sign-off, independently configured.",
                "results": "Content that stays current without the manual grind. Your team focuses on decisions, not repetitive updates."
            },
            {
                "client": "Organisations Needing Verified Data at Scale",
                "title": "Structured Data Collection with Built-In Verification",
                "summary": "You need accurate data about businesses, practices, suppliers, or competitors in your industry — but manual research doesn't scale and bought datasets go stale. Our agents discover entities, extract structured information, and verify against authoritative sources like Companies House or industry registries. Humans can review flagged records, approve batches, or let the pipeline run end-to-end with verification built in.",
                "results": "Continuously updated, verified datasets with full provenance. Flag anything uncertain for human review rather than publishing bad data."
            }
        ]
    }'::jsonb,
    'manual', 'manual', 'manual',
    'Real case studies from framework build — agent orchestration, site deployment, tools, news, data collection'
);

-- ============================================================================
-- finetuning.uk — update existing portfolio spec
-- ============================================================================
UPDATE site_specs SET is_current = false, superseded_at = NOW()
WHERE site_id = (SELECT id FROM sites WHERE domain = 'finetuning.uk')
  AND aspect = 'portfolio' AND is_current = true;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes)
VALUES (
    (SELECT id FROM sites WHERE domain = 'finetuning.uk'),
    'portfolio',
    '{
        "case_studies": [
            {
                "client": "Industry Portal Network",
                "title": "AI-Powered Website Generation for SMEs",
                "summary": "Built an AI system that takes a domain name and autonomously creates a complete business website — researching the industry, writing targeted content, designing appropriate visuals, generating interactive tools, and deploying to production. The system understands different industries and adapts its output accordingly, from gas wholesalers to consulting firms.",
                "results": "Multiple production sites running autonomously. Each site receives industry-specific content, relevant interactive tools, and ongoing quality improvements without manual intervention. SME owners get a professional web presence without the typical agency timeline or cost."
            },
            {
                "client": "Veterinary Industry",
                "title": "Automated Data Collection and Business Verification",
                "summary": "Deployed AI agents to collect, verify, and enrich veterinary practice data across the UK. The system discovers practices by area, extracts structured information, then cross-references against Companies House to verify legitimacy — catching dissolved businesses and incomplete registrations automatically.",
                "results": "Thousands of verified practice records with financial enrichment data. What would take a research team weeks runs continuously and autonomously, with built-in quality checks at every stage."
            },
            {
                "client": "Fuel Distribution Sector",
                "title": "Intelligent Tool Suggestion and Generation",
                "summary": "The AI evaluates what interactive tools would genuinely help a website''s visitors based on their industry and needs — then builds and deploys those tools automatically. A gas wholesaler gets fuel cost calculators and unit converters; a consulting firm gets ROI estimators. No irrelevant suggestions, no manual development.",
                "results": "Industry-specific tools deployed across multiple sites with companion guides and automatic cross-referencing from related content pages. The system handles everything from tool concept through to live deployment."
            },
            {
                "client": "News-Driven Businesses",
                "title": "Automated News Collection with Credibility Filtering",
                "summary": "Built a multi-source news pipeline that collects from RSS, web search, and AI-powered search simultaneously. An AI triage layer scores every item for relevance to the business and credibility of the source — distinguishing Reuters from anonymous social media posts before anything reaches the live site.",
                "results": "Automated news sections running on production sites with six-hour refresh cycles. Source credibility tracking means businesses only display verified, relevant industry news — building authority without editorial overhead."
            }
        ],
        "use_cases": [
            {
                "client": "Service Businesses (5-50 employees)",
                "title": "A Professional Website That Looks After Itself",
                "summary": "You need a website that actually represents your business — not a generic template, and not a six-week agency project. Our AI researches your industry, writes content that speaks to your customers, generates useful tools like calculators or estimators, and keeps everything updated. You can review and approve anything before it goes live, or let it run hands-off. Your call, page by page.",
                "results": "A professional, maintained website that stays current. You control what matters without doing the legwork."
            },
            {
                "client": "Businesses with Websites That Go Stale",
                "title": "Continuous Site Improvement Without the Overhead",
                "summary": "Your website was fine when it launched, but now the content is dated, the tools are broken, and nobody has time to fix it. Our improvement pipeline runs continuously — auditing content quality, checking tools work properly, identifying gaps, and fixing what it finds. Every change can be reviewed before it goes live, or you can let the system handle routine maintenance autonomously while you approve the bigger changes.",
                "results": "A website that gets better over time instead of slowly going stale. Routine maintenance handled automatically, significant changes flagged for your approval."
            },
            {
                "client": "Businesses Drowning in Manual Research",
                "title": "Automated Data Collection You Can Trust",
                "summary": "You spend hours every week checking competitor prices, tracking supplier details, or updating a spreadsheet of contacts. Our agents do that collection automatically — finding sources, extracting the data, and checking it against official records. Anything the system isn''t sure about gets flagged for you to confirm rather than silently publishing something wrong.",
                "results": "Structured, verified data on your schedule. You review the exceptions instead of doing all the work."
            },
            {
                "client": "Businesses That Need Fresh Content",
                "title": "Industry News Grounded in Credible Sources",
                "summary": "Your website looks static because nobody has time to curate industry news. Our pipeline collects relevant news from multiple sources and checks the credibility of each item — is this from a recognised publication or an unverified social media post? AI handles the research, source-checking, and first drafts. A human reviews and finalises the quality before anything goes live on your site. The result is authoritative content, not AI slop.",
                "results": "Credible, well-sourced industry content on your site. AI does the research and heavy lifting, humans ensure the quality."
            },
            {
                "client": "Business Owners Not Sure Where AI Fits",
                "title": "Honest Assessment, Then Practical Implementation",
                "summary": "You keep hearing about AI but you''re not sure what''s real and what''s hype for your business. I research your specific operations and constraints, then propose something concrete — sometimes that''s automation, sometimes it''s a better workflow, sometimes the honest answer is that AI isn''t the right fit yet. If we build something, you have approval at every step. No black boxes.",
                "results": "A clear answer on where AI helps your business, followed by implementation that pays for itself — or an honest ''not yet'' that saves you money."
            }
        ]
    }'::jsonb,
    'manual', 'manual', 'manual',
    'Real case studies from framework build — framed for SME audience'
);

-- ============================================================================
-- leopardessconsulting.co.uk — portfolio spec
-- ============================================================================
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes)
VALUES (
    (SELECT id FROM sites WHERE domain = 'leopardessconsulting.co.uk'),
    'portfolio',
    '{
        "case_studies": [
            {
                "client": "Multi-Site Content Platform",
                "title": "Hierarchical Agent Architecture for Autonomous Website Operations",
                "summary": "Designed and delivered a production agent orchestration platform running on Kubernetes, Kafka, and Postgres. Over 30 specialised agent types coordinate through hierarchical workflows — from domain research through content generation, design, tool deployment, and ongoing quality improvement. The system deploys complete websites from a domain name input and maintains them autonomously.",
                "results": "Six production sites deployed and self-maintaining. Stateless agent pods scale horizontally with zero-downtime workflow updates via database configuration. Full observability through processing history, error routing, and orchestration audit trails."
            },
            {
                "client": "Veterinary Data Aggregator",
                "title": "Multi-Stage Data Pipeline with Companies House Verification",
                "summary": "Built a structured data collection pipeline spanning area discovery, entity scraping, and automated business verification against Companies House. Each pipeline stage operates as independent agents with their own error handling and retry logic, coordinated through work items and Kafka messaging.",
                "results": "Thousands of verified business records with financial enrichment. Automated matching cascade handles name variations, dissolved companies, and registration gaps without manual intervention."
            },
            {
                "client": "Industry News Aggregation",
                "title": "Real-Time Multi-Source News Pipeline with Credibility Scoring",
                "summary": "Delivered a content feed system ingesting from RSS, news search APIs, and the xAI Responses API. An LLM-powered triage agent scores relevance and credibility per item, tracking provenance from original publisher through discovery channel. Rendered output commits via GitHub Actions to Cloudflare-hosted sites.",
                "results": "Four ingestion source types operational with automated credibility classification. Production deployment with six-hour refresh cycles and source health monitoring."
            },
            {
                "client": "Interactive Tool Platform",
                "title": "LLM-Driven Tool Generation and Deployment Pipeline",
                "summary": "Engineered an autonomous pipeline where an LLM evaluates what interactive tools benefit a site''s specific audience, generates self-contained HTML/CSS/JS tools, deploys them with navigation entries and companion guides, then creates cross-references from related content pages — closing the loop between tool creation and content integration.",
                "results": "End-to-end autonomous tool lifecycle: evaluation, generation, deployment, content integration, and quality auditing. Tools deployed across multiple verticals with industry-appropriate suggestions and zero manual development."
            }
        ],
        "use_cases": [
            {
                "client": "Scale-ups with Agent Infrastructure Needs",
                "title": "Production Agent Orchestration, Built and Delivered",
                "summary": "You need agents that run reliably in production — not a framework that works in demos. I build the orchestration layer on Kubernetes, Kafka, and Postgres: workflow-driven agents with proper state management, timeout recovery, and horizontal scaling. Human review gates are configurable per workflow stage — fully autonomous where you''re confident, human-in-the-loop where you''re not. The system runs thousands of agent instances in production today.",
                "results": "Production agent infrastructure delivered on proven stack. Runs unattended for weeks, pauses for human input when configured to."
            },
            {
                "client": "Teams Needing Automated Content or Data Pipelines",
                "title": "Multi-Stage Pipelines with Configurable Human Oversight",
                "summary": "You have a pipeline problem: collecting data from multiple sources, validating it against authoritative records, transforming it, and publishing — and right now too much of that is manual. I build agent pipelines where each stage runs independently with its own error handling and retry logic. Every stage can be set to autonomous or human-approval, independently. Your team reviews what matters and lets the rest flow.",
                "results": "Autonomous pipelines that handle the volume while your team handles the judgment calls."
            },
            {
                "client": "Engineering Teams Evaluating AI for Their Operations",
                "title": "Problem-First Assessment and Implementation",
                "summary": "You describe the problem — whether that''s operational bottlenecks, data quality issues, or scaling challenges. I research your domain, your constraints, and what''s been tried. Then I propose something specific: sometimes that''s an agent system, sometimes it''s infrastructure, sometimes it''s a workflow change that doesn''t need AI at all. I build what we agree on, with your team having visibility and approval at every step.",
                "results": "An honest technical assessment followed by delivery. No predetermined solution — the recommendation fits the problem."
            },
            {
                "client": "Businesses with Deployed Systems That Need Ongoing Improvement",
                "title": "Continuous Quality Improvement Pipeline",
                "summary": "Your site or platform is live but quality degrades over time — content goes stale, tools break, new gaps appear. I build improvement loops that run continuously: content audits, tool health checks, structural analysis, and automated fixes. Each finding is classified by severity and routed appropriately — automated fixes for routine issues, human review for anything that needs judgment. The system gets smarter about your specific quality standards over time.",
                "results": "Systems that improve autonomously between human review cycles. Routine maintenance handled, significant issues escalated with full context."
            }
        ]
    }'::jsonb,
    'manual', 'manual', 'manual',
    'Real case studies from framework build — framed for CTO/engineering audience'
);