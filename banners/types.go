package banners

import "context"

// Repository defines an interface to how we'll access our storage
type Repository interface {
	SaveBanner(Banner) error
	// TODO improvement: potentially have a GetUnexpiredBanners that only gets banners that haven't been expired.
	GetBanners() ([]Banner, error)
}

// Banner is a banner interface (or what I think is one)
// time is assumed to be stored using RFC3339. Banners won't work if they're not stored using this RFC.
type Banner interface {
	Display() error
	GetExpiration() string
	GetStart() string
	IsExpired() bool
}

// Manager describes an interface for managing banners
type Manager interface {
	GetValidBanners(ctx context.Context) ([]Banner, error)
	AddBanner(Banner) error
	DisplayAppropriateBanner(ctx context.Context) (bool, error)
}
