package types

// MessageTracer interface to avoid cyclic import
type MessageTracer interface {
	TraceMessage(execCtx interface{}, direction, topic string, payload interface{})
	TraceAwaitedSteps(execCtx *ExecutionContext, awaitedSteps []string, action string)
	TraceError(execCtx *ExecutionContext, err error, context string)
	DumpTrace(correlationID string)
}
