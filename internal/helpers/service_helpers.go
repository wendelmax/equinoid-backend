package helpers

import (
	"context"

	"github.com/equinoid/backend/internal/models"
	"github.com/equinoid/backend/internal/modules/equinos"
	apperrors "github.com/equinoid/backend/pkg/errors"
	"github.com/equinoid/backend/pkg/logging"
)

func FindEquinoOrError(ctx context.Context, equinoRepo equinos.Repository, logger *logging.Logger, equinoidID string, serviceName string) (*models.Equino, error) {
	equino, err := equinoRepo.FindByEquinoid(ctx, equinoidID)
	if err != nil {
		if apperrors.IsNotFound(err) {
			return nil, &apperrors.NotFoundError{
				Resource: "equino",
				Message:  "equino não encontrado",
				ID:       equinoidID,
			}
		}
		logger.LogError(err, serviceName, logging.Fields{"equinoid": equinoidID})
		return nil, apperrors.NewDatabaseError("find_equino", "erro ao buscar equino", err)
	}
	return equino, nil
}

func HandleServiceError(err error, logger *logging.Logger, serviceName string, operation string, fields logging.Fields) error {
	if apperrors.IsNotFound(err) {
		return err
	}
	logger.LogError(err, serviceName, fields)
	return apperrors.NewDatabaseError(operation, "erro ao executar operação", err)
}
