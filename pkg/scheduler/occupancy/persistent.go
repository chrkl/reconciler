package occupancy

import (
	"fmt"
	"github.com/kyma-incubator/reconciler/pkg/db"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/kyma-incubator/reconciler/pkg/repository"
	"time"
)

type PersistentOccupancyRepository struct {
	*repository.Repository
}

func NewPersistentOccupancyRepository(conn db.Connection, debug bool) (Repository, error) {
	repo, err := repository.NewRepository(conn, debug)
	if err != nil {
		return nil, err
	}
	return &PersistentOccupancyRepository{repo}, nil
}

func (r *PersistentOccupancyRepository) WithTx(tx *db.TxConnection) (Repository, error) {
	return NewPersistentOccupancyRepository(tx, r.Debug)
}

func (r *PersistentOccupancyRepository) CreateWorkerPoolOccupancy(poolID, component string, poolSize int) (*model.WorkerPoolOccupancyEntity, error) {

	dbOps := func(tx *db.TxConnection) (interface{}, error) {

		occupancyEntity := &model.WorkerPoolOccupancyEntity{
			WorkerPoolID:       poolID,
			Component:          component,
			WorkerPoolCapacity: int64(poolSize),
			Created:            time.Now().UTC(),
		}

		createOccupancyQ, err := db.NewQuery(tx, occupancyEntity, r.Logger)
		if err != nil {
			return nil, err
		}
		if err = createOccupancyQ.Insert().Exec(); err != nil {
			r.Logger.Errorf("ReconRepo failed to create new worker-pool occupancy entity: %s", err)
			return nil, err
		}

		r.Logger.Debugf("ReconRepo created new worker-pool occupancy entity with poolID '%s'", poolID)
		return occupancyEntity, err
	}
	occupancyEntity, err := db.TransactionResult(r.Conn, dbOps, r.Logger)
	if err != nil {
		return nil, err
	}
	return occupancyEntity.(*model.WorkerPoolOccupancyEntity), nil
}

func (r *PersistentOccupancyRepository) GetComponentList() ([]string, error) {

	q, err := db.NewQuery(r.Conn, &model.WorkerPoolOccupancyEntity{}, r.Logger)
	if err != nil {
		return nil, err
	}

	databaseEntities, err := q.Select().GetMany()
	if err != nil {
		return nil, err
	}
	if len(databaseEntities) == 0 {
		return nil, fmt.Errorf("unable to get component list: no record was found")
	}
	var componentList []string
	for _, occupancy := range databaseEntities {
		occupancyEntity := occupancy.(*model.WorkerPoolOccupancyEntity)
		componentList = append(componentList, occupancyEntity.Component)
	}
	return componentList, nil
}

func (r *PersistentOccupancyRepository) GetMeanWorkerPoolOccupancyByComponent(component string) (float64, error) {

	q, err := db.NewQuery(r.Conn, &model.WorkerPoolOccupancyEntity{}, r.Logger)
	if err != nil {
		return 0, err
	}

	whereCond := map[string]interface{}{"Component": component}

	databaseEntities, err := q.Select().Where(whereCond).GetMany()
	if err != nil {
		return 0, err
	}

	if len(databaseEntities) == 0 {
		return 0, fmt.Errorf("unable to calculate worker pool capacity: no record was found for component: %s", component)
	}

	var aggregatedCapacity int64
	var aggregatedUsage int64
	for _, occupancy := range databaseEntities {
		occupancyEntity := occupancy.(*model.WorkerPoolOccupancyEntity)
		aggregatedUsage += occupancyEntity.RunningWorkers
		aggregatedCapacity += occupancyEntity.WorkerPoolCapacity
	}
	aggregatedOccupancy := 100 * float64(aggregatedUsage) / float64(aggregatedCapacity)
	return aggregatedOccupancy, nil
}

func (r *PersistentOccupancyRepository) GetWorkerPoolOccupancies() ([]*model.WorkerPoolOccupancyEntity, error) {

	q, err := db.NewQuery(r.Conn, &model.WorkerPoolOccupancyEntity{}, r.Logger)
	if err != nil {
		return nil, err
	}

	databaseEntities, err := q.Select().GetMany()
	if err != nil {
		return nil, err
	}
	if len(databaseEntities) == 0 {
		return nil, fmt.Errorf("unable to get occupancies list: no record was found")
	}
	var occupancyEntities []*model.WorkerPoolOccupancyEntity
	for _, occupancy := range databaseEntities {
		occupancyEntity := occupancy.(*model.WorkerPoolOccupancyEntity)
		occupancyEntities = append(occupancyEntities, occupancyEntity)
	}
	return occupancyEntities, nil
}

func (r *PersistentOccupancyRepository) FindWorkerPoolOccupancyByID(poolID string) (*model.WorkerPoolOccupancyEntity, error) {

	q, err := db.NewQuery(r.Conn, &model.WorkerPoolOccupancyEntity{}, r.Logger)
	if err != nil {
		return nil, err
	}

	whereCond := map[string]interface{}{"WorkerPoolID": poolID}

	databaseEntity, err := q.Select().Where(whereCond).GetOne()
	if err != nil {
		return nil, err
	}

	return databaseEntity.(*model.WorkerPoolOccupancyEntity), nil
}

func (r *PersistentOccupancyRepository) UpdateWorkerPoolOccupancy(poolID string, runningWorkers int) error {

	dbOps := func(tx *db.TxConnection) error {

		rTx, err := r.WithTx(tx)
		if err != nil {
			return err
		}
		occupancyEntity, err := rTx.FindWorkerPoolOccupancyByID(poolID)
		if err != nil {
			return err
		}

		cvtdRunningWorkers := int64(runningWorkers)
		if occupancyEntity.RunningWorkers == cvtdRunningWorkers {
			r.Logger.Warnf("Same number of running workers is already persisted for occupancy entity with poolID '%s'", poolID)
			return nil
		}

		if cvtdRunningWorkers > occupancyEntity.WorkerPoolCapacity {
			return fmt.Errorf("invalid number of running workers, should be less that worker pool capacity: "+
				"(running: %d, capacity:%d)", runningWorkers, occupancyEntity.WorkerPoolCapacity)
		}
		occupancyEntity.RunningWorkers = int64(runningWorkers)
		updateOccupancyQ, err := db.NewQuery(tx, occupancyEntity, r.Logger)
		if err != nil {
			return err
		}

		whereCond := map[string]interface{}{"WorkerPoolID": poolID}
		if err = updateOccupancyQ.Update().Where(whereCond).Exec(); err != nil {
			r.Logger.Errorf("ReconRepo failed to update occupancy entity with poolID '%s': %s", occupancyEntity.WorkerPoolID, err)
			return err
		}

		r.Logger.Debugf("ReconRepo updated workersCnt of occupancy entity with poolID '%s' to '%d'", occupancyEntity.WorkerPoolID, runningWorkers)
		return err
	}
	return db.Transaction(r.Conn, dbOps, r.Logger)
}

func (r *PersistentOccupancyRepository) RemoveWorkerPoolOccupancy(poolID string) error {

	dbOps := func(tx *db.TxConnection) error {

		deleteOccupancyQ, err := db.NewQuery(tx, &model.WorkerPoolOccupancyEntity{}, r.Logger)
		if err != nil {
			return err
		}

		whereCond := map[string]interface{}{"WorkerPoolID": poolID}
		if err != nil {
			return err
		}

		deletionCnt, err := deleteOccupancyQ.Delete().Where(whereCond).Exec()
		if err != nil {
			r.Logger.Errorf("ReconRepo failed to delete occupancy entity with poolID '%s': %s", poolID, err)
			return err
		}

		r.Logger.Debugf("ReconRepo deleted '%d' occupancy entity with poolID '%s'", deletionCnt, poolID)
		return err
	}
	return db.Transaction(r.Conn, dbOps, r.Logger)
}
