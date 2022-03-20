package repo

import (
	"context"
	"database/sql"
	"github.com/ATenderholt/lambda-router/pkg/database"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type RuntimeRepository struct {
	db database.Database
}

func NewRuntimeRepository(db database.Database) *RuntimeRepository {
	return &RuntimeRepository{db}
}

func (r RuntimeRepository) RuntimeExistsByName(ctx context.Context, runtime types.Runtime) (bool, error) {
	logger.Infof("Querying for Lambda Runtime %s.", runtime)
	var id int
	var name string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, name from lambda_runtime WHERE name = ?`,
		runtime,
	).Scan(&id, &name)

	switch {
	case err == sql.ErrNoRows:
		logger.Infof("Runtime %s not found", runtime)
		return false, nil
	case err != nil:
		e := Error{"unable to query runtime " + string(runtime), err}
		logger.Error(e)
		return false, e
	}

	logger.Infof("Runtime %s found with id=%s", name, id)
	return true, nil
}

func (r RuntimeRepository) RuntimeIDsByNames(ctx context.Context, runtimes []types.Runtime) (map[types.Runtime]int, error) {
	results := make(map[types.Runtime]int, len(runtimes))
	var resultError error = nil
	for _, runtime := range runtimes {
		var id int
		var name string
		err := r.db.QueryRowContext(
			ctx,
			`SELECT id, name from lambda_runtime WHERE name = ?`,
			runtime,
		).Scan(&id, &name)

		switch {
		case err == sql.ErrNoRows:
			logger.Errorf("unable to find Layer Runtime %s", runtime)
			resultError = sql.ErrNoRows
			results[runtime] = -1
		case err != nil:
			e := Error{"unable to query runtime " + string(runtime), err}
			logger.Error(e)
			return nil, e
		default:
			logger.Infof("Found Layer Runtime id=%d name=%s", id, name)
			results[runtime] = id
		}
	}

	return results, resultError
}
