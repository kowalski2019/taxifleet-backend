package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"taxifleet/backend/internal/service"
)

type TaxiHandler struct {
	service *service.TaxiService
}

func NewTaxiHandler(service *service.TaxiService) *TaxiHandler {
	return &TaxiHandler{service: service}
}

func (h *TaxiHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	taxis, err := h.service.List(tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, taxis)
}

func (h *TaxiHandler) Create(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")

	var req service.CreateTaxiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taxi, err := h.service.Create(tenantID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, taxi)
}

func (h *TaxiHandler) Get(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	taxi, err := h.service.GetByID(uint(id), tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, taxi)
}

func (h *TaxiHandler) Update(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req service.UpdateTaxiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taxi, err := h.service.Update(uint(id), tenantID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, taxi)
}

func (h *TaxiHandler) Delete(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.service.Delete(uint(id), tenantID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Taxi deleted successfully"})
}

