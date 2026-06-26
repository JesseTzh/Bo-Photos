package site

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	apiresponse "github.com/besscroft/bophotos/backend/internal/api"
	"github.com/go-chi/chi/v5"
)

type Settings struct {
	SiteTitle              string   `json:"site_title"`
	SiteAuthor             string   `json:"site_author"`
	SiteFaviconURL         string   `json:"site_favicon_url"`
	AboutIntro             string   `json:"about_intro"`
	AboutInstagram         string   `json:"about_social_instagram"`
	AboutXiaohongshu       string   `json:"about_social_xiaohongshu"`
	AboutWeibo             string   `json:"about_social_weibo"`
	AboutGithub            string   `json:"about_social_github"`
	AboutGalleryAssetIDs   []string `json:"about_gallery_asset_ids"`
	GalleryLayout          string   `json:"gallery_layout"`
	PublicOriginalDownload bool     `json:"public_original_download"`
	AdminImagesPerPage     int      `json:"admin_images_per_page"`
	MaxUploadFiles         int      `json:"max_upload_files"`
	PreviewQuality         int      `json:"preview_quality"`
	PreviewMaxWidth        int      `json:"preview_max_width"`
	AnalyticsEnabled       bool     `json:"analytics_enabled"`
	AnalyticsRetentionDays int      `json:"analytics_retention_days"`
	AnalyticsTimezone      string   `json:"analytics_timezone"`
}

func Defaults() Settings {
	return Settings{SiteTitle: "BoPhoto", GalleryLayout: "grid", PublicOriginalDownload: true, AdminImagesPerPage: 20, MaxUploadFiles: 5, PreviewQuality: 80, PreviewMaxWidth: 2560, AnalyticsEnabled: true, AnalyticsRetentionDays: 90, AnalyticsTimezone: "Asia/Shanghai"}
}

type Repository struct {
	db  *sql.DB
	now func() time.Time
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db, now: time.Now} }
func (r *Repository) currentTime() time.Time {
	if r.now == nil {
		return time.Now()
	}
	return r.now()
}
func (r *Repository) Get(ctx context.Context) (Settings, error) {
	s := Defaults()
	rows, err := r.db.QueryContext(ctx, `SELECT config_key,COALESCE(config_value,'') FROM configs`)
	if err != nil {
		return s, err
	}
	defer rows.Close()
	m := map[string]string{}
	for rows.Next() {
		var k, v string
		if rows.Scan(&k, &v) == nil {
			m[k] = v
		}
	}
	s.SiteTitle = value(m, "site_title", s.SiteTitle)
	s.SiteAuthor = m["site_author"]
	s.SiteFaviconURL = m["site_favicon_url"]
	s.AboutIntro = m["about_intro"]
	s.AboutInstagram = m["about_social_instagram"]
	s.AboutXiaohongshu = m["about_social_xiaohongshu"]
	s.AboutWeibo = m["about_social_weibo"]
	s.AboutGithub = m["about_social_github"]
	_ = json.Unmarshal([]byte(value(m, "about_gallery_asset_ids", "[]")), &s.AboutGalleryAssetIDs)
	s.GalleryLayout = value(m, "gallery_layout", s.GalleryLayout)
	s.PublicOriginalDownload = boolValue(m, "public_original_download", s.PublicOriginalDownload)
	s.AdminImagesPerPage = intValue(m, "admin_images_per_page", s.AdminImagesPerPage)
	s.MaxUploadFiles = intValue(m, "max_upload_files", s.MaxUploadFiles)
	s.PreviewQuality = intValue(m, "preview_quality", s.PreviewQuality)
	s.PreviewMaxWidth = intValue(m, "preview_max_width", s.PreviewMaxWidth)
	s.AnalyticsEnabled = boolValue(m, "analytics_enabled", s.AnalyticsEnabled)
	s.AnalyticsRetentionDays = intValue(m, "analytics_retention_days", s.AnalyticsRetentionDays)
	s.AnalyticsTimezone = value(m, "analytics_timezone", s.AnalyticsTimezone)
	return s, rows.Err()
}
func (r *Repository) Put(ctx context.Context, s Settings) error {
	if s.GalleryLayout != "grid" && s.GalleryLayout != "single" {
		return errors.New("invalid gallery layout")
	}
	if s.AdminImagesPerPage < 1 || s.AdminImagesPerPage > 200 || s.MaxUploadFiles < 1 || s.MaxUploadFiles > 100 || s.PreviewQuality < 1 || s.PreviewQuality > 100 || s.PreviewMaxWidth < 320 || s.PreviewMaxWidth > 16384 || s.AnalyticsRetentionDays < 1 || s.AnalyticsRetentionDays > 3650 {
		return errors.New("setting out of range")
	}
	if _, err := time.LoadLocation(s.AnalyticsTimezone); err != nil {
		return errors.New("invalid analytics timezone")
	}
	gallery, _ := json.Marshal(s.AboutGalleryAssetIDs)
	values := map[string]string{"site_title": s.SiteTitle, "site_author": s.SiteAuthor, "site_favicon_url": s.SiteFaviconURL, "about_intro": s.AboutIntro, "about_social_instagram": s.AboutInstagram, "about_social_xiaohongshu": s.AboutXiaohongshu, "about_social_weibo": s.AboutWeibo, "about_social_github": s.AboutGithub, "about_gallery_asset_ids": string(gallery), "gallery_layout": s.GalleryLayout, "public_original_download": strconv.FormatBool(s.PublicOriginalDownload), "admin_images_per_page": strconv.Itoa(s.AdminImagesPerPage), "max_upload_files": strconv.Itoa(s.MaxUploadFiles), "preview_quality": strconv.Itoa(s.PreviewQuality), "preview_max_width": strconv.Itoa(s.PreviewMaxWidth), "analytics_enabled": strconv.FormatBool(s.AnalyticsEnabled), "analytics_retention_days": strconv.Itoa(s.AnalyticsRetentionDays), "analytics_timezone": s.AnalyticsTimezone}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := r.currentTime().UTC().Format(time.RFC3339Nano)
	for k, v := range values {
		_, err = tx.ExecContext(ctx, `INSERT INTO configs(config_key,config_value,created_at,updated_at) VALUES(?,?,?,?) ON CONFLICT(config_key) DO UPDATE SET config_value=excluded.config_value,updated_at=excluded.updated_at`, k, v, now, now)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
func (r *Repository) salt(ctx context.Context) (string, error) {
	var v string
	err := r.db.QueryRowContext(ctx, `SELECT config_value FROM configs WHERE config_key='analytics_ip_salt'`).Scan(&v)
	if err == nil {
		return v, nil
	}
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	v = hex.EncodeToString(b)
	now := r.currentTime().UTC().Format(time.RFC3339Nano)
	_, err = r.db.ExecContext(ctx, `INSERT INTO configs(config_key,config_value,created_at,updated_at) VALUES('analytics_ip_salt',?,?,?)`, v, now, now)
	return v, err
}
func (r *Repository) Log(ctx context.Context, path, pageType, ip, ua, ref string) error {
	s, err := r.Get(ctx)
	if err != nil || !s.AnalyticsEnabled {
		return err
	}
	salt, err := r.salt(ctx)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, []byte(salt))
	_, _ = mac.Write([]byte(ip))
	hash := hex.EncodeToString(mac.Sum(nil))
	source := "direct"
	if ref != "" {
		source = "referer"
		if strings.Contains(ref, "google.") || strings.Contains(ref, "bing.") || strings.Contains(ref, "baidu.") {
			source = "search"
		}
	}
	now := r.currentTime()
	id := hash[:16] + strconv.FormatInt(now.UnixNano(), 36)
	_, err = r.db.ExecContext(ctx, `INSERT INTO visit_logs(id,path,page_type,ip_hash,user_agent,referrer,source,created_at) VALUES(?,?,?,?,?,?,?,?)`, id, path, pageType, hash, ua, ref, source, now.UTC().Format(time.RFC3339Nano))
	return err
}
func (r *Repository) Cleanup(ctx context.Context) error {
	s, err := r.Get(ctx)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM visit_logs WHERE created_at<?`, r.currentTime().UTC().AddDate(0, 0, -s.AnalyticsRetentionDays).Format(time.RFC3339Nano))
	return err
}

type Dashboard struct {
	ImagesTotal, ImagesPublic, AlbumsTotal, GuidesTotal, GuidesPublic int
	VisitsTotal, VisitsToday, VisitsYesterday                         int
	CamerasTotal, LensesTotal                                         int
	Last7Days                                                         []Point `json:"last_7_days"`
	TopCameras, TopLenses                                             []NamedCount
	PhotosByYear                                                      []NamedCount
}
type Point struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
type NamedCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (r *Repository) Dashboard(ctx context.Context) (Dashboard, error) {
	var d Dashboard
	location := r.analyticsLocation(ctx)
	now := r.currentTime().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	yesterday := today.AddDate(0, 0, -1)
	tomorrow := today.AddDate(0, 0, 1)
	err := r.db.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM assets WHERE status!='purged'),(SELECT COUNT(*) FROM assets WHERE status='ready' AND visible=1),(SELECT COUNT(*) FROM albums WHERE deleted_at IS NULL),(SELECT COUNT(*) FROM guides WHERE deleted_at IS NULL),(SELECT COUNT(*) FROM guides WHERE deleted_at IS NULL AND published=1),(SELECT COUNT(*) FROM visit_logs),(SELECT COUNT(*) FROM visit_logs WHERE created_at>=? AND created_at<?),(SELECT COUNT(*) FROM visit_logs WHERE created_at>=? AND created_at<?)`, today.UTC().Format(time.RFC3339Nano), tomorrow.UTC().Format(time.RFC3339Nano), yesterday.UTC().Format(time.RFC3339Nano), today.UTC().Format(time.RFC3339Nano)).Scan(&d.ImagesTotal, &d.ImagesPublic, &d.AlbumsTotal, &d.GuidesTotal, &d.GuidesPublic, &d.VisitsTotal, &d.VisitsToday, &d.VisitsYesterday)
	if err != nil {
		return d, err
	}
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT camera) FROM assets WHERE status!='purged' AND camera IS NOT NULL AND camera!=''`).Scan(&d.CamerasTotal)
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT lens) FROM assets WHERE status!='purged' AND lens IS NOT NULL AND lens!=''`).Scan(&d.LensesTotal)
	for dayOffset := -6; dayOffset <= 0; dayOffset++ {
		start := today.AddDate(0, 0, dayOffset)
		end := start.AddDate(0, 0, 1)
		d.Last7Days = append(d.Last7Days, Point{
			Date:  start.Format("2006-01-02"),
			Count: r.visitCount(ctx, start, end),
		})
	}
	d.TopCameras = r.named(ctx, `SELECT COALESCE(camera,'Unknown'),COUNT(*) FROM assets WHERE status!='purged' GROUP BY 1 ORDER BY 2 DESC LIMIT 5`)
	d.TopLenses = r.named(ctx, `SELECT COALESCE(lens,'Unknown'),COUNT(*) FROM assets WHERE status!='purged' GROUP BY 1 ORDER BY 2 DESC LIMIT 5`)
	d.PhotosByYear = r.named(ctx, `SELECT COALESCE(substr(shoot_at,1,4),'Unknown'),COUNT(*) FROM assets WHERE status!='purged' GROUP BY 1 ORDER BY 1 DESC LIMIT 10`)
	return d, nil
}
func (r *Repository) named(ctx context.Context, q string) []NamedCount {
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []NamedCount
	for rows.Next() {
		var x NamedCount
		_ = rows.Scan(&x.Name, &x.Count)
		out = append(out, x)
	}
	return out
}
func (r *Repository) Analytics(ctx context.Context) (map[string]any, error) {
	d, err := r.Dashboard(ctx)
	if err != nil {
		return nil, err
	}
	var unique int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT ip_hash) FROM visit_logs WHERE ip_hash IS NOT NULL`).Scan(&unique)
	sources := r.named(ctx, `SELECT source,COUNT(*) FROM visit_logs GROUP BY source ORDER BY 2 DESC`)
	pages := r.named(ctx, `SELECT page_type,COUNT(*) FROM visit_logs GROUP BY page_type ORDER BY 2 DESC`)
	return map[string]any{"dashboard": d, "unique_visitors": unique, "hourly": r.hourly(ctx), "sources": sources, "pages": pages}, nil
}
func (r *Repository) hourly(ctx context.Context) []Point {
	location := r.analyticsLocation(ctx)
	now := r.currentTime().In(location)
	currentHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, location)
	start := currentHour.Add(-23 * time.Hour)
	out := make([]Point, 0, 24)
	for i := 0; i < 24; i++ {
		hour := start.Add(time.Duration(i) * time.Hour)
		out = append(out, Point{
			Date:  hour.Format(time.RFC3339),
			Count: r.visitCount(ctx, hour, hour.Add(time.Hour)),
		})
	}
	return out
}
func (r *Repository) analyticsLocation(ctx context.Context) *time.Location {
	settings, err := r.Get(ctx)
	if err == nil {
		if location, loadErr := time.LoadLocation(settings.AnalyticsTimezone); loadErr == nil {
			return location
		}
	}
	return time.UTC
}
func (r *Repository) visitCount(ctx context.Context, start, end time.Time) int {
	var count int
	_ = r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM visit_logs WHERE created_at>=? AND created_at<?`,
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	).Scan(&count)
	return count
}
func Disk(path string) (map[string]uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(filepath.Clean(path), &stat); err != nil {
		return nil, err
	}
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	return map[string]uint64{"total": total, "free": free, "used": total - free}, nil
}

type Handler struct {
	repo    *Repository
	dataDir string
}

func NewHandler(repo *Repository, dataDir string) *Handler {
	return &Handler{repo: repo, dataDir: dataDir}
}
func (h *Handler) RegisterPublic(r chi.Router) {
	r.Get("/settings", h.publicSettings)
	r.Post("/visits", h.visit)
}
func (h *Handler) RegisterAdmin(r chi.Router) {
	r.Get("/dashboard", h.dashboard)
	r.Get("/analytics", h.analytics)
	r.Get("/settings", h.adminSettings)
	r.Put("/settings", h.updateSettings)
	r.Get("/disk", h.disk)
}
func (h *Handler) publicSettings(w http.ResponseWriter, r *http.Request) {
	s, err := h.repo.Get(r.Context())
	if err != nil {
		write(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"site_title": s.SiteTitle, "site_author": s.SiteAuthor, "site_favicon_url": s.SiteFaviconURL, "about_intro": s.AboutIntro, "about_social_instagram": s.AboutInstagram, "about_social_xiaohongshu": s.AboutXiaohongshu, "about_social_weibo": s.AboutWeibo, "about_social_github": s.AboutGithub, "about_gallery_asset_ids": s.AboutGalleryAssetIDs, "gallery_layout": s.GalleryLayout, "public_original_download": s.PublicOriginalDownload})
}
func (h *Handler) adminSettings(w http.ResponseWriter, r *http.Request) {
	s, e := h.repo.Get(r.Context())
	if e != nil {
		write(w, 500, e)
		return
	}
	writeJSON(w, 200, s)
}
func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var s Settings
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if dec.Decode(&s) != nil {
		write(w, 400, errors.New("invalid settings"))
		return
	}
	if e := h.repo.Put(r.Context(), s); e != nil {
		write(w, 422, e)
		return
	}
	writeJSON(w, 200, nil)
}
func (h *Handler) visit(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Path     string `json:"path"`
		PageType string `json:"page_type"`
		Referrer string `json:"referrer"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		write(w, 400, errors.New("invalid visit"))
		return
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if e := h.repo.Log(r.Context(), in.Path, in.PageType, ip, r.UserAgent(), in.Referrer); e != nil {
		write(w, 500, e)
		return
	}
	writeJSON(w, 200, nil)
}
func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	d, e := h.repo.Dashboard(r.Context())
	if e != nil {
		write(w, 500, e)
		return
	}
	writeJSON(w, 200, d)
}
func (h *Handler) analytics(w http.ResponseWriter, r *http.Request) {
	d, e := h.repo.Analytics(r.Context())
	if e != nil {
		write(w, 500, e)
		return
	}
	writeJSON(w, 200, d)
}
func (h *Handler) disk(w http.ResponseWriter, r *http.Request) {
	d, e := Disk(h.dataDir)
	if e != nil {
		write(w, 500, e)
		return
	}
	writeJSON(w, 200, d)
}
func write(w http.ResponseWriter, status int, e error) {
	apiresponse.WriteError(w, &http.Request{}, status, "SITE_ERROR", e.Error())
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	apiresponse.WriteJSON(w, status, v)
}
func value(m map[string]string, k, d string) string {
	if m[k] != "" {
		return m[k]
	}
	return d
}
func boolValue(m map[string]string, k string, d bool) bool {
	v, ok := m[k]
	if !ok {
		return d
	}
	b, e := strconv.ParseBool(v)
	if e != nil {
		return d
	}
	return b
}
func intValue(m map[string]string, k string, d int) int {
	v, e := strconv.Atoi(m[k])
	if e != nil {
		return d
	}
	return v
}

var _ = fmt.Sprint
var _ = os.ErrNotExist
