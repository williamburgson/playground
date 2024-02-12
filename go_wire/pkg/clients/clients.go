package clients

// Example of nested client initialization
type FirstClient struct {
	Name string
}

func NewFirstClient() FirstClient {
	return FirstClient{Name: "first client"}
}

type SecondClient struct {
	Name  string
	First FirstClient
}

func NewSecondClient(first FirstClient) SecondClient {
	return SecondClient{Name: "second client", First: first}
}

type ThirdClient struct {
	Name   string
	Second SecondClient
}

func NewThirdClient(second SecondClient) ThirdClient {
	return ThirdClient{Name: "third client", Second: second}
}

type FourthClient struct {
	Name  string
	Third ThirdClient
}

func NewFourthClient(third ThirdClient) FourthClient {
	return FourthClient{Name: "fourth client", Third: third}
}

type MainClient struct {
	Name   string
	Fourth FourthClient
}

func NewMainClient(four FourthClient) MainClient {
	return MainClient{Name: "main client", Fourth: four}
}

// Example of initialization with interfaces
type IClientWithManyFields interface {
	Fields() string
}

type ClientWithManyFields struct {
	Str    string
	Num    int
	Double float64
}

func NewClientWithManyFields(a string, n int, d float64) IClientWithManyFields {
	return &ClientWithManyFields{Str: a, Num: n, Double: d}
}

func (c *ClientWithManyFields) Fields() string {
	return "Str, Num, Double"
}

type IClientWithManyClients interface {
	GetMain() MainClient
}

type ClientWithManyClients struct {
	CF   IClientWithManyFields
	Main MainClient
}

func NewClientWithManyClients(c1 IClientWithManyFields, main MainClient) IClientWithManyClients {
	return &ClientWithManyClients{CF: c1, Main: main}
}

func (c *ClientWithManyClients) GetMain() MainClient {
	return c.Main
}
