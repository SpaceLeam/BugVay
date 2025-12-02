package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kokuroshesh/bugvay/internal/api/handlers"
	"github.com/kokuroshesh/bugvay/internal/api/middleware"
	"github.com/kokuroshesh/bugvay/internal/database"
	"github.com/kokuroshesh/bugvay/internal/queue"
	"github.com/kokuroshesh/bugvay/internal/services"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(pg *database.PostgresDB, ch *database.ClickHouseDB, q *queue.Client) *Router {
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler())

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// Initialize services
	endpointService := services.NewEndpointService(pg, ch, q)
	scanService := services.NewScanService(pg, ch, q)
	findingService := services.NewFindingService(pg, ch)
	programService := services.NewProgramService(pg)
	assetService := services.NewAssetService(pg)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Programs
		programs := v1.Group("/programs")
		{
			programs.GET("", handlers.ListPrograms(programService))
			programs.POST("", handlers.CreateProgram(programService))
			programs.GET("/:id", handlers.GetProgram(programService))
		}

		// Assets
		assets := v1.Group("/assets")
		{
			assets.GET("", handlers.ListAssets(assetService))
			assets.POST("", handlers.CreateAsset(assetService))
			assets.GET("/:id", handlers.GetAsset(assetService))
			assets.DELETE("/:id", handlers.DeleteAsset(assetService))
		}

		// Endpoints
		endpoints := v1.Group("/endpoints")
		{
			endpoints.POST("/upload", handlers.UploadEndpoints(endpointService))
			endpoints.GET("", handlers.ListEndpoints(endpointService))
			endpoints.GET("/:id", handlers.GetEndpoint(endpointService))
		}

		// Scans
		scans := v1.Group("/scans")
		{
			scans.POST("", handlers.CreateScan(scanService))
			scans.GET("", handlers.ListScans(scanService))
			scans.GET("/:id", handlers.GetScan(scanService))
		}

		// Findings
		findings := v1.Group("/findings")
		{
			findings.GET("", handlers.ListFindings(findingService))
			findings.GET("/:id", handlers.GetFinding(findingService))
			findings.PATCH("/:id/triage", handlers.TriageFinding(findingService))
		}

		// Jobs (Asynq status)
		jobs := v1.Group("/jobs")
		{
			jobs.GET("", handlers.ListJobs(q))
			jobs.GET("/:id", handlers.GetJobStatus(q))
		}
	}

	return &Router{engine: r}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}
