package site

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/besscroft/bophotos/backend/internal/repository"
)

func TestSettingsDefaultsAndHashedVisitLogging(t *testing.T) {
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := repository.Migrate(db); err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	settings, err := repo.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !settings.AnalyticsEnabled || settings.AnalyticsRetentionDays != 90 || !settings.PublicOriginalDownload {
		t.Fatalf("defaults = %#v", settings)
	}
	if err := repo.Log(context.Background(), "/", "home", "203.0.113.1", "test", ""); err != nil {
		t.Fatal(err)
	}
	var hash string
	if err := db.QueryRow(`SELECT ip_hash FROM visit_logs`).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if hash == "" || hash == "203.0.113.1" {
		t.Fatalf("ip_hash = %q", hash)
	}
	settings.AnalyticsEnabled = false
	if err := repo.Put(context.Background(), settings); err != nil {
		t.Fatal(err)
	}
	if err := repo.Log(context.Background(), "/", "home", "203.0.113.2", "test", ""); err != nil {
		t.Fatal(err)
	}
	var count int
	_ = db.QueryRow(`SELECT COUNT(*) FROM visit_logs`).Scan(&count)
	if count != 1 {
		t.Fatalf("visit count = %d, want 1", count)
	}
}

func TestAnalyticsIncludesHourlySourcesAndPages(t *testing.T) {
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := repository.Migrate(db); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Hour)
	visits := []struct {
		id, path, pageType, source string
		createdAt                  time.Time
	}{
		{"visit-1", "/gallery", "gallery", "direct", now.Add(10 * time.Minute)},
		{"visit-2", "/gallery/travel", "gallery", "search", now.Add(20 * time.Minute)},
		{"visit-3", "/guides", "guide", "direct", now.Add(-time.Hour)},
	}
	for _, visit := range visits {
		_, err = db.Exec(
			`INSERT INTO visit_logs(id,path,page_type,ip_hash,user_agent,referrer,source,created_at) VALUES(?,?,?,?,?,?,?,?)`,
			visit.id,
			visit.path,
			visit.pageType,
			visit.id+"-hash",
			"test",
			"",
			visit.source,
			visit.createdAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	result, err := NewRepository(db).Analytics(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	hourly, ok := result["hourly"].([]Point)
	if !ok {
		t.Fatalf("hourly aggregate missing: %#v", result)
	}
	if len(hourly) != 24 {
		t.Fatalf("hourly points = %d, want 24", len(hourly))
	}
	if got := namedCount(result["sources"].([]NamedCount), "direct"); got != 2 {
		t.Fatalf("direct source count = %d, want 2", got)
	}
	if got := namedCount(result["pages"].([]NamedCount), "gallery"); got != 2 {
		t.Fatalf("gallery page count = %d, want 2", got)
	}
}

func TestDashboardUsesConfiguredAnalyticsTimezone(t *testing.T) {
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := repository.Migrate(db); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	settings, err := repo.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	settings.AnalyticsTimezone = "Asia/Shanghai"
	if err := repo.Put(context.Background(), settings); err != nil {
		t.Fatal(err)
	}
	repo.now = func() time.Time {
		return time.Date(2026, 6, 23, 20, 0, 0, 0, time.UTC)
	}

	for _, visit := range []struct {
		id        string
		createdAt time.Time
	}{
		{"local-today", time.Date(2026, 6, 23, 17, 0, 0, 0, time.UTC)},
		{"local-yesterday", time.Date(2026, 6, 23, 15, 0, 0, 0, time.UTC)},
	} {
		_, err = db.Exec(
			`INSERT INTO visit_logs(id,path,page_type,ip_hash,user_agent,referrer,source,created_at) VALUES(?,?,?,?,?,?,?,?)`,
			visit.id,
			"/",
			"home",
			visit.id+"-hash",
			"test",
			"",
			"direct",
			visit.createdAt.Format(time.RFC3339Nano),
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	dashboard, err := repo.Dashboard(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.VisitsToday != 1 || dashboard.VisitsYesterday != 1 {
		t.Fatalf("today/yesterday = %d/%d, want 1/1", dashboard.VisitsToday, dashboard.VisitsYesterday)
	}
}

func TestSettingsRejectInvalidAnalyticsTimezone(t *testing.T) {
	db, err := repository.Open(context.Background(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := repository.Migrate(db); err != nil {
		t.Fatal(err)
	}

	repo := NewRepository(db)
	settings, err := repo.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	settings.AnalyticsTimezone = "Mars/Olympus_Mons"
	if err := repo.Put(context.Background(), settings); err == nil {
		t.Fatal("Put() accepted an invalid analytics timezone")
	}
}

func namedCount(values []NamedCount, name string) int {
	for _, value := range values {
		if value.Name == name {
			return value.Count
		}
	}
	return 0
}
