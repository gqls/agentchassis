Complete Flow Sequence Diagram
Phase 1: Initial Request Processing

Client                Generic Agent                   Database              Kafka
|                        |                              |                   |
|--[Request]------------>|                              |                   |
|  action: calculate     |                              |                   |
|  first_calc: ...       |                              |                   |
|  second_calc: ...      |                              |                   |
|                        |                              |                   |
|                    [processRequests()]                |                   |
|                    [processMessage()]                 |                   |
|                    [ProcessMessage()]                 |                   |
|                    [process()]                        |                   |
|                    [loadAgentDefinition()]----------->|                   |
|                        |<---------[workflow config]---|                   |
|                    [executeWorkflow()]                |                   |
|                    [ExecuteWorkflow()]                |                   |
|                    [getOrCreateState()]-------------->|                   |
|                        |<-------[new state created]---|                   |
|                    [handleOrchestrationStatus()]      |                   |
|                    [continueExecution()]              |                   |


Phase 2: Spawn Agent Execution

Generic Agent           Database            Kafka              Calculator Agent
     |                        |                  |                      |
[executeStep()]               |                  |                      |
[executeLocalAction()]        |                  |                      |
[SpawnAgentAction()]          |                  |                      |
     |                        |                  |                      |
     |--[getAgentDefinition]->|                  |                      |
     |<--[calculator def]--   |                  |                      |
     |                        |                  |                      |
     |--[createAgentInDB]->   |                  |                      |
     |<-------[OK]---------   |                  |                      |
     |                        |                  |                      |
[spawnAgentKubernetesJob()]   |                  |                      |
     |--[Create K8s Job]---   ------------------>|                      |
     |                        |                  |                      |
     |--[spawn message]----   |----------------->|                      |
     |  to: calculator.requests                  |                      |
     |  request_id: fd1b94ed                     |                      |
     |  action: initialize                       |                      |
     |  responses_topic: generic.responses        |                      |
     |                        |                  |                      |
[Add to AwaitedRequests]      |                  |                      |
     |--[UpdateState]----->   |                  |                      |
     |  Status: AWAITING      |                  |                      |
     |<-------[OK]---------   |                  |                      |
     |                        |                  |                      |
[continueExecution returns]   |                  |                      |
[Workflow pauses]             |                  |                      |


Phase 3: Calculator Initialization

Calculator Pod         Kafka            Generic Agent         Database
     |                   |                    |                  |
[Pod Starts]             |                    |                  |
[Agent.Run()]            |                    |                  |
[processRequests()]      |                    |                  |
     |<--[consume]-------|                    |                  |
     |  initialize msg   |                    |                  |
     |                   |                    |                  |
[processMessage()]       |                    |                  |
[ProcessMessage()]       |                    |                  |
     |                   |                    |                  |
[SendInitializationResponse()]                |                  |
     |--[response]------>|                    |                  |
     |  to: generic.responses                 |                  |
     |  in_response_to_request_id: fd1b94ed   |                  |
     |  status: complete                      |                  |
     |  message_type: response                |                  |


Phase 4: Response Processing

Kafka                  Generic Agent              Database           Coordinator
|                          |                        |                   |
|--[response waiting]----->|                        |                   |
|  on generic.responses    |                        |                   |
|                          |                        |                   |
|                    [processResponses()]           |                   |
|                    [responseConsumer.Consume()]   |                   |
|                          |                        |                   |
|                    *** NOTHING HAPPENS ***        |                   |
|                    (Response not consumed?)       |                   |
|                          |                        |                   |
|                          |                        |                   |
[2 minutes pass...]        |                        |                   |
|                          |                        |                   |
|                    [handleRequestTimeout()]       |                   |
|                    [Request fd1b94ed timeout]     |                   |
|                    [Retry attempt]                |                   |

Expected Flow

Generic Agent           Coordinator              Database
     |                      |                        |
[processResponses()]        |                        |
[processMessage()]          |                        |
[ProcessMessage()]          |                        |
     |                      |                        |
     |--[ProcessResponse]-->|                        |
     |  execCtx.MessageType="response"               |
     |                      |                        |
     |              [FindByAwaitedRequestID]-------->|
     |                      |<---[found state]-------|
     |                      |                        |
     |              [handleCompleteResponse()]       |
     |              [RemoveAwaitedRequest()]-------->|
     |                      |                        |
     |              [state.CurrentStep = "first_calculation"]
     |              [continueExecution()]            |
     |              [executeStep("first_calculation")]
     |              [CallAgentAction()]              |
     |                      |                        |

