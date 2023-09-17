package waiter

import "sync"

type EngineWaiter struct {
	Program        *sync.WaitGroup
	HttpConnection *sync.WaitGroup
	Write          *sync.WaitGroup
}

func (w *EngineWaiter) Wait(target string) {
	switch target {
	case "program":
		w.Program.Wait()
	case "http":
		w.HttpConnection.Wait()
	case "write":
		w.Write.Wait()
	}
}

func (w *EngineWaiter) Add(target string, amount int) {
	switch target {
	case "program":
		w.Program.Add(amount)
	case "http":
		w.HttpConnection.Add(amount)
	case "write":
		w.Write.Add(amount)
	}
}

func (w *EngineWaiter) Done(target string) {
	switch target {
	case "program":
		w.Program.Done()
	case "http":
		w.HttpConnection.Done()
	case "write":
		w.Write.Done()
	}
}
