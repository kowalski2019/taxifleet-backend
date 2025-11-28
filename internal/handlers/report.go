package handlers

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"taxifleet/backend/internal/permissions"
	"taxifleet/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type ReportHandler struct {
	service *service.ReportService
}

func NewReportHandler(service *service.ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

func (h *ReportHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	userID, _ := c.Get("userID")
	permission, _ := c.Get("permission")

	reports, err := h.service.List(tenantID.(uint), userID.(uint), permission.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

func (h *ReportHandler) Create(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	userID, _ := c.Get("userID")

	var req service.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.Create(tenantID.(uint), userID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, report)
}

func (h *ReportHandler) Get(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	report, err := h.service.GetByID(uint(id), tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Update(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	userID, _ := c.Get("userID")
	permission, _ := c.Get("permission")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req service.UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.Update(uint(id), tenantID.(uint), userID.(uint), permission.(int), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Submit(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	userID, _ := c.Get("userID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	report, err := h.service.Submit(uint(id), tenantID.(uint), userID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Approve(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	approvedByID, _ := c.Get("userID")
	permission, _ := c.Get("permission")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	report, err := h.service.Approve(uint(id), tenantID.(uint), approvedByID.(uint), permission.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Delete(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	userID, _ := c.Get("userID")
	permission, _ := c.Get("permission")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	err = h.service.Delete(uint(id), tenantID.(uint), userID.(uint), permission.(int))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}

func (h *ReportHandler) Reject(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	report, err := h.service.Reject(uint(id), tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

func (h *ReportHandler) Export(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	userID, _ := c.Get("userID")
	permission, _ := c.Get("permission")

	// Check if user has permission to export (owner or manager)
	userPerm := permission.(int)
	if !permissions.HasPermission(userPerm, permissions.PermissionViewReports) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to export reports"})
		return
	}

	reports, err := h.service.List(tenantID.(uint), userID.(uint), userPerm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "csv"
	}

	// Generate filename with random ID and date
	rand.Seed(time.Now().UnixNano())
	randomID := rand.Intn(1000000)
	dateStr := time.Now().Format("20060102")
	filename := fmt.Sprintf("reports-%d-%s.%s", randomID, dateStr, format)

	// Helper function to format date as dd/month/year
	formatDateDDMMYYYY := func(t time.Time) string {
		return fmt.Sprintf("%02d/%02d/%d", t.Day(), t.Month(), t.Year())
	}

	if format == "csv" {
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write header
		writer.Write([]string{"ID", "Week Start", "Taxi", "Driver", "Earnings", "Expenses", "Status", "Notes", "Created At"})

		// Write data
		for _, report := range reports {
			writer.Write([]string{
				strconv.Itoa(int(report.ID)),
				formatDateDDMMYYYY(report.WeekStartDate),
				report.Taxi.LicensePlate,
				report.Driver.FirstName + " " + report.Driver.LastName,
				strconv.FormatFloat(report.Earnings, 'f', 2, 64),
				strconv.FormatFloat(report.TotalExpenses, 'f', 2, 64),
				report.Status,
				report.Notes,
				formatDateDDMMYYYY(report.CreatedAt),
			})
		}
	} else if format == "xlsx" {
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}()

		sheetName := "Reports"
		index, err := f.NewSheet(sheetName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		f.SetActiveSheet(index)

		// Write header
		headers := []string{"ID", "Week Start", "Taxi", "Driver", "Earnings", "Expenses", "Status", "Notes", "Created At"}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Write data
		for rowIdx, report := range reports {
			row := rowIdx + 2
			cell1, _ := excelize.CoordinatesToCellName(1, row)
			cell2, _ := excelize.CoordinatesToCellName(2, row)
			cell3, _ := excelize.CoordinatesToCellName(3, row)
			cell4, _ := excelize.CoordinatesToCellName(4, row)
			cell5, _ := excelize.CoordinatesToCellName(5, row)
			cell6, _ := excelize.CoordinatesToCellName(6, row)
			cell7, _ := excelize.CoordinatesToCellName(7, row)
			cell8, _ := excelize.CoordinatesToCellName(8, row)
			cell9, _ := excelize.CoordinatesToCellName(9, row)
			f.SetCellValue(sheetName, cell1, report.ID)
			f.SetCellValue(sheetName, cell2, formatDateDDMMYYYY(report.WeekStartDate))
			f.SetCellValue(sheetName, cell3, report.Taxi.LicensePlate)
			f.SetCellValue(sheetName, cell4, report.Driver.FirstName+" "+report.Driver.LastName)
			f.SetCellValue(sheetName, cell5, report.Earnings)
			f.SetCellValue(sheetName, cell6, report.TotalExpenses)
			f.SetCellValue(sheetName, cell7, report.Status)
			f.SetCellValue(sheetName, cell8, report.Notes)
			f.SetCellValue(sheetName, cell9, formatDateDDMMYYYY(report.CreatedAt))
		}

		// Remove default sheet
		f.DeleteSheet("Sheet1")

		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		if err := f.Write(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported format. Use 'csv' or 'xlsx'"})
	}
}
