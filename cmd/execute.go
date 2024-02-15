package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// executeCmd represents the execute command
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute commands on specified hosts",
	Long: `This command executes specified commands on a list of hosts defined in a configuration file.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		viper.SetConfigName("config") // name of config file (without extension)
		viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
		viper.AddConfigPath(".")      // optionally look for config in the working directory
		err := viper.ReadInConfig()   // Find and read the config file
		if err != nil {               // Handle errors reading the config file
			fmt.Printf("Fatal error config file: %s \n", err)
			os.Exit(1)
		}

		hosts := viper.GetStringSlice("hosts")
		restrictedCommands := viper.GetStringSlice("restricted_commands")

		for _, cmd := range args {
			for _, restricted := range restrictedCommands {
				if strings.Contains(cmd, restricted) {
					fmt.Printf("Restricted command detected: %s\n", cmd)
					os.Exit(1)
				}
			}
		}

		var wg sync.WaitGroup
		for _, host := range hosts {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()

				fmt.Printf("Executing on host: %s\n", host)
				sshCommand := fmt.Sprintf("ssh %s %s", host, strings.Join(args, " "))
				cmd := exec.Command("bash", "-c", sshCommand)
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Error executing command on host %s: %v\n", host, err)
				}
				fmt.Println(string(output))
			}(host)
		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(executeCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// executeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// executeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

