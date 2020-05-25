package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/stuartleeks/devcontainer-cli/internal/pkg/devcontainers"
	ioutil2 "github.com/stuartleeks/devcontainer-cli/internal/pkg/ioutil"
)

func createTemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "template",
		Short: "work with templates",
		Long:  "Use subcommands to work with devcontainer templates",
	}
	cmd.AddCommand(createTemplateListCommand())
	cmd.AddCommand(createTemplateAddCommand())
	cmd.AddCommand(createTemplateAddLinkCommand())
	return cmd
}

func createTemplateListCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "list",
		Short: "list templates",
		Long:  "List devcontainer templates",
		Run: func(cmd *cobra.Command, args []string) {

			templates, err := devcontainers.GetTemplates()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, template := range templates {
				fmt.Println(template.Name)
			}
		},
	}
	return cmd
}

func createTemplateAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add TEMPLATE_NAME",
		Short: "add devcontainer from template",
		Long:  "Add a devcontainer definition to the current folder using the specified template",
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) != 1 {
				cmd.Usage()
				os.Exit(1)
			}
			name := args[0]

			template, err := devcontainers.GetTemplateByName(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if template == nil {
				fmt.Printf("Template '%s' not found\n", name)
			}

			info, err := os.Stat("./.devcontainer")
			if info != nil && err == nil {
				fmt.Println("Current folder already contains a .devcontainer folder - exiting")
				os.Exit(1)
			}

			currentDirectory, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error reading current directory: %s\n", err)
			}
			if err = ioutil2.CopyFolder(template.Path, currentDirectory+"/.devcontainer"); err != nil {
				fmt.Printf("Error copying folder: %s\n", err)
				os.Exit(1)
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// only completing the first arg  (template name)
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			templates, err := devcontainers.GetTemplates()
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
			names := []string{}
			for _, template := range templates {
				names = append(names, template.Name)
			}
			sort.Strings(names)
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}

func createTemplateAddLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-link TEMPLATE_NAME",
		Short: "add-link devcontainer from template",
		Long:  "Symlink a devcontainer definition to the current folder using the specified template",
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) != 1 {
				cmd.Usage()
				os.Exit(1)
			}
			name := args[0]

			template, err := devcontainers.GetTemplateByName(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if template == nil {
				fmt.Printf("Template '%s' not found\n", name)
			}

			info, err := os.Stat("./.devcontainer")
			if info != nil && err == nil {
				fmt.Println("Current folder already contains a .devcontainer folder - exiting")
				os.Exit(1)
			}

			currentDirectory, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error reading current directory: %s\n", err)
			}
			if err = ioutil2.LinkFolder(template.Path, currentDirectory+"/.devcontainer"); err != nil {
				fmt.Printf("Error linking folder: %s\n", err)
				os.Exit(1)
			}

			content := []byte("*\n")
			if err := ioutil.WriteFile(currentDirectory+"/.devcontainer/.gitignore", content, 0644); err != nil { // -rw-r--r--
				fmt.Printf("Error writing .gitignore: %s\n", err)
				os.Exit(1)
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// only completing the first arg  (template name)
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			templates, err := devcontainers.GetTemplates()
			if err != nil {
				fmt.Printf("Error: %v", err)
				os.Exit(1)
			}
			names := []string{}
			for _, template := range templates {
				names = append(names, template.Name)
			}
			sort.Strings(names)
			return names, cobra.ShellCompDirectiveNoFileComp
		},
	}
	return cmd
}