// FILE: platform/orchestration/datahelpers/slot_pairing.go
//
// ONE canonical answer to the question the save path asks three times:
// "does this incoming section represent this stored row?"
//
// The three askers — save_page_sections' matchLockedRow (the 058 lock guard),
// MergeLockedPageSlots (LOCK-008's list merge), and matchPreservedSectionIdx
// (the Layer 2 interactive-tool carry) — each grew their own copy of this
// relation, and the copies drifted: matchLockedRow was given a component-
// identity arm (bugs 182/189/204), the merge was born mirroring it, and the
// Layer 2 matcher kept comparing slot-name strings until a build-arm rebuild
// "rescued" a locked calculator that was in the composition under its function
// name and duplicated it (bugs_open/385 §5c). The council's reuse seat gated
// the per-site fix on exactly that history (corr ece638fb): three copies of
// near-identical arm logic will drift again the next time one gains an arm.
//
// So the RELATION lives here, once, and the askers adapt to it.
// ⚠ A NEW pairing site MUST call PairIncomingToStored or PairStoredToIncoming
// rather than hand-rolling a comparison — a fourth private copy is precisely
// how 385 happened, and the arm it forgets will be the one that bites.
//
// THE ARMS, in priority order — each tried across ALL candidates before the
// next arm is considered (so a weaker match on an early candidate never beats
// a stronger match on a later one):
//   1. component IDENTITY — both sides carry a component id and they agree.
//      Names lie on decomposed pages (a positional slot_name is no component's
//      name); ids do not.
//   2. slot name, exact.
//   3. slot name, kebab-normalised (the bugs_open/041 naming landmine: older
//      rows and plans carry snake_case/CamelCase variants of one slot).
//   4. component function or component name against the incoming name — the
//      arm that pairs a plan's function-named entry with a positionally-named
//      stored row when the incoming side carries no id. matchLockedRow's
//      candidates deliberately leave these fields empty (its loader does not
//      join content_components), which makes this arm a structural no-op
//      there — stated at its call site, not silently assumed.
//
// Every arm is guarded on non-empty on BOTH sides: an empty id or name never
// pairs, or every unresolved section would claim the first idless row.
//
// Consumption is the caller's: both functions take a predicate for "already
// claimed" and return an index; marking the claim stays at the call site,
// because the three askers store consumption differently (a struct field, a
// bool slice, a map).

package datahelpers

// SlotIdentity is the stored side of the pairing relation: a page_components
// row (or its list-side projection) with every identity the arms can use.
// Leave a field empty and the arms that need it stand down.
type SlotIdentity struct {
	Slot              string // page_components.slot_name
	ComponentID       string // content_components id the row renders ('' if none)
	ComponentFunction string // that component's function ('' if not loaded)
	ComponentName     string // that component's name ('' if not loaded)
}

// IncomingSection is the incoming side: a proposed section's name and (when
// enrichment has resolved one) its component id.
type IncomingSection struct {
	Name        string
	ComponentID string
}

const slotPairArmCount = 4

// slotPairArmMatches reports whether incoming (name, componentID) pairs with
// stored on the given arm. The switch IS the relation — change it here and
// all three askers change together, which is the point.
func slotPairArmMatches(arm int, name, componentID string, stored SlotIdentity) bool {
	switch arm {
	case 1: // component identity
		return componentID != "" && stored.ComponentID != "" && stored.ComponentID == componentID
	case 2: // slot exact
		return name != "" && stored.Slot != "" && stored.Slot == name
	case 3: // slot kebab-normalised
		if name == "" || stored.Slot == "" {
			return false
		}
		norm := NormalizeComponentFunction(name)
		return norm != "" && NormalizeComponentFunction(stored.Slot) == norm
	case 4: // component function or name
		return name != "" && ((stored.ComponentFunction != "" && stored.ComponentFunction == name) ||
			(stored.ComponentName != "" && stored.ComponentName == name))
	}
	return false
}

// PairIncomingToStored pairs ONE incoming section against many stored rows:
// the lock guard's and the list merge's direction. Returns the index of the
// first unconsumed stored row the incoming pairs with — arms in priority
// order, candidates in slice order within each arm — or -1.
func PairIncomingToStored(name, componentID string, stored []SlotIdentity, consumed func(int) bool) int {
	for arm := 1; arm <= slotPairArmCount; arm++ {
		for i := range stored {
			if consumed != nil && consumed(i) {
				continue
			}
			if slotPairArmMatches(arm, name, componentID, stored[i]) {
				return i
			}
		}
	}
	return -1
}

// PairStoredToIncoming pairs ONE stored row against many incoming sections:
// the Layer 2 carry's direction ("is this stored tool already represented in
// what this save is about to write?"). Same relation, same arm order; only
// the iteration is inverted. Returns the index of the first unclaimed
// incoming section, or -1 — and -1 is the verdict that licenses a re-append,
// so an arm missing HERE is how bugs_open/385's duplicate was minted.
func PairStoredToIncoming(stored SlotIdentity, incoming []IncomingSection, claimed func(int) bool) int {
	for arm := 1; arm <= slotPairArmCount; arm++ {
		for i := range incoming {
			if claimed != nil && claimed(i) {
				continue
			}
			if slotPairArmMatches(arm, incoming[i].Name, incoming[i].ComponentID, stored) {
				return i
			}
		}
	}
	return -1
}
