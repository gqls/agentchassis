Content Validation & Review Flow
Complete Flow

page-content-writer generates HTML
│
▼
┌─────────────────────────────────────────────────────────────┐
│                    CONTENT-REVIEWER                          │
│                                                              │
│  validate_content ─────────────────────────────────────────┐ │
│      │                                                     │ │
│      ▼ validation_result: {issues, error_count}            │ │
│                                                            │ │
│  determine_review_mode ────────────────────────────────────┤ │
│      │                    │                                │ │
│      ▼                    ▼                                │ │
│  auto-eval            HITL mode                            │ │
│      │                    │                                │ │
│      │                    ▼                                │ │
│      │           prepare_hitl_review                       │ │
│      │           (formats issues for display)              │ │
│      │                    │                                │ │
│      ▼                    ▼                                │ │
│  check_auto_approval  request_human_review                 │ │
│  (errors → HITL)      (can edit HTML, see issues)          │ │
│      │                    │                                │ │
│      └────────────────────┴──────────────────┐             │ │
│                                              │             │ │
│                                              ▼             │ │
│                                    process_human_response  │ │
│                                              │             │ │
│                                              ▼             │ │
│                                    check_rejection         │ │
│                                        /        \          │ │
│                                approved        rejected    │ │
│                                   │               │        │ │
│                                   ▼               ▼        │ │
│                           finalize_hitl    mark_page       │ │
│                           (use edits)     needs_attention  │ │
└──────────────────────────────────┼───────────────┼─────────┘
                                    │               │
                                    ▼               ▼
                                    approved=true    approved=false
                                    │               │
                                    └───────┬───────┘
                                            │
        ┌─────────────────────────────────┘
        │
        ▼
        ┌────────────────────────────────────────────────────────────┐
        │                    PAGEFLOW-BUILDER                         │
        │                                                             │
        │  check_review_approved ────────────────────────────────────┤
        │      │                    │                                │
        │   approved             rejected                            │
        │      │                    │                                │
        │      ▼                    ▼                                │
        │  assemble_page      complete_page                          │
        │      │              (skip this page)                       │
        │      ▼                                                     │
        │  deploy_page                                               │
        │      │                                                     │
        │      ▼                                                     │
        │  update_page_status                                        │
        │      │                                                     │
        │      ▼                                                     │
        │  complete_page ──────→ next page in loop                   │
        └────────────────────────────────────────────────────────────┘