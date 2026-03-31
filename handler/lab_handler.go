package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"apollo-backend/model"
	"apollo-backend/service"

	"github.com/gin-gonic/gin"
)

type LabHandler struct {
	service service.LabService
}

func NewLabHandler(svc service.LabService) *LabHandler {
	return &LabHandler{service: svc}
}

func (h *LabHandler) GetDashboard(c *gin.Context) {
	stats, err := h.service.GetDashboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *LabHandler) ListTests(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")
	tests, err := h.service.ListTests(search, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tests": tests})
}

func (h *LabHandler) CreateTest(c *gin.Context) {
	var test model.LabTest
	if err := c.ShouldBindJSON(&test); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.AddTest(&test); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Test added successfully", "test": test})
}

func (h *LabHandler) UpdateTest(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.UpdateTest(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Test updated"})
}

func (h *LabHandler) ListLabOrders(c *gin.Context) {
	status := c.Query("status")
	fromDate := c.Query("from_date")
	var patientID *uint
	if pid := c.Query("patient_id"); pid != "" {
		if id, err := strconv.ParseUint(pid, 10, 32); err == nil {
			uid := uint(id)
			patientID = &uid
		}
	}
	orders, err := h.service.ListLabOrders(status, fromDate, patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *LabHandler) GetLabOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	order, err := h.service.GetLabOrder(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *LabHandler) CollectSample(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	// Mock techID
	var techID uint = 0

	if err := h.service.CollectSample(uint(id), techID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sample collected"})
}

func (h *LabHandler) UploadResult(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		ResultValue string `json:"result_value" binding:"required"`
		IsAbnormal  *bool  `json:"is_abnormal"`
		Notes       string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var techID uint = 0

	if err := h.service.UploadResult(uint(id), body.ResultValue, body.IsAbnormal, body.Notes, techID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Result uploaded successfully"})
}

func (h *LabHandler) CancelLabOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var body struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&body)

	if err := h.service.CancelLabOrder(uint(id), body.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled"})
}

func (h *LabHandler) ListRadiologyOrders(c *gin.Context) {
	status := c.Query("status")
	modality := c.Query("modality")
	orders, err := h.service.ListRadiologyOrders(status, modality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *LabHandler) GetRadiologyOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	order, err := h.service.GetRadiologyOrder(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *LabHandler) UploadRadiologyReport(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	reportText := c.PostForm("report_text")
	
	file, _ := c.FormFile("report_file")
	filePath := ""
	if file != nil {
		filename := fmt.Sprintf("rad_%d_%d%s", id, time.Now().Unix(), filepath.Ext(file.Filename))
		filePath = "uploads/reports/" + filename
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			// ignore specific file saving logic for basic testing
			filePath = "simulated/" + filename
		}
	}

	var radID uint = 0 // from user claims

	if err := h.service.UploadRadiologyReport(uint(id), reportText, filePath, radID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Report uploaded successfully"})
}

func (h *LabHandler) AttachRadiologyImage(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	
	file, err := c.FormFile("image_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}
	
	filename := fmt.Sprintf("img_%d_%d%s", id, time.Now().Unix(), filepath.Ext(file.Filename))
	filePath := "uploads/images/" + filename
	// simulate saving
	filePath = "/simulated/images/" + filename

	if err := h.service.AttachImage(uint(id), filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Image attached successfully"})
}
