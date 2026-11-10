package popup

import "sync/atomic"

// DefaultShortcut is the key the global shortcut starts with, and what the page shows until the real
// binding is known.
const DefaultShortcut = "Meta+Alt+G"

// Info is what the page needs to know about how the popup runs.
type Info struct {
	// Layer is true when the popup is a full-screen layer-shell surface: the page then has to treat a
	// click outside the card as "close".
	Layer bool `json:"layer"`
	// Shortcut is the current global shortcut, as text such as "Meta+Alt+G".
	Shortcut string `json:"shortcut"`
}

// Service is bound to the page by Wails: the page calls Hide on Escape or a click outside the card,
// and Info when it is shown.
type Service struct {
	popup    atomic.Pointer[Popup]
	shortcut func() string
}

// NewService returns a service that is connected to a popup by New. shortcut returns the current
// global shortcut as text; nil, or an empty result, means DefaultShortcut. (It is a constructor
// argument and not a method because every exported method of a service is bound to the page.)
func NewService(shortcut func() string) *Service { return &Service{shortcut: shortcut} }

func (s *Service) attach(p *Popup) { s.popup.Store(p) }

// Hide hides the popup.
func (s *Service) Hide() {
	if p := s.popup.Load(); p != nil {
		p.Hide()
	}
}

// Info returns how the popup runs.
func (s *Service) Info() Info {
	info := Info{Shortcut: DefaultShortcut}
	if s.shortcut != nil {
		if t := s.shortcut(); t != "" {
			info.Shortcut = t
		}
	}
	if p := s.popup.Load(); p != nil {
		info.Layer = p.LayerShell()
	}
	return info
}
