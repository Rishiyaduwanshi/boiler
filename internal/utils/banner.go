package utils

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/rishiyaduwanshi/boiler/pkg/version"
)

var (
	primaryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	secondaryStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42"))

	subtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	authorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		Italic(true)

	versionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)
)

func ShowBanner() {
	art := `
██████╗  ██████╗ ██╗██╗     ███████╗██████╗ 
██╔══██╗██╔═══██╗██║██║     ██╔════╝██╔══██╗
██████╔╝██║   ██║██║██║     █████╗  ██████╔╝
██╔══██╗██║   ██║██║██║     ██╔══╝  ██╔══██╗
██████╔╝╚██████╔╝██║███████╗███████╗██║  ██║
╚═════╝  ╚═════╝ ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝`

	fmt.Println(primaryStyle.Render(art))
	fmt.Println(subtitleStyle.Render("Code Snippet & Stack Manager"))
	fmt.Print(secondaryStyle.Render("Version: "))
	fmt.Println(versionStyle.Render("v" + version.Version))
	fmt.Println(authorStyle.Render("by Abhinav Prakash"))
}

func ShowQuickHelp() {
	fmt.Println("\nQuick Commands:")
	fmt.Println("  bl add <resource>        Add snippet or stack")
	fmt.Println("  bl store <path>          Store as snippet/stack")
	fmt.Println("  bl ls --all              List all resources")
	fmt.Println("  bl --help                Show full help")
	fmt.Println()
}
