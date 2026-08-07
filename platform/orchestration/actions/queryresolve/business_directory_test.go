// FILE: platform/orchestration/actions/queryresolve/business_directory_test.go
//
// bugs_open/206: pins the escaping contract (this package's templates don't
// auto-escape, same class of risk news_items.go's tests already pin) and the
// two DB-touching behaviours that matter — no config for this site yields
// empty+no error rather than a fabricated/default vertical, and a configured
// site's query carries the vertical it actually configured.

package queryresolve

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestProjectBusinessDirectoryRowEscapes(t *testing.T) {
	row := projectBusinessDirectoryRow(
		`Acme <script>alert(1)</script> Vets`,
		`AB1 2CD`,
		`Town & County`,
		`https://example.com/?a=1&b=2`,
		true,
	)
	for field, v := range map[string]string{
		"name":     row["name"].(string),
		"location": row["location"].(string),
		"website":  row["website"].(string),
	} {
		if strings.ContainsAny(v, "<>") {
			t.Errorf("%s not escaped: %q", field, v)
		}
	}
	if !strings.Contains(row["name"].(string), "&lt;script&gt;") {
		t.Errorf("name should carry escaped markup, got %q", row["name"])
	}
	if row["is_claimed"] != true {
		t.Errorf("is_claimed = %v, want true", row["is_claimed"])
	}
}

func TestProjectBusinessDirectoryRowHandlesEmptyLocation(t *testing.T) {
	row := projectBusinessDirectoryRow("Acme Vets", "AB1 2CD", "", "https://example.com", false)
	if row["location"] != "" {
		t.Errorf("empty location should stay empty, got %q", row["location"])
	}
}

func TestResolveBusinessDirectory_NoConfig_ReturnsEmptyNotError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT st.input_data->>'vertical'").
		WithArgs(siteID).
		WillReturnError(sql.ErrNoRows)

	got, err := resolveBusinessDirectory(context.Background(), db, siteID, 0, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items, ok := got.([]map[string]interface{})
	if !ok || len(items) != 0 {
		t.Fatalf("expected empty slice for a site with no exporter config, got %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestResolveBusinessDirectory_UsesSitesOwnVertical(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT st.input_data->>'vertical'").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"vertical", "business_type_ilike"}).
			AddRow("veterinary", "%vet%"))

	rows := sqlmock.NewRows([]string{"name", "postcode", "location", "website_url", "is_claimed"}).
		AddRow("Acme Vets", "AB1 2CD", "Anytown, Anyshire", "https://acmevets.example", false)
	mock.ExpectQuery("FROM business_intel.businesses").
		WithArgs("veterinary", "%vet%", 60).
		WillReturnRows(rows)

	got, err := resolveBusinessDirectory(context.Background(), db, siteID, 0, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	items := got.([]map[string]interface{})
	if len(items) != 1 || items[0]["name"] != "Acme Vets" {
		t.Fatalf("items = %#v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestResolveBusinessDirectory_CapsLimit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT st.input_data->>'vertical'").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"vertical", "business_type_ilike"}).
			AddRow("veterinary", ""))

	mock.ExpectQuery("FROM business_intel.businesses").
		WithArgs("veterinary", 100). // requested 500, hard cap clamps to 100
		WillReturnRows(sqlmock.NewRows([]string{"name", "postcode", "location", "website_url", "is_claimed"}))

	if _, err := resolveBusinessDirectory(context.Background(), db, siteID, 500, zap.NewNop()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (limit not clamped as expected): %v", err)
	}
}
