package utils

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	cadenceClient "go.uber.org/cadence/client"

	"github.com/bitmark-inc/autonomy-api/external/cadence"
)

// FIXME: there will be an import cycle if we use `github.com/bitmark-inc/autonomy-api/background/score`
const TaskListName = "autonomy-score-tasks"

// TriggerAccountUpdate is a helper function to send a signal to
// trigger the workflow to update scores.
func TriggerAccountUpdate(client cadence.CadenceClient, c context.Context, accountNumbers []string) error {
	for _, a := range accountNumbers {
		if _, err := client.SignalWithStartWorkflow(c,
			fmt.Sprintf("account-state-%s", a), "accountCheckSignal", nil,
			cadenceClient.StartWorkflowOptions{
				ID:                           fmt.Sprintf("account-state-%s", a),
				TaskList:                     TaskListName,
				ExecutionStartToCloseTimeout: time.Hour,
				WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
			}, "AccountStateUpdateWorkflow", a); err != nil {
			return err
		}
	}
	return nil
}

// TriggerPOIUpdate is a helper function to send a signal to
// trigger the workflow to update scores.
func TriggerPOIUpdate(client cadence.CadenceClient, c context.Context, poiIDs []primitive.ObjectID) error {
	for _, id := range poiIDs {
		poiID := id.Hex()
		if _, err := client.SignalWithStartWorkflow(c,
			fmt.Sprintf("poi-state-%s", poiID), "poiCheckSignal", nil,
			cadenceClient.StartWorkflowOptions{
				ID:                           fmt.Sprintf("poi-state-%s", poiID),
				TaskList:                     TaskListName,
				ExecutionStartToCloseTimeout: time.Hour,
				WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
			}, "POIStateUpdateWorkflow", poiID); err != nil {
			return err
		}
	}
	return nil
}
