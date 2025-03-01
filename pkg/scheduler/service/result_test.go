package service

import (
	"testing"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/logger"
	"github.com/kyma-incubator/reconciler/pkg/model"
	"github.com/stretchr/testify/require"
)

type testCase struct {
	operations              []*model.OperationEntity
	expectedResultReconcile model.Status
	expectedResultDelete    model.Status
	expectedOrphans         []string //contains correlation IDs
}

func TestReconciliationResult(t *testing.T) {
	testCases := []*testCase{
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateNew,
					Updated:       time.Now().Add(-1999 * time.Millisecond),
				},
			},
			expectedResultReconcile: model.ClusterStatusReconciling,
			expectedResultDelete:    model.ClusterStatusDeleting,
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateNew,
					Updated:       time.Now().Add(-1999 * time.Millisecond),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateError,
					Updated:       time.Now().Add(-2000 * time.Millisecond),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.3",
					State:         model.OperationStateInProgress,
					Updated:       time.Now().Add(-2001 * time.Millisecond),
				},
			},
			expectedResultReconcile: model.ClusterStatusReconciling,
			expectedResultDelete:    model.ClusterStatusDeleting,
			expectedOrphans:         []string{"1.3"},
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateNew,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateError,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.3",
					State:         model.OperationStateOrphan,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.4",
					State:         model.OperationStateDone,
					Updated:       time.Now(),
				},
			},
			expectedResultReconcile: model.ClusterStatusReconcileError,
			expectedResultDelete:    model.ClusterStatusDeleteError,
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateFailed,
					Updated:       time.Now().Add(-3 * time.Second),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateNew,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.3",
					State:         model.OperationStateInProgress,
					Updated:       time.Now(),
				},
			},
			expectedResultReconcile: model.ClusterStatusReconciling,
			expectedResultDelete:    model.ClusterStatusDeleting,
			expectedOrphans:         []string{"1.1"},
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateDone,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateDone,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.3",
					State:         model.OperationStateInProgress,
					Updated:       time.Now(),
				},
			},
			expectedResultReconcile: model.ClusterStatusReconciling,
			expectedResultDelete:    model.ClusterStatusDeleting,
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateError,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "2.1",
					State:         model.OperationStateError,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateInProgress,
					Updated:       time.Now(),
				},
			},
			expectedResultReconcile: model.ClusterStatusReconciling,
			expectedResultDelete:    model.ClusterStatusDeleting,
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateError,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "2.1",
					State:         model.OperationStateError,
					Updated:       time.Now(),
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateNew,
					Updated:       time.Now(),
				},
			},
			expectedResultReconcile: model.ClusterStatusReconcileError,
			expectedResultDelete:    model.ClusterStatusDeleteError,
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateDone,
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateDone,
				},
			},
			expectedResultReconcile: model.ClusterStatusReady,
			expectedResultDelete:    model.ClusterStatusDeleted,
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateDone,
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateError,
				},
			},
			expectedResultReconcile: model.ClusterStatusReconcileError,
			expectedResultDelete:    model.ClusterStatusDeleteError,
		},
		{
			operations: []*model.OperationEntity{
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.1",
					State:         model.OperationStateError,
				},
				{
					Priority:      1,
					SchedulingID:  "schedulingID",
					CorrelationID: "1.2",
					State:         model.OperationStateError,
				},
			},
			expectedResultReconcile: model.ClusterStatusReconcileError,
			expectedResultDelete:    model.ClusterStatusDeleteError,
		},
	}

	//test reconcile result
	for _, testCase := range testCases {
		reconResult := newReconciliationResult(&model.ReconciliationEntity{
			RuntimeID:    "runtimeID",
			SchedulingID: "schedulingID",
		}, logger.NewLogger(true))

		require.NoError(t, reconResult.AddOperations(testCase.operations))
		require.Equal(t, testCase.expectedResultReconcile, reconResult.GetResult())
		require.ElementsMatch(t, testCase.operations, reconResult.GetOperations())

		//check detected orphans
		allDetectedOrphans := make(map[string]*model.OperationEntity)
		detectedOrphans := reconResult.GetOrphans(1 * time.Second)
		for _, detectedOrphan := range detectedOrphans {
			allDetectedOrphans[detectedOrphan.CorrelationID] = detectedOrphan
		}
		for _, correlationID := range testCase.expectedOrphans {
			_, ok := allDetectedOrphans[correlationID]
			require.True(t, ok)
		}
	}

	//test delete result
	for _, testCase := range testCases {
		reconResult := newReconciliationResult(&model.ReconciliationEntity{
			RuntimeID:    "runtimeID",
			SchedulingID: "schedulingID",
		}, logger.NewLogger(true))

		//mark all operations as delete op
		for _, op := range testCase.operations {
			op.Type = model.OperationTypeDelete
		}

		require.NoError(t, reconResult.AddOperations(testCase.operations))
		require.Equal(t, reconResult.GetResult(), testCase.expectedResultDelete)
	}
}
