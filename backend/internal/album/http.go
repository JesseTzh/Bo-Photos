package album

import (
	"encoding/json"
	"errors"
	"net/http"

	apiresponse "github.com/besscroft/bophotos/backend/internal/api"
	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	service *Service
	repo    *Repository
}

func NewHTTPHandler(service *Service, repo *Repository) *HTTPHandler {
	return &HTTPHandler{service: service, repo: repo}
}

func (h *HTTPHandler) PublicRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.listPublic)
	r.Get("/{value}", h.getPublic)
	return r
}

func (h *HTTPHandler) AdminRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.listAdmin)
	r.Post("/", h.create)
	r.Put("/sort", h.updateOrder)
	r.Get("/{id}", h.getAdmin)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Put("/{id}/cover", h.setCover)
	r.Get("/{id}/assets", h.assets)
	r.Put("/{id}/assets", h.replaceAssets)
	r.Put("/{id}/assets/sort", h.updateAssetSort)
	r.Post("/{id}/assets/sort/reset", h.resetAssetSort)
	return r
}

func (h *HTTPHandler) listPublic(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context(), true)
	h.writeList(w, items, err)
}
func (h *HTTPHandler) listAdmin(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context(), false)
	h.writeList(w, items, err)
}
func (h *HTTPHandler) writeList(w http.ResponseWriter, items []Album, err error) {
	if err != nil {
		writeError(w, 500, "ALBUM_LIST_FAILED", "读取相册失败")
		return
	}
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, dto(item))
	}
	writeJSON(w, 200, map[string]any{"items": result})
}
func (h *HTTPHandler) getPublic(w http.ResponseWriter, r *http.Request) {
	item, err := h.repo.GetByValue(r.Context(), chi.URLParam(r, "value"), true)
	h.writeOne(w, item, err)
}
func (h *HTTPHandler) getAdmin(w http.ResponseWriter, r *http.Request) {
	item, err := h.repo.Get(r.Context(), chi.URLParam(r, "id"), false)
	h.writeOne(w, item, err)
}
func (h *HTTPHandler) writeOne(w http.ResponseWriter, item Album, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, 404, "ALBUM_NOT_FOUND", "相册不存在")
		return
	}
	if err != nil {
		writeError(w, 500, "ALBUM_READ_FAILED", "读取相册失败")
		return
	}
	writeJSON(w, 200, dto(item))
}
func (h *HTTPHandler) create(w http.ResponseWriter, r *http.Request) {
	var input Input
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	item, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeError(w, 422, "ALBUM_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, dto(item))
}
func (h *HTTPHandler) update(w http.ResponseWriter, r *http.Request) {
	var input Input
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	item, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		writeError(w, 422, "ALBUM_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, dto(item))
}
func (h *HTTPHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, 422, "ALBUM_DELETE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) updateOrder(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []string `json:"ids"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil || len(input.IDs) == 0 {
		writeError(w, 400, "INVALID_REQUEST", "ids 不能为空")
		return
	}
	if err := h.repo.UpdateOrder(r.Context(), input.IDs); err != nil {
		writeError(w, 422, "ALBUM_SORT_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) setCover(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AssetID string `json:"asset_id"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repo.SetCover(r.Context(), chi.URLParam(r, "id"), input.AssetID); err != nil {
		writeError(w, 422, "ALBUM_COVER_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) assets(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.Assets(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, 500, "ALBUM_ASSETS_FAILED", "读取相册图片失败")
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (h *HTTPHandler) replaceAssets(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AssetIDs []string `json:"asset_ids"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	id := chi.URLParam(r, "id")
	if _, err := h.repo.Get(r.Context(), id, false); err != nil {
		writeError(w, 404, "ALBUM_NOT_FOUND", "相册不存在")
		return
	}
	if err := h.repo.ReplaceAssets(r.Context(), id, input.AssetIDs); err != nil {
		writeError(w, 422, "ALBUM_ASSETS_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) updateAssetSort(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Orders []struct {
			AssetID string `json:"asset_id"`
			Sort    int    `json:"sort"`
		} `json:"orders"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil || len(input.Orders) == 0 {
		writeError(w, 400, "INVALID_REQUEST", "orders 不能为空")
		return
	}
	orders := make(map[string]int, len(input.Orders))
	for _, order := range input.Orders {
		orders[order.AssetID] = order.Sort
	}
	if err := h.repo.UpdateAssetSort(r.Context(), chi.URLParam(r, "id"), orders); err != nil {
		writeError(w, 422, "ALBUM_ASSET_SORT_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) resetAssetSort(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.ResetAssetSort(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, 422, "ALBUM_ASSET_SORT_RESET_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func dto(item Album) map[string]any {
	cover := ""
	if item.EffectiveCoverAssetID != "" {
		cover = "/media/assets/" + item.EffectiveCoverAssetID + "/preview"
	}
	return map[string]any{"id": item.ID, "name": item.Name, "album_value": item.Value,
		"detail": item.Detail, "theme": item.Theme, "visible": item.Visible, "sort": item.Sort,
		"random_show": item.RandomShow, "license": item.License, "cover_asset_id": item.CoverAssetID,
		"cover_url": cover, "image_sorting": item.ImageSorting, "asset_ids": item.AssetIDs,
		"asset_count": item.AssetCount}
}
func writeJSON(w http.ResponseWriter, status int, payload any) {
	apiresponse.WriteJSON(w, status, payload)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	apiresponse.WriteError(w, nilRequest(), status, code, message)
}

func nilRequest() *http.Request {
	return &http.Request{}
}
