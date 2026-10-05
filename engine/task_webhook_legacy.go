package main

import "encoding/json"

// Compatibility with the previous PR-specific checkpoint and journal structure.
// These fields are historical evidence, never provider-specific runtime policy.
type legacyTaskPR struct {
	RepositoryID     int64  `json:"repositoryId"`
	Repository       string `json:"repository"`
	Number           int    `json:"number"`
	URL              string `json:"url"`
	BaseRepositoryID int64  `json:"baseRepositoryId"`
	HeadRepositoryID int64  `json:"headRepositoryId"`
	BaseSHA          string `json:"baseSha"`
	HeadSHA          string `json:"headSha"`
	Draft            bool   `json:"draft"`
	DeliveryID       string `json:"deliveryId,omitempty"`
}

func taskRunInput(run TaskRun) TaskInput {
	if run.Input != nil {
		return *run.Input
	}
	body, _ := json.Marshal(run.LegacyPR)
	return TaskInput{ContentType: "application/json", Body: string(body)}
}

// Run only after journal replay, through the normal durable transaction gate.
// Retain old keys/fields so checkpoints preceding this transaction still replay.
func (a *app) migrateWebhookSettings() error {
	needed := false
	for _, b := range a.state.WebhookBindings {
		if b.Revision == 0 {
			needed = true
		}
	}
	if !needed {
		return nil
	}
	return a.commitLocked(func(d *diskState) {
		for i := range d.WebhookBindings {
			b := &d.WebhookBindings[i]
			if b.Revision != 0 {
				continue
			}
			b.Revision = 1
			b.SignatureHeader = "X-Hub-Signature-256"
			b.DeliveryHeader = "X-GitHub-Delivery"
		}
	})
}
