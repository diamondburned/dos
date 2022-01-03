package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/fivemoreminix/dos"
	"github.com/gdamore/tcell/v2"
)

func main() {
	// Handle SIGINT.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	main := newMainWidget()

	funcCh := make(chan func())

	app := dos.App{
		ClearStyle: tcell.Style{}.
			Background(tcell.ColorDefault).
			Foreground(tcell.ColorDefault),
		MainWidget: &dos.Center{Child: main},
		CallbackCh: funcCh,
		OnKeyEvent: func(ev *tcell.EventKey) bool {
			switch ev.Key() {
			case tcell.KeyEsc:
				cancel()
				return true
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					cancel()
					return true
				default:
					return false
				}
			default:
				return false
			}
		},
	}

	go func() {
		for range time.Tick(time.Second / 2) {
			funcCh <- func() {
				// One way to synchronize this.
				main.increase()
				main.refresh()
			}
		}
	}()

	if err := app.RunNewScreen(ctx); err != nil {
		log.Fatalln(err)
	}
}

type mainWidget struct {
	*dos.Label
	counter int
	time    time.Time
}

func newMainWidget() *mainWidget {
	return &mainWidget{
		Label: &dos.Label{
			Align: dos.AlignCenter,
		},
	}
}

func (m *mainWidget) increase() {
	m.counter++
	m.time = time.Now()
}

func (m *mainWidget) refresh() {
	m.Label.Text = fmt.Sprintf(
		"We're at %d now.\nThe time is %s.",
		m.counter, m.time.Format("Jan 2 15:04:05"),
	)
}
