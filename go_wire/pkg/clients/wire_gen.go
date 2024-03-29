// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package clients

// Injectors from wire.go:

func InitializeMainClient(first FirstClient) (MainClient, error) {
	secondClient := NewSecondClient(first)
	thirdClient := NewThirdClient(secondClient)
	fourthClient := NewFourthClient(thirdClient)
	mainClient := NewMainClient(fourthClient)
	return mainClient, nil
}

func InitializeClientWithManyClients(s1 string, n1 int, f1 float64, first FirstClient) (IClientWithManyClients, error) {
	iClientWithManyFields := NewClientWithManyFields(s1, n1, f1)
	secondClient := NewSecondClient(first)
	thirdClient := NewThirdClient(secondClient)
	fourthClient := NewFourthClient(thirdClient)
	mainClient := NewMainClient(fourthClient)
	iClientWithManyClients := NewClientWithManyClients(iClientWithManyFields, mainClient)
	return iClientWithManyClients, nil
}
