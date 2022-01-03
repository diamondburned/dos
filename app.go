package dos

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type App struct {
	ClearRune       rune
	ClearStyle      tcell.Style // Style used when clearing the screen
	MainWidget      Widget
	CustomEventLoop func(app *App, ctx context.Context, s tcell.Screen)
	// CallbackCh is an optional channel for synchronizing callbacks into the
	// event loop. The user can use this to execute callbacks after running
	// goroutines.
	CallbackCh chan func()
	// OnResize is called when the screen is resized and before it's
	// synchronized.
	OnResize func(width, height int)
	// OnKeyEvent is called before the MainWidget's handler, and if this function
	// returns true, then the event is never passed onto the main widget.
	OnKeyEvent func(ev *tcell.EventKey) bool
	// OnMouseEvent is called before the MainWidget's handler, and if this function
	// returns true, then the event is never passed onto the main widget.
	OnMouseEvent func(ev *tcell.EventMouse) bool
}

func (app *App) Run(ctx context.Context, s tcell.Screen) {
	s.EnableMouse()
	s.EnablePaste()
	app.MainWidget.SetFocused(true)

	if app.CustomEventLoop != nil {
		app.CustomEventLoop(app, ctx, s)
	} else {
		DefaultEventLoop(app, ctx, s)
	}

	s.Fini()
}

// RunNewScreen is a convenient function around Run that automatically creates a
// new screen.
func (app *App) RunNewScreen(ctx context.Context) error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create tcell screen: %w", err)
	}
	if err = screen.Init(); err != nil {
		return fmt.Errorf("failed to initialize screen: %w", err)
	}

	app.Run(ctx, screen)
	return nil
}

func DefaultEventLoop(app *App, ctx context.Context, s tcell.Screen) {
	screenEvents := make(chan tcell.Event)

	go func() {
		for {
			ev := s.PollEvent()
			if ev == nil {
				return
			}

			select {
			case screenEvents <- ev:
				// ok
			case <-ctx.Done():
				return
			}
		}
	}()

	currentW, currentH := s.Size()
	rect := Rect{0, 0, currentW, currentH}

	for {
		select {
		case <-ctx.Done():
			return

		case fn := <-app.CallbackCh:
			fn()
			redraw(app, s, rect)

		case ev := <-screenEvents:
			redraw(app, s, rect)

			switch ev := ev.(type) {
			case *tcell.EventResize:
				rect.W, rect.H = ev.Size()
				if app.OnResize != nil {
					app.OnResize(rect.W, rect.H)
				}
				s.Sync() // Redraw the entire screen
			case *tcell.EventKey:
				if app.OnKeyEvent != nil {
					if app.OnKeyEvent(ev) {
						break
					}
				}
				_ = app.MainWidget.HandleKey(ev)
			case *tcell.EventMouse:
				if app.OnMouseEvent != nil {
					if app.OnMouseEvent(ev) {
						break
					}
				}
				_ = app.MainWidget.HandleMouse(rect, ev)
			}
		}
	}
}

func redraw(app *App, s tcell.Screen, rect Rect) {
	s.Clear()
	s.Fill(app.ClearRune, app.ClearStyle)
	app.MainWidget.Draw(rect, s)
	s.Show() // Renders all changed cells
}
