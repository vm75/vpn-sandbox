package webserver

import (
	"encoding/json"
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
	previous := statusUpdates
	statusUpdates = newStatusNotifier()
	t.Cleanup(func() {
		statusUpdates = previous
	})

	first, unsubscribeFirst := statusUpdates.subscribe()
	defer unsubscribeFirst()
	second, unsubscribeSecond := statusUpdates.subscribe()
	defer unsubscribeSecond()

	statusEventListener{}.HandleEvent(utils.Event{Name: "proxy-up"})

	for index, updates := range []<-chan string{first, second} {
		select {
		case event := <-updates:
			if event != "proxy-up" {
				t.Fatalf("subscriber %d event = %q, want proxy-up", index, event)
			}
		default:
			t.Fatalf("subscriber %d was not notified", index)
		}
	}
}

func TestForceRefreshHandlerRefreshesIpInfo(t *testing.T) {
	previous := ipInfo
	previousUpdates := statusUpdates
	ipInfo = &ipInfoCache{
		lookup: func(output IpInfo) error {
			output["ip"] = "198.51.100.1"
			return nil
		},
	}
	t.Cleanup(func() {
		ipInfo = previous
		statusUpdates = previousUpdates
	})
	statusUpdates = newStatusNotifier()

	response := httptest.NewRecorder()
	forceRefreshHandler(response, httptest.NewRequest(http.MethodGet, "/api/force-refresh", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
	var responseStatus ipInfoStatus
	if err := json.Unmarshal(response.Body.Bytes(), &responseStatus); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if responseStatus.Stale {
		t.Fatal("refresh response should contain fresh IP info")
	}
	if got := responseStatus.Output["ip"]; got != "198.51.100.1" {
		t.Fatalf("refresh response IP info = %v, want refreshed output", got)
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
