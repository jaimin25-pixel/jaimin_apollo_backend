package handler

import (
	"net/http"
	"strconv"

	"apollo-backend/model"
	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type OTHandler struct {
	service service.OTService
}

func NewOTHandler(svc service.OTService) *OTHandler {
	return &OTHandler{service: svc}
}

func (h *OTHandler) GetDashboard(c *gin.Context) {
	date := c.Query("date")
	stats, err := h.service.GetDashboard(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *OTHandler) ListSchedules(c *gin.Context) {
	date := c.Query("date")
	status := c.Query("status")
	var roomID *uint
	if rid := c.Query("room_id"); rid != "" {
		if id, err := strconv.ParseUint(rid, 10, 32); err == nil {
			uid := uint(id)
			roomID = &uid
		}
	}
	schedules, err := h.service.ListSchedules(date, roomID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

func (h *OTHandler) CreateSchedule(c *gin.Context) {
	var sched model.OTSchedule
	if err := c.ShouldBindJSON(&sched); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.CreateSchedule(&sched); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "OT booked successfully", "schedule": sched})
}

func (h *OTHandler) GetSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	sched, err := h.service.GetSchedule(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}
	c.JSON(http.StatusOK, sched)
}

func (h *OTHandler) UpdateSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateSchedule(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Schedule updated"})
}

func (h *OTHandler) AdvanceStatus(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.AdvanceStatus(uint(id), body.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status advanced"})
}

func (h *OTHandler) CancelSchedule(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&body)
	if err := h.service.CancelSchedule(uint(id), body.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Schedule cancelled"})
}

func (h *OTHandler) AddNotes(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		SurgicalNotes    string `json:"surgical_notes"`
		AnesthesiaRecord string `json:"anesthesia_record"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.AddNotes(uint(id), body.SurgicalNotes, body.AnesthesiaRecord); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notes added"})
}

func (h *OTHandler) SterilizeRoom(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.SterilizeRoom(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Room sterilized"})
}

func (h *OTHandler) ListRooms(c *gin.Context) {
	date := c.Query("date")
	rooms, err := h.service.ListRooms(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}
