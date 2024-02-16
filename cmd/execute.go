package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
)

// executeCmd はコマンド実行を担当するcobraコマンドです。
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute commands on specified hosts",
	Long: `This command executes specified commands on a list of hosts defined in a configuration file.`,
	Run: execute,
}

func init() {
	rootCmd.AddCommand(executeCmd)
}

func execute(cmd *cobra.Command, args []string) {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	for _, host := range config.Hosts {
		executeOnHost(host, args)
	}
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // または他の設定ファイルのパス

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func executeOnHost(host HostConfig, commands []string) {
	// 公開鍵認証の設定
	key, err := ioutil.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read private key: %v\n", err)
		return
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse private key: %v\n", err)
		return
	}

	config := &ssh.ClientConfig{
		User: host.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 実際の使用ではより安全な方法を選択してください
	}

	// SSH接続
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.Name, host.Port), config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to dial: %v\n", err)
		return
	}
	defer client.Close()

	// セッションの作成
	session, err := client.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create session: %v\n", err)
		return
	}
	defer session.Close()

	// コマンドの実行
	cmd := strings.Join(commands, " ")
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute command: %v\n", err)
		return
	}

	fmt.Printf("Output from %s: %s\n", host.Name, output)
}
