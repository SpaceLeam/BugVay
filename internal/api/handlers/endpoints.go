package handlers

import (
	"bufio"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kokuroshesh/bugvay/internal/services"
)

type UploadRequest struct {
	AssetID int `form:"asset_id" binding:"required"`
}

func UploadEndpoints(service *services.EndpointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UploadRequest
		if err := c.ShouldBind(&req); err != nil {
			c.Error(err)
			return
		}

		file, _, err := c.Request.FormFile("file")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "file required"})
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		created := 0
		skipped := 0

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			_, err := service.CreateEndpoint(c.Request.Context(), req.AssetID, line, "upload")
			if err != nil {
				skipped++
				continue
			}
			created++
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "endpoints uploaded",
			"created": created,
			"skipped": skipped,
		})
	}
}

func ListEndpoints(service *services.EndpointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		assetID, _ := strconv.Atoi(c.Query("asset_id"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		endpoints, err := service.ListEndpoints(c.Request.Context(), assetID, limit, offset)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": endpoints})
	}
}

func GetEndpoint(service *services.EndpointService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		endpoint, err := service.GetEndpoint(c.Request.Context(), id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": endpoint})
	}
}
