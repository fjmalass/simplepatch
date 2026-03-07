package format

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for help output
type Theme struct {
	Title    lipgloss.Style
	Command  lipgloss.Style
	Flag     lipgloss.Style
	Arg      lipgloss.Style
	Desc     lipgloss.Style
	Default  lipgloss.Style
	Required lipgloss.Style
	Muted    lipgloss.Style
}

func DefaultTheme() Theme {
	r := lipgloss.NewRenderer(os.Stdout)

	// Adaptive Palette: {Light Mode Hex, Dark Mode Hex}
	var (
		TitleColor    = lipgloss.AdaptiveColor{Light: "#8839ef", Dark: "#cba6f7"} // Deep Purple / Mauve
		CommandColor  = lipgloss.AdaptiveColor{Light: "#1e66f5", Dark: "#89b4fa"} // Blue / Sky
		FlagColor     = lipgloss.AdaptiveColor{Light: "#ea76cb", Dark: "#f5c2e7"} // Pink / Flamingo
		ArgColor      = lipgloss.AdaptiveColor{Light: "#40a02b", Dark: "#a6e3a1"} // Deep Green / Green
		DescColor     = lipgloss.AdaptiveColor{Light: "#4c4f69", Dark: "#bac2de"} // Dark Grey / Subtext
		MutedColor    = lipgloss.AdaptiveColor{Light: "#9ca0b0", Dark: "#585b70"} // Light Grey / Surface
		RequiredColor = lipgloss.AdaptiveColor{Light: "#fe640b", Dark: "#fab387"} // Orange / Peach
	)

	return Theme{
		Title:    r.NewStyle().Foreground(TitleColor).Bold(true),
		Command:  r.NewStyle().Foreground(CommandColor),
		Flag:     r.NewStyle().Foreground(FlagColor),
		Arg:      r.NewStyle().Foreground(ArgColor),
		Desc:     r.NewStyle().Foreground(DescColor),
		Default:  r.NewStyle().Foreground(MutedColor).Italic(true),
		Required: r.NewStyle().Foreground(RequiredColor).Bold(true),
		Muted:    r.NewStyle().Foreground(MutedColor),
	}
}

// Printer renders styled help output
type Printer struct {
	w     io.Writer
	theme Theme
}

// New creates a new Printer with the given theme
func New(w io.Writer, theme Theme) *Printer {
	return &Printer{w: w, theme: theme}
}

// KongOption returns a kong.Option that uses this printer
func KongOption(theme Theme) kong.Option {
	return kong.Help(func(options kong.HelpOptions, kongCtx *kong.Context) error {
		p := New(os.Stdout, theme)
		return p.Print(kongCtx)
	})
}

func (p *Printer) Print(kongCtx *kong.Context) error {
	var node *kong.Node
	if selected := kongCtx.Selected(); selected != nil {
		node = selected
	} else {
		node = kongCtx.Model.Node
	}

	p.header(node, kongCtx.Model)
	p.usage(node, kongCtx.Model)
	p.arguments(node)
	p.commands(node)
	p.flags(node)

	return nil
}

func (p *Printer) header(node *kong.Node, app *kong.Application) {
	name := node.Name
	if node.Parent != nil {
		name = app.Name + " " + node.Name
	}

	fmt.Fprintln(p.w, p.theme.Title.Render(name))
	if node.Help != "" {
		fmt.Fprintln(p.w, p.theme.Desc.Render(node.Help))
	}
	fmt.Fprintln(p.w)
}

func (p *Printer) usage(node *kong.Node, app *kong.Application) {
	fmt.Fprintln(p.w, p.theme.Title.Render("Usage:"))

	var usage strings.Builder
	usage.WriteString("  " + app.Name)
	if node.Parent != nil {
		usage.WriteString(" " + node.Name)
	}

	for _, pos := range node.Positional {
		arg := pos.Name
		if !pos.Required {
			arg = "[" + arg + "]"
		} else {
			arg = "<" + arg + ">"
		}
		usage.WriteString(" " + p.theme.Arg.Render(arg))
	}

	if len(node.Children) > 0 {
		usage.WriteString(" " + p.theme.Command.Render("<command>"))
	}
	if len(node.Flags) > 0 {
		usage.WriteString(" " + p.theme.Flag.Render("[flags]"))
	}

	fmt.Fprintln(p.w, usage.String())
	fmt.Fprintln(p.w)
}

func (p *Printer) arguments(node *kong.Node) {
	if len(node.Positional) == 0 {
		return
	}

	fmt.Fprintln(p.w, p.theme.Title.Render("Arguments:"))
	for _, arg := range node.Positional {
		name := p.theme.Arg.Render(arg.Name)
		if arg.Required {
			name += " " + p.theme.Required.Render("(required)")
		}
		fmt.Fprintf(p.w, "  %s\n", name)
		if arg.Help != "" {
			fmt.Fprintf(p.w, "      %s\n", p.theme.Desc.Render(arg.Help))
		}
	}
	fmt.Fprintln(p.w)
}

func (p *Printer) commands(node *kong.Node) {
	var cmds []*kong.Node
	for _, c := range node.Children {
		if !c.Hidden {
			cmds = append(cmds, c)
		}
	}
	if len(cmds) == 0 {
		return
	}

	fmt.Fprintln(p.w, p.theme.Title.Render("Commands:"))

	maxW := 0
	for _, c := range cmds {
		if len(c.Name) > maxW {
			maxW = len(c.Name)
		}
	}

	for _, c := range cmds {
		pad := strings.Repeat(" ", maxW-len(c.Name)+2)
		fmt.Fprintf(p.w, "  %s%s%s\n",
			p.theme.Command.Render(c.Name),
			pad,
			p.theme.Desc.Render(c.Help),
		)
	}

	fmt.Fprintln(p.w)
	fmt.Fprintf(p.w, "  %s\n\n", p.theme.Muted.Render("Run \"<command> --help\" for command help."))
}

func (p *Printer) flags(node *kong.Node) {
	var flags []*kong.Flag
	for n := node; n != nil; n = n.Parent {
		for _, f := range n.Flags {
			if !f.Hidden {
				flags = append(flags, f)
			}
		}
	}
	if len(flags) == 0 {
		return
	}

	fmt.Fprintln(p.w, p.theme.Title.Render("Flags:"))

	for _, f := range flags {
		var parts []string
		if f.Short != 0 {
			parts = append(parts, fmt.Sprintf("-%c", f.Short))
		}
		long := "--" + f.Name
		if !f.IsBool() {
			if f.PlaceHolder != "" {
				long += "=" + f.PlaceHolder
			} else {
				long += "=<value>"
			}
		}
		parts = append(parts, long)

		flagStr := p.theme.Flag.Render(strings.Join(parts, ", "))

		desc := f.Help
		if f.HasDefault && !f.IsBool() {
			desc += " " + p.theme.Default.Render(fmt.Sprintf("(default: %v)", f.Default))
		}
		if f.Required {
			desc += " " + p.theme.Required.Render("(required)")
		}
		if len(f.Envs) > 0 {
			desc += " " + p.theme.Muted.Render(fmt.Sprintf("[$%s]", f.Envs[0]))
		}

		fmt.Fprintf(p.w, "  %-30s  %s\n", flagStr, p.theme.Desc.Render(desc))
	}
	fmt.Fprintln(p.w)
}
