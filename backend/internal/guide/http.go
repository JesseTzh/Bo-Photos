package guide

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
	r.Get("/{id}", h.getPublic)
	return r
}
func (h *HTTPHandler) AdminRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.listAdmin)
	r.Post("/", h.create)
	r.Put("/sort", h.sort)
	r.Get("/{id}", h.getAdmin)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	r.Get("/{id}/modules", h.modules)
	r.Post("/{id}/modules", h.createModule)
	r.Put("/{id}/modules/sort", h.sortModules)
	r.Put("/{id}/modules/{moduleID}", h.updateModule)
	r.Delete("/{id}/modules/{moduleID}", h.deleteModule)
	r.Post("/{id}/modules/{moduleID}/blocks", h.createBlock)
	r.Put("/{id}/modules/{moduleID}/blocks/sort", h.sortBlocks)
	r.Put("/{id}/modules/{moduleID}/blocks/{blockID}", h.updateBlock)
	r.Delete("/{id}/modules/{moduleID}/blocks/{blockID}", h.deleteBlock)
	r.Get("/{id}/toc", h.toc)
	r.Put("/{id}/toc", h.replaceTOC)
	r.Post("/{id}/toc/auto-generate", h.autoTOC)
	r.Get("/{id}/albums", h.albums)
	r.Put("/{id}/albums", h.replaceAlbums)
	return r
}
func (h *HTTPHandler) listPublic(w http.ResponseWriter, r *http.Request) { h.list(w, r, true) }
func (h *HTTPHandler) listAdmin(w http.ResponseWriter, r *http.Request)  { h.list(w, r, false) }
func (h *HTTPHandler) list(w http.ResponseWriter, r *http.Request, public bool) {
	items, err := h.repo.List(r.Context(), public)
	if err != nil {
		writeErr(w, 500, "GUIDE_LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"items": guideDTOs(items)})
}
func (h *HTTPHandler) getPublic(w http.ResponseWriter, r *http.Request) { h.get(w, r, true) }
func (h *HTTPHandler) getAdmin(w http.ResponseWriter, r *http.Request)  { h.get(w, r, false) }
func (h *HTTPHandler) get(w http.ResponseWriter, r *http.Request, public bool) {
	item, err := h.repo.Get(r.Context(), chi.URLParam(r, "id"), public)
	if errors.Is(err, ErrNotFound) {
		writeErr(w, 404, "GUIDE_NOT_FOUND", "Guide 不存在")
		return
	}
	if err != nil {
		writeErr(w, 500, "GUIDE_READ_FAILED", err.Error())
		return
	}
	modules, _ := h.repo.ListModules(r.Context(), item.ID, public)
	toc, _ := h.repo.ListTOC(r.Context(), item.ID, public)
	albums, _ := h.repo.ListAlbums(r.Context(), item.ID, public)
	writeJSON(w, 200, map[string]any{"guide": guideDTO(item), "modules": modules, "toc": toc, "albums": albums})
}
func (h *HTTPHandler) create(w http.ResponseWriter, r *http.Request) {
	var in GuideInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	item, err := h.service.Create(r.Context(), in)
	if err != nil {
		writeErr(w, 422, "GUIDE_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, guideDTO(item))
}
func (h *HTTPHandler) update(w http.ResponseWriter, r *http.Request) {
	var in GuideInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	item, err := h.service.Update(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		writeErr(w, 422, "GUIDE_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, guideDTO(item))
}
func (h *HTTPHandler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeErr(w, 422, "GUIDE_DELETE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) sort(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs []string `json:"ids"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repo.Sort(r.Context(), in.IDs); err != nil {
		writeErr(w, 422, "GUIDE_SORT_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) modules(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListModules(r.Context(), chi.URLParam(r, "id"), false)
	if err != nil {
		writeErr(w, 500, "MODULE_LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (h *HTTPHandler) createModule(w http.ResponseWriter, r *http.Request) {
	var in ModuleInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	item, err := h.repo.CreateModule(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		writeErr(w, 422, "MODULE_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, item)
}
func (h *HTTPHandler) createBlock(w http.ResponseWriter, r *http.Request) {
	var in BlockInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	item, err := h.repo.CreateBlock(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "moduleID"), in)
	if err != nil {
		writeErr(w, 422, "BLOCK_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, item)
}
func (h *HTTPHandler) updateModule(w http.ResponseWriter, r *http.Request) {
	var in ModuleInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repo.UpdateModule(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "moduleID"), in); err != nil {
		writeErr(w, 422, "MODULE_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) deleteModule(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteModule(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "moduleID")); err != nil {
		writeErr(w, 422, "MODULE_DELETE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) sortModules(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs []string `json:"ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if err := h.repo.SortModules(r.Context(), chi.URLParam(r, "id"), in.IDs); err != nil {
		writeErr(w, 422, "MODULE_SORT_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) updateBlock(w http.ResponseWriter, r *http.Request) {
	var in BlockInput
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repo.UpdateBlock(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "moduleID"), chi.URLParam(r, "blockID"), in); err != nil {
		writeErr(w, 422, "BLOCK_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) deleteBlock(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.DeleteBlock(r.Context(), chi.URLParam(r, "moduleID"), chi.URLParam(r, "blockID")); err != nil {
		writeErr(w, 422, "BLOCK_DELETE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) sortBlocks(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs []string `json:"ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	if err := h.repo.SortBlocks(r.Context(), chi.URLParam(r, "moduleID"), in.IDs); err != nil {
		writeErr(w, 422, "BLOCK_SORT_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) toc(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListTOC(r.Context(), chi.URLParam(r, "id"), false)
	if err != nil {
		writeErr(w, 500, "TOC_LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (h *HTTPHandler) replaceTOC(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Items []TOCInput `json:"items"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repo.ReplaceTOC(r.Context(), chi.URLParam(r, "id"), in.Items); err != nil {
		writeErr(w, 422, "TOC_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) autoTOC(w http.ResponseWriter, r *http.Request) {
	if err := h.repo.AutoGenerateTOC(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeErr(w, 422, "TOC_GENERATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func (h *HTTPHandler) albums(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.ListAlbums(r.Context(), chi.URLParam(r, "id"), false)
	if err != nil {
		writeErr(w, 500, "ALBUM_LIST_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"items": items})
}
func (h *HTTPHandler) replaceAlbums(w http.ResponseWriter, r *http.Request) {
	var in struct {
		IDs []string `json:"ids"`
	}
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		writeErr(w, 400, "INVALID_REQUEST", "请求格式不正确")
		return
	}
	if err := h.repo.ReplaceAlbums(r.Context(), chi.URLParam(r, "id"), in.IDs); err != nil {
		writeErr(w, 422, "ALBUM_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, nil)
}
func guideDTOs(items []Guide) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, x := range items {
		out = append(out, guideDTO(x))
	}
	return out
}
func guideDTO(x Guide) map[string]any {
	cover := ""
	if x.CoverAssetID != "" {
		cover = "/media/assets/" + x.CoverAssetID + "/preview"
	}
	return map[string]any{"id": x.ID, "title": x.Title, "country": x.Country, "city": x.City, "days": x.Days, "start_date": x.StartDate, "end_date": x.EndDate, "cover_asset_id": x.CoverAssetID, "cover_url": cover, "published": x.Published, "sort": x.Sort, "created_at": x.CreatedAt.Format(time.RFC3339)}
}
func writeJSON(w http.ResponseWriter, status int, payload any) {
	apiresponse.WriteJSON(w, status, payload)
}
func writeErr(w http.ResponseWriter, status int, code, message string) {
	apiresponse.WriteError(w, nilRequest(), status, code, message)
}

func nilRequest() *http.Request {
	return &http.Request{}
}
