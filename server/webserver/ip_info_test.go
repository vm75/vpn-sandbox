package webserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"vpn-sandbox/utils"
)

func TestIpInfoCacheStaleness(t *testing.T) {
	cache := &ipInfoCache{}
	startupAt := time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC)

	cache.markEvent("startup", startupAt)
	status := cache.snapshot()
	if !status.Stale {
		t.Fatal("IP info should be stale before the startup lookup completes")
	}
	if status.ExecutedAt != nil {
		t.Fatalf("unexpected execution time before lookup: %v", status.ExecutedAt)
	}

	executedAt := startupAt.Add(time.Second)
	cache.store(IpInfo{"ip": "192.0.2.1"}, executedAt)
	status = cache.snapshot()
	if status.Stale {
		t.Fatal("IP info should be fresh after a lookup completes after startup")
	}
	if status.ExecutedAt == nil || !status.ExecutedAt.Equal(executedAt) {
		t.Fatalf("execution time = %v, want %v", status.ExecutedAt, executedAt)
	}

	cache.markEvent("vpn-up", executedAt.Add(time.Second))
	status = cache.snapshot()
	if !status.Stale {
		t.Fatal("IP info should become stale after a tunnel event")
	}
	if got := status.Output["ip"]; got != "192.0.2.1" {
		t.Fatalf("cached output = %v, want previous output", got)
	}
}

func TestIpInfoCacheSnapshotCopiesOutput(t *testing.T) {
	cache := &ipInfoCache{}
	cache.store(IpInfo{"ip": "192.0.2.1"}, time.Now())

	status := cache.snapshot()
	status.Output["ip"] = "198.51.100.1"

	if got := cache.snapshot().Output["ip"]; got != "192.0.2.1" {
		t.Fatalf("cached output changed through snapshot: %v", got)
	}
}

func TestStatusEventListenerNotifiesStatusStream(t *testing.T) {
	for len(nwChangedChannel) > 0 {
		<-nwChangedChannel
	}

	statusEventListener{}.HandleEvent(utils.Event{Name: "proxy-up"})

	select {
	case event := <-nwChangedChannel:
		if event != "proxy-up" {
			t.Fatalf("event = %q, want proxy-up", event)
		}
	default:
		t.Fatal("proxy event did not notify status stream")
	}
}

func TestForceRefreshHandlerRefreshesIpInfo(t *testing.T) {
	previous := ipInfo
	ipInfo = &ipInfoCache{
		lookup: func(output IpInfo) error {
			output["ip"] = "198.51.100.1"
			return nil
		},
	}
	t.Cleanup(func() {
		ipInfo = previous
	})

	for len(nwChangedChannel) > 0 {
		<-nwChangedChannel
	}

	response := httptest.NewRecorder()
	forceRefreshHandler(response, httptest.NewRequest(http.MethodGet, "/api/force-refresh", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}

	status := ipInfo.snapshot()
	if status.Stale {
		t.Fatal("IP info should be fresh after a successful force refresh")
	}
	if status.Event != "force" {
		t.Fatalf("event = %q, want force", status.Event)
	}
	if got := status.Output["ip"]; got != "198.51.100.1" {
		t.Fatalf("IP info = %v, want refreshed output", got)
	}
}
