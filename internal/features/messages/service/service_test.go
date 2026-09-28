package messages_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Rics69/rics-chat/internal/core/domain"
	core_errors "github.com/Rics69/rics-chat/internal/core/errors"
)

type fakeMessagesRepository struct {
	// сообщения диалога от новых к старым, как отдаёт БД
	history   []domain.Message
	lastLimit int
	createErr error
}

func (r *fakeMessagesRepository) CreateMessage(_ context.Context, message domain.Message) (domain.Message, error) {
	if r.createErr != nil {
		return domain.Message{}, r.createErr
	}

	message.ID = 100

	return message, nil
}

func (r *fakeMessagesRepository) ListMessages(
	_ context.Context,
	_ int64,
	_ int64,
	beforeID *int64,
	limit int,
) ([]domain.Message, error) {
	r.lastLimit = limit

	var result []domain.Message
	for _, message := range r.history {
		if beforeID != nil && message.ID >= *beforeID {
			continue
		}

		if len(result) == limit {
			break
		}

		result = append(result, message)
	}

	return result, nil
}

func (r *fakeMessagesRepository) ListDialogs(context.Context, int64, int) ([]domain.Dialog, error) {
	return nil, nil
}

type fakeNotifier struct {
	notified []domain.Message
}

func (n *fakeNotifier) NotifyNewMessage(_ context.Context, message domain.Message) {
	n.notified = append(n.notified, message)
}

func TestListMessagesPagination(t *testing.T) {
	repo := &fakeMessagesRepository{}
	for id := int64(5); id >= 1; id-- {
		repo.history = append(repo.history, domain.NewMessage(id, 1, 2, "hi", time.Now()))
	}

	service := NewMessagesService(repo, &fakeNotifier{})
	ctx := context.Background()

	page, next, err := service.ListMessages(ctx, 1, 2, nil, 2)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}

	if repo.lastLimit != 3 {
		t.Fatalf("repository limit = %d, want limit+1 = 3", repo.lastLimit)
	}

	assertIDs(t, page, 5, 4)
	if next == nil || *next != 4 {
		t.Fatalf("next_before_id = %v, want 4", next)
	}

	page, next, _ = service.ListMessages(ctx, 1, 2, next, 2)
	assertIDs(t, page, 3, 2)

	page, next, _ = service.ListMessages(ctx, 1, 2, next, 2)
	assertIDs(t, page, 1)
	if next != nil {
		t.Fatalf("last page next_before_id = %d, want nil", *next)
	}
}

func TestListMessagesLimit(t *testing.T) {
	repo := &fakeMessagesRepository{}
	service := NewMessagesService(repo, &fakeNotifier{})
	ctx := context.Background()

	_, _, _ = service.ListMessages(ctx, 1, 2, nil, 0)
	if repo.lastLimit != defaultMessagesLimit+1 {
		t.Fatalf("default: repository limit = %d, want %d", repo.lastLimit, defaultMessagesLimit+1)
	}

	_, _, _ = service.ListMessages(ctx, 1, 2, nil, 100500)
	if repo.lastLimit != maxMessagesLimit+1 {
		t.Fatalf("max: repository limit = %d, want %d", repo.lastLimit, maxMessagesLimit+1)
	}

	_, _, err := service.ListMessages(ctx, 1, 2, nil, -1)
	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("negative limit error = %v, want ErrInvalidArgument", err)
	}
}

func TestSendMessageNotifiesOnlyAfterSave(t *testing.T) {
	ctx := context.Background()

	notifier := &fakeNotifier{}
	service := NewMessagesService(&fakeMessagesRepository{}, notifier)

	message, err := service.SendMessage(ctx, 1, 2, "привет")
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}

	if len(notifier.notified) != 1 || notifier.notified[0].ID != message.ID {
		t.Fatalf("notified = %+v, want saved message id=%d", notifier.notified, message.ID)
	}

	failing := &fakeNotifier{}
	service = NewMessagesService(&fakeMessagesRepository{createErr: core_errors.ErrNotFound}, failing)
	if _, err := service.SendMessage(ctx, 1, 999, "привет"); !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}

	invalid := &fakeNotifier{}
	service = NewMessagesService(&fakeMessagesRepository{}, invalid)
	if _, err := service.SendMessage(ctx, 1, 1, "себе"); !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("error = %v, want ErrInvalidArgument", err)
	}

	if len(failing.notified)+len(invalid.notified) != 0 {
		t.Fatal("notifier must not be called when message was not saved")
	}
}

func assertIDs(t *testing.T, messages []domain.Message, want ...int64) {
	t.Helper()

	if len(messages) != len(want) {
		t.Fatalf("got %d messages, want %d", len(messages), len(want))
	}

	for i, message := range messages {
		if message.ID != want[i] {
			t.Fatalf("messages[%d].ID = %d, want %d", i, message.ID, want[i])
		}
	}
}
