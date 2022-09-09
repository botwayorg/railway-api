package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/botwayorg/railway-api/cmd"
	"github.com/botwayorg/railway-api/constants"
	"github.com/botwayorg/railway-api/entity"
	"github.com/botwayorg/railway-api/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "railway",
	SilenceUsage:  true,
	SilenceErrors: true,
	Version:       constants.Version,
	Short:         "🚅 Railway. Infrastructure, Instantly.",
	Long:          "Interact with 🚅 Railway via CLI \n\n Deploy infrastructure, instantly. Docs: https://docs.railway.app",
}

func addRootCmd(cmd *cobra.Command) *cobra.Command {
	rootCmd.AddCommand(cmd)

	return cmd
}

// contextualize converts a HandlerFunction to a cobra function
func contextualize(fn entity.HandlerFunction, panicFn entity.PanicFunction) entity.CobraFunction {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		defer func() {
			// Skip recover during development, so we can see the panic stack traces instead of going
			// through the "send to Railway" flow and hiding the stack from the user
			if constants.IsDevVersion() {
				return
			}

			if r := recover(); r != nil {
				err := panicFn(ctx, fmt.Sprint(r), string(debug.Stack()), cmd.Name(), args)
				if err != nil {
					fmt.Println("Unable to relay panic to server. Are you connected to the internet?")
				}
			}
		}()

		req := &entity.CommandRequest{
			Cmd:  cmd,
			Args: args,
		}

		err := fn(ctx, req)

		if err != nil {
			fmt.Println(ui.AlertDanger(err.Error()))
			os.Exit(1) // Set non-success exit code on error
		}

		return nil
	}
}

func init() {
	// Initializes all commands
	handler := cmd.New()

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Print verbose output")

	loginCmd := addRootCmd(&cobra.Command{
		Use:   "login",
		Short: "Login to your Railway account",
		RunE:  contextualize(handler.Login, handler.Panic),
	})

	loginCmd.Flags().Bool("browserless", false, "--browserless")

	addRootCmd(&cobra.Command{
		Use:               "init",
		Short:             "Create a new Railway project",
		PersistentPreRunE: contextualize(handler.CheckVersion, handler.Panic),
		RunE:              contextualize(handler.Init, handler.Panic),
	})

	addRootCmd(&cobra.Command{
		Use:               "link",
		Short:             "Associate existing project with current directory, may specify projectId as an argument",
		PersistentPreRunE: contextualize(handler.CheckVersion, handler.Panic),
		RunE:              contextualize(handler.Link, handler.Panic),
	})

	addRootCmd(&cobra.Command{
		Use:        "env",
		RunE:       contextualize(handler.Variables, handler.Panic),
		Deprecated: "Please use 'railway variables' instead", /**/
	})

	variablesCmd := addRootCmd(&cobra.Command{
		Use:     "variables",
		Aliases: []string{"vars"},
		Short:   "Show variables for active environment",
		RunE:    contextualize(handler.Variables, handler.Panic),
	})

	variablesCmd.Flags().StringP("service", "s", "", "Fetch variables accessible to a specific service")

	variablesGetCmd := &cobra.Command{
		Use:     "get key",
		Short:   "Get the value of a variable",
		RunE:    contextualize(handler.VariablesGet, handler.Panic),
		Args:    cobra.MinimumNArgs(1),
		Example: "  railway variables get MY_KEY",
	}

	variablesCmd.AddCommand(variablesGetCmd)
	variablesGetCmd.Flags().StringP("service", "s", "", "Fetch variables accessible to a specific service")

	variablesSetCmd := &cobra.Command{
		Use:     "set key=value",
		Short:   "Create or update the value of a variable",
		RunE:    contextualize(handler.VariablesSet, handler.Panic),
		Args:    cobra.MinimumNArgs(1),
		Example: "  railway variables set NODE_ENV=prod NODE_VERSION=12",
	}

	variablesCmd.AddCommand(variablesSetCmd)
	variablesSetCmd.Flags().StringP("service", "s", "", "Fetch variables accessible to a specific service")
	variablesSetCmd.Flags().Bool("skip-redeploy", false, "Skip redeploying the specified service after changing the variables")
	variablesSetCmd.Flags().Bool("replace", false, "Fully replace all previous variables instead of updating them")
	variablesSetCmd.Flags().Bool("yes", false, "Skip all confirmation dialogs")

	variablesDeleteCmd := &cobra.Command{
		Use:     "delete key",
		Short:   "Delete a variable",
		RunE:    contextualize(handler.VariablesDelete, handler.Panic),
		Example: "  railway variables delete MY_KEY",
	}

	variablesCmd.AddCommand(variablesDeleteCmd)
	variablesDeleteCmd.Flags().StringP("service", "s", "", "Fetch variables accessible to a specific service")
	variablesDeleteCmd.Flags().Bool("skip-redeploy", false, "Skip redeploying the specified service after changing the variables")

	addRootCmd(&cobra.Command{
		Use:   "environment",
		Short: "Change the active environment",
		RunE:  contextualize(handler.Environment, handler.Panic),
	})

	runCmd := addRootCmd(&cobra.Command{
		Use:                "run",
		Short:              "Run a local command using variables from the active environment",
		PersistentPreRunE:  contextualize(handler.CheckVersion, handler.Panic),
		RunE:               contextualize(handler.Run, handler.Panic),
		DisableFlagParsing: true,
	})

	runCmd.Flags().Bool("ephemeral", false, "Run the local command in an ephemeral environment")
	runCmd.Flags().String("service", "", "Run the command using variables from the specified service")

	addRootCmd(&cobra.Command{
		Use:               "version",
		Short:             "Get the version of the Railway CLI",
		PersistentPreRunE: contextualize(handler.CheckVersion, handler.Panic),
		RunE:              contextualize(handler.Version, handler.Panic),
	})

	upCmd := addRootCmd(&cobra.Command{
		Use:   "up",
		Short: "Upload and deploy project from the current directory",
		RunE:  contextualize(handler.Up, handler.Panic),
	})

	upCmd.Flags().BoolP("detach", "d", false, "Detach from cloud build/deploy logs")
	upCmd.Flags().StringP("environment", "e", "", "Specify an environment to up onto")
	upCmd.Flags().StringP("service", "s", "", "Fetch variables accessible to a specific service")

	downCmd := addRootCmd(&cobra.Command{
		Use:   "down",
		Short: "Remove the most recent deployment",
		RunE:  contextualize(handler.Down, handler.Panic),
	})

	downCmd.Flags().StringP("environment", "e", "", "Specify an environment to delete from")

	addRootCmd(&cobra.Command{
		Use:   "connect",
		Short: "Open an interactive shell to a database",
		RunE:  contextualize(handler.Connect, handler.Panic),
	})

	addRootCmd(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

	Bash:

	  $ source <(railway completion bash)

	  # To load completions for each session, execute once:
	  # Linux:
	  $ railway completion bash > /etc/bash_completion.d/railway
	  # macOS:
	  $ railway completion bash > /usr/local/etc/bash_completion.d/railway

	Zsh:

	  # If shell completion is not already enabled in your environment,
	  # you will need to enable it.  You can execute the following once:

	  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

	  # To load completions for each session, execute once:
	  $ railway completion zsh > "${fpath[1]}/_railway"

	  # You will need to start a new shell for this setup to take effect.

	fish:

	  $ railway completion fish | source

	  # To load completions for each session, execute once:
	  $ railway completion fish > ~/.config/fish/completions/railway.fish

	PowerShell:

	  PS> railway completion powershell | Out-String | Invoke-Expression

	  # To load completions for every new session, run:
	  PS> railway completion powershell > railway.ps1
	  # and source this file from your PowerShell profile.
	`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE:                  contextualize(handler.Completion, handler.Panic),
	})
}

func main() {
	if _, err := os.Stat("/proc/version"); !os.IsNotExist(err) && runtime.GOOS == "windows" {
		fmt.Printf("%s : Running in Non standard shell!\n Please consider using something like WSL!\n", ui.YellowText(ui.Bold("[WARNING!]").String()).String())
	}

	if err := rootCmd.Execute(); err != nil {
		if strings.Contains(err.Error(), "unknown command") {
			suggStr := "\nS"

			suggestions := rootCmd.SuggestionsFor(os.Args[1])

			if len(suggestions) > 0 {
				suggStr = fmt.Sprintf(" Did you mean \"%s\"?\nIf not, s", suggestions[0])
			}

			fmt.Println(fmt.Sprintf("Unknown command \"%s\" for \"%s\".%s"+
				"ee \"railway --help\" for available commands.",
				os.Args[1], rootCmd.CommandPath(), suggStr))
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}
