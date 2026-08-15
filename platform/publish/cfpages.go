// FILE: platform/publish/cfpages.go
//
// Cloudflare Pages Direct Upload — the ruled primary backend (PLAN
// 2026-08-14 Part 2d) — NOT YET ARMED. This is a deliberate, loud gap, not
// an oversight:
//
// The Direct Upload protocol is multi-step and partly undocumented (project
// ensure → upload-token JWT → assets/check-missing → assets/upload →
// deployment create with a hash manifest; wrangler is the reference client),
// and no CF_PAGES_API_TOKEN exists yet — minting one is an owner key
// (PLAN §"owner's keys"). A protocol client written blind, with no API to
// verify against, ships plausible-but-unproven code behind an APPROVED
// verdict: exactly the trap the register records against MDL-040 ("a
// capability with no live caller has an untested dependency on its
// ENVIRONMENT"). So the backend registers, refuses loudly, and gets built
// verify-as-you-go the day the token lands in personae-platform-secrets —
// at which point the spawner also needs to inject it for site-publisher
// pods (platform/orchestration/actions/spawn_actions.go, the storage-env
// block), which is part of the same arming step.
//
// Everything around it — the seam, drift detection, the reconciler, and the
// caller's served-bytes acceptance — is backend-agnostic and already proven
// via b2worker, so arming cfpages changes this file and nothing else.
package publish

import (
	"context"
	"fmt"
)

type CFPages struct{}

func NewCFPages() *CFPages { return &CFPages{} }

func (c *CFPages) Name() string { return TargetCFPages }

func (c *CFPages) Publish(ctx context.Context, req Request) (Result, error) {
	return Result{}, fmt.Errorf(
		"cfpages: backend not yet armed — CF_PAGES_API_TOKEN is an owner key that does not exist yet; " +
			"the Direct Upload client is built and live-verified together once it does (see cfpages.go header). " +
			"Use publish_target='b2worker' until then")
}
