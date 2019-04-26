package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/mittwald/protodep/helper"
	"github.com/mittwald/protodep/logger"
	"github.com/mittwald/protodep/service"
	"github.com/spf13/cobra"
)

var (
	authProvider helper.AuthProvider
)

type protoResource struct {
	source       string
	relativeDest string
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Populate .proto vendors existing protodep.toml and lock",
	RunE: func(cmd *cobra.Command, args []string) error {

		isForceUpdate, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}
		logger.Info("force update = %t", isForceUpdate)

		identityFile, err := cmd.Flags().GetString("ssh-identity-file")
		if err != nil {
			return err
		}
		logger.Info("identity file = %s", identityFile)

		password, err := cmd.Flags().GetString("ssh-identity-file-passphrase")
		if err != nil {
			return err
		}
		if password != "" {
			logger.Info("password = %s", strings.Repeat("x", len(password))) // Do not display the password.
		}

		pwd, err := os.Getwd()
		if err != nil {
			return err
		}

		homeDir, err := homedir.Dir()
		if err != nil {
			return err
		}

		idFilePath := filepath.Join(homeDir, ".ssh", identityFile)
		if _, err := os.Stat(idFilePath); os.IsNotExist(err) {
			idFilePath = ""
		}

		authProvider = helper.NewAuthProvider(idFilePath, password)
		updateService := service.NewSync(authProvider, homeDir, pwd, pwd)
		return updateService.Resolve(isForceUpdate)
	},
}

func initDepCmd() {
	upCmd.PersistentFlags().BoolP("force", "f", false, "update locked file and .proto vendors")
	upCmd.PersistentFlags().BoolP("https-only", "o", false, "use https only for downloading dependencies")
	upCmd.PersistentFlags().StringP("ssh-identity-file", "i", "id_rsa", "set name identity file for ssh-connection")
	upCmd.PersistentFlags().StringP("ssh-identity-file-passphrase", "p", "", "set the passphrase for ssh-identity-file")
}
