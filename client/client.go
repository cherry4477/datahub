package client

import (
	"fmt"
	"github.com/asiainfoLDP/datahub/cmd"
	"os"
	"strings"
)

func RunClient() {

	if len(os.Args) < 2 {
		ShowUsage()
		os.Exit(2)
	}

	command := os.Args[1]

	commandFound := false
	for _, v := range cmd.Cmd {
		if strings.EqualFold(v.Name, command) {
			commandFound = true
			if len(os.Args) > 2 && len(os.Args[2]) > 0 && os.Args[2][0] != '-' {
				subCmdFound := false
				for _, vv := range v.SubCmd {
					if strings.EqualFold(vv.Name, os.Args[2]) {
						command += os.Args[2]
						subCmdFound = true
						vv.Handler(v.NeedLogin, os.Args[3:])
					}
				}
				if !subCmdFound {
					v.Handler(v.NeedLogin, os.Args[2:])
				}
			} else {
				v.Handler(v.NeedLogin, os.Args[2:])
			}
		}
	}

	if command == "help" {
		cmd.Help(false, os.Args[2:])
		commandFound = true
	}

	if command == "stop" {
		if err := cmd.StopP2P(); err != nil {
			fmt.Println(err)
		}
		commandFound = true
	}
	if !commandFound {
		fmt.Println(command, "not found")
		ShowUsage()
	}

	return

}

func ShowUsage() {
	fmt.Println("Usage:\tdatahub COMMAND [arg...]")
	fmt.Println("\tdatahub COMMAND [ --help ]")
	fmt.Println("\tdatahub help [COMMAND]\n")
	fmt.Println("Commands:")
	for _, v := range cmd.Cmd {
		fmt.Printf("    %-10s%s\n", v.Name, v.Desc)
	}
	fmt.Printf("\nRun '%s COMMAND --help' for details on a command.\n", os.Args[0])
}
