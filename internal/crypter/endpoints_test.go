package crypter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.binggl.net/monorepo/internal/crypter"
)

type MockEncryptionSvc struct{}

func (m *MockEncryptionSvc) Encrypt(ctx context.Context, req crypter.Request) ([]byte, error) {
	payload := make([]byte, 1)
	return payload, nil
}

func Test_Encryption_Endpoints(t *testing.T) {
	payload := make([]byte, 1)
	endpoint := crypter.NewEndpoints(&MockEncryptionSvc{}, logger)
	assert.NotNil(t, endpoint)

	p, err := endpoint.Encrypt(context.TODO(), crypter.Request{})
	if err != nil {
		t.Fatalf("error encrypting: %v", err)
	}
	assert.Equal(t, payload, p)
}
