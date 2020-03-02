package banners

import (
	"bytes"
	"context"
	"net"
	"time"
)

var (
	internalIPBegin = net.ParseIP("10.0.0.0")
	internalIPEnd   = net.ParseIP("10.255.255.255")
)

type contextKey string

func (c contextKey) String() string {
	return "banners " + string(c)
}

var contextKeyIPAddress = contextKey("ip-address")

func isInternalIP(ctx context.Context) bool {
	i, ok := ctx.Value(contextKeyIPAddress).(string)
	if !ok {
		return false
	}

	ip := net.ParseIP(i)
	if ip.To4() == nil {
		return false
	}
	return bytes.Compare(ip, internalIPBegin) >= 0 && bytes.Compare(ip, internalIPEnd) <= 0
}

func withinPeriod(ctx context.Context, b Banner) bool {
	location := time.Now().Location()

	exp, err := time.ParseInLocation(time.RFC3339, b.GetExpiration(), location)
	if err != nil {
		// timestamp is stored incorrectly, assume it's not within the period.
		return false
	}
	start, err := time.ParseInLocation(time.RFC3339, b.GetStart(), location)
	if err != nil {
		// timestamp is stored incorrectly, assume it's not within the period.
		return false
	}
	now := time.Now()

	// we know the timestamps from above are valid now. Now we'll return as long as it's before an expiration period and
	// IP is internal.
	if isInternalIP(ctx) && now.Before(exp) {
		return true
	}
	return now.After(start) && now.Before(exp)
}
