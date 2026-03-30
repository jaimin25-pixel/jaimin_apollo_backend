package handler

import (
	"net/http"
	"strconv"

	"apollo-backend/model"
	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type PharmacyHandler struct {
	svc *service.PharmacyService
}

func NewPharmacyHandler(svc *service.PharmacyService) *PharmacyHandler {
	return &PharmacyHandler{svc: svc}
}

// ── Dashboard ──────────────────────────────────────────────────────────

func (h *PharmacyHandler) GetDashboard(c *gin.Context) {
	data, err := h.svc.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// ── Prescriptions ──────────────────────────────────────────────────────

func (h *PharmacyHandler) ListPrescriptions(c *gin.Context) {
	status := c.Query("status")
	doctorID := c.Query("doctor_id")
	date := c.Query("date")

	prescriptions, err := h.svc.ListPrescriptions(status, doctorID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if prescriptions == nil {
		prescriptions = []model.Prescription{}
	}
	c.JSON(http.StatusOK, prescriptions)
}

func (h *PharmacyHandler) GetPrescription(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	rx, err := h.svc.GetPrescription(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "prescription not found"})
		return
	}

	c.JSON(http.StatusOK, rx)
}

func (h *PharmacyHandler) DispensePrescription(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	pharmacistID, _ := c.Get("userID")
	var req service.DispenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one item required"})
		return
	}

	err := h.svc.DispensePrescription(id, pharmacistID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prescription dispensed successfully"})
}

func (h *PharmacyHandler) CancelPrescription(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.svc.CancelPrescription(id, body.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "prescription cancelled"})
}

// ── Medicines ──────────────────────────────────────────────────────────

func (h *PharmacyHandler) ListMedicines(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("search")

	medicines, err := h.svc.ListMedicines(category, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if medicines == nil {
		medicines = []model.Medicine{}
	}
	c.JSON(http.StatusOK, medicines)
}

func (h *PharmacyHandler) GetMedicine(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	med, err := h.svc.GetMedicine(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "medicine not found"})
		return
	}

	c.JSON(http.StatusOK, med)
}

func (h *PharmacyHandler) CreateMedicine(c *gin.Context) {
	var req service.CreateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validation
	if req.GenericName == "" || req.Category == "" || req.Unit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "generic_name, category, and unit are required"})
		return
	}

	med, err := h.svc.CreateMedicine(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, med)
}

func (h *PharmacyHandler) UpdateMedicine(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req service.UpdateMedicineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.svc.UpdateMedicine(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "medicine updated"})
}

// ── Medicine Batches ───────────────────────────────────────────────────

func (h *PharmacyHandler) ListBatches(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	batches, err := h.svc.ListBatchesForMedicine(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if batches == nil {
		batches = []model.MedicineBatch{}
	}
	c.JSON(http.StatusOK, batches)
}

func (h *PharmacyHandler) CreateBatch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req service.CreateBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validation
	if req.BatchNumber == "" || req.ExpiryDate == "" || req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "batch_number, expiry_date, and quantity are required"})
		return
	}

	batch, err := h.svc.CreateBatch(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, batch)
}

// ── Stock Adjustments ──────────────────────────────────────────────────

func (h *PharmacyHandler) AdjustStock(c *gin.Context) {
	pharmacistID, _ := c.Get("userID")

	var body struct {
		MedicineID     uint    `json:"medicine_id" binding:"required"`
		AdjustmentType string  `json:"adjustment_type" binding:"required"`
		Quantity       float64 `json:"quantity" binding:"required"`
		Reason         string  `json:"reason"`
		BatchID        *uint   `json:"batch_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validation
	if body.Quantity <= 0 || body.AdjustmentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity and adjustment_type are required"})
		return
	}

	req := service.AdjustStockRequest{
		AdjustmentType: body.AdjustmentType,
		Quantity:       body.Quantity,
		Reason:         body.Reason,
		BatchID:        body.BatchID,
	}

	err := h.svc.AdjustStock(body.MedicineID, pharmacistID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock adjusted"})
}

// ── Alerts ─────────────────────────────────────────────────────────────

func (h *PharmacyHandler) GetLowStockAlerts(c *gin.Context) {
	alerts, err := h.svc.GetLowStockAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if alerts == nil {
		alerts = []model.StockAlert{}
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *PharmacyHandler) GetExpiringAlerts(c *gin.Context) {
	days := 60
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	alerts, err := h.svc.GetExpiringAlerts(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if alerts == nil {
		alerts = []model.StockAlert{}
	}
	c.JSON(http.StatusOK, alerts)
}

// ── Partner Pharmacies ─────────────────────────────────────────────────

func (h *PharmacyHandler) ListPartnerPharmacies(c *gin.Context) {
	partners, err := h.svc.ListPartnerPharmacies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if partners == nil {
		partners = []model.PartnerPharmacy{}
	}
	c.JSON(http.StatusOK, partners)
}

// ── Transfer Requests ──────────────────────────────────────────────────

func (h *PharmacyHandler) CreateTransferRequest(c *gin.Context) {
	pharmacistID, _ := c.Get("userID")

	var req service.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validation
	if req.PartnerID == 0 || req.MedicineID == 0 || req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "partner_id, medicine_id, and quantity are required"})
		return
	}

	transfer, err := h.svc.CreateTransferRequest(pharmacistID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transfer)
}

func (h *PharmacyHandler) ListTransferRequests(c *gin.Context) {
	status := c.Query("status")

	transfers, err := h.svc.ListTransferRequests(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if transfers == nil {
		transfers = []model.TransferRequest{}
	}
	c.JSON(http.StatusOK, transfers)
}

func (h *PharmacyHandler) UpdateTransferRequest(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	err := h.svc.UpdateTransferRequestStatus(id, body.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer request updated"})
}
