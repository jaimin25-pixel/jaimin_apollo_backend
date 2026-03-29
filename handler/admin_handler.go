package handler

import (
	"net/http"
	"strconv"
	"time"

	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id), true
}

func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	from, _ := time.Parse("2006-01-02", c.Query("from_date"))
	to, _ := time.Parse("2006-01-02", c.Query("to_date"))
	if from.IsZero() {
		from = time.Now().AddDate(0, -1, 0)
	}
	if to.IsZero() {
		to = time.Now()
	}
	return from, to
}

// ── 1. GET /api/admin/dashboard ─────────────────────────────────────

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	data, err := h.svc.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load dashboard"})
		return
	}
	c.JSON(http.StatusOK, data)
}

// ── 2-7. Doctors CRUD ───────────────────────────────────────────────

func (h *AdminHandler) ListDoctors(c *gin.Context) {
	deptID, _ := strconv.ParseUint(c.Query("dept_id"), 10, 32)
	status := c.Query("status")
	search := c.Query("search")

	docs, err := h.svc.ListDoctors(uint(deptID), status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *AdminHandler) GetDoctor(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	doc, appts, rxs, err := h.svc.GetDoctor(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"doctor":        doc,
		"appointments":  appts,
		"prescriptions": rxs,
	})
}

func (h *AdminHandler) CreateDoctor(c *gin.Context) {
	var input service.CreateDoctorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := h.svc.CreateDoctor(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, doc)
}

func (h *AdminHandler) UpdateDoctor(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.UpdateDoctorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc, err := h.svc.UpdateDoctor(id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, doc)
}

func (h *AdminHandler) UpdateDoctorStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.StatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateDoctorStatus(id, input.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *AdminHandler) DeleteDoctor(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.svc.DeleteDoctor(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "doctor deleted"})
}

// ── 8-10. Departments ───────────────────────────────────────────────

func (h *AdminHandler) ListDepartments(c *gin.Context) {
	depts, err := h.svc.ListDepartments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, depts)
}

func (h *AdminHandler) CreateDepartment(c *gin.Context) {
	var input service.CreateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept, err := h.svc.CreateDepartment(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dept)
}

func (h *AdminHandler) UpdateDepartment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.UpdateDepartmentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dept, err := h.svc.UpdateDepartment(id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dept)
}

func (h *AdminHandler) UpdateDepartmentStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input service.StatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateDepartmentStatus(id, input.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *AdminHandler) DeleteDepartment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteDepartment(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "department deleted"})
}

// ── 11-15. Staff ────────────────────────────────────────────────────

func (h *AdminHandler) ListStaff(c *gin.Context) {
	role := c.Query("role")
	deptID, _ := strconv.ParseUint(c.Query("dept_id"), 10, 32)
	status := c.Query("status")

	staff, err := h.svc.ListStaff(role, uint(deptID), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

func (h *AdminHandler) GetStaff(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	staff, err := h.svc.GetStaff(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

func (h *AdminHandler) CreateStaff(c *gin.Context) {
	var input service.CreateStaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staff, err := h.svc.CreateStaff(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, staff)
}

func (h *AdminHandler) UpdateStaff(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.UpdateStaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staff, err := h.svc.UpdateStaff(id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, staff)
}

func (h *AdminHandler) UpdateStaffStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.StatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdateStaffStatus(id, input.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

func (h *AdminHandler) DeleteStaff(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteStaff(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "staff deleted"})
}

// ── 16-18. Pharmacists ──────────────────────────────────────────────

func (h *AdminHandler) ListPharmacists(c *gin.Context) {
	status := c.Query("status")
	list, err := h.svc.ListPharmacists(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *AdminHandler) CreatePharmacist(c *gin.Context) {
	var input service.CreatePharmacistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.svc.CreatePharmacist(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *AdminHandler) UpdatePharmacist(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.UpdatePharmacistInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.svc.UpdatePharmacist(id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

// ── 19-22. Partner Pharmacies ───────────────────────────────────────

func (h *AdminHandler) ListPartnerPharmacies(c *gin.Context) {
	status := c.Query("status")
	list, err := h.svc.ListPartnerPharmacies(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *AdminHandler) CreatePartnerPharmacy(c *gin.Context) {
	var input service.CreatePartnerPharmacyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.svc.CreatePartnerPharmacy(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *AdminHandler) UpdatePartnerPharmacy(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.UpdatePartnerPharmacyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.svc.UpdatePartnerPharmacy(id, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *AdminHandler) UpdatePartnerPharmacyStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var input service.StatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdatePartnerPharmacyStatus(id, input.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

// ── 23-24. Hospital Config ──────────────────────────────────────────

func (h *AdminHandler) GetConfig(c *gin.Context) {
	cfg, err := h.svc.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var input service.UpdateConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg, err := h.svc.UpdateConfig(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// ── 25-27. Reports ──────────────────────────────────────────────────

func (h *AdminHandler) FinancialReport(c *gin.Context) {
	from, to := parseDateRange(c)
	report := h.svc.FinancialReport(from, to)
	c.JSON(http.StatusOK, report)
}

func (h *AdminHandler) OccupancyReport(c *gin.Context) {
	from, to := parseDateRange(c)
	deptID, _ := strconv.ParseUint(c.Query("dept_id"), 10, 32)
	report := h.svc.OccupancyReport(from, to, uint(deptID))
	c.JSON(http.StatusOK, report)
}

func (h *AdminHandler) PrescriptionReport(c *gin.Context) {
	from, to := parseDateRange(c)
	report := h.svc.PrescriptionReport(from, to)
	c.JSON(http.StatusOK, report)
}

// ── 28. Export ───────────────────────────────────────────────────────

func (h *AdminHandler) Export(c *gin.Context) {
	var input service.ExportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.Export(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
