// Package sim is the Phase 5 simulation harness: golden packet replay,
// opt-in live scenario checks, and CT119 budget constants.
//
// The default `go test` suite stays offline and lean. Live TEST runs need
// GORO_LIVE_TEST=1 and a kgap pool account file. Slow and real-data tests
// use the `slow` or `live` build tags.
package sim
