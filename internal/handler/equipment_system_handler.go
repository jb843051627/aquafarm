package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"aquafarm/internal/model"
)

// --- Equipment Handlers ---

func (h *Handler) ListEquipment(c *gin.Context) {
	equipment, err := h.svc.ListEquipment(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, equipment)
}

func (h *Handler) CreateEquipment(c *gin.Context) {
	var e model.Equipment
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.CreateEquipment(c.Request.Context(), &e)
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

func (h *Handler) GetEquipment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	eq, err := h.svc.GetEquipment(c.Request.Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "equipment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, eq)
}

func (h *Handler) UpdateEquipment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var e model.Equipment
	if err := c.ShouldBindJSON(&e); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	e.ID = id
	updated, err := h.svc.UpdateEquipment(c.Request.Context(), &e)
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

func (h *Handler) DeleteEquipment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteEquipment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) ListEquipmentByStatus(c *gin.Context) {
	status := c.Param("status")
	equipment, err := h.svc.ListEquipmentByStatus(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, equipment)
}

func (h *Handler) GetEquipmentHealth(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	score, err := h.svc.GetEquipmentHealth(c.Request.Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "equipment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"health_score": score})
}

// --- Maintenance Handlers ---

func (h *Handler) ListMaintenanceTasks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	tasks, err := h.svc.ListMaintenanceTasks(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *Handler) CreateMaintenanceTask(c *gin.Context) {
	var t model.MaintenanceTask
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.CreateMaintenanceTask(c.Request.Context(), &t)
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

func (h *Handler) GetMaintenanceTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	task, err := h.svc.GetMaintenanceTask(c.Request.Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) CompleteMaintenanceTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Technician string `json:"technician"`
	}
	_ = c.ShouldBindJSON(&body)
	if err := h.svc.CompleteMaintenanceTask(c.Request.Context(), id, body.Technician); err != nil {
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
	c.JSON(http.StatusOK, gin.H{"message": "completed"})
}

func (h *Handler) ListOverdueTasks(c *gin.Context) {
	tasks, err := h.svc.ListOverdueTasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// --- Batch Handlers ---

func (h *Handler) ListBatches(c *gin.Context) {
	batches, err := h.svc.ListBatches(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batches)
}

func (h *Handler) CreateBatch(c *gin.Context) {
	var b model.Batch
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.svc.CreateBatch(c.Request.Context(), &b)
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

func (h *Handler) ListBatchesByTank(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	batches, err := h.svc.ListBatchesByTank(c.Request.Context(), tankID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batches)
}

func (h *Handler) RecordMortality(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body struct {
		Count int    `json:"count"`
		Cause string `json:"cause"`
		Note  string `json:"note"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.RecordMortality(c.Request.Context(), tankID, body.Count, body.Cause, body.Note); err != nil {
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
	c.JSON(http.StatusOK, gin.H{"message": "recorded"})
}

func (h *Handler) ListMortalityLogs(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	logs, err := h.svc.ListMortalityLogs(c.Request.Context(), tankID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}

// --- Water Change Handlers ---

func (h *Handler) RecordWaterChange(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var w model.WaterChange
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	w.TankID = tankID
	if w.Timestamp.IsZero() {
		w.Timestamp = time.Now()
	}
	created, err := h.svc.RecordWaterChange(c.Request.Context(), &w)
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

func (h *Handler) ListWaterChanges(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	changes, err := h.svc.ListWaterChanges(c.Request.Context(), tankID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, changes)
}

// --- Threshold Handlers ---

func (h *Handler) SaveThreshold(c *gin.Context) {
	var t model.ThresholdConfig
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saved, err := h.svc.SaveThreshold(c.Request.Context(), &t)
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
	c.JSON(http.StatusCreated, saved)
}

func (h *Handler) GetThresholdsByTank(c *gin.Context) {
	tankID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	thresholds, err := h.svc.GetThresholdsByTank(c.Request.Context(), tankID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, thresholds)
}

func (h *Handler) DeleteThreshold(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteThreshold(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// --- System Handlers ---

func (h *Handler) GetSystemOverview(c *gin.Context) {
	overview, err := h.svc.GetSystemOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *Handler) ListSystemConfig(c *gin.Context) {
	config, err := h.svc.ListSystemConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *Handler) SetSystemConfig(c *gin.Context) {
	key := c.Param("key")
	var body struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SetSystemConfig(c.Request.Context(), key, body.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) GetTanksNeedingAttention(c *gin.Context) {
	tankIDs, err := h.svc.GetTanksNeedingAttention(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tank_ids": tankIDs})
}
