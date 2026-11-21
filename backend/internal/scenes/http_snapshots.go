package scenes

import (
	"net/http"
)

func (h *Handlers) registerSnapshots(mux *http.ServeMux, base string) {
	mux.Handle("GET "+base+"/{sceneId}/snapshots", h.cmp.RequireGM(http.HandlerFunc(h.listSnapshots)))
	mux.Handle("POST "+base+"/{sceneId}/snapshots", h.cmp.RequireGM(http.HandlerFunc(h.saveSnapshot)))
	mux.Handle("DELETE "+base+"/{sceneId}/snapshots/{snapshotId}", h.cmp.RequireGM(http.HandlerFunc(h.deleteSnapshot)))
}

func (h *Handlers) listSnapshots(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListSnapshots(r.Context(), r.PathValue("id"), r.PathValue("sceneId"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handlers) saveSnapshot(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !decode(w, r, &in) {
		return
	}
	sn, err := h.svc.SaveSnapshot(r.Context(), r.PathValue("id"), r.PathValue("sceneId"), in.Name)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, sn)
}

func (h *Handlers) deleteSnapshot(w http.ResponseWriter, r *http.Request) {
	h.done(w, h.svc.DeleteSnapshot(r.Context(), r.PathValue("id"), r.PathValue("sceneId"), r.PathValue("snapshotId")))
}
