package worker

import (
	"log"
	"os"
	"time"
)

type Task struct {
	F      func() error
	Dt     time.Duration
	Repeat *int
}

type Worker struct {
	Tasks []Task
}

func (w *Worker) Start() {
	l := log.New(os.Stderr, "TASK: ", log.Default().Flags())
	for _, t := range w.Tasks {
		go func() {
			for {
				err := t.F()
				if err != nil {
					l.Println("Task error: ", err)
				}
				time.Sleep(t.Dt)

				if t.Repeat != nil {
					cnntr := *t.Repeat - 1
					t.Repeat = &cnntr
				}

				if t.Repeat != nil && *t.Repeat < 0 {
					break
				}
			}
		}()
	}
}
