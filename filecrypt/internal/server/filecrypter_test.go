package server

import (
	"context"
	"testing"

	"github.com/bihe/monorepo/filecrypt/proto"
	log "github.com/sirupsen/logrus"
)

func TestFileCrypter_Expect_Missing_TokenError(t *testing.T) {
	crypter := &FileCrypter{
		Logger: log.New().WithField("mode", "test"),
	}

	_, err := crypter.Encrypt(context.TODO(), &proto.CrypterRequest{})
	if err == nil {
		t.Errorf("epected empty token error")
	}
}
