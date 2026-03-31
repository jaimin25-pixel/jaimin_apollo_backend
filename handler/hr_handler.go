package handler

import (
	"net/http"
	"strconv"

	"apollo-backend/model"
	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type HRHandler struct {
	service service.HRService
}

func NewHRHandler(svc service.HRService) *HRHandler {
	return &HRHandler{service: svc}
}

func (h *HRHandler) GetDashboard(c *gin.Context) {
	stats, err := h.service.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *HRHandler) ListStaff(c *gin.Context) {
	var deptID *uint
	if did := c.Query("dept_id"); did != "" {
		if id, err := strconv.ParseUint(did, 10, 32); err == nil {
			uid := uint(id)
			deptID = &uid
		}
	}
	role := c.Query("role")

	staff, err := h.service.ListStaff(deptID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"staff": staff})
}

func (h *HRHandler) CreateStaff(c *gin.Context) {
	var req service.StaffCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreateStaff(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Staff profile created successfully"})
}

func (h *HRHandler) UpdateStaff(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateStaff(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Staff updated"})
}

func (h *HRHandler) ListAttendance(c *gin.Context) {
	date := c.Query("date")
	var staffID *uint
	if sid := c.Query("staff_id"); sid != "" {
		if id, err := strconv.ParseUint(sid, 10, 32); err == nil {
			uid := uint(id)
			staffID = &uid
		}
	}
	shifts, err := h.service.ListShifts(staffID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"attendance": shifts})
}

func (h *HRHandler) ClockIn(c *gin.Context) {
	var body struct {
		StaffID uint `json:"staff_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ClockInOut(body.StaffID, "in"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Clocked in"})
}

func (h *HRHandler) ClockOut(c *gin.Context) {
	var body struct {
		StaffID uint `json:"staff_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ClockInOut(body.StaffID, "out"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Clocked out"})
}

func (h *HRHandler) ListLeaves(c *gin.Context) {
	status := c.Query("status")
	var staffID *uint
	if sid := c.Query("staff_id"); sid != "" {
		if id, err := strconv.ParseUint(sid, 10, 32); err == nil {
			uid := uint(id)
			staffID = &uid
		}
	}
	leaves, err := h.service.ListLeaves(staffID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"leaves": leaves})
}

func (h *HRHandler) ApplyLeave(c *gin.Context) {
	var req model.LeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.ApplyLeave(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Leave applied successfully", "leave": req})
}

func (h *HRHandler) ProcessLeave(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Mock approver
	var approverID uint = 0
	
	if err := h.service.ProcessLeave(uint(id), body.Status, body.Notes, &approverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Leave processed"})
}

func (h *HRHandler) ListPayroll(c *gin.Context) {
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	var staffID *uint
	if sid := c.Query("staff_id"); sid != "" {
		if id, err := strconv.ParseUint(sid, 10, 32); err == nil {
			uid := uint(id)
			staffID = &uid
		}
	}
	payrolls, err := h.service.ListPayroll(month, year, staffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"payrolls": payrolls})
}

func (h *HRHandler) GeneratePayroll(c *gin.Context) {
	var body struct {
		Month int `json:"month" binding:"required"`
		Year  int `json:"year" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	u, _ := c.Get("user")
	user := u.(*model.User)
	
	if err := h.service.GeneratePayroll(body.Month, body.Year, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Payroll generated successfully"})
}
