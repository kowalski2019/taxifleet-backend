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

type ExpenseHandler struct {
	service *service.ExpenseService
}

func NewExpenseHandler(service *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

func (h *ExpenseHandler) List(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	expenses, err := h.service.List(tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expenses)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	userID, _ := c.Get("userID")

	var req service.CreateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service.Create(tenantID.(uint), userID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, expense)
}

func (h *ExpenseHandler) Get(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	expense, err := h.service.GetByID(uint(id), tenantID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req service.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	expense, err := h.service.Update(uint(id), tenantID.(uint), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := h.service.Delete(uint(id), tenantID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expense deleted successfully"})
}

func (h *ExpenseHandler) Export(c *gin.Context) {
	tenantID, _ := c.Get("tenantID")
	permission, _ := c.Get("permission")

	// Check if user has permission to export (owner or manager)
	userPerm := permission.(int)
	if !permissions.HasPermission(userPerm, permissions.PermissionViewExpenses) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to export expenses"})
		return
	}

	expenses, err := h.service.List(tenantID.(uint))
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
	filename := fmt.Sprintf("expenses-%d-%s.%s", randomID, dateStr, format)

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
		writer.Write([]string{"ID", "Date", "Category", "Amount", "Taxi", "Reason", "Created At"})

		// Write data
		for _, expense := range expenses {
			taxiPlate := ""
			if expense.Taxi != nil {
				taxiPlate = expense.Taxi.LicensePlate
			}
			writer.Write([]string{
				strconv.Itoa(int(expense.ID)),
				formatDateDDMMYYYY(expense.Date),
				expense.Category,
				strconv.FormatFloat(expense.Amount, 'f', 2, 64),
				taxiPlate,
				expense.Reason,
				formatDateDDMMYYYY(expense.CreatedAt),
			})
		}
	} else if format == "xlsx" {
		f := excelize.NewFile()
		defer func() {
			if err := f.Close(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
		}()

		sheetName := "Expenses"
		index, err := f.NewSheet(sheetName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		f.SetActiveSheet(index)

		// Write header
		headers := []string{"ID", "Date", "Category", "Amount", "Taxi", "Reason", "Created At"}
		for i, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, header)
		}

		// Write data
		for rowIdx, expense := range expenses {
			row := rowIdx + 2
			taxiPlate := ""
			if expense.Taxi != nil {
				taxiPlate = expense.Taxi.LicensePlate
			}
			cell1, _ := excelize.CoordinatesToCellName(1, row)
			cell2, _ := excelize.CoordinatesToCellName(2, row)
			cell3, _ := excelize.CoordinatesToCellName(3, row)
			cell4, _ := excelize.CoordinatesToCellName(4, row)
			cell5, _ := excelize.CoordinatesToCellName(5, row)
			cell6, _ := excelize.CoordinatesToCellName(6, row)
			cell7, _ := excelize.CoordinatesToCellName(7, row)
			f.SetCellValue(sheetName, cell1, expense.ID)
			f.SetCellValue(sheetName, cell2, formatDateDDMMYYYY(expense.Date))
			f.SetCellValue(sheetName, cell3, expense.Category)
			f.SetCellValue(sheetName, cell4, expense.Amount)
			f.SetCellValue(sheetName, cell5, taxiPlate)
			f.SetCellValue(sheetName, cell6, expense.Reason)
			f.SetCellValue(sheetName, cell7, formatDateDDMMYYYY(expense.CreatedAt))
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
