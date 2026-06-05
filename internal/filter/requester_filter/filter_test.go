package requesterfilter

import (
	"context"
	"testing"
	"time"

	"github.com/jon4hz/jellysweep/internal/api/models"
	"github.com/jon4hz/jellysweep/internal/config"
	"github.com/jon4hz/jellysweep/internal/engine/arr"
	"github.com/jon4hz/jellysweep/pkg/jellyseerr"
)

type fakeRequester struct {
	requests map[int32]*jellyseerr.RequestInfo
}

func (f fakeRequester) GetRequestInfo(_ context.Context, tmdbID int32, _ string) (*jellyseerr.RequestInfo, error) {
	if requestInfo, ok := f.requests[tmdbID]; ok {
		return requestInfo, nil
	}
	return &jellyseerr.RequestInfo{}, nil
}

func TestApplyOnlyIncludesUnrequestedItems(t *testing.T) {
	requestTime := time.Now()
	cfg := &config.Config{
		Libraries: map[string]*config.CleanupConfig{
			"Movies": {
				Filter: config.FilterConfig{
					OnlyUnrequested: true,
				},
			},
		},
	}
	requester := fakeRequester{
		requests: map[int32]*jellyseerr.RequestInfo{
			1: {
				RequestTime: &requestTime,
				UserEmail:   "user@example.com",
			},
		},
	}
	items := []arr.MediaItem{
		{
			Title:       "Requested Movie",
			LibraryName: "Movies",
			TmdbId:      1,
			MediaType:   models.MediaTypeMovie,
		},
		{
			Title:       "Unrequested Movie",
			LibraryName: "Movies",
			TmdbId:      2,
			MediaType:   models.MediaTypeMovie,
		},
	}

	filtered, err := New(cfg, requester).Apply(context.Background(), items)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered item, got %d", len(filtered))
	}
	if filtered[0].Title != "Unrequested Movie" {
		t.Fatalf("expected unrequested movie to pass filter, got %q", filtered[0].Title)
	}
}

func TestApplyPassesThroughWhenOnlyUnrequestedDisabled(t *testing.T) {
	requestTime := time.Now()
	cfg := &config.Config{
		Libraries: map[string]*config.CleanupConfig{
			"Movies": {},
		},
	}
	requester := fakeRequester{
		requests: map[int32]*jellyseerr.RequestInfo{
			1: {
				RequestTime: &requestTime,
				UserEmail:   "user@example.com",
			},
		},
	}
	items := []arr.MediaItem{
		{
			Title:       "Requested Movie",
			LibraryName: "Movies",
			TmdbId:      1,
			MediaType:   models.MediaTypeMovie,
		},
	}

	filtered, err := New(cfg, requester).Apply(context.Background(), items)
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if len(filtered) != 1 {
		t.Fatalf("expected item to pass through when filter is disabled, got %d items", len(filtered))
	}
}

func TestApplyErrorsWhenOnlyUnrequestedEnabledWithoutRequester(t *testing.T) {
	cfg := &config.Config{
		Libraries: map[string]*config.CleanupConfig{
			"Movies": {
				Filter: config.FilterConfig{
					OnlyUnrequested: true,
				},
			},
		},
	}
	items := []arr.MediaItem{
		{
			Title:       "Movie",
			LibraryName: "Movies",
			TmdbId:      1,
			MediaType:   models.MediaTypeMovie,
		},
	}

	_, err := New(cfg, nil).Apply(context.Background(), items)
	if err == nil {
		t.Fatal("expected error when only_unrequested is enabled without Jellyseerr")
	}
}
