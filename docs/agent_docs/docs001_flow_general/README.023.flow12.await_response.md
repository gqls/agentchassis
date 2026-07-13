
The Flow When await_response: true (call_agent - working correctly)

Line 855: Action handler completes, returns result with await_response: true
Line 916: Result is stored in CollectedData
Line 977: "Action requires waiting for response" log appears
Line 1005: Gets responses_topic from environment
Between lines 977-1005, the code checks await_response and if true, adds request_id to state.AwaitedRequests map
Line 644: "Execution paused - waiting for responses" - workflow stops here
Later, when child responds, ProcessResponse uses FindOrchestrationAwaitingRequest to find the parent
Parent workflow resumes

The Flow When await_response: false (spawn_agent - currently broken)

Line 855: Action handler completes, returns result with await_response: false
Line 916: Result is stored in CollectedData
Line 937: "Added agent to subtree"
Line 650: "currentStepConfig.NextStep was not blank" - immediately continues to next step
No waiting - workflow proceeds immediately
Child init responses arrive but can't find parent in AwaitedRequests, so they're ignored