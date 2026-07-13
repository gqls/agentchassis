// FILE: platform/orchestration/actions/ch_vertical_profiles.go
// Vertical-specific configuration for CH matching heuristics.
// Add a new entry here when onboarding a new industry vertical.
// The matching, scoring, and collection actions read from this registry
// via GetCHVerticalProfile(slug).

package actions

// CHVerticalProfile holds industry-specific config for CH matching.
type CHVerticalProfile struct {
	Slug string

	// SIC codes to collect from CH advanced search
	SICCodes []string

	// Words in CH company names that signal the right industry.
	// Used for scoring bonus in postcode+name matching.
	IndustryKeywords []string

	// Bonus score for industry keyword in company title (pass 1)
	IndustryKeywordBonus float64

	// Words too common in this industry to count as "distinctive"
	// for the trigram name-only matching post-check.
	GenericWords map[string]bool

	// Suffixes to strip from business names before searching/matching.
	// Keeps industry terms (e.g. "Veterinary"), strips type words
	// (e.g. "Surgery", "Clinic") that vary between trading and registered names.
	NameStripSuffixes []string

	// Vertical slug used to filter businesses table
	BusinessVerticalSlug string
}

// chVerticalProfiles is the registry of all vertical profiles.
var chVerticalProfiles = map[string]*CHVerticalProfile{
	"veterinary": {
		Slug:                 "veterinary",
		SICCodes:             []string{"75000"},
		IndustryKeywords:     []string{"veterinary", "vet", "vets"},
		IndustryKeywordBonus: 0.10,
		GenericWords: map[string]bool{
			// Common English
			"the": true, "and": true, "of": true, "for": true, "in": true, "at": true, "a": true,
			// Industry terms
			"veterinary": true, "vet": true, "vets": true, "vets4pets": true,
			"animal": true, "pet": true, "pets": true, "paws": true,
			"equine": true, "farm": true, "emergency": true, "referrals": true,
			// Business type terms
			"mobile": true, "services": true, "service": true,
			"clinic": true, "centre": true, "center": true, "surgery": true,
			"practice": true, "hospital": true, "group": true,
			// Legal suffixes
			"limited": true, "ltd": true, "llp": true, "plc": true,
		},
		NameStripSuffixes: []string{
			" Limited", " Ltd", " LLP", " PLC", " Group",
			" Surgery", " Centre", " Center", " Clinic",
			" Hospital", " Practice",
		},
		BusinessVerticalSlug: "veterinary",
	},

	// Template for next vertical — uncomment and fill in:
	//
	// "seaweed-farming": {
	// 	Slug:                 "seaweed-farming",
	// 	SICCodes:             []string{"01210", "10200"},
	// 	IndustryKeywords:     []string{"seaweed", "algae", "kelp", "aquaculture"},
	// 	IndustryKeywordBonus: 0.10,
	// 	GenericWords: map[string]bool{
	// 		"the": true, "and": true, "of": true, "for": true, "in": true,
	// 		"farming": true, "farm": true, "marine": true, "aquaculture": true,
	// 		"services": true, "group": true, "limited": true, "ltd": true,
	// 	},
	// 	NameStripSuffixes: []string{
	// 		" Limited", " Ltd", " LLP", " PLC", " Group",
	// 		" Farm", " Farms", " Marine",
	// 	},
	// 	BusinessVerticalSlug: "seaweed-farming",
	// },
}

// GetCHVerticalProfile returns the profile for a vertical slug.
// Returns a sensible default if the slug is not found.
func GetCHVerticalProfile(slug string) *CHVerticalProfile {
	if p, ok := chVerticalProfiles[slug]; ok {
		return p
	}

	// Default profile — no industry-specific bonuses or word lists.
	// Matching will work on postcode + name similarity only.
	return &CHVerticalProfile{
		Slug:                 slug,
		SICCodes:             []string{},
		IndustryKeywords:     []string{},
		IndustryKeywordBonus: 0.0,
		GenericWords: map[string]bool{
			"the": true, "and": true, "of": true, "for": true, "in": true, "a": true,
			"services": true, "service": true, "group": true,
			"limited": true, "ltd": true, "llp": true, "plc": true,
		},
		NameStripSuffixes: []string{
			" Limited", " Ltd", " LLP", " PLC", " Group",
		},
		BusinessVerticalSlug: slug,
	}
}
