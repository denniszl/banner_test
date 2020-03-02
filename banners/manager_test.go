package banners

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	Repository
	banners []Banner
}

func (m mockRepo) SaveBanner(_ Banner) error {
	return nil
}

func (m mockRepo) GetBanners() ([]Banner, error) {
	return m.banners, nil
}

type mockBanner struct {
	Banner
	expired      bool
	start        string
	expiration   string
	displayed    bool
	displayError bool
}

func (mb *mockBanner) IsExpired() bool {
	return mb.expired
}

func (mb *mockBanner) GetExpiration() string {
	return mb.expiration
}

func (mb *mockBanner) GetStart() string {
	return mb.start
}

func (mb *mockBanner) Display() error {
	if mb.displayError {
		return fmt.Errorf("fake error")
	}
	mb.displayed = true
	return nil
}

// tests the time parsing logic for Banners works
func TestGetBanners(t *testing.T) {
	testCases := []struct {
		testName        string
		banners         []Banner
		errors          bool
		expectedBanners []Banner
		ctxValues       func(context.Context) context.Context
	}{
		{
			testName:        "no banners",
			banners:         []Banner{},
			errors:          false,
			expectedBanners: []Banner{},
		},
		{
			testName: "one banner that hasn't expired",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
			errors: false,
			expectedBanners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
		},
		{
			testName: "two banners that haven't expired",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
			errors: false,
			expectedBanners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
		},
		{
			testName: "a banner that isn't in its promotional period yet",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
			},
			errors:          false,
			expectedBanners: []Banner{},
		},
		{
			testName: "a banner that isn't in its promotional period yet, but with an internal IP.",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
			},
			errors: false,
			expectedBanners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
			},
			ctxValues: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, contextKeyIPAddress, "10.0.0.0")
			},
		},
		{
			testName: "an expired banner",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
			errors:          false,
			expectedBanners: []Banner{},
		},
		{
			testName: "one banner that hasn't expired, and one that has.",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
				&mockBanner{
					expiration: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
			errors: false,
			expectedBanners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
		},
		{
			testName: "an expired banner with an internal ip address",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
			ctxValues: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, contextKeyIPAddress, "10.0.0.0")
			},
			errors:          false,
			expectedBanners: []Banner{},
		},
		{
			testName: "an expired and a banner before its promotional period with an internal ip address",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
			},
			ctxValues: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, contextKeyIPAddress, "10.0.255.1")
			},
			errors: false,
			expectedBanners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
			},
		},
	}

	for _, testCase := range testCases {
		// wrap in closure to prevent too many file (descriptiors) being open
		t.Run(testCase.testName, func(t *testing.T) {
			manager := NewManager(mockRepo{
				banners: testCase.banners,
			})

			ctx := context.Background()
			if testCase.ctxValues != nil {
				ctx = testCase.ctxValues(ctx)
			}

			results, err := manager.GetValidBanners(ctx)
			require.Equal(t, testCase.errors, err != nil)
			require.ElementsMatch(t, testCase.expectedBanners, results)
		})
	}
}

func TestDisplayAppropriateBanner(t *testing.T) {
	testCases := []struct {
		testName        string
		banners         []Banner
		errors          bool
		expectedBanners []Banner
		ctxValues       func(context.Context) context.Context
		displayed       bool
	}{
		{
			testName:        "no banners",
			banners:         []Banner{},
			errors:          false,
			expectedBanners: []Banner{},
		},
		{
			testName: "one banner that hasn't expired",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
			},
			errors:    false,
			displayed: true,
		},
		{
			testName: "two banners that haven't expired",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				},
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
			},
			errors:    false,
			displayed: true,
		},
		{
			testName: "display error",
			banners: []Banner{
				&mockBanner{
					expiration:   time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					start:        time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					displayError: true,
				},
			},
			errors: true,
		},
		{
			testName: "an expired and a banner before its promotional period without an internal ip address",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
			},
			ctxValues: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, contextKeyIPAddress, "1.1.1.1")
			},
			errors:    false,
			displayed: false,
		},
		{
			testName: "an expired and a banner before its promotional period with an internal ip address",
			banners: []Banner{
				&mockBanner{
					expiration: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				},
				&mockBanner{
					expiration: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					start:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				},
			},
			ctxValues: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, contextKeyIPAddress, "10.0.255.1")
			},
			errors:    false,
			displayed: true,
		},
	}

	for _, testCase := range testCases {
		// wrap in closure to prevent too many file (descriptiors) being open
		t.Run(testCase.testName, func(t *testing.T) {
			manager := NewManager(mockRepo{
				banners: testCase.banners,
			})

			ctx := context.Background()
			if testCase.ctxValues != nil {
				ctx = testCase.ctxValues(ctx)
			}

			d, err := manager.DisplayAppropriateBanner(ctx)
			require.Equal(t, testCase.errors, err != nil)
			require.Equal(t, testCase.displayed, d)
			if testCase.displayed {
				displayedCount := 0
				validSeenThusFar := 0
				// assert there is only _one_ displayed, and that it is the earliest valid one.
				for _, banner := range testCase.banners {
					b := banner.(*mockBanner)
					if b.displayed {
						displayedCount++
						require.Equal(t, 0, validSeenThusFar)
					}

					if !b.IsExpired() && withinPeriod(ctx, b) {
						validSeenThusFar++
					}
				}
				require.Equal(t, 1, displayedCount)
			}
		})
	}
}
