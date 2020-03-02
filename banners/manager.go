package banners

import (
	"context"
	"sort"
	"time"
)

// manager is the internal implementation of a banner manager.
type manager struct {
	repo Repository
}

// NewManager returns a banner manager
func NewManager(r Repository) Manager {
	return &manager{
		r,
	}
}

// GetValidBanners returns banners that are valid right now (e.g. not before or after the promotional period).
// If the banner has been expired before and is marked as such by the Banner interface, we will not display it, even if
// it is within the promotional period.
// It is up to the caller how they wish to consume the returned banners.
// Note: Arguably this doesn't have to be exposed in the interface.
func (m *manager) GetValidBanners(ctx context.Context) ([]Banner, error) {
	validBanners := []Banner{}

	banners, err := m.repo.GetBanners()
	if err != nil {
		return nil, err
	}

	for _, b := range banners {
		if !b.IsExpired() && withinPeriod(ctx, b) {
			validBanners = append(validBanners, b)
		}
	}

	return validBanners, nil
}

// AddBanner saves a banner to the repository. It returns an error if it fails to do so.
func (m *manager) AddBanner(b Banner) error {
	return m.repo.SaveBanner(b)
}

// DisplayAppropriateBanner looks for an appropriate banner to display.
// If there is no appropriate banner to display, it returns false, nil.
// If there is an error displaying a banner, it will return an error.
// Otherwise, this function returns true, nil.
func (m *manager) DisplayAppropriateBanner(ctx context.Context) (bool, error) {
	banners, err := m.GetValidBanners(ctx)
	if err != nil {
		return false, err
	}

	if len(banners) == 0 {
		return false, nil
	}

	// Get the earliest expiration and only display that one.
	sort.Slice(banners, func(i, j int) bool {
		// we know these won't error because GetValidBanners only returns these if they're valid timestamps.
		location := time.Now().Location()

		t1, _ := time.ParseInLocation(time.RFC3339, banners[i].GetExpiration(), location)
		t2, _ := time.ParseInLocation(time.RFC3339, banners[j].GetExpiration(), location)

		return t1.Before(t2)
	})

	err = banners[0].Display()
	if err != nil {
		return false, err
	}

	return true, nil
}
