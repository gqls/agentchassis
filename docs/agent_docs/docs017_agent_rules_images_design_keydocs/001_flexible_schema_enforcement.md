This is a thoughtful architectural point. Let me think through the different modes and when each applies:

## The Two Worlds

| Mode | Use Case | Behavior |
|------|----------|----------|
| **Flexible** | Initial build, creative exploration | LLM can generate any structure; renderer does best-effort substitution |
| **Strict** | Editing, maintenance, approved designs | Must match component's `input_schema` exactly; fail if mismatch |

## When Each Mode Applies

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ INITIAL BUILD (flexible)                                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│ 1. Domain comes in → Site created                                           │
│ 2. LLM analyzes industry, generates content plan                            │
│ 3. Components selected (or LLM generates structure)                         │
│ 4. Content generated with best-effort rendering                             │
│ 5. Human reviews, approves                                                  │
│                                                                             │
│ At approval: SNAPSHOT the schema + content structure                        │
│              Site moves to "strict" mode                                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ EDITING MODE (strict)                                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│ - page_components records have locked schema_snapshot                       │
│ - Edits must provide all required fields                                    │
│ - Component template upgrades don't break existing pages                    │
│ - Clear validation errors if content doesn't match                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Proposed Schema

```sql
-- On sites table
ALTER TABLE sites ADD COLUMN IF NOT EXISTS
    schema_mode TEXT DEFAULT 'flexible';
COMMENT ON COLUMN sites.schema_mode IS 
    'flexible: best-effort rendering; strict: enforce input_schema validation';

-- On page_components table (per-section granularity)
ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    schema_snapshot JSONB;
COMMENT ON COLUMN page_components.schema_snapshot IS 
    'Locked input_schema at approval time - edits must match this';

ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    component_version_id UUID;
COMMENT ON COLUMN page_components.component_version_id IS 
    'Specific component version this was built with (for template consistency)';

ALTER TABLE page_components ADD COLUMN IF NOT EXISTS
    content_snapshot JSONB;
COMMENT ON COLUMN page_components.content_snapshot IS 
    'The actual content values used - for edit comparison and rollback';
```

## How Rendering Would Work

```go
type RenderOptions struct {
    SchemaMode      string                 // "flexible" or "strict"
    SchemaSnapshot  map[string]interface{} // Locked schema (if strict)
    ContentSnapshot map[string]interface{} // Previous content (for diff)
}

func RenderComponentWithOptions(
    template string,
    ctx *RenderContext,
    contentData map[string]interface{},
    opts RenderOptions,
    logger *zap.Logger,
) (string, error) {
    
    if opts.SchemaMode == "strict" && opts.SchemaSnapshot != nil {
        // Validate content matches schema
        if err := validateAgainstSchema(contentData, opts.SchemaSnapshot); err != nil {
            return "", fmt.Errorf("content doesn't match locked schema: %w", err)
        }
    }
    
    // Merge content into render context
    mergeContentIntoContext(ctx, contentData)
    
    // Convert to template data
    data := contextToMap(ctx)
    
    if opts.SchemaMode == "strict" {
        // Check all required fields are present
        if required, ok := opts.SchemaSnapshot["required"].([]interface{}); ok {
            for _, field := range required {
                fieldName := field.(string)
                if _, exists := data[fieldName]; !exists {
                    return "", fmt.Errorf("required field missing: %s", fieldName)
                }
            }
        }
    }
    
    // Render template
    result := template
    result = renderGoStyleSubstitutions(result, data)
    result = renderHandlebarsSubstitutions(result, data)
    
    // In flexible mode, warn about unsubstituted placeholders
    // In strict mode, fail on unsubstituted placeholders
    if hasUnsubstitutedPlaceholders(result) {
        if opts.SchemaMode == "strict" {
            return "", fmt.Errorf("template has unsubstituted placeholders")
        }
        logger.Warn("Flexible mode: template has unsubstituted placeholders",
            zap.String("preview", result[:min(200, len(result))]))
    }
    
    return result, nil
}
```

## The Approval Transition

When a page is approved (HITL or auto-eval), lock in the structure:

```go
func ApprovePageComponent(ctx context.Context, db *sql.DB, componentID uuid.UUID, contentData map[string]interface{}) error {
    // Get the component's current input_schema
    var inputSchema []byte
    err := db.QueryRowContext(ctx, `
        SELECT cc.input_schema 
        FROM page_components pc
        JOIN content_components cc ON pc.component_id = cc.id
        WHERE pc.id = $1
    `, componentID).Scan(&inputSchema)
    if err != nil {
        return err
    }
    
    // Lock in schema + content at approval time
    _, err = db.ExecContext(ctx, `
        UPDATE page_components SET
            build_status = 'approved',
            schema_snapshot = $2,
            content_snapshot = $3,
            reviewed_at = now()
        WHERE id = $1
    `, componentID, inputSchema, contentData)
    
    return err
}
```

## Benefits

| Aspect | Flexible Mode | Strict Mode |
|--------|---------------|-------------|
| **Validation** | Warn only | Fail on mismatch |
| **Missing fields** | Use defaults, leave placeholders | Error |
| **Component updates** | Uses latest template | Uses locked version |
| **Rollback** | N/A | Can restore from content_snapshot |
| **Editing UI** | Freeform | Schema-driven form |

## Questions for You

1. **Granularity**: Should strict mode be at site level, page level, or section level? (I suggested section level for maximum flexibility)

2. **Component versioning**: When we update a component template, should we create a new version record, or is the schema_snapshot enough to preserve the contract?

3. **Transition trigger**: What marks the transition from flexible → strict?
    - Human approval in HITL?
    - First successful deploy?
    - Explicit "lock design" action?

4. **Override capability**: Should there be a way to "unlock" a strict page for redesign while preserving content?

This feels like the right architectural direction - it separates the "creative building" phase from the "maintenance" phase and gives you clear contracts for editing without breaking existing designs.



