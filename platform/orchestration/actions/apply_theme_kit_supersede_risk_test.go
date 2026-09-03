package actions

import (
	"context"
	"errors"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// classifierDesignIntentState is the guard the council gate asked for twice
// (correlation bed139b2, rounds 2 and 3). It marks an apply whose design_intent
// write will be SUPERSEDED by domain-research-classifier on the fresh path —
// see bugs_open/438 §6d. These tests exist because the failure mode is SILENCE:
// nothing errors, the write lands, and the values are gone later. A guard whose
// only evidence is "it compiles" is exactly the shape this lane keeps logging.
//
// PROVEN BY MUTATION 2026-09-03 — a passing test is not evidence until it has
// been shown it can fail:
//
//	mutation                                              result
//	classifierDesignIntentState always returns            FAIL — "state =
//	  riskClassifierAlreadyWrote (the guard goes blind)      classifier_already_wrote,
//	                                                        want at_risk_…"
//	the err!=nil arm returns riskClassifierAlreadyWrote    FAIL — "want unknown"
//	  instead of riskUnknown (a confident wrong answer)
//	restored                                              PASS
//
// The SQL predicate is separately checked against live data: [MEASURED
// 2026-09-03] of 39 sites carrying a design_intent, 38 have one written by
// domain-research-classifier — so it is true on an established site and false
// on a fresh one, which is the distinction the marker has to draw. That number
// is the disconfirmable half: had it been 39 or 0, the predicate would not
// discriminate and this guard would be decoration.

func TestClassifierDesignIntentState_AllThreeStates(t *testing.T) {
	const q = "SELECT count\\(\\*\\) FROM site_specs"
	siteID := uuid.New()

	cases := []struct {
		name  string
		rows  func(m sqlmock.Sqlmock)
		want  string
		whyIt string
	}{
		{
			name: "classifier has never written — the fresh path, values WILL be superseded",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(q).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			want:  riskNoClassifierWriteYet,
			whyIt: "this is the case the guard exists for",
		},
		{
			name: "classifier has written before — the fresh-path certainty is gone",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(q).WithArgs(siteID).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
			},
			want:  riskClassifierAlreadyWrote,
			whyIt: "must NOT be reported as at_risk, or the warning stops discriminating",
		},
		{
			name: "read fails — must be UNKNOWN, never a confident no-risk",
			rows: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(q).WithArgs(siteID).WillReturnError(errors.New("boom"))
			},
			want:  riskUnknown,
			whyIt: "recording a false 'no risk' is the false-structured-fact class this lane has already fixed twice",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock: %v", err)
			}
			defer db.Close()

			mock.ExpectBegin()
			tc.rows(mock)

			tx, err := db.Begin()
			if err != nil {
				t.Fatalf("begin: %v", err)
			}
			got := classifierDesignIntentState(context.Background(), tx, siteID)
			if got != tc.want {
				t.Fatalf("state = %q, want %q — %s", got, tc.want, tc.whyIt)
			}
		})
	}
}

// The three markers must be DISTINCT strings. A copy-paste that made two of them
// equal would make the guard silently unable to report the case it exists for,
// and every test above would still pass on the collapsed pair.
func TestSupersedeRiskConstantsAreDistinct(t *testing.T) {
	all := map[string]string{
		"riskNoClassifierWriteYet":   riskNoClassifierWriteYet,
		"riskClassifierAlreadyWrote": riskClassifierAlreadyWrote,
		"riskUnknown":                riskUnknown,
	}
	seen := map[string]string{}
	for name, val := range all {
		if val == "" {
			t.Fatalf("%s is empty — an empty marker reads as an absent field in the adoption spec", name)
		}
		if prev, dup := seen[val]; dup {
			t.Fatalf("%s and %s are both %q — the guard cannot distinguish the states it reports", prev, name, val)
		}
		seen[val] = name
	}
	// The at-risk marker must SAY it is a risk. It is read by a human staring at
	// a theme_kit_adoption row, not only by code comparing constants.
	if !strings.Contains(riskNoClassifierWriteYet, "at_risk") {
		t.Fatalf("riskNoClassifierWriteYet = %q — it must be self-describing in the spec row", riskNoClassifierWriteYet)
	}
	if strings.Contains(riskClassifierAlreadyWrote, "at_risk") || strings.Contains(riskUnknown, "at_risk") {
		t.Fatal("only the at-risk state may contain 'at_risk' — a grep for it must not match the safe states")
	}
}
