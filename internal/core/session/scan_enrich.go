package session

import (
	"context"
	"sort"
)

// ScanAndEnrich is a convenience function that scans root for sessions,
// enriches each with metadata from workspace.yaml and events.jsonl, and
// returns them sorted by UpdatedAt descending (newest first).
//
// It respects context cancellation between enrichment calls.
// Non-fatal metadata errors are embedded in each Session.MetadataErr.
func ScanAndEnrich(ctx context.Context, root string) ([]Session, error) {
	scanner := NewScanner(root)
	sessions, err := scanner.Scan(ctx)
	if err != nil {
		return nil, err
	}

	for i := range sessions {
		if ctx.Err() != nil {
			return sessions[:i], ctx.Err()
		}
		EnrichMetadata(&sessions[i])
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}
