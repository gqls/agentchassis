┌─────────────────────────────────────────────────────────┐
│ Initial Request Arrives                                  │
│ - ExecutionContext created from headers                 │
│ - Stored in state.CollectedData["__execution_context__"]│
└────────────────┬────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────┐
│ ExecuteWorkflow called                                  │
│ - Uses ExecutionContext from CollectedData              │
│ - Executes first step                                   │
└────────────────┬────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────┐
│ Step spawns/calls child                                 │
│ - State becomes AWAITING_RESPONSES                      │
│ - ExecutionContext still in CollectedData               │
└────────────────┬────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────┐
│ Child response arrives via ProcessResponse              │
│ - Brings child's ExecutionContext (response type)      │
│ - handleCompleteResponse called with child's context   │
└────────────────┬────────────────────────────────────────┘
│
▼
┌─────────────────────────────────────────────────────────┐
│ continueExecution called                                │
│ - Needs PARENT's ExecutionContext, not child's!        │
│ - Should get it from state.CollectedData?              │
│ - Or create fresh from state fields?                   │
└─────────────────────────────────────────────────────────┘


---


Excellent question! This is a fundamental conceptual point about **message flow direction** vs **execution mode**. Let me explain this carefully.

## The Confusion

You're thinking: "We just received a response, so shouldn't `MessageType` be 'response'?"

**The key insight:** `MessageType` indicates **what THIS agent is doing NOW**, not what just happened.

## The Mental Model

Think of it like a conversation between two people:

```
Parent: "Hey child, go do X" (sends REQUEST)
  ↓
Child: receives request, processes it
  ↓
Child: "Done! Here's the result" (sends RESPONSE)
  ↓
Parent: receives response, stores it
  ↓
Parent: NOW WHAT? → Parent needs to continue their OWN work
```

When the parent continues their own work, they're not "responding" - they're making new requests or executing new actions.

## The Flow in Detail

```
┌─────────────────────────────────────────────────────────────┐
│ Step 1: Parent calls child                                   │
├─────────────────────────────────────────────────────────────┤
│ Parent's ExecutionContext:                                   │
│   MessageType: "request"         ← Making a request          │
│   Action: "call_agent"                                       │
│   Sender: Parent                                             │
└────────────────────────┬────────────────────────────────────┘
                         │ Sends message
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 2: Child receives and processes                         │
├─────────────────────────────────────────────────────────────┤
│ Child's ExecutionContext:                                    │
│   MessageType: "request"         ← Received a request        │
│   Action: "process"                                          │
│   Sender: Child                                              │
│   InResponseTo: (parent's request)                           │
└────────────────────────┬────────────────────────────────────┘
                         │ Does work
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 3: Child sends response                                 │
├─────────────────────────────────────────────────────────────┤
│ Child's ExecutionContext:                                    │
│   MessageType: "response"        ← Sending a response        │
│   Sender: Child                                              │
│   InResponseTo: (parent's request)                           │
│   Status: "complete"                                         │
└────────────────────────┬────────────────────────────────────┘
                         │ Sends message
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 4: Parent receives response (handleCompleteResponse)    │
├─────────────────────────────────────────────────────────────┤
│ Incoming ExecutionContext: (from child)                      │
│   MessageType: "response"        ← Message we received       │
│   Sender: Child                                              │
│                                                              │
│ This context is ONLY used to:                                │
│   - Identify which request this responds to                  │
│   - Extract the response data                                │
│   - Store it in CollectedData                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Step 5: Parent continues workflow (continueExecution)        │
├─────────────────────────────────────────────────────────────┤
│ Parent's NEW ExecutionContext:                               │
│   MessageType: "request"         ← Making NEW actions        │
│   Action: "execute_llm_prompt" (next step)                   │
│   Sender: Parent                                             │
│   InResponseTo: null             ← Not responding anymore    │
│                                                              │
│ Why "request"?                                               │
│   - Parent is now the ACTIVE agent                           │
│   - Parent is executing ITS OWN workflow steps               │
│   - Parent might call MORE agents                            │
│   - Parent might execute actions                             │
│   - This is internal execution, not a response               │
└─────────────────────────────────────────────────────────────┘
```

## Real Example from Your Logs

Looking at your logs:

```json
// When content-creator-hero receives child's response:
"MessageType from ExecutionContext":"response"  // ← Child's message

// When content-creator-hero continues:
"exec_ctx_message_type":"request"  // ← Parent resuming
"exec_ctx_sender_type":"content-creator-hero"
"exec_ctx_in_response_to":null
```

## Why Not "response"?

If we set `MessageType: "response"` when continuing:

```go
// WRONG
freshExecCtx := &types.ExecutionContext{
    MessageType: "response",  // ← WRONG!
    Sender: Parent,
}
```

**Problems:**
1. **Who are we responding TO?** The parent isn't responding to the child - the child responded to the parent!
2. **Routing breaks:** Response messages route to `InResponseTo.ParentOrchestrationID`, but parent has no parent in this context
3. **executeStep expects request context:** Actions are designed to be executed in "request" mode
4. **Semantic confusion:** A response is a terminating message, but parent is continuing work

## The Correct Mental Model

**MessageType indicates the agent's current role:**

- `"request"` = "I am actively executing my workflow / making requests / doing work"
- `"response"` = "I am packaging up results to send back to my caller"

When the parent continues after receiving a response:
- ✅ Parent is actively executing (request mode)
- ❌ Parent is not responding to anyone (not response mode)

## Analogy: A Manager and Employee

```
Manager (Parent):        Employee (Child):
1. "File these reports"  
   [request mode]        2. Receives task
                            [request mode: working]
                         
                         3. "Done! Here they are"
                            [response mode: reporting back]

4. Receives files
   [temporarily in response handling]

5. "Now I'll review them and prepare presentation"
   [back to request mode: doing my own work]
```

The manager doesn't enter "response mode" just because they received a response. They go back to "request mode" to continue their own work.

## In Code Terms

```go
// Child sends response:
childContext.MessageType = "response"
childContext.InResponseTo = parentRequest
childContext.Status = "complete"

// Parent receives response:
func handleCompleteResponse(response) {
    // response.ExecutionContext.MessageType == "response" (child's)
    // Store the data
    state.CollectedData[stepName] = response.Body
    
    // Continue parent's work
    continueExecution(state)
}

// Parent continues:
func continueExecution(state) {
    freshContext.MessageType = "request"  // ← Parent doing work
    freshContext.InResponseTo = nil       // ← Not responding
    freshContext.Sender = Parent          // ← Parent is active
    
    executeStep(state, nextStep, freshContext)
}
```

## Summary

- **Response MessageType** = "I'm sending results back to someone who called me"
- **Request MessageType** = "I'm actively doing work (might be calling others, executing actions, etc.)"

When a parent continues after receiving a child's response, the parent is **resuming active work**, not **responding**. Hence: `MessageType = "request"`.

Does this clarify the distinction?