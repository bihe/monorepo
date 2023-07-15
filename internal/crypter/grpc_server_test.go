package crypter_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"golang.binggl.net/monorepo/internal/crypter"
	"golang.binggl.net/monorepo/pkg/config"
	"golang.binggl.net/monorepo/pkg/logging"
	"golang.binggl.net/monorepo/proto"

	"google.golang.org/grpc"
)

var log = logging.NewNop()

const jwt_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE5OTIxNDIwNjYsImp0aSI6IjhjMmRlNzIzLTUwOTItNDQxOC05ZTcyLTZkNDlkYjZiMGI1ZSIsImlhdCI6MTYwMTUzNzI2NiwiaXNzIjoiaXNzdWVyIiwic3ViIjoiYS5iQGMuZGUiLCJUeXBlIjoibG9naW4uVXNlciIsIkRpc3BsYXlOYW1lIjoiVGVzdCBVc2VyIiwiRW1haWwiOiJhLmJAYy5kZSIsIlVzZXJJZCI6IjEyMzQ1IiwiVXNlck5hbWUiOiJVc2VybmFtZSIsIkdpdmVuTmFtZSI6IlRlc3QiLCJTdXJuYW1lIjoiVXNlciIsIkNsYWltcyI6WyJjcnlwdGVyfGh0dHA6Ly9sb2NhbGhvc3Q6MzAwNHxVc2VyIl19.BBfONdlk7-TVOvWcZX7YxeaiXt8hPFNxk_j1N1Pfyt0"

func Test_Grpc_Server_Encrypt(t *testing.T) {
	// testing grpc is difficult: this test is more of an integration test because
	// we spin up the whole server

	// SERVER
	var (
		service = crypter.NewService(log, config.Security{
			JwtIssuer: "issuer",
			JwtSecret: "secret",
			Claim: config.Claim{
				Name:  "crypter",
				URL:   "http://localhost:3004",
				Roles: []string{"User"},
			},
		})
		crypterServer = crypter.NewServer(service, logger)
		grpcServer    = grpc.NewServer()
	)
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("could not get free open port: %v", err)
	}

	grpcListener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		t.Fatalf("could not start grpc listener: %v", err)
	}
	proto.RegisterCrypterServer(grpcServer, crypterServer)

	go func() {
		if err := grpcServer.Serve(grpcListener); err != http.ErrServerClosed {
			return
		}
	}()

	// CLIENT
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := crypter.NewClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	payload, err := os.ReadFile("../../testdata/unencrypted.pdf")
	if err != nil {
		t.Errorf("could not read payload: %v", err)
	}
	res, err := client.Encrypt(ctx, crypter.Request{
		AuthToken: jwt_token,
		Payload:   payload,
		Type:      crypter.PDF,
		NewPass:   "test",
	})
	if err != nil {
		t.Errorf("could not encrypt given payload: %v", err)
	}
	os.WriteFile("./testresult.pdf", res, 0666)
	grpcServer.Stop()
}

// gist: https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
