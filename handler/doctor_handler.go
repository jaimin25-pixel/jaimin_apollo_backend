package handler

import (
	"net/http"
	"strconv"

	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type DoctorHandler struct {
	svc *service.DoctorService
}

func NewDoctorHandler(svc *service.DoctorService) *DoctorHandler {
	return &DoctorHandler{svc: svc}
}

// resolveDoctorID returns the doctor ID from the JWT (role=doctor) or from the
// optional "doctor_id" query param (role=admin). Returns (0, true) for admin
// with no query param — meaning no filter. Returns (0, false) on parse error.
func resolveDoctorID(c *gin.Context) (uint, bool) {
	role, _ := c.Get("role")
	if role.(string) == "doctor" {
		uid, _ := c.Get("userID")
		return uid.(uint), true
	}
	raw := c.Query("doctor_id")
	if raw == "" {
		return 0, true
	}
	id, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor_id"})
		return 0, false
	}
	return uint(id), true
}

// ── Dashboard ─────────────────────────────────────────────────────────

func (h *DoctorHandler) GetDashboard(c *gin.Context) {
	doctorID, _ := c.Get("userID")
	data, err := h.svc.GetDashboard(doctorID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// ── Appointments ──────────────────────────────────────────────────────

func (h *DoctorHandler) ListAppointments(c *gin.Context) {
	doctorID, ok := resolveDoctorID(c)
	if !ok {
		return
	}
	var f service.AppointmentFilters
	_ = c.ShouldBindQuery(&f)

	appts, err := h.svc.ListAppointments(doctorID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, appts)
}

func (h *DoctorHandler) GetAppointment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	doctorID, _ := c.Get("userID")
	role, _ := c.Get("role")

	appt, err := h.svc.GetAppointment(id, doctorID.(uint), role.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, appt)
}

func (h *DoctorHandler) UpdateAppointmentStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input service.UpdateAppointmentStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doctorID, _ := c.Get("userID")
	if err := h.svc.UpdateAppointmentStatus(id, doctorID.(uint), input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "appointment status updated"})
}

// ── Patients ──────────────────────────────────────────────────────────

func (h *DoctorHandler) ListPatients(c *gin.Context) {
	doctorID, ok := resolveDoctorID(c)
	if !ok {
		return
	}
	role, _ := c.Get("role")
	var f service.PatientFilters
	_ = c.ShouldBindQuery(&f)

	patients, err := h.svc.ListPatients(doctorID, role.(string), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, patients)
}

func (h *DoctorHandler) GetPatientEHR(c *gin.Context) {
	patientID, ok := parseID(c)
	if !ok {
		return
	}
	doctorID, _ := c.Get("userID")
	role, _ := c.Get("role")

	ehr, err := h.svc.GetPatientEHR(patientID, doctorID.(uint), role.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ehr)
}

func (h *DoctorHandler) RecordVital(c *gin.Context) {
	patientID, ok := parseID(c)
	if !ok {
		return
	}
	var input service.RecordVitalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recorderID, _ := c.Get("userID")
	role, _ := c.Get("role")

	vital, err := h.svc.RecordVital(patientID, recorderID.(uint), role.(string), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, vital)
}

func (h *DoctorHandler) CreateClinicalNote(c *gin.Context) {
	patientID, ok := parseID(c)
	if !ok {
		return
	}
	var input service.CreateClinicalNoteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doctorID, _ := c.Get("userID")
	note, err := h.svc.CreateClinicalNote(patientID, doctorID.(uint), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, note)
}

// ── Prescriptions ─────────────────────────────────────────────────────

func (h *DoctorHandler) ListPrescriptions(c *gin.Context) {
	doctorID, ok := resolveDoctorID(c)
	if !ok {
		return
	}
	role, _ := c.Get("role")
	var f service.PrescriptionFilters
	_ = c.ShouldBindQuery(&f)

	rxs, err := h.svc.ListPrescriptions(doctorID, role.(string), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rxs)
}

func (h *DoctorHandler) CreatePrescription(c *gin.Context) {
	var input service.CreatePrescriptionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doctorID, _ := c.Get("userID")
	rx, err := h.svc.CreatePrescription(doctorID.(uint), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rx)
}

func (h *DoctorHandler) GetPrescription(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	doctorID, _ := c.Get("userID")
	role, _ := c.Get("role")

	rx, err := h.svc.GetPrescription(id, doctorID.(uint), role.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rx)
}

// ── Lab Orders ────────────────────────────────────────────────────────

func (h *DoctorHandler) ListLabOrders(c *gin.Context) {
	doctorID, _ := c.Get("userID")
	var f service.LabOrderFilters
	_ = c.ShouldBindQuery(&f)

	orders, err := h.svc.ListLabOrders(doctorID.(uint), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *DoctorHandler) CreateLabOrder(c *gin.Context) {
	var input service.CreateLabOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doctorID, _ := c.Get("userID")
	order, err := h.svc.CreateLabOrder(doctorID.(uint), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *DoctorHandler) GetLabOrder(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	doctorID, _ := c.Get("userID")
	order, err := h.svc.GetLabOrder(id, doctorID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// ── Radiology Orders ──────────────────────────────────────────────────

func (h *DoctorHandler) ListRadiologyOrders(c *gin.Context) {
	doctorID, _ := c.Get("userID")
	var f service.RadiologyOrderFilters
	_ = c.ShouldBindQuery(&f)

	orders, err := h.svc.ListRadiologyOrders(doctorID.(uint), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *DoctorHandler) CreateRadiologyOrder(c *gin.Context) {
	var input service.CreateRadiologyOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doctorID, _ := c.Get("userID")
	order, err := h.svc.CreateRadiologyOrder(doctorID.(uint), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *DoctorHandler) GetRadiologyOrder(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	doctorID, _ := c.Get("userID")
	order, err := h.svc.GetRadiologyOrder(id, doctorID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

// ── Admissions ────────────────────────────────────────────────────────

func (h *DoctorHandler) CreateAdmission(c *gin.Context) {
	var input service.CreateAdmissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doctorID, _ := c.Get("userID")
	adm, err := h.svc.CreateAdmission(doctorID.(uint), input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, adm)
}

func (h *DoctorHandler) DischargeAdmission(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input service.DischargeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	doctorID, _ := c.Get("userID")
	if err := h.svc.DischargeAdmission(id, doctorID.(uint), input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "patient discharged successfully"})
}
