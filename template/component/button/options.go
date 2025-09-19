package button

import common "github.com/zackarysantana/goview/template/component"

type ButtonProps struct {
	Size  ButtonSize
	Class string

	Name     string
	Disabled bool
	// Type can be "button", "submit", or "reset". Defaults to "button".
	// It cannot be used when Href is set.
	Type string
	// Href makes the button render as an anchor tag.
	Href string
}

func (p *ButtonProps) size() ButtonSize {
	if p.Size == "" {
		return Md
	}
	return p.Size
}

type ButtonVariant string

const (
	primary ButtonVariant = "primary"
	plain   ButtonVariant = "plain"
)

type ButtonSize string

const (
	Md   ButtonSize = "md"
	Sm   ButtonSize = "sm"
	Lg   ButtonSize = "lg"
	Icon ButtonSize = "icon"
)

func Class(p ButtonProps, variant ButtonVariant) string {
	base := "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors " +
		"focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50"

	v := map[ButtonVariant]string{
		primary: "bg-primary text-primary-foreground hover:bg-primary/90",
		plain:   "bg-transparent text-foreground/70 hover:bg-foreground/10 hover:text-foreground transition-opacity",
	}[variant]

	size := map[ButtonSize]string{
		Md:   "h-9 px-4 py-2",
		Sm:   "h-8 px-3",
		Lg:   "h-10 px-6",
		Icon: "h-9 w-9 p-0",
	}[p.size()]

	return common.Cn(base, v, size, p.Class)
}

func Type(t string) string {
	if t == "" {
		return "button"
	}
	return t
}
