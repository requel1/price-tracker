package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/requel1/price-tracker/internal/model"
	"github.com/requel1/price-tracker/internal/repository"
)

type ProductHandler struct {
	productRepo *repository.ProductRepo
	priceRepo   *repository.PriceRepo
}

func NewProductHandler(productRepo *repository.ProductRepo, priceRepo *repository.PriceRepo) *ProductHandler {
	return &ProductHandler{
		productRepo: productRepo,
		priceRepo:   priceRepo,
	}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL  string `json:"url"`
		Name string `json:"name"`
		Site string `json:"site"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.URL == "" || req.Name == "" || req.Site == "" {
		writeError(w, http.StatusBadRequest, "url,name,site are required")
		return
	}

	p := &model.Product{
		URL:  req.URL,
		Name: req.Name,
		Site: req.Site,
	}
	created, err := h.productRepo.Create(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	products, err := h.productRepo.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if products == nil {
		products = []*model.Product{}
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := h.productRepo.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	err = h.productRepo.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) GetPrices(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	prices, err := h.priceRepo.GetByProductID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if prices == nil {
		prices = []*model.Price{}

	}
	writeJSON(w, http.StatusOK, prices)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
