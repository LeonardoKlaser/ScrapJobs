package repository

import (
	"database/sql"
	"fmt"
	"web-scrapper/model"

	"github.com/lib/pq"
)

type PlanRepository struct {
	connection *sql.DB
}

func NewPlanRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{
		connection: db,
	}
}

func (pr *PlanRepository) GetAllPlans() ([]model.Plan, error) {
	query := `SELECT id, name, price, max_sites, max_ai_analyses, features FROM plans`
	rows, err := pr.connection.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar planos: %w", err)
	}
	defer rows.Close()

	var plans []model.Plan
	for rows.Next() {
		var plan model.Plan
		var features pq.StringArray
		if err := rows.Scan(&plan.ID, &plan.Name, &plan.Price, &plan.MaxSites, &plan.MaxAIAnalyses, &features); err != nil {
			return nil, fmt.Errorf("erro ao ler dados do plano: %w", err)
		}
		plan.Features = features
		plans = append(plans, plan)
	}

	return plans, nil
}

func (pr *PlanRepository) GetPlanByUserID(userID int) (*model.Plan, error) {
	query := `
        SELECT p.id, p.name, p.price, p.max_sites, p.max_ai_analyses, p.features
        FROM plans p
        JOIN users u ON u.plan_id = p.id
        WHERE u.id = $1`

	row := pr.connection.QueryRow(query, userID)

	var plan model.Plan
	var features pq.StringArray
	err := row.Scan(&plan.ID, &plan.Name, &plan.Price, &plan.MaxSites, &plan.MaxAIAnalyses, &features)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil 
		}
		return nil, fmt.Errorf("erro ao buscar plano do usuário: %w", err)
	}
	plan.Features = features

	return &plan, nil
}