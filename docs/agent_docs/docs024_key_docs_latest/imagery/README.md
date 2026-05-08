Mental model for the lock-expiry project when picked up later:

One migration adds lock_type + lock_expires_at to all four Pattern A tables (page_components, site_components, site_plan_directives, assets) at once.
Auto-lock writers default to the policy table — 'admin' permanent, 'deploy' timed/30, auditor-approvals timed/90.
Filter expansion across ~8-10 callsites: locked_at IS NULL → (locked_at IS NULL OR (lock_type = 'timed' AND lock_expires_at < NOW())).
CheckComponentLock extended to return LockType and LockExpiresAt.
New expired_review_locks discovery check creates HITL items for 'review'-type locks past expiry.

