package domain

import (
	"strings"
	"testing"
)

func TestMessageValidate(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		wantErr bool
	}{
		{name: "ok", message: NewMessageUnitialized(1, 2, "привет")},
		{name: "4096 runes", message: NewMessageUnitialized(1, 2, strings.Repeat("я", 4096))},
		{name: "4097 runes", message: NewMessageUnitialized(1, 2, strings.Repeat("я", 4097)), wantErr: true},
		{name: "empty", message: NewMessageUnitialized(1, 2, ""), wantErr: true},
		{name: "only spaces", message: NewMessageUnitialized(1, 2, " \n\t "), wantErr: true},
		{name: "to yourself", message: NewMessageUnitialized(1, 1, "hi"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.Validate()
			checkInvalidArgument(t, err, tt.wantErr)
		})
	}
}

func TestMessagePeerID(t *testing.T) {
	message := NewMessageUnitialized(1, 2, "hi")

	if got := message.PeerID(1); got != 2 {
		t.Fatalf("PeerID(sender) = %d, want 2", got)
	}

	if got := message.PeerID(2); got != 1 {
		t.Fatalf("PeerID(recipient) = %d, want 1", got)
	}
}
