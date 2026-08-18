package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"aquafarm/internal/model"
	"aquafarm/internal/monitor"
	"aquafarm/internal/service"
)

// Handler wraps the service and provides HTTP handlers.
type Handler struct {
	svc *service.Service
	mon *monitor.Monitor
}

// New creates a new handler.
func New(svc *service.Service, mon *monitor.Monitor) *Handler {
	return &Handler{svc: svc, mon: mon}
}

// Router sets up the gin router with all routes.
func (h *Handler) Router() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// API routes
	api := r.Group("/api")
	{
		// Tank routes
		api.GET("/tanks", h.ListTanks)
		api.POST("/tanks", h.CreateTank)
		api.GET("/tanks/:id", h.GetTank)
		api.PUT("/tanks/:id", h.UpdateTank)
		api.DELETE("/tanks/:id", h.DeleteTank)
		api.GET("/tanks/search", h.SearchTanks)
		api.GET("/tanks/:id/summary", h.GetTankSummary)
		api.PUT("/tanks/:id/stock", h.UpdateTankStock)

		// Reading routes
		api.GET("/tanks/:id/readings", h.ListReadings)
		api.POST("/tanks/:id/readings", h.IngestReading)
		api.POST("/tanks/:id/readings/batch", h.BatchIngestReadings)
		api.GET("/tanks/:id/readings/latest", h.GetLatestReadings)
		api.GET("/tanks/:id/readings/trend", h.GetReadingTrend)

		// Alert routes
		api.GET("/alerts", h.ListAlerts)
		api.GET("/alerts/unresolved", h.ListUnresolvedAlerts)
		api.GET("/tanks/:id/alerts", h.ListAlertsByTank)
		api.POST("/alerts/:id/resolve", h.ResolveAlert)
		api.POST("/tanks/:id/alerts/resolve", h.ResolveAlertsByTank)
		api.GET("/alerts/stats", h.GetAlertStats)

		// Feed routes
		api.GET("/feed-plans", h.ListFeedPlans)
		api.POST("/feed-plans", h.CreateFeedPlan)
		api.GET("/tanks/:id/feed-plans", h.ListFeedPlansByTank)
		api.PUT("/feed-plans/:id", h.UpdateFeedPlan)
		api.DELETE("/feed-plans/:id", h.DeleteFeedPlan)
		api.POST("/feed-plans/:id/execute", h.ExecuteScheduledFeed)
		api.POST("/tanks/:id/feed/manual", h.RecordManualFeed)
		api.GET("/tanks/:id/feed-logs", h.ListFeedLogs)

		// Equipment routes
		api.GET("/equipment", h.ListEquipment)
		api.POST("/equipment", h.CreateEquipment)
		api.GET("/equipment/:id", h.GetEquipment)
		api.PUT("/equipment/:id", h.UpdateEquipment)
		api.DELETE("/equipment/:id", h.DeleteEquipment)
		api.GET("/equipment/status/:status", h.ListEquipmentByStatus)
		api.GET("/equipment/:id/health", h.GetEquipmentHealth)

		// Maintenance routes
		api.GET("/maintenance", h.ListMaintenanceTasks)
		api.POST("/maintenance", h.CreateMaintenanceTask)
		api.GET("/maintenance/:id", h.GetMaintenanceTask)
		api.POST("/maintenance/:id/complete", h.CompleteMaintenanceTask)
		api.GET("/maintenance/overdue", h.ListOverdueTasks)

		// Batch routes
		api.GET("/batches", h.ListBatches)
		api.POST("/batches", h.CreateBatch)
		api.GET("/tanks/:id/batches", h.ListBatchesByTank)
		api.POST("/tanks/:id/mortality", h.RecordMortality)
		api.GET("/tanks/:id/mortality", h.ListMortalityLogs)

		// Water change routes
		api.POST("/tanks/:id/water-changes", h.RecordWaterChange)
		api.GET("/tanks/:id/water-changes", h.ListWaterChanges)

		// Threshold routes
		api.POST("/thresholds", h.SaveThreshold)
		api.GET("/tanks/:id/thresholds", h.GetThresholdsByTank)
		api.DELETE("/thresholds/:id", h.DeleteThreshold)

		// System routes
		api.GET("/overview", h.GetSystemOverview)
		api.GET("/config", h.ListSystemConfig)
		api.PUT("/config/:key", h.SetSystemConfig)
		api.GET("/tanks/attention", h.GetTanksNeedingAttention)
	}

	// Serve embedded frontend
	r.StaticFile("/", "./web/index.html")
	r.Static("/web/static", "./web/static")

	return r
}

// --- Tank Handlers ---

func (h *Handler) ListTanks(c *gin.Context) {
	tanks, err := h.svc.ListTanks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tanks)
}

func (h *Handler) CreateTank(c *gin.Context) {
	var t model.Tank
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.CreateTank(c.Request.Context(), &t)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) GetTank(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	tank, err := h.svc.GetTank(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tank not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tank)
}

func (h *Handler) UpdateTank(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var t model.Tank
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t.ID = id
	updated, err := h.svc.UpdateTank(c.Request.Context(), &t)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeleteTank(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteTank(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) SearchTanks(c *gin.Context) {
	q := c.Query("q")
	tanks, err := h.svc.SearchTanks(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tanks)
}

func (h *Handler) GetTankSummary(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	summary, err := h.svc.GetTankSummary(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "tank not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

func (h *Handler) UpdateTankStock(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Stock int `json:"stock"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.UpdateTankStock(c.Request.Context(), id, body.Stock); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
