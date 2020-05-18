package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/stuartleeks/devcontainer-cli/internal/pkg/devcontainers"
)

func main() {
	var listIncludeContainerNames bool
	var listVerbose bool

	rootCmd := &cobra.Command{Use: "devcontainer"}

	cmdList := &cobra.Command{
		Use:   "list",
		Short: "List devcontainers",
		Long:  "Lists devcontainers that are currently running",
		Run: func(cmd *cobra.Command, args []string) {
			runListCommand(cmd, args, listIncludeContainerNames, listVerbose)
		},
	}
	cmdList.Flags().BoolVar(&listIncludeContainerNames, "include-container-names", false, "Also include container names in the list")
	cmdList.Flags().BoolVarP(&listVerbose, "verbose", "v", false, "Verbose output")
	rootCmd.AddCommand(cmdList)

	cmdExec := &cobra.Command{
		Use:   "exec DEVCONTAINER_NAME COMMAND",
		Short: "Execute a command in a devcontainer",
		Long:  "Execute a command in a devcontainer, similar to `docker exec`.",
		Run: func(cmd *cobra.Command, args []string) {
			runExecCommand(cmd, args)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// only completing the first arg  (devcontainer name)
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			devcontainers, err := devcontainers.ListDevcontainers()
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
			names := []string{}
			for _, devcontainer := range devcontainers {
				names = append(names, devcontainer.DevcontainerName)
				names = append(names, devcontainer.ContainerName)
			}
			sort.Strings(names)
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	rootCmd.AddCommand(cmdExec)

	cmdCompletion := &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: `To load completion run
	
	. <(devcontainer completion)
	
	To configure your bash shell to load completions for each session add to your bashrc
	
	# ~/.bashrc or ~/.profile
	. <(devcontainer completion)
	`,
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}
	rootCmd.AddCommand(cmdCompletion)

	rootCmd.Execute()
}

func runListCommand(cmd *cobra.Command, args []string, listIncludeContainerNames bool, listVerbose bool) {
	if listIncludeContainerNames && listVerbose {
		fmt.Println("Can't use both verbose and include-container-names")
		os.Exit(1)
	}
	devcontainers, err := devcontainers.ListDevcontainers()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
	if listVerbose {
		sort.Slice(devcontainers, func(i, j int) bool { return devcontainers[i].DevcontainerName < devcontainers[j].DevcontainerName })

		w := new(tabwriter.Writer)
		// minwidth, tabwidth, padding, padchar, flags
		w.Init(os.Stdout, 8, 8, 0, '\t', 0)
		defer w.Flush()

		fmt.Fprintf(w, "%s\t%s\n", "DEVCONTAINER NAME", "CONTAINER NAME")
		fmt.Fprintf(w, "%s\t%s\n", "-----------------", "--------------")

		for _, devcontainer := range devcontainers {
			fmt.Fprintf(w, "%s\t%s\n", devcontainer.DevcontainerName, devcontainer.ContainerName)
		}
		return
	}
	names := []string{}
	for _, devcontainer := range devcontainers {
		names = append(names, devcontainer.DevcontainerName)
		if listIncludeContainerNames {
			names = append(names, devcontainer.ContainerName)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
}

func runExecCommand(cmd *cobra.Command, args []string) {
	// TODO argument validation!

	fmt.Printf("TODO - exec: %v\n", args)

	devcontainerName := args[0]
	devcontainers, err := devcontainers.ListDevcontainers()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	containerID := ""
	for _, devcontainer := range devcontainers {
		if devcontainer.ContainerName == devcontainerName || devcontainer.DevcontainerName == devcontainerName {
			containerID = devcontainer.ContainerID
			break
		}
	}
	if containerID == "" {
		cmd.Usage()
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
		os.Exit(1)
	}

	dockerArgs := []string{"exec", "-it", containerID}
	dockerArgs = append(dockerArgs, args[1:]...)

	dockerCmd := exec.Command("docker", dockerArgs...)
	dockerCmd.Stdin = os.Stdin
	dockerCmd.Stdout = os.Stdout

	err = dockerCmd.Start()
	if err != nil {
		fmt.Printf("Exec: start error: %s\n", err)
	}
	err = dockerCmd.Wait()
	if err != nil {
		fmt.Printf("Exec: wait error: %s\n", err)
	}
}
