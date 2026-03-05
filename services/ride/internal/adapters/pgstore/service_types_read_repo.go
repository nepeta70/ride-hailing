package pgstore

import (
	"context"

	"github.com/nepeta70/ride-hailing/internal/pkg/adapters/pgstore"
	"github.com/nepeta70/ride-hailing/services/ride/internal/config"
	"github.com/nepeta70/ride-hailing/services/ride/internal/core/domain"
	"github.com/nepeta70/ride-hailing/services/ride/internal/ports"
)

type ServiceTypeReadEntity struct {
	ID            string `db:"id"`
	DisplayName   string `db:"display_name"`
	MaxPassengers int    `db:"max_passengers"`
}

type ServiceTypeReadRepo struct {
	config *config.Config
	db     *pgstore.PostgresDB
}

func NewServiceTypeReadRepo(config *config.Config, db *pgstore.PostgresDB) *ServiceTypeReadRepo {
	return &ServiceTypeReadRepo{
		config: config,
		db:     db,
	}
}

func (r *ServiceTypeReadRepo) GetAllEnabled(ctx context.Context) (map[string]*domain.ServiceType, error) {
	query := `
		SELECT id, display_name, max_passengers
		FROM service_types 
		WHERE is_active = true 
		ORDER BY sort_order ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*domain.ServiceType, 10)
	for rows.Next() {
		var s ServiceTypeReadEntity
		err := rows.Scan(
			&s.ID,
			&s.DisplayName,
			&s.MaxPassengers,
		)
		if err != nil {
			return nil, err
		}

		results[s.ID] = &domain.ServiceType{
			Code:          s.ID,
			Name:          s.DisplayName,
			MaxPassengers: s.MaxPassengers,
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Ensure implementation matches the interface
var _ ports.ServiceTypeReadRepository = (*ServiceTypeReadRepo)(nil)
