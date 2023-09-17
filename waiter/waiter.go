package waiter

import "sync"

type EngineWaiter struct {
	Program        *sync.WaitGroup
	HttpConnection *sync.WaitGroup
	Write          *sync.WaitGroup
}

type Waiter = string

var (
	Program        Waiter = "program"
	HttpConnection Waiter = "http-connection"
	Write          Waiter = "write"
)

func (w *EngineWaiter) Wait(target string) {
	switch target {
	case Program:
		w.Program.Wait()
	case HttpConnection:
		w.HttpConnection.Wait()
	case Write:
		w.Write.Wait()
	}
}

func (w *EngineWaiter) Add(target string, amount int) {
	switch target {
	case Program:
		w.Program.Add(amount)
	case HttpConnection:
		w.HttpConnection.Add(amount)
	case Write:
		w.Write.Add(amount)
	}
}

func (w *EngineWaiter) Done(target string) {
	switch target {
	case Program:
		w.Program.Done()
	case HttpConnection:
		w.HttpConnection.Done()
	case Write:
		w.Write.Done()
	}
}
