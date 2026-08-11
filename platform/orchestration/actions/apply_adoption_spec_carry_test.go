package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// carryForwardStructureSpecKeys (bugs_open/241 round-2 REVISE; LANDMINES
// 2026-08-11 "Re-adopting a site silently drops the structure spec's opt-in
// flags"). Adoption supersede+INSERTs a fresh marshal of the structure aspect,
// so operator-seeded keys it does not write (url_shape, the PLAN-048 identity
// gates) vanished on re-adoption. These pin the carry: unknown keys survive,
// adoption's own keys win, and every failure arm returns the fresh data
// unchanged (the pre-fix behaviour).

func TestCarryForwardStructureSpecKeys_OptInFlagsSurviveReadoption(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).
			AddRow([]byte(`{"pages":["old-a","old-b"],"source":"adoption","adopted_from":"x.co.uk","url_shape":"flat","honour_realised_identity":true}`)))

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	fresh := map[string]interface{}{
		"pages":        []interface{}{"new-a"},
		"source":       "adoption",
		"adopted_from": "x.co.uk",
	}
	out := carryForwardStructureSpecKeys(context.Background(), tx, uuid.New(), fresh, zap.NewNop())

	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("result is not a map: %T", out)
	}
	if m["url_shape"] != "flat" {
		t.Errorf("url_shape dropped on re-adoption: %v", m["url_shape"])
	}
	if m["honour_realised_identity"] != true {
		t.Errorf("honour_realised_identity dropped on re-adoption: %v", m["honour_realised_identity"])
	}
	pages, _ := m["pages"].([]interface{})
	if len(pages) != 1 || pages[0] != "new-a" {
		t.Errorf("adoption's own pages list must win, got %v", m["pages"])
	}
}

func TestCarryForwardStructureSpecKeys_FirstAdoptionUnchanged(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"})) // no current row

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	fresh := map[string]interface{}{"pages": []interface{}{"a"}, "source": "adoption"}
	out := carryForwardStructureSpecKeys(context.Background(), tx, uuid.New(), fresh, zap.NewNop())
	m, _ := out.(map[string]interface{})
	if len(m) != 2 {
		t.Errorf("first adoption must pass fresh data through unchanged, got %v", m)
	}
}

func TestCarryForwardStructureSpecKeys_UnreadableCurrentFailsOpen(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT data FROM site_specs").
		WillReturnRows(sqlmock.NewRows([]string{"data"}).AddRow([]byte(`not json`)))

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	fresh := map[string]interface{}{"pages": []interface{}{"a"}}
	out := carryForwardStructureSpecKeys(context.Background(), tx, uuid.New(), fresh, zap.NewNop())
	m, _ := out.(map[string]interface{})
	if len(m) != 1 {
		t.Errorf("unreadable current row must fail open to fresh data, got %v", m)
	}
}
