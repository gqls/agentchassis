Navigation structure:

Duplicate header — page_components contain a full <header> inside <main>, then InjectHeader adds another after <body>
Duplicate footer — same pattern, one inside <main>, one after </main>
Tool pages listed individually in primary nav instead of under a single "Tools" link to /tools.html or similar
Tool labels too long when they do appear ("LLM Provider Cost Comparison Calculator")
rerenderSimplifyNavLabel truncates long labels to first word ("AI Agent") — meaningless
Mismatched nav items between duplicate headers (7 vs 8 items)
Mismatched CTAs between duplicate headers ("Get Started" vs "Discuss Your Architecture")
max_header_items: 8 is too generous — leads to crowded nav

Styling:
9. Hover color #0f3460 on #1a1a2e background — nearly invisible
10. Responsive CSS injected 4 times by component-template-fixer (no idempotency check)
11. First footer has background: ; — empty value, renders transparent
12. Logo missing from injected header — logo_url not reaching RenderContext
    Data/content:
13. Placeholder email "agents@contactforsales.com"
14. Footer "Our Services" column has junk entries ("Tools / Password Strength Physics", nonexistent pricing page)
15. Stale pages from previous builds remain in nav (general problem, SyncPagesToDBAction doesn't deactivate)
    Discovery/fix loop:
16. component-template-fixer appends CSS without checking if already present
17. No discovery check detects duplicate headers inside <main>
18. addToolToNav adds tools to primary group — should add to a "tools" group or not add to header nav at all
