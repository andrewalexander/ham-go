package ui

import (
	"log"
	"time"

	"github.com/tzneal/ham-go/cmd/termlog/input"

	termbox "github.com/nsf/termbox-go"
)

type Controller interface {
	AddWidget(w Widget)
	FocusNext() bool
	FocusPrevious() bool
	Theme() Theme
}
type FocusController struct {
	focusIdx  int
	focusable []Focusable
}
type MainController struct {
	widgets  []Widget
	commands map[input.Key]func()
	shutdown chan struct{}
	theme    Theme
	FocusController
}

func NewController(thm Theme) *MainController {
	err := termbox.Init()
	if err != nil {
		log.Fatalf("error initializing termbox: %s", err)
	}
	termbox.SetInputMode(termbox.InputAlt)
	termbox.SetOutputMode(termbox.Output256)

	c := &MainController{
		commands: make(map[input.Key]func()),
		shutdown: make(chan struct{}),
		theme:    thm,
	}
	return c
}

func (c *MainController) AddCommand(key input.Key, fn func()) {
	c.commands[key] = fn
}
func (c *MainController) AddWidget(w Widget) {
	c.widgets = append(c.widgets, w)
	if f, ok := w.(Focusable); ok {
		c.focusable = append(c.focusable, f)
	}
	w.SetController(c)
}
func (c *MainController) Redraw() {
	w, h := termbox.Size()
	termbox.Clear(termbox.ColorWhite, termbox.ColorDefault)
	_ = w
	_ = h
	for _, w := range c.widgets {
		w.Redraw()
	}
	termbox.Flush()
}

func (c *MainController) HandleEvent(key input.Key) bool {
	if fn, ok := c.commands[key]; ok {
		fn()
		return true
	}

	switch key {
	case input.KeyCtrlQ:
		return false
	default:
		if c.focusIdx < len(c.focusable) {
			c.focusable[c.focusIdx].HandleEvent(key)
		}
	}
	return true
}

func (c *MainController) RefreshEvery(duration time.Duration) {
	go func() {
		ticker := time.NewTicker(duration)
	lfor:
		for {
			select {
			case <-ticker.C:
				termbox.Interrupt()
			case <-c.shutdown:
				ticker.Stop()
				break lfor
			}
		}
	}()
}
func (c *MainController) Shutdown() {
	close(c.shutdown)
	defer termbox.Close()
}

func (c *MainController) Theme() Theme {
	return c.theme
}

func (c *FocusController) Focus(w Focusable) {
	if c.focusIdx < len(c.focusable) {
		c.focusable[c.focusIdx].Focus(false)
	}
	for i, f := range c.focusable {
		if f == w {
			c.focusIdx = i
			w.Focus(true)
		}
	}
}

func (p *FocusController) Unfocus() {
	for _, f := range p.focusable {
		f.Focus(false)
	}
}

func (c *FocusController) FocusNext() bool {
	// nothing to focus
	if len(c.focusable) == 0 {
		return true
	}

	// un-focus last widget
	if c.focusIdx < len(c.focusable) {
		c.focusable[c.focusIdx].Focus(false)
	}

	// focus next
	c.focusIdx++
	looped := false
	if c.focusIdx >= len(c.focusable) {
		c.focusIdx = 0
		looped = true
	}
	c.focusable[c.focusIdx].Focus(true)
	return looped
}

func (c *FocusController) FocusPrevious() bool {
	// nothing to focus
	if len(c.focusable) == 0 {
		return true
	}

	// un-focus last widget
	if c.focusIdx < len(c.focusable) {
		c.focusable[c.focusIdx].Focus(false)
	}

	c.focusIdx--
	looped := false
	if c.focusIdx < 0 {
		c.focusIdx = len(c.focusable) - 1
		looped = true
	}
	c.focusable[c.focusIdx].Focus(true)
	return looped
}
