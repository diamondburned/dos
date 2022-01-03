package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/fivemoreminix/dos"
	"github.com/gdamore/tcell/v2"
)

func main() {
	// Handle SIGINT.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	main := newMainWidget(ctx)

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
		// Refresh and redraw the screen at 15fps.
		for range time.Tick(time.Second / 15) {
			funcCh <- func() {
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
	state *mainState
}

func newMainWidget(ctx context.Context) *mainWidget {
	return &mainWidget{
		Label: &dos.Label{
			Align: dos.AlignCenter,
		},
		state: newMainState(ctx),
	}
}

func (m *mainWidget) refresh() {
	m.state.RLock()
	defer m.state.RUnlock()

	m.Label.Text = fmt.Sprintf(
		"We're at %d now.\nThe time is %s.",
		m.state.counter, m.state.time.Format("Jan 2 15:04:05"),
	)
}

type mainState struct {
	sync.RWMutex
	counter int
	time    time.Time
}

// newMainState creates a new active main state.
func newMainState(ctx context.Context) *mainState {
	s := &mainState{}

	go func() {
		ticker := time.NewTicker(time.Second / 2)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				s.Lock()
				s.counter++
				s.time = now
				s.Unlock()
			}
		}
	}()

	return s
}
