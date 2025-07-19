package worker

import (
	"log"
	"os"
	"time"
)

type Task struct {
	F      func() error
	Repeat *int
	Dt     time.Duration
}

type Worker struct {
	Tasks []Task
}

func runTask(t Task) {
	l := log.New(os.Stderr, "TASK: ", log.Default().Flags())
	for {
		err := t.F()
		if err != nil {
			l.Println("Task error: ", err)
		}

		if t.Repeat != nil {
			cnntr := *t.Repeat - 1
			t.Repeat = &cnntr
		}

		if t.Repeat != nil && *t.Repeat < 0 {
			break
		}

		time.Sleep(t.Dt)
	}
}

func (w *Worker) Start() {
	for _, t := range w.Tasks {
		go runTask(t)
	}
}
