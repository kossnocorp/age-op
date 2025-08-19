package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type model struct {
	spinner spinner.Model
	done    bool
	timeout time.Duration
}

type finishedMsg struct{}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Tick(m.timeout, func(t time.Time) tea.Msg {
			return finishedMsg{}
		}),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle Ctrl+C within Bubbletea
		if msg.String() == "ctrl+c" {
			cleanup()
			return m, tea.Quit
		}
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

			// Set up signal handling for SIGTERM (Ctrl+C is handled by Bubble Tea)
			sigTermChan := make(chan os.Signal, 1)
			signal.Notify(sigTermChan, syscall.SIGTERM)
			defer signal.Stop(sigTermChan)

			p := tea.NewProgram(model{
				spinner: s,
				timeout: time.Duration(timeout) * time.Second,
			})

			errChan := make(chan error, 1)
			go func() {
				_, err := p.Run()
				errChan <- err
			}()

			// Wait for completion or SIGTERM
			select {
			case sig := <-sigTermChan:
				if verbose {
					fmt.Printf("\nReceived signal: %v\n", sig)
				}
				cleanup()

				// Wait for Bubble Tea
				p.Send(tea.Quit())
				<-errChan

				os.Exit(0)

			case err := <-errChan:
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
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

func cleanup() {
	fmt.Println("TODO: Clean up")
}
