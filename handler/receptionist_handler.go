package handler

import (
	"net/http"
	"strconv"
	"time"

	"apollo-backend/model"
	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type ReceptionistHandler struct {
	svc *service.ReceptionistService
}

func NewReceptionistHandler(svc *service.ReceptionistService) *ReceptionistHandler {
	return &ReceptionistHandler{svc: svc}
}

// ─── Response DTOs (REQ-REC-009: no consultation_notes / clinical data) ───────

type patientInfo struct {
	PatientID             uint   `json:"patient_id"`
	PatCode               string `json:"pat_code"`
	FullName              string `json:"full_name"`
	Gender                string `json:"gender"`
	BloodGroup            string `json:"blood_group,omitempty"`
	ContactNumber         string `json:"contact_number"`
	Address               string `json:"address,omitempty"`
	EmergencyContactName  string `json:"emergency_contact_name,omitempty"`
	EmergencyContactPhone string `json:"emergency_contact_phone,omitempty"`
	InsuranceID           string `json:"insurance_id,omitempty"`
	InsuranceProvider     string `json:"insurance_provider,omitempty"`
	DateOfBirth           string `json:"date_of_birth"`
	CreatedAt             string `json:"created_at"`
}

type doctorInfo struct {
	DoctorID       uint   `json:"doctor_id"`
	FullName       string `json:"full_name"`
	Specialization string `json:"specialization,omitempty"`
}

type deptInfo struct {
	DeptID uint   `json:"dept_id"`
	Name   string `json:"name"`
}

type apptResponse struct {
	ApptID           uint        `json:"appt_id"`
	PatientID        uint        `json:"patient_id"`
	DoctorID         uint        `json:"doctor_id"`
	DeptID           uint        `json:"dept_id"`
	ScheduledAt      time.Time   `json:"scheduled_at"`
	QueueToken       string      `json:"queue_token,omitempty"`
	Status           string      `json:"status"`
	ChiefComplaint   string      `json:"chief_complaint,omitempty"`
	CreatedByStaffID *uint       `json:"created_by_staff_id,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Patient          *patientInfo `json:"patient,omitempty"`
	Doctor           *doctorInfo  `json:"doctor,omitempty"`
	Department       *deptInfo    `json:"department,omitempty"`
}

// parseUintParam parses a Gin URL param as uint; returns (0, false) on failure.
func parseUintParam(c *gin.Context, key string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + key})
		return 0, false
	}
	return uint(v), true
}

// staffIDFromCtx extracts the authenticated staff's userID from JWT claims.
func staffIDFromCtx(c *gin.Context) uint {
	if id, ok := c.Get("userID"); ok {
		if uid, ok := id.(uint); ok {
			return uid
		}
	}
	return 0
}

// ─── Dashboard ─────────────────────────────────────────────────────────────────

func (h *ReceptionistHandler) GetDashboard(c *gin.Context) {
	stats, err := h.svc.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ─── Patient ───────────────────────────────────────────────────────────────────

func (h *ReceptionistHandler) RegisterPatient(c *gin.Context) {
	var input service.RegisterPatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.svc.RegisterPatient(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *ReceptionistHandler) SearchPatients(c *gin.Context) {
	q := c.Query("q")
	bloodGroup := c.Query("blood_group")
	insurance := c.Query("insurance")
	patients, err := h.svc.SearchPatients(q, bloodGroup, insurance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, patients)
}

func (h *ReceptionistHandler) GetPatient(c *gin.Context) {
	id, ok := parseUintParam(c, "patient_id")
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

func (h *ReceptionistHandler) UpdatePatientContact(c *gin.Context) {
	id, ok := parseUintParam(c, "patient_id")
	if !ok {
		return
	}
	var input service.UpdateContactInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdatePatientContact(id, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "contact updated"})
}

func (h *ReceptionistHandler) GetPatientRegistrationCard(c *gin.Context) {
	id, ok := parseUintParam(c, "patient_id")
	if !ok {
		return
	}
	p, err := h.svc.GetPatient(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	// Return a clean registration card payload (subset of patient data)
	card := gin.H{
		"pat_code":                p.PatCode,
		"full_name":               p.FullName,
		"date_of_birth":           p.DateOfBirth,
		"gender":                  p.Gender,
		"blood_group":             p.BloodGroup,
		"contact_number":          p.ContactNumber,
		"address":                 p.Address,
		"emergency_contact_name":  p.EmergencyContactName,
		"emergency_contact_phone": p.EmergencyContactPhone,
		"insurance_id":            p.InsuranceID,
		"insurance_provider":      p.InsuranceProvider,
		"registered_on":           p.CreatedAt,
	}
	c.JSON(http.StatusOK, card)
}

// ─── Appointment ───────────────────────────────────────────────────────────────

func (h *ReceptionistHandler) BookAppointment(c *gin.Context) {
	var input service.BookAppointmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	staffID := staffIDFromCtx(c)
	appt, err := h.svc.BookAppointment(staffID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toApptResponse(appt))
}

func (h *ReceptionistHandler) ListAppointments(c *gin.Context) {
	type query struct {
		Status   string `form:"status"`
		Date     string `form:"date"`
		DeptID   uint   `form:"dept_id"`
		DoctorID uint   `form:"doctor_id"`
	}
	var q query
	_ = c.ShouldBindQuery(&q)
	appts, err := h.svc.ListAppointments(q.Status, q.Date, q.DeptID, q.DoctorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	out := make([]apptResponse, len(appts))
	for i, a := range appts {
		out[i] = toApptResponse(&a)
	}
	c.JSON(http.StatusOK, out)
}

func (h *ReceptionistHandler) GetAppointment(c *gin.Context) {
	id, ok := parseUintParam(c, "appt_id")
	if !ok {
		return
	}
	appt, err := h.svc.GetAppointment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toApptResponse(appt))
}

func (h *ReceptionistHandler) RescheduleAppointment(c *gin.Context) {
	id, ok := parseUintParam(c, "appt_id")
	if !ok {
		return
	}
	var input service.RescheduleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.RescheduleAppointment(id, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "appointment rescheduled"})
}

func (h *ReceptionistHandler) CancelAppointment(c *gin.Context) {
	id, ok := parseUintParam(c, "appt_id")
	if !ok {
		return
	}
	if err := h.svc.CancelAppointment(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "appointment cancelled"})
}

func (h *ReceptionistHandler) CheckInAppointment(c *gin.Context) {
	id, ok := parseUintParam(c, "appt_id")
	if !ok {
		return
	}
	appt, err := h.svc.CheckInAppointment(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toApptResponse(appt))
}

func (h *ReceptionistHandler) WalkIn(c *gin.Context) {
	var input service.WalkInInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	staffID := staffIDFromCtx(c)
	appt, err := h.svc.WalkIn(staffID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toApptResponse(appt))
}

func (h *ReceptionistHandler) GetAppointmentSlip(c *gin.Context) {
	id, ok := parseUintParam(c, "appt_id")
	if !ok {
		return
	}
	appt, err := h.svc.GetAppointment(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	slip := gin.H{
		"appt_id":         appt.ApptID,
		"queue_token":     appt.QueueToken,
		"scheduled_at":    appt.ScheduledAt,
		"status":          appt.Status,
		"chief_complaint": appt.ChiefComplaint,
	}
	if appt.Patient.PatientID != 0 {
		slip["patient"] = gin.H{
			"patient_id":     appt.Patient.PatientID,
			"pat_code":       appt.Patient.PatCode,
			"full_name":      appt.Patient.FullName,
			"contact_number": appt.Patient.ContactNumber,
			"blood_group":    appt.Patient.BloodGroup,
		}
	}
	if appt.Doctor.DoctorID != 0 {
		slip["doctor"] = gin.H{"full_name": appt.Doctor.FullName}
	}
	if appt.Department.DeptID != 0 {
		slip["department"] = gin.H{"name": appt.Department.Name}
	}
	c.JSON(http.StatusOK, slip)
}

// ─── Visitor Log ───────────────────────────────────────────────────────────────

func (h *ReceptionistHandler) LogVisitor(c *gin.Context) {
	var input service.LogVisitorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	staffID := staffIDFromCtx(c)
	v, err := h.svc.LogVisitor(staffID, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *ReceptionistHandler) ListVisitors(c *gin.Context) {
	date := c.Query("date")
	var patientID uint
	if pid := c.Query("patient_id"); pid != "" {
		if v, err := strconv.ParseUint(pid, 10, 64); err == nil {
			patientID = uint(v)
		}
	}
	visitors, err := h.svc.ListVisitors(date, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, visitors)
}

func (h *ReceptionistHandler) CheckoutVisitor(c *gin.Context) {
	id, ok := parseUintParam(c, "visitor_id")
	if !ok {
		return
	}
	if err := h.svc.CheckoutVisitor(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "visitor checked out"})
}

// ─── Helper ────────────────────────────────────────────────────────────────────

func toApptResponse(a *model.Appointment) apptResponse {
	r := apptResponse{
		ApptID:           a.ApptID,
		PatientID:        a.PatientID,
		DoctorID:         a.DoctorID,
		DeptID:           a.DeptID,
		ScheduledAt:      a.ScheduledAt,
		QueueToken:       a.QueueToken,
		Status:           a.Status,
		ChiefComplaint:   a.ChiefComplaint,
		CreatedByStaffID: a.CreatedByStaffID,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
	if a.Patient.PatientID != 0 {
		r.Patient = &patientInfo{
			PatientID:             a.Patient.PatientID,
			PatCode:               a.Patient.PatCode,
			FullName:              a.Patient.FullName,
			Gender:                a.Patient.Gender,
			BloodGroup:            a.Patient.BloodGroup,
			ContactNumber:         a.Patient.ContactNumber,
			Address:               a.Patient.Address,
			EmergencyContactName:  a.Patient.EmergencyContactName,
			EmergencyContactPhone: a.Patient.EmergencyContactPhone,
			InsuranceID:           a.Patient.InsuranceID,
			InsuranceProvider:     a.Patient.InsuranceProvider,
			DateOfBirth:           a.Patient.DateOfBirth.Format("2006-01-02"),
			CreatedAt:             a.Patient.CreatedAt.Format(time.RFC3339),
		}
	}
	if a.Doctor.DoctorID != 0 {
		r.Doctor = &doctorInfo{
			DoctorID:       a.Doctor.DoctorID,
			FullName:       a.Doctor.FullName,
			Specialization: a.Doctor.Specialization,
		}
	}
	if a.Department.DeptID != 0 {
		r.Department = &deptInfo{DeptID: a.Department.DeptID, Name: a.Department.Name}
	}
	return r
}
