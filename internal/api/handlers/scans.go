package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kokuroshesh/bugvay/internal/services"
)

func CreateScan(service *services.ScanService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.ScanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		scan, err := service.CreateScan(c.Request.Context(), &req)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": scan})
	}
}

func ListScans(service *services.ScanService) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
	}
}

func GetScan(service *services.ScanService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		scan, err := service.GetScanStatus(c.Request.Context(), id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": scan})
	}
}
