The initial request starts an orchestration with a wait step, which uses the await_approval action.

Step 1: AwaitApprovalAction Executes (First Time)

    File: internal/agents/orchestration/actions/await_approval.go

    Function: (a *AwaitApprovalAction) Execute

    Logic:

        The action checks step.Context.ApprovalStatus. It's not "approved" or "rejected".

        It checks step.Context.ApprovalSent. It's false.

        The action logs: "Sending approval request".

        It creates a new approvalRequestMessage with Action: "request_approval".

        Crucially, it sets ResponsesTopic: originalMessage.Headers.ResponsesTopic. This is the response topic of the original requestor (the kcat command), which is system.generic.responses.

        It sends this new message to the system.hitl.requests topic.

        It updates the step context: step.Context.ApprovalSent = true and step.Context.ApprovalStatus = "pending".

        It saves this updated step state via agent.GetStateManager().SaveStep().

        It returns false, indicating the step is not complete.

    Result: The orchestrator logs "Step not complete, awaiting further messages" (from orchestrator.go -> ProcessCurrentStep) and stops processing this workflow. It is now waiting.

Step 2: The Gap (Manual Approval)

    At this point, some external system (a "human in the loop" UI or service) is expected to be listening to the system.hitl.requests topic.

    This system would receive the request_approval message.

    A human would review it and send a new message (e.g., to the orchestrator's main request topic, system.agent.generic.requests) with an Action of "approved" or "rejected".

Step 3: Receiving the Approval Response

    File: internal/agents/orchestration/agent.go

    Function: (a *Agent) processMessage

    Logic:

        A new message arrives on the agent's listener topic (system.agent.generic.requests).

        processMessage checks the headers. It sees MessageType == "response" and Action == "approved".

        It logs: "Received approval response".

        It calls a.handleApprovalResponse(ctx, msg).

Step 4: Handling the Approval Response

    File: internal/agents/orchestration/agent.go

    Function: (a *Agent) handleApprovalResponse

    Logic:

        It loads the orchestrator state using the OrchestrationID from the approval message.

        It gets the waitingStepName from the approval message's step_name header.

        It compares this to the orchestrator's currentStep.Name. They should both be "wait".

        It updates the step context: currentStep.Context.ApprovalStatus = "approved".

        It saves the updated step state: a.stateManager.SaveStep(ctx, currentStep).

        It logs: "Triggering orchestrator re-evaluation after approval".

        It calls orchestrator.ProcessCurrentStep(ctx) to "wake up" the workflow.

Step 5: AwaitApprovalAction Executes (Second Time)

    File: internal/agents/orchestration/actions/await_approval.go

    Function: (a *AwaitApprovalAction) Execute

    Logic:

        The orchestrator re-evaluates the same step ("wait").

        This time, step.Context.ApprovalStatus is "approved".

        The action logs: "Approval already received, skipping wait".

        It returns true, indicating the step is now complete.

    Result: The orchestrator logs "Step completed" and calls o.AdvanceToNextStep(ctx, currentStep) to move to the "done" step.