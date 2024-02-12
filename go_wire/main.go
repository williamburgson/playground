package main

import (
	"fmt"
	"playground/go-wire/pkg/clients"
)

func main() {
	c, err := clients.InitializeMainClient(clients.NewFirstClient())
	if err != nil {
		panic(err)
	}
	fmt.Println(c.Name)
	fmt.Println(c.Fourth.Name)
	fmt.Println(c.Fourth.Third.Name)
	fmt.Println(c.Fourth.Third.Second.Name)
	fmt.Println(c.Fourth.Third.Second.First.Name)

	cf, err := clients.InitializeClientWithManyClients("client with many clients", 0, 0.1, clients.NewFirstClient())
	if err != nil {
		panic(err)
	}
	fmt.Println(cf.GetMain())
}
