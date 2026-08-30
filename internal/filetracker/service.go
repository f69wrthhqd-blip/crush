// Package filetracker provides functionality to track file reads in sessions.
package filetracker

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/crush/internal/db"
)

// Service defines the interface for tracking file reads in sessions.
type Service interface {
	// RecordRead records when a file was read.
	RecordRead(ctx context.Context, sessionID, path string)

	// LastReadTime returns when a file was last read.
	// Returns zero time if never read.
	LastReadTime(ctx context.Context, sessionID, path string) time.Time

	// ListReadFiles returns the paths of all files read in a session.
	ListReadFiles(ctx context.Context, sessionID string) ([]string, error)
}

type service struct {
	q *db.Queries
	// workingDir is the workspace directory relative paths are recorded
	// against. It must be injected rather than derived from the process
	// cwd so that multiple workspaces hosted by one process (client/server
	// mode) each attribute files to their own directory.
	workingDir string
}

// NewService creates a new file tracker service. The workingDir is the
// workspace directory used to relativize recorded paths; when empty the
// process working directory is used as a fallback.
func NewService(q *db.Queries, workingDir string) Service {
	return &service{q: q, workingDir: workingDir}
}

// RecordRead records when a file was read.
func (s *service) RecordRead(ctx context.Context, sessionID, path string) {
	if err := s.q.RecordFileRead(ctx, db.RecordFileReadParams{
		SessionID: sessionID,
		Path:      s.relpath(path),
	}); err != nil {
		slog.Error("Error recording file read", "error", err, "file", path)
	}
}

// LastReadTime returns when a file was last read.
// Returns zero time if never read.
func (s *service) LastReadTime(ctx context.Context, sessionID, path string) time.Time {
	readFile, err := s.q.GetFileRead(ctx, db.GetFileReadParams{
		SessionID: sessionID,
		Path:      s.relpath(path),
	})
	if err != nil {
		return time.Time{}
	}

	return time.Unix(readFile.ReadAt, 0)
}

// relpath cleans path and makes it relative to the service's working
// directory. Paths outside the working directory are stored as absolute.
func (s *service) relpath(path string) string {
	path = filepath.Clean(path)
	basepath := s.workingDir
	if basepath == "" {
		var err error
		basepath, err = os.Getwd()
		if err != nil {
			slog.Warn("Error getting basepath", "error", err)
			return path
		}
	}
	relpath, err := filepath.Rel(basepath, path)
	if err != nil {
		slog.Warn("Error getting relpath", "error", err)
		return path
	}
	return relpath
}

// ListReadFiles returns the paths of all files read in a session.
func (s *service) ListReadFiles(ctx context.Context, sessionID string) ([]string, error) {
	readFiles, err := s.q.ListSessionReadFiles(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("listing read files: %w", err)
	}

	basepath := s.workingDir
	if basepath == "" {
		basepath, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getting working directory: %w", err)
		}
	}

	paths := make([]string, 0, len(readFiles))
	for _, rf := range readFiles {
		// Skip paths that were recorded relative to a different working
		// directory: joining an absolute path onto a basepath would
		// produce a corrupt path.
		if filepath.IsAbs(rf.Path) {
			paths = append(paths, rf.Path)
			continue
		}
		paths = append(paths, filepath.Join(basepath, rf.Path))
	}
	return paths, nil
}
