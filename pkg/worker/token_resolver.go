package worker

import (
	"context"
	"errors"
)

// ResolveInstallation fetches the installation record for the event's state_id.
func ResolveInstallation(ctx context.Context, evt *Event, client *InstallationsClient) (*InstallationRecord, error) {
	if evt == nil {
		return nil, errors.New("event is required")
	}
	if client == nil {
		return nil, errors.New("installations client is required")
	}
	stateID := ""
	if evt.Metadata != nil {
		stateID = evt.Metadata["state_id"]
	}
	if stateID == "" {
		return nil, errors.New("state_id missing from metadata")
	}
	return client.GetByStateID(ctx, evt.Provider, stateID)
}
