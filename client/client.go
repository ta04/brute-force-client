package client

import (
	"github.com/micro/go-micro"
	"github.com/micro/go-plugins/registry/consul"
	proto "github.com/ta04/auth-service/model/proto"
)

// NewAuthSC creates a new auth service client
func NewAuthSC() proto.AuthServiceClient {
	registry := consul.NewRegistry()

	s := micro.NewService(
		micro.Name("com.ta04.cli.auth"),
		micro.Registry(registry),
	)
	s.Init()

	authServiceClient := proto.NewAuthServiceClient("com.ta04.srv.auth", s.Client())
	return authServiceClient
}
