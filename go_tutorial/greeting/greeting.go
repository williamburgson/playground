package greeting

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

func Hello(name string) (string, error) {
	if name == "" {
		return "", errors.New("Username cannot be empty")
	}
	// := declares and inits a var, the long way round:
	// var msg string
	// msg = fmt.Sprintf("Hello, %v", name)
	msg := fmt.Sprintf(randGreeting(), name)
	return msg, nil
}

func Hellos(names []string) (map[string]string, error) {
	msgs := make(map[string]string)
	for _, name := range names {
		msg, err := Hello(name)
		if err != nil {
			return nil, err
		}
		msgs[name] = msg
	}
	return msgs, nil
}

func randGreeting() string {
	msgs := []string{
		"Hello, %v",
		"Hi, %v",
		"Welcome, %v",
	}

	return msgs[rand.Intn(len(msgs))]
}

// init func is reserved func, it will be executed upon import
// it is also guaranteed to execute before the main func
func init() {
	rand.Seed(time.Now().UnixNano())
}
