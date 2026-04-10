Unresolved mechanism
Located in load_work_item_actions.go line ~893, in insertWorkItem:

Discovery check finds an issue, calls insertWorkItem
Action queries for terminal items (complete/failed) with same item_key in last 7 days
If newest terminal item < 3 hours old → suppress entirely (don't create)
If 2+ terminal items exist → create item with status unresolved and prefix summary with "[unresolved after N attempts]"
unresolved items are not dispatched — they sit visible for investigation

This correctly catches handlers that "succeed" without fixing the underlying issue. The attempt_count = 0 is expected — these are new items born as unresolved, not retried items.

