package requesterfilter

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/jon4hz/jellysweep/internal/config"
	"github.com/jon4hz/jellysweep/internal/engine/arr"
	"github.com/jon4hz/jellysweep/internal/filter"
	"github.com/jon4hz/jellysweep/pkg/jellyseerr"
)

type requestInfoGetter interface {
	GetRequestInfo(ctx context.Context, tmdbID int32, mediaType string) (*jellyseerr.RequestInfo, error)
}

// Filter implements the filter.Filterer interface.
type Filter struct {
	cfg       *config.Config
	requester requestInfoGetter
}

var _ filter.Filterer = (*Filter)(nil)

// New creates a new requester Filter instance.
func New(cfg *config.Config, requester requestInfoGetter) *Filter {
	return &Filter{
		cfg:       cfg,
		requester: requester,
	}
}

// String returns the name of the filter.
func (f *Filter) String() string { return "Requester Filter" }

// Apply filters media items based on Jellyseerr requester information.
func (f *Filter) Apply(ctx context.Context, mediaItems []arr.MediaItem) ([]arr.MediaItem, error) {
	filteredItems := make([]arr.MediaItem, 0, len(mediaItems))

	for _, item := range mediaItems {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		libraryConfig := f.cfg.GetLibraryConfig(item.LibraryName)
		if libraryConfig == nil || !libraryConfig.Filter.OnlyUnrequested {
			filteredItems = append(filteredItems, item)
			continue
		}

		if f.requester == nil {
			return nil, fmt.Errorf("filter.only_unrequested is enabled for library %q but Jellyseerr is not configured", item.LibraryName)
		}

		requestInfo, err := f.requester.GetRequestInfo(ctx, item.TmdbId, string(item.MediaType))
		if err != nil {
			return nil, fmt.Errorf("failed to get request info for %q: %w", item.Title, err)
		}
		if requestInfo == nil || requestInfo.RequestTime == nil {
			filteredItems = append(filteredItems, item)
			log.Debug("including unrequested item", "title", item.Title, "library", item.LibraryName)
			continue
		}

		log.Debug("excluding requested item", "title", item.Title, "library", item.LibraryName, "requestedBy", requestInfo.UserEmail)
	}

	return filteredItems, nil
}
