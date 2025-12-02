package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kokuroshesh/bugvay/internal/queue"
)

func ListJobs(client *queue.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement Asynq inspector to list jobs
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
	}
}

func GetJobStatus(client *queue.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		jobID := c.Param("id")
		status, err := client.GetJobStatus(c.Request.Context(), jobID)
		if err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"job_id": jobID,
			"status": status,
		})
	}
}
