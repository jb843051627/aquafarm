package service

import (
	"context"
	"fmt"
	"time"

	"aquafarm/internal/model"
	"aquafarm/internal/store"
)

// Service wraps the repository and provides business logic.
type Service struct {
	repo *store.Repo
}

// New creates a new Service.
func New(repo *store.Repo) *Service {
	return &Service{repo: repo}
}

// CreateTank creates a new fish tank.
func (s *Service) CreateTank(ctx context.Context, t *model.Tank) (*model.Tank, error) {
	if t.Status == "" {
		t.Status = model.TankStatusActive
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = t.CreatedAt

	if err := model.ValidateTank(t); err != nil {
		return nil, err
	}

	if err := s.repo.Tanks().Create(t); err != nil {
		return nil, fmt.Errorf("create tank: %w", err)
	}
	return t, nil
}

// GetTank retrieves a tank by ID.
func (s *Service) GetTank(ctx context.Context, id int64) (*model.Tank, error) {
	tank, err := s.repo.Tanks().GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("get tank %d: %w", id, err)
	}
	return tank, nil
}

// ListTanks returns all tanks.
func (s *Service) ListTanks(ctx context.Context) ([]model.Tank, error) {
	return s.repo.Tanks().List()
}

// ListTanksByStatus returns tanks filtered by status.
func (s *Service) ListTanksByStatus(ctx context.Context, status string) ([]model.Tank, error) {
	return s.repo.Tanks().ListByStatus(status)
}

// UpdateTank updates a tank.
func (s *Service) UpdateTank(ctx context.Context, t *model.Tank) (*model.Tank, error) {
	if err := model.ValidateTank(t); err != nil {
		return nil, err
	}
	if err := s.repo.Tanks().Update(t); err != nil {
		return nil, fmt.Errorf("update tank %d: %w", t.ID, err)
	}
	return t, nil
}

// UpdateTankStock updates the stock quantity of a tank.
func (s *Service) UpdateTankStock(ctx context.Context, id int64, stock int) error {
	if stock < 0 {
		return &model.ValidationError{Field: "stock_qty", Message: "stock cannot be negative"}
	}
	if err := s.repo.Tanks().UpdateStock(id, stock); err != nil {
		return fmt.Errorf("update tank stock: %w", err)
	}
	return nil
}

// DeleteTank removes a tank and all associated data.
func (s *Service) DeleteTank(ctx context.Context, id int64) error {
	if err := s.repo.Cleanup(id); err != nil {
		return fmt.Errorf("delete tank %d: %w", id, err)
	}
	return nil
}

// SearchTanks finds tanks by name or species.
func (s *Service) SearchTanks(ctx context.Context, query string) ([]model.Tank, error) {
	return s.repo.SearchTanks(query)
}

// GetTankSummary returns a summary of the tank's state.
func (s *Service) GetTankSummary(ctx context.Context, id int64) (*store.TankSummary, error) {
	return s.repo.GetTankSummary(id)
}

// GetSystemOverview returns system-wide metrics.
func (s *Service) GetSystemOverview(ctx context.Context) (*store.SystemOverview, error) {
	return s.repo.GetSystemOverview()
}

// GetTanksNeedingAttention returns IDs of tanks with unresolved alerts or faults.
func (s *Service) GetTanksNeedingAttention(ctx context.Context) ([]int64, error) {
	return s.repo.GetTanksNeedingAttention()
}

// CreateFeedPlan creates a new feed plan.
func (s *Service) CreateFeedPlan(ctx context.Context, f *model.FeedPlan) (*model.FeedPlan, error) {
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now()
	}
	f.UpdatedAt = f.CreatedAt
	f.Active = true // Active defaults to true via DB schema

	if err := model.ValidateFeedPlan(f); err != nil {
		return nil, err
	}

	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(f.TankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}

	if err := s.repo.Feeds().CreatePlan(f); err != nil {
		return nil, fmt.Errorf("create feed plan: %w", err)
	}
	return f, nil
}

// ListFeedPlans returns all feed plans.
func (s *Service) ListFeedPlans(ctx context.Context) ([]model.FeedPlan, error) {
	return s.repo.Feeds().ListPlans()
}

// ListFeedPlansByTank returns feed plans for a tank.
func (s *Service) ListFeedPlansByTank(ctx context.Context, tankID int64) ([]model.FeedPlan, error) {
	return s.repo.Feeds().ListPlansByTank(tankID)
}

// ListActiveFeedPlans returns active feed plans.
func (s *Service) ListActiveFeedPlans(ctx context.Context) ([]model.FeedPlan, error) {
	return s.repo.Feeds().ListActivePlans()
}

// UpdateFeedPlan updates a feed plan.
func (s *Service) UpdateFeedPlan(ctx context.Context, f *model.FeedPlan) (*model.FeedPlan, error) {
	f.UpdatedAt = time.Now()
	if err := model.ValidateFeedPlan(f); err != nil {
		return nil, err
	}
	if err := s.repo.Feeds().UpdatePlan(f); err != nil {
		return nil, fmt.Errorf("update feed plan: %w", err)
	}
	return f, nil
}

// DeactivateFeedPlan deactivates a feed plan.
func (s *Service) DeactivateFeedPlan(ctx context.Context, id int64) error {
	return s.repo.Feeds().DeactivatePlan(id)
}

// DeleteFeedPlan removes a feed plan.
func (s *Service) DeleteFeedPlan(ctx context.Context, id int64) error {
	return s.repo.Feeds().DeletePlan(id)
}

// RecordManualFeed records a manual feeding event.
func (s *Service) RecordManualFeed(ctx context.Context, tankID int64, feedType string, amount float64) (*model.FeedLog, error) {
	if amount <= 0 {
		return nil, &model.ValidationError{Field: "amount", Message: "amount must be positive"}
	}

	// Verify tank exists
	tank, err := s.repo.Tanks().GetByID(tankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}

	log := &model.FeedLog{
		TankID:    tankID,
		FeedType:  feedType,
		Amount:    amount,
		Source:    model.FeedSourceManual,
		Timestamp: time.Now(),
	}

	if err := s.repo.Feeds().LogFeed(log); err != nil {
		return nil, fmt.Errorf("log feed: %w", err)
	}
	return log, nil
}

// ExecuteScheduledFeed executes a scheduled feed plan.
func (s *Service) ExecuteScheduledFeed(ctx context.Context, planID int64) (*model.FeedLog, error) {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("context cancelled before feed: %w", ctxErr)
	}

	plan, err := s.repo.Feeds().GetPlanByID(planID)
	if err != nil {
		return nil, fmt.Errorf("get feed plan %d: %w", planID, err)
	}
	if plan == nil {
		return nil, model.ErrNotFound
	}
	if !plan.Active {
		return nil, &model.ValidationError{Field: "active", Message: "feed plan is not active"}
	}

	// Verify tank still exists and is active
	tank, err := s.repo.Tanks().GetByID(plan.TankID)
	if err != nil {
		return nil, fmt.Errorf("check tank: %w", err)
	}
	if tank == nil {
		return nil, model.ErrNotFound
	}
	if tank.Status == model.TankStatusMaintenance {
		return nil, &model.ValidationError{Field: "tank", Message: "tank is in maintenance mode"}
	}

	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, fmt.Errorf("context cancelled before logging feed: %w", ctxErr)
	}

	log := &model.FeedLog{
		TankID:    plan.TankID,
		PlanID:    planID,
		FeedType:  plan.FeedType,
		Amount:    plan.Amount,
		Source:    model.FeedSourceAuto,
		Timestamp: time.Now(),
	}

	if err := s.repo.Feeds().LogFeed(log); err != nil {
		return nil, fmt.Errorf("log scheduled feed: %w", err)
	}
	return log, nil
}

// ListFeedLogs returns feed logs for a tank.
func (s *Service) ListFeedLogs(ctx context.Context, tankID int64, limit int) ([]model.FeedLog, error) {
	return s.repo.Feeds().ListFeedLogs(tankID, limit)
}

// GetFeedTotal returns total feed amount for a tank.
func (s *Service) GetFeedTotal(ctx context.Context, tankID int64) (float64, error) {
	return s.repo.Feeds().GetFeedTotalByTank(tankID)
}
