package handler

import (
	"net/http"
	"strconv"

	"apollo-backend/model"
	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type FinanceHandler struct {
	service service.FinanceService
}

func NewFinanceHandler(svc service.FinanceService) *FinanceHandler {
	return &FinanceHandler{service: svc}
}

func (h *FinanceHandler) GetDashboard(c *gin.Context) {
	stats, err := h.service.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *FinanceHandler) ListInvoices(c *gin.Context) {
	status := c.Query("status")
	var patientID *uint
	if pid := c.Query("patient_id"); pid != "" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			uid := uint(id)
			patientID = &uid
		}
	}
	invoices, err := h.service.ListInvoices(status, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

func (h *FinanceHandler) CreateInvoice(c *gin.Context) {
	var inv model.Invoice
	if err := c.ShouldBindJSON(&inv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Mock staff record using it
	var createdBy uint = 1 
	inv.CreatedBy = &createdBy

	if err := h.service.CreateInvoice(&inv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Invoice created", "invoice": inv})
}

func (h *FinanceHandler) GetInvoice(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	inv, err := h.service.GetInvoice(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}
	c.JSON(http.StatusOK, inv)
}

func (h *FinanceHandler) UpdateInvoiceStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Status        string `json:"status" binding:"required"`
		PaymentMethod string `json:"payment_method"`
		Reference     string `json:"reference"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var staffID uint = 1

	if err := h.service.MarkInvoiceStatus(uint(id), body.Status, body.PaymentMethod, body.Reference, staffID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invoice status updated"})
}

func (h *FinanceHandler) ListClaims(c *gin.Context) {
	status := c.Query("status")
	var patientID *uint
	if pid := c.Query("patient_id"); pid != "" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			uid := uint(id)
			patientID = &uid
		}
	}
	claims, err := h.service.ListClaims(status, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"claims": claims})
}

func (h *FinanceHandler) CreateClaim(c *gin.Context) {
	var claim model.InsuranceClaim
	if err := c.ShouldBindJSON(&claim); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if err := h.service.CreateClaim(&claim); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Claim initiated", "claim": claim})
}

func (h *FinanceHandler) UpdateClaimStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateClaimStatus(uint(id), body.Status, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Claim status updated"})
}

func (h *FinanceHandler) ListLedger(c *gin.Context) {
	txType := c.Query("transaction_type")
	ledger, err := h.service.ListLedger(txType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ledger": ledger})
}
