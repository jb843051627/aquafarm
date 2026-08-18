package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"aquafarm/internal/model"
)

// --- Reading Handlers ---

func (h *Handler) IngestReading(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var r model.SensorReading
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r.TankID = tankID
	created, err := h.svc.IngestReading(c.Request.Context(), &r)
	if err != nil {
		status := http.StatusInternalServerError
		if err == model.ErrNotFound {
			status = http.StatusNotFound
		}
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) BatchIngestReadings(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var readings []model.SensorReading
	if err := c.ShouldBindJSON(&readings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for i := range readings {
		readings[i].TankID = tankID
	}
	count, err := h.svc.BatchIngestReadings(c.Request.Context(), readings)
	if err != nil {
		status := http.StatusInternalServerError
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ingested": count})
}

func (h *Handler) ListReadings(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	readings, err := h.svc.ListReadingsByTank(c.Request.Context(), tankID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, readings)
}

func (h *Handler) GetLatestReadings(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	readings, err := h.svc.GetLatestReadings(c.Request.Context(), tankID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, readings)
}

func (h *Handler) GetReadingTrend(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	sensorType := c.Query("type")
	if sensorType == "" {
		sensorType = "temperature"
	}
	trend, err := h.svc.GetReadingTrend(c.Request.Context(), tankID, sensorType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}

// --- Alert Handlers ---

func (h *Handler) ListAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	alerts, err := h.svc.ListAlerts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *Handler) ListUnresolvedAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	alerts, err := h.svc.ListUnresolvedAlerts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *Handler) ListAlertsByTank(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	alerts, err := h.svc.ListAlertsByTank(c.Request.Context(), tankID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, alerts)
}

func (h *Handler) ResolveAlert(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.ResolveAlert(c.Request.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if err == model.ErrNotFound {
			status = http.StatusNotFound
		}
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "resolved"})
}

func (h *Handler) ResolveAlertsByTank(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	count, err := h.svc.ResolveAlertsByTank(c.Request.Context(), tankID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"resolved": count})
}

func (h *Handler) GetAlertStats(c *gin.Context) {
	stats, err := h.svc.GetAlertStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// --- Feed Handlers ---

func (h *Handler) ListFeedPlans(c *gin.Context) {
	plans, err := h.svc.ListFeedPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *Handler) CreateFeedPlan(c *gin.Context) {
	var f model.FeedPlan
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.CreateFeedPlan(c.Request.Context(), &f)
	if err != nil {
		status := http.StatusInternalServerError
		if err == model.ErrNotFound {
			status = http.StatusNotFound
		}
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (h *Handler) ListFeedPlansByTank(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	plans, err := h.svc.ListFeedPlansByTank(c.Request.Context(), tankID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *Handler) UpdateFeedPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var f model.FeedPlan
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f.ID = id
	updated, err := h.svc.UpdateFeedPlan(c.Request.Context(), &f)
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

func (h *Handler) DeleteFeedPlan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteFeedPlan(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) ExecuteScheduledFeed(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	log, err := h.svc.ExecuteScheduledFeed(c.Request.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if err == model.ErrNotFound {
			status = http.StatusNotFound
		}
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, log)
}

func (h *Handler) RecordManualFeed(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		FeedType string  `json:"feed_type"`
		Amount   float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log, err := h.svc.RecordManualFeed(c.Request.Context(), tankID, body.FeedType, body.Amount)
	if err != nil {
		status := http.StatusInternalServerError
		if err == model.ErrNotFound {
			status = http.StatusNotFound
		}
		if _, ok := err.(*model.ValidationError); ok {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, log)
}

func (h *Handler) ListFeedLogs(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	logs, err := h.svc.ListFeedLogs(c.Request.Context(), tankID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
