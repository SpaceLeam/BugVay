package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kokuroshesh/bugvay/internal/services"
)

func ListAssets(service *services.AssetService) gin.HandlerFunc {
	return func(c *gin.Context) {
		programID, _ := strconv.Atoi(c.Query("program_id"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		assets, err := service.ListAssets(c.Request.Context(), programID, limit, offset)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": assets})
	}
}

func CreateAsset(service *services.AssetService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.CreateAssetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		asset, err := service.CreateAsset(c.Request.Context(), &req)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": asset})
	}
}

func GetAsset(service *services.AssetService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		asset, err := service.GetAsset(c.Request.Context(), id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": asset})
	}
}

func DeleteAsset(service *services.AssetService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		if err := service.DeleteAsset(c.Request.Context(), id); err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "asset deleted"})
	}
}
