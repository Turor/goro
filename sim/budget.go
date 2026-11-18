package sim

import "time"

// CT119 budget targets from the goro port plan (Phase 5.3). Override them
// with environment variables in scripts/release/budget-gate.sh when a
// slower machine runs the opt-in gate. Default `go test` does not enforce
// these; it follows the 10s lean-suite policy (Phase 5.4).
const (
	CleanBuildBudget      = 15 * time.Second
	TestSuiteBudget       = 15 * time.Second
	HeadlessLoginBudget   = 3 * time.Second
	LeanSuitePolicyBudget = 10 * time.Second
)

const (
	EnvCleanBuildBudget    = "CLEAN_BUILD_BUDGET_S"
	EnvTestSuiteBudget     = "TEST_SUITE_BUDGET_S"
	EnvHeadlessLoginBudget = "HEADLESS_LOGIN_BUDGET_S"
)
