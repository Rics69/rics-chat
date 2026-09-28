package core_ws

import (
	"sync"
	"testing"
	"time"

	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
)

func newTestHub(t *testing.T) *Hub {
	t.Helper()

	log, err := core_logger.NewLogger(core_logger.Config{Level: "ERROR", Folder: t.TempDir()})
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	t.Cleanup(log.Close)

	return NewHub(log)
}

// conn не нужен: hub работает только с userID и каналом send
func newTestClient(hub *Hub, userID int64) *Client {
	return newClient(hub, nil, userID, time.Now().Add(time.Hour), nil)
}

func TestHubSendsToAllUserTabs(t *testing.T) {
	hub := newTestHub(t)

	tab1 := newTestClient(hub, 1)
	tab2 := newTestClient(hub, 1)
	other := newTestClient(hub, 2)
	for _, c := range []*Client{tab1, tab2, other} {
		hub.register(c)
	}

	hub.SendToUser(1, []byte("hi"))

	for name, c := range map[string]*Client{"tab1": tab1, "tab2": tab2} {
		select {
		case msg := <-c.send:
			if string(msg) != "hi" {
				t.Fatalf("%s got %q", name, msg)
			}
		default:
			t.Fatalf("%s got nothing", name)
		}
	}

	if len(other.send) != 0 {
		t.Fatal("message leaked to another user")
	}
}

func TestHubDisconnectsSlowClient(t *testing.T) {
	hub := newTestHub(t)

	slow := newTestClient(hub, 1)
	hub.register(slow)

	for range sendBufferSize {
		hub.SendToUser(1, []byte("x"))
	}

	// буфер полон: следующее сообщение не должно заблокировать отправителя
	done := make(chan struct{})
	go func() {
		hub.SendToUser(1, []byte("overflow"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("SendToUser blocked on slow client")
	}

	for range sendBufferSize {
		<-slow.send
	}

	if _, ok := <-slow.send; ok {
		t.Fatal("slow client channel must be closed")
	}
}

func TestHubUnregisterAndCloseAreIdempotent(t *testing.T) {
	hub := newTestHub(t)

	c := newTestClient(hub, 1)
	hub.register(c)

	hub.unregister(c)
	hub.unregister(c)
	hub.Close()
	hub.Close()

	if hub.register(newTestClient(hub, 2)) {
		t.Fatal("register after Close must fail")
	}
}

func TestHubConcurrentAccess(t *testing.T) {
	hub := newTestHub(t)

	var wg sync.WaitGroup
	for userID := range int64(20) {
		wg.Add(1)
		go func() {
			defer wg.Done()

			c := newTestClient(hub, userID%5)
			hub.register(c)

			go func() {
				for range c.send {
				}
			}()

			for range 50 {
				hub.SendToUser((userID+1)%5, []byte("x"))
			}

			hub.unregister(c)
		}()
	}

	wg.Wait()
	hub.Close()
}
