package handler

import (
	"net/http"
	"strconv"

	"apollo-backend/model"
	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type NursingHandler struct {
	service service.NursingService
}

func NewNursingHandler(svc service.NursingService) *NursingHandler {
	return &NursingHandler{service: svc}
}

func (h *NursingHandler) GetDashboard(c *gin.Context) {
	var wardID *uint
	if idStr := c.Query("ward_id"); idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			uid := uint(id)
			wardID = &uid
		}
	}
	stats, err := h.service.GetDashboard(wardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *NursingHandler) ListWards(c *gin.Context) {
	var deptID *uint
	if idStr := c.Query("dept_id"); idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			uid := uint(id)
			deptID = &uid
		}
	}
	wards, err := h.service.ListWards(deptID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wards": wards})
}

func (h *NursingHandler) GetBeds(c *gin.Context) {
	wardID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	status := c.Query("status")
	beds, err := h.service.GetBedsByWard(uint(wardID), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"beds": beds})
}

func (h *NursingHandler) UpdateBedStatus(c *gin.Context) {
	bedID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateBedStatus(uint(bedID), body.Status, body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bed status updated"})
}

func (h *NursingHandler) TransferBedRequest(c *gin.Context) {
	var req model.BedTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// user, _ := c.Get("user")
	// auto-populate requested_by here if needed

	if err := h.service.TransferBed(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Transfer requested", "transfer": req})
}

func (h *NursingHandler) ConfirmTransfer(c *gin.Context) {
	transferID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Notes string `json:"notes"`
	}
	c.ShouldBindJSON(&body)

	if err := h.service.ConfirmTransfer(uint(transferID), body.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Transfer confirmed"})
}

func (h *NursingHandler) MarkBedHousekeeping(c *gin.Context) {
	bedID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.MarkBedClean(uint(bedID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Bed marked cleanly available"})
}

func (h *NursingHandler) ListVitals(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	fromDate := c.Query("from_date")
	vitals, err := h.service.ListVitals(uint(patientID), fromDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"vitals": vitals})
}

func (h *NursingHandler) RecordVital(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var vital model.Vital
	if err := c.ShouldBindJSON(&vital); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	vital.PatientID = uint(patientID)

	u, exists := c.Get("user")
	if exists {
		user := u.(*model.User)
		if user.Role == "nurse" {
			// Actually need staff profile for nurse. We will just use the active staff account if linked.
			// Defaulting recordedRole:
		}
		vital.RecordedByRole = user.Role
	}

	if err := h.service.RecordVital(&vital); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Vitals recorded", "vital": vital})
}

func (h *NursingHandler) GetMAR(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	date := c.Query("date")
	mars, err := h.service.GetMAR(uint(patientID), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"mar": mars})
}

func (h *NursingHandler) UpdateMARStatus(c *gin.Context) {
	marID, _ := strconv.ParseUint(c.Param("item_id"), 10, 32)
	var body struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var nurseID uint = 0 // ideal would be extracting from JWT

	if err := h.service.UpdateMARStatus(uint(marID), body.Status, body.Reason, body.Notes, nurseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "MAR updated"})
}

func (h *NursingHandler) AddNursingNote(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var note model.NursingNote
	if err := c.ShouldBindJSON(&note); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	note.PatientID = uint(patientID)
	
	// NurseID should come from token, mocking a 1 for now if needed, or if frontend passes it.

	if err := h.service.AddNursingNote(&note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Nursing note added", "note": note})
}
