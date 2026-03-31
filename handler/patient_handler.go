package handler

import (
	"net/http"
	"strconv"

	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

// ─── PatientHandler ───────────────────────────────────────────────────────────

type PatientHandler struct {
	svc *service.PatientService
}

func NewPatientHandler(svc *service.PatientService) *PatientHandler {
	return &PatientHandler{svc: svc}
}

// patientParseID parses a Gin URL param as uint.
func patientParseID(c *gin.Context, key string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + key})
		return 0, false
	}
	return uint(v), true
}

// ─── 1. SearchPatients — GET /api/patient/search ─────────────────────────────

func (h *PatientHandler) SearchPatients(c *gin.Context) {
	q := c.Query("q")
	patients, err := h.svc.SearchPatients(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, patients)
}

// ─── 2. RegisterPatient — POST /api/patient/register ─────────────────────────

func (h *PatientHandler) RegisterPatient(c *gin.Context) {
	var input service.PatientRegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, duplicates, err := h.svc.RegisterPatient(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp := gin.H{"patient": p}
	if len(duplicates) > 0 {
		resp["duplicate_warning"] = "Potential duplicate patients found"
		resp["duplicates"] = duplicates
	}
	c.JSON(http.StatusCreated, resp)
}

// ─── 3. GetPatient — GET /api/patient/:id ────────────────────────────────────

func (h *PatientHandler) GetPatient(c *gin.Context) {
	id, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.GetPatient(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// ─── 4. UpdatePatient — PUT /api/patient/:id ─────────────────────────────────

func (h *PatientHandler) UpdatePatient(c *gin.Context) {
	id, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	var input service.PatientUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdatePatient(id, input); err != nil {
		if err.Error() == "patient not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "patient updated"})
}

// ─── 5. ListAppointments — GET /api/patient/:id/appointments ─────────────────

func (h *PatientHandler) ListAppointments(c *gin.Context) {
	id, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	status := c.Query("status")
	fromDate := c.Query("from_date")
	appts, err := h.svc.ListAppointments(id, status, fromDate)
	if err != nil {
		if err.Error() == "patient not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, appts)
}

// ─── 6. BookAppointment — POST /api/patient/:id/appointments ─────────────────

func (h *PatientHandler) BookAppointment(c *gin.Context) {
	id, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	var input service.PatientBookAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	appt, err := h.svc.BookAppointment(id, input)
	if err != nil {
		if err.Error() == "patient not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, appt)
}

// ─── 7. UpdateAppointment — PATCH /api/patient/:id/appointments/:appt_id ─────

func (h *PatientHandler) UpdateAppointment(c *gin.Context) {
	_, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	apptID, ok := patientParseID(c, "appt_id")
	if !ok {
		return
	}
	var input service.PatientRescheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateAppointment(apptID, input); err != nil {
		if err.Error() == "appointment not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "appointment updated"})
}

// ─── 8. CheckInAppointment — PATCH /api/patient/:id/appointments/:appt_id/checkin

func (h *PatientHandler) CheckInAppointment(c *gin.Context) {
	_, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	apptID, ok := patientParseID(c, "appt_id")
	if !ok {
		return
	}
	appt, err := h.svc.CheckInAppointment(apptID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":     "checked in",
		"queue_token": appt.QueueToken,
		"appointment": appt,
	})
}

// ─── 9. ListAdmissions — GET /api/patient/:id/admissions ─────────────────────

func (h *PatientHandler) ListAdmissions(c *gin.Context) {
	id, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	status := c.Query("status")
	admissions, err := h.svc.ListAdmissions(id, status)
	if err != nil {
		if err.Error() == "patient not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, admissions)
}

// ─── 10. GetAdmission — GET /api/patient/:id/admissions/:adm_id ──────────────

func (h *PatientHandler) GetAdmission(c *gin.Context) {
	_, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	admID, ok := patientParseID(c, "adm_id")
	if !ok {
		return
	}
	adm, err := h.svc.GetAdmission(admID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, adm)
}

// ─── 11. GetFullEHR — GET /api/patient/:id/ehr ───────────────────────────────

func (h *PatientHandler) GetFullEHR(c *gin.Context) {
	id, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	ehr, err := h.svc.GetFullEHR(id)
	if err != nil {
		if err.Error() == "patient not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ehr)
}

// ─── 12. ListInvoices — GET /api/patient/:id/invoices ────────────────────────

func (h *PatientHandler) ListInvoices(c *gin.Context) {
	id, ok := patientParseID(c, "id")
	if !ok {
		return
	}
	status := c.Query("status")
	invoices, err := h.svc.ListInvoices(id, status)
	if err != nil {
		if err.Error() == "patient not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invoices)
}
