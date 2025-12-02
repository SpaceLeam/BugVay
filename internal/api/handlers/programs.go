package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kokuroshesh/bugvay/internal/services"
)

func ListPrograms(service *services.ProgramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		programs, err := service.ListPrograms(c.Request.Context(), limit, offset)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": programs})
	}
}

func CreateProgram(service *services.ProgramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.CreateProgramRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			return
		}

		program, err := service.CreateProgram(c.Request.Context(), &req)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"data": program})
	}
}

func GetProgram(service *services.ProgramService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		program, err := service.GetProgram(c.Request.Context(), id)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": program})
	}
}
