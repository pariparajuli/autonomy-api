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
const ScoreTaskListName = "autonomy-score-tasks"
const NudgeTaskListName = "autonomy-nudge-tasks"

// TriggerAccountUpdate is a helper function to send a signal to
// trigger the workflow to update scores.
func TriggerAccountUpdate(client cadence.CadenceClient, c context.Context, accountNumbers []string) error {
	for _, a := range accountNumbers {
		if _, err := client.SignalWithStartWorkflow(c,
			fmt.Sprintf("account-state-%s", a), "accountCheckSignal", nil,
			cadenceClient.StartWorkflowOptions{
				ID:                           fmt.Sprintf("account-state-%s", a),
				TaskList:                     ScoreTaskListName,
				ExecutionStartToCloseTimeout: time.Hour,
				WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
			}, "AccountStateUpdateWorkflow", a); err != nil {
			return err
		}
	}
	return nil
}

// TriggerAccountSymptomFollowUpNudge is a helper function to send a signal to
// trigger symptoms report following up.
func TriggerAccountSymptomFollowUpNudge(client cadence.CadenceClient, c context.Context, accountNumber string) error {
	_, err := client.StartWorkflow(c,
		cadenceClient.StartWorkflowOptions{
			ID:                           fmt.Sprintf("account-nudge-symptom-follow-up-%s", accountNumber),
			TaskList:                     NudgeTaskListName,
			ExecutionStartToCloseTimeout: 24 * time.Hour,
			WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
		}, "SymptomFollowUpNudgeWorkflow", accountNumber)

	return err
}

// TriggerAccountHighRiskFollowUpNudge is a helper function to send a signal to
// trigger account following up behaviors on self risk.
func TriggerAccountHighRiskFollowUpNudge(client cadence.CadenceClient, c context.Context, accountNumber string) error {
	_, err := client.StartWorkflow(c,
		cadenceClient.StartWorkflowOptions{
			ID:                           fmt.Sprintf("account-nudge-behavior-follow-up-on-risk-%s", accountNumber),
			TaskList:                     NudgeTaskListName,
			ExecutionStartToCloseTimeout: 24 * time.Hour,
			WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
		}, "AccountSelfReportedHighRiskFollowUpWorkflow", accountNumber)

	return err
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
				TaskList:                     ScoreTaskListName,
				ExecutionStartToCloseTimeout: time.Hour,
				WorkflowIDReusePolicy:        cadenceClient.WorkflowIDReusePolicyAllowDuplicate,
			}, "POIStateUpdateWorkflow", poiID); err != nil {
			return err
		}
	}
	return nil
}
