package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kokuroshesh/bugvay/internal/services"
)

func ListFindings(service *services.FindingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		filters := map[string]interface{}{
			"severity": c.Query("severity"),
			"status":   c.Query("status"),
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		findings, err := service.ListFindings(c.Request.Context(), filters, limit, offset)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": findings})
	}
}

func GetFinding(service *services.FindingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		finding, err := service.GetFinding(c.Request.Context(), id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": finding})
	}
}

func TriageFinding(service *services.FindingService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req services.TriageRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		if err := service.TriageFinding(c.Request.Context(), id, &req); err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "finding updated"})
	}
}
