package ghapi

import "github.com/go-playground/webhooks/github"

func GetFileChangesFromPushPayload(payload github.PushPayload) (fileChanges []FileChangeInfo) {
	for _, commit := range payload.Commits {
		for _, added := range commit.Added {
			fileChanges = append(fileChanges, FileChangeInfo{Filename: added, Additions: 1, Deletions: 0})
		}
		for _, removed := range commit.Removed {
			fileChanges = append(fileChanges, FileChangeInfo{Filename: removed, Additions: 0, Deletions: 1})
		}
		for _, modified := range commit.Modified {
			fileChanges = append(fileChanges, FileChangeInfo{Filename: modified, Additions: 0, Deletions: 0})
		}
	}
	return fileChanges
}
