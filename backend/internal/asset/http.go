package asset

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	apiresponse "github.com/besscroft/bophotos/backend/internal/api"
	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	service    *Service
	repository *Repository
}

type AssetDTO struct {
	ID             string     `json:"id"`
	Status         Status     `json:"status"`
	OriginalName   string     `json:"original_name"`
	MIMEType       string     `json:"mime_type,omitempty"`
	ByteSize       int64      `json:"byte_size,omitempty"`
	Width          int        `json:"width"`
	Height         int        `json:"height"`
	Title          string     `json:"title,omitempty"`
	Description    string     `json:"description,omitempty"`
	Longitude      *float64   `json:"longitude,omitempty"`
	Latitude       *float64   `json:"latitude,omitempty"`
	BlurHash       string     `json:"blurhash,omitempty"`
	EXIFJSON       string     `json:"exif_json,omitempty"`
	ShootAt        *time.Time `json:"shoot_at,omitempty"`
	Camera         string     `json:"camera,omitempty"`
	Lens           string     `json:"lens,omitempty"`
	ExposureTime   string     `json:"exposure_time,omitempty"`
	Aperture       string     `json:"aperture,omitempty"`
	ISO            string     `json:"iso,omitempty"`
	FocalLength    string     `json:"focal_length,omitempty"`
	Visible        bool       `json:"visible"`
	ShowOnHomepage bool       `json:"show_on_homepage"`
	Featured       bool       `json:"featured"`
	Sort           int        `json:"sort"`
	ThumbnailURL   string     `json:"thumbnail_url,omitempty"`
	PreviewURL     string     `json:"preview_url,omitempty"`
	OriginalURL    string     `json:"original_url,omitempty"`
	ErrorCode      string     `json:"error_code,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Private        bool       `json:"private"`
}

func NewHTTPHandler(service *Service, repository *Repository) *HTTPHandler {
	return &HTTPHandler{service: service, repository: repository}
}

func (h *HTTPHandler) PublicRoutes() http.Handler {
	router := chi.NewRouter()
	router.Get("/", h.listPublic)
	router.Get("/filters", h.filters)
	router.Get("/{id}", h.getPublic)
	return router
}

func (h *HTTPHandler) AdminRoutes(register ...func(chi.Router)) http.Handler {
	router := chi.NewRouter()
	router.Post("/", h.upload)
	router.Get("/", h.listAdmin)
	router.Get("/filters", h.adminFilters)
	router.Get("/private", h.listPrivate)
	router.Put("/sort", h.updateSort)
	router.Patch("/{id}", h.update)
	router.Delete("/{id}", h.delete)
	router.Get("/{id}/albums", h.assetAlbums)
	router.Put("/{id}/albums", h.saveAssetAlbums)
	router.Post("/{id}/restore", h.restore)
	router.Post("/{id}/purge", h.purge)
	router.Post("/{id}/retry", h.retry)
	router.Delete("/", h.batchDelete)
	router.Get("/{id}", h.getAdmin)
	for _, attach := range register {
		attach(router)
	}
	return router
}

func (h *HTTPHandler) adminFilters(w http.ResponseWriter, r *http.Request) {
	options, err := h.repository.AdminFilterOptions(r.Context())
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_FILTERS_FAILED", "读取筛选项失败")
		return
	}
	writeAssetJSON(w, http.StatusOK, options)
}

func (h *HTTPHandler) filters(w http.ResponseWriter, r *http.Request) {
	cameras, lenses, err := h.repository.FilterOptions(r.Context())
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_FILTERS_FAILED", "读取筛选项失败")
		return
	}
	writeAssetJSON(w, http.StatusOK, map[string]any{"cameras": cameras, "lenses": lenses})
}

func (h *HTTPHandler) listAdmin(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := AdminFilter{
		Status: Status(query.Get("status")), Camera: query.Get("camera"),
		Lens: query.Get("lens"), ExposureTime: query.Get("exposure_time"),
		Aperture: query.Get("aperture"), ISO: query.Get("iso"),
		Album: query.Get("album"), TagsOperator: query.Get("tags_operator"),
		TagIDs: splitCSV(query.Get("tags")), Title: query.Get("title"),
		Page: intQuery(query.Get("page"), 1), PageSize: intQuery(query.Get("page_size"), 20),
	}
	if value := query.Get("visible"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			writeAssetError(w, r, http.StatusBadRequest, "INVALID_FILTER", "visible 筛选无效")
			return
		}
		filter.Visible = &parsed
	}
	if value := query.Get("private"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			writeAssetError(w, r, http.StatusBadRequest, "INVALID_FILTER", "private 筛选无效")
			return
		}
		filter.Private = &parsed
	}
	if value := query.Get("featured"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			writeAssetError(w, r, http.StatusBadRequest, "INVALID_FILTER", "featured 筛选无效")
			return
		}
		filter.Featured = &parsed
	}
	items, total, err := h.repository.ListAdmin(r.Context(), filter)
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_LIST_FAILED", "读取图片列表失败")
		return
	}
	dtos := make([]AssetDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, toDTO(item))
	}
	writeAssetJSON(w, http.StatusOK, map[string]any{
		"page": filter.Page, "page_size": filter.PageSize, "total": total, "items": dtos,
	})
}

func (h *HTTPHandler) listPrivate(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := PrivateFilter{
		Page:     intQuery(query.Get("page"), 1),
		PageSize: intQuery(query.Get("page_size"), 16),
	}
	items, total, err := h.repository.ListPrivate(r.Context(), filter)
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_LIST_FAILED", "读取隐私相册失败")
		return
	}
	dtos := make([]AssetDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, toDTO(item))
	}
	writeAssetJSON(w, http.StatusOK, map[string]any{
		"page": filter.Page, "page_size": filter.PageSize, "total": total, "items": dtos,
	})
}

func (h *HTTPHandler) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeAssetError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repository.Update(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_UPDATE_FAILED", err.Error())
		return
	}
	writeAssetJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) assetAlbums(w http.ResponseWriter, r *http.Request) {
	ids, err := h.repository.AlbumIDsForAsset(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_ALBUMS_FAILED", "读取图片相册失败")
		return
	}
	writeAssetJSON(w, http.StatusOK, map[string]any{"album_ids": ids})
}

func (h *HTTPHandler) saveAssetAlbums(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AlbumIDs []string `json:"album_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeAssetError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repository.ReplaceAssetAlbums(r.Context(), chi.URLParam(r, "id"), input.AlbumIDs); err != nil {
		writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_ALBUMS_FAILED", err.Error())
		return
	}
	writeAssetJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_DELETE_FAILED", err.Error())
		return
	}
	writeAssetJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) restore(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Restore(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_RESTORE_FAILED", err.Error())
		return
	}
	writeAssetJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) purge(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Purge(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_PURGE_FAILED", err.Error())
		return
	}
	writeAssetJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) retry(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Retry(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_RETRY_FAILED", err.Error())
		return
	}
	writeAssetJSON(w, http.StatusAccepted, nil)
}

func (h *HTTPHandler) batchDelete(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || len(input.IDs) == 0 {
		writeAssetError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "ids 不能为空")
		return
	}
	for _, id := range input.IDs {
		if err := h.service.Delete(r.Context(), id); err != nil {
			writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_DELETE_FAILED", err.Error())
			return
		}
	}
	writeAssetJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) updateSort(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Orders []struct {
			ID   string `json:"id"`
			Sort int    `json:"sort"`
		} `json:"orders"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || len(input.Orders) == 0 {
		writeAssetError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "orders 不能为空")
		return
	}
	orders := make(map[string]int, len(input.Orders))
	for _, order := range input.Orders {
		orders[order.ID] = order.Sort
	}
	if err := h.repository.UpdateSort(r.Context(), orders); err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_SORT_FAILED", "更新排序失败")
		return
	}
	writeAssetJSON(w, http.StatusOK, nil)
}

func (h *HTTPHandler) upload(w http.ResponseWriter, r *http.Request) {
	reader, err := r.MultipartReader()
	if err != nil {
		writeAssetError(w, r, http.StatusBadRequest, "INVALID_MULTIPART", "上传请求格式不正确")
		return
	}
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeAssetError(w, r, http.StatusBadRequest, "INVALID_MULTIPART", "读取上传文件失败")
			return
		}
		if part.FormName() != "file" || part.FileName() == "" {
			_ = part.Close()
			continue
		}
		result, err := h.service.Upload(r.Context(), part.FileName(), part)
		_ = part.Close()
		if err != nil {
			writeAssetError(w, r, http.StatusUnprocessableEntity, "ASSET_UPLOAD_FAILED", err.Error())
			return
		}
		writeAssetJSON(w, http.StatusAccepted, map[string]any{
			"id":                  result.Asset.ID,
			"status":              result.Asset.Status,
			"duplicate_asset_ids": result.DuplicateAssetIDs,
		})
		return
	}
	writeAssetError(w, r, http.StatusBadRequest, "FILE_REQUIRED", "缺少上传文件")
}

func (h *HTTPHandler) listPublic(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := PublicFilter{
		Cameras:       splitCSV(query.Get("cameras")),
		Lenses:        splitCSV(query.Get("lenses")),
		Tags:          splitCSV(query.Get("tags")),
		TagsOperator:  query.Get("tags_operator"),
		Album:         query.Get("album"),
		ShootTimeSort: SortDirection(query.Get("sort_by_shoot_time")),
		HomepageOnly:  query.Get("homepage") == "true",
		Page:          intQuery(query.Get("page"), 1),
		PageSize:      intQuery(query.Get("page_size"), 16),
	}
	if value := query.Get("featured"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			writeAssetError(w, r, http.StatusBadRequest, "INVALID_FILTER", "featured 筛选无效")
			return
		}
		filter.Featured = &parsed
	}
	items, total, err := h.repository.ListPublic(r.Context(), filter)
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_LIST_FAILED", "读取图库失败")
		return
	}
	dtos := make([]AssetDTO, 0, len(items))
	for _, item := range items {
		dtos = append(dtos, toDTO(item))
	}
	writeAssetJSON(w, http.StatusOK, map[string]any{
		"page": filter.Page, "page_size": filter.PageSize, "total": total, "items": dtos,
	})
}

func (h *HTTPHandler) getPublic(w http.ResponseWriter, r *http.Request) {
	item, err := h.repository.Get(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, ErrNotFound) || item.Status != StatusReady || !item.Visible || item.Private {
		writeAssetError(w, r, http.StatusNotFound, "ASSET_NOT_FOUND", "图片不存在")
		return
	}
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_READ_FAILED", "读取图片失败")
		return
	}
	writeAssetJSON(w, http.StatusOK, toDTO(item))
}

func (h *HTTPHandler) getAdmin(w http.ResponseWriter, r *http.Request) {
	item, err := h.repository.Get(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, ErrNotFound) {
		writeAssetError(w, r, http.StatusNotFound, "ASSET_NOT_FOUND", "图片不存在")
		return
	}
	if err != nil {
		writeAssetError(w, r, http.StatusInternalServerError, "ASSET_READ_FAILED", "读取图片失败")
		return
	}
	writeAssetJSON(w, http.StatusOK, toDTO(item))
}

func toDTO(item Asset) AssetDTO {
	base := "/media/assets/" + item.ID
	dto := AssetDTO{
		ID: item.ID, Status: item.Status, OriginalName: item.OriginalName,
		MIMEType: item.MIMEType, ByteSize: item.ByteSize, Width: item.Width,
		Height: item.Height, Title: item.Title, Description: item.Description,
		Longitude: item.Longitude, Latitude: item.Latitude, BlurHash: item.BlurHash,
		EXIFJSON: item.EXIFJSON, ShootAt: item.ShootAt, Camera: item.Camera,
		Lens: item.Lens, ExposureTime: item.ExposureTime, Aperture: item.Aperture,
		ISO: item.ISO, FocalLength: item.FocalLength, Visible: item.Visible,
		Private: item.Private, ShowOnHomepage: item.ShowOnHomepage, Featured: item.Featured, Sort: item.Sort,
		ErrorCode: item.ErrorCode, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
	}
	if item.ThumbnailKey != "" {
		dto.ThumbnailURL = base + "/thumbnail?v=" + strconv.Itoa(item.DerivativeVersion)
	}
	if item.PreviewKey != "" {
		dto.PreviewURL = base + "/preview?v=" + strconv.Itoa(item.DerivativeVersion)
	}
	if item.OriginalKey != "" {
		dto.OriginalURL = base + "/original"
	}
	return dto
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func intQuery(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func writeAssetJSON(w http.ResponseWriter, status int, payload any) {
	apiresponse.WriteJSON(w, status, payload)
}

func writeAssetError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	apiresponse.WriteError(w, r, status, code, message)
}
