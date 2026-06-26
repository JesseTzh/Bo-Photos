package tag

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
	r.Get("/", h.list)
	return r
}
func (h *HTTPHandler) AdminRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/", h.create)
	r.Put("/{id}", h.update)
	r.Put("/{id}/parent", h.move)
	r.Delete("/{id}", h.delete)
	return r
}

func (h *HTTPHandler) RegisterAssetRoutes(router chi.Router) {
	router.Get("/{assetID}/tags", h.assetTags)
	router.Put("/{assetID}/tags", h.assignAsset)
}
func (h *HTTPHandler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		writeError(w, 500, "TAG_LIST_FAILED", "读取标签失败")
		return
	}
	writeJSON(w, 200, map[string]any{"items": Tree(items)})
}
func (h *HTTPHandler) create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	item, err := h.service.Create(r.Context(), input)
	if err != nil {
		writeError(w, 422, "TAG_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, item)
}
func (h *HTTPHandler) update(w http.ResponseWriter, r *http.Request) {
	var input UpdateInput
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.service.Update(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		writeError(w, 422, "TAG_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) move(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ParentID string `json:"parent_id"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.service.Move(r.Context(), chi.URLParam(r, "id"), input.ParentID); err != nil {
		writeError(w, 422, "TAG_MOVE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) delete(w http.ResponseWriter, r *http.Request) {
	err := h.service.Delete(r.Context(), chi.URLParam(r, "id"))
	if errors.Is(err, ErrHasChildren) {
		writeError(w, 409, "TAG_HAS_CHILDREN", "请先移动或删除子标签")
		return
	}
	if err != nil {
		writeError(w, 422, "TAG_DELETE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) assetTags(w http.ResponseWriter, r *http.Request) {
	ids, err := h.repo.AssetTagIDs(r.Context(), chi.URLParam(r, "assetID"))
	if err != nil {
		writeError(w, 500, "ASSET_TAGS_FAILED", "读取图片标签失败")
		return
	}
	writeJSON(w, 200, map[string]any{"tag_ids": ids})
}
func (h *HTTPHandler) assignAsset(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TagIDs []string `json:"tag_ids"`
	}
	if json.NewDecoder(r.Body).Decode(&input) != nil {
		writeError(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.service.AssignAsset(r.Context(), chi.URLParam(r, "assetID"), input.TagIDs); err != nil {
		writeError(w, 422, "ASSET_TAGS_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
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
