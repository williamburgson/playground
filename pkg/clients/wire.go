//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.
package clients

import "github.com/google/wire"

func InitializeMainClient(first FirstClient) (MainClient, error) {
	wire.Build(NewSecondClient, NewThirdClient, NewFourthClient, NewMainClient)
	return MainClient{}, nil
}

// params for the init function cannot be the same type: https://github.com/google/wire/issues/206
// e.g., if two clients takes strings to initialize, then string needs to aliased
func InitializeClientWithManyClients(s1 string, n1 int, f1 float64, first FirstClient) (IClientWithManyClients, error) {
	wire.Build(NewClientWithManyFields, NewSecondClient, NewThirdClient, NewFourthClient, NewMainClient, NewClientWithManyClients)
	return &ClientWithManyClients{}, nil
}
