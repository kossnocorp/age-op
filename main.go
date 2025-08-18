package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type model struct {
	spinner spinner.Model
	done    bool
}

type finishedMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return finishedMsg{}
		}),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case finishedMsg:
		m.done = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if m.done {
		return "Hello, cruel world!\n"
	}
	return fmt.Sprintf("%s Loading...\n", m.spinner.View())
}

func main() {
	var verbose bool
	var timeout int
	var versionFlag bool

	rootCmd := &cobra.Command{
		Use:   "age-op",
		Short: "1Password CLI ❤️ age",
		Long:  "age-op is a CLI tool that integrates 1Password with age encryption.",
		Run: func(cmd *cobra.Command, args []string) {
			if versionFlag {
				version()
				return
			}

			if verbose {
				fmt.Println("Verbose mode enabled")
				fmt.Printf("Timeout: %d seconds\n", timeout)
			}

			s := spinner.New()
			s.Spinner = spinner.Dot

			p := tea.NewProgram(model{spinner: s})
			if _, err := p.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 2, "Timeout in seconds")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Print version information")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			version()
		},
	}

	var format string
	listCmd := &cobra.Command{
		Use:   "list [items]",
		Short: "List items",
		Long:  "List items in various formats",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Listing items in %s format:\n", format)
			for _, item := range args {
				fmt.Printf("  - %s\n", item)
			}
		},
	}
	listCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format (text, json, yaml)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func version() {
	fmt.Println("age-op version 0.1.0")
}
