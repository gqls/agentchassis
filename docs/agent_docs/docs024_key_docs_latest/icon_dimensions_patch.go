// ============================================================================
// generate_image_actions.go — Phase 2H icon dimensions fix
// ============================================================================
//
// Single edit to platform/orchestration/actions/generate_image_actions.go.
//
// Locate the `"icon"` entry in kindDefaults (current line 73-79). Replace
// the entire block with the version below. Everything else in the map
// (logo, hero, illustration, infographic) is unchanged.
//
// Why: SDXL v1.0 accepts only a fixed list of dimensions:
//   1024x1024, 1152x896, 896x1152, 1216x832, 832x1216,
//   1344x768, 768x1344, 1536x640, 640x1536
//
// 512x512 returned HTTP 400 `invalid_sdxl_v1_dimensions` during the
// first icon test (2026-05-14 orchestration 660a8c86). SDXL doesn't
// support sub-1024 native dimensions; if smaller icons are needed,
// downscale on the asset-deployer step rather than at generation.

// ── Old (lines 73-79) ──
//
//    "icon": {
//        NegativePrompt: "background, shadows, photorealistic, text, complex, busy",
//        CfgScale:       7,
//        Steps:          25,
//        Width:          512,    // <-- SDXL rejects: not in allowed list
//        Height:         512,    // <-- SDXL rejects: not in allowed list
//    },

// ── New (replacement) ──

	"icon": {
		// SDXL v1.0 only supports a fixed dimension whitelist (see
		// invalid_sdxl_v1_dimensions API error). 1024x1024 is the
		// smallest allowed square. Asset-deployer can downscale at
		// deploy time if smaller files are wanted on disk.
		NegativePrompt: "background, shadows, photorealistic, text, complex, busy",
		CfgScale:       7,
		Steps:          25,
		Width:          1024,
		Height:         1024,
	},
