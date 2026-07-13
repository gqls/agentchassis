package messaging

import "github.com/gqls/agentchassis/platform/orchestration/types"

// Initializer defines the contract for an object that can send an
// agent initialization confirmation response.
type Initializer interface {
	SendInitializationResponse(spawnRequest *types.RequestMessage) error
}
