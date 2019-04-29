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

const (
	forceUpdateFlag               = "force"
	httpsOnlyFlag                 = "https-only"
	httpsUsernameFlag             = "https-username"
	httpsPasswordFlag             = "https-password"
	sshIdentityFileFlag           = "ssh-identity-file"
	sshIdentityFilePassphraseFlag = "ssh-identity-file-passphrase"
	sshPortFlag                   = "ssh-port"
)

var (
	authProvider helper.AuthProvider
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Populate .proto vendors existing protodep.toml and lock",
	RunE: func(cmd *cobra.Command, args []string) error {
		var authProvider helper.AuthProvider

		isForceUpdate, err := cmd.Flags().GetBool(forceUpdateFlag)
		if err != nil {
			return err
		}
		logger.Info("force update = %t", isForceUpdate)

		httpsOnly, err := cmd.Flags().GetBool(httpsOnlyFlag)
		if err != nil {
			return err
		}
		logger.Info("https only = %s", httpsOnly)

		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		logger.Info("current dir is = %s", pwd)

		homeDir, err := homedir.Dir()
		if err != nil {
			return err
		}
		logger.Info("home dir is = %s", homeDir)

		identityFile, err := cmd.Flags().GetString(sshIdentityFileFlag)
		if err != nil {
			return err
		}

		passphrase, err := cmd.Flags().GetString(sshIdentityFilePassphraseFlag)
		if err != nil {
			return err
		}

		idFilePath := filepath.Join(homeDir, ".ssh", identityFile)
		if _, err := os.Stat(idFilePath); os.IsNotExist(err) {
			idFilePath = ""
		}

		sshPort, err := cmd.Flags().GetString(sshPortFlag)
		if err != nil {
			return err
		}

		httpsUsername, err := cmd.Flags().GetString(httpsUsernameFlag)
		if err != nil {
			return err
		}

		httpsPassword, err := cmd.Flags().GetString(httpsPasswordFlag)
		if err != nil {
			return err
		}

		if len(idFilePath) > 0 && !httpsOnly {
			authProvider = helper.NewSSHAuthProvider(idFilePath, passphrase, sshPort)
			logger.Info("identity file = %s", idFilePath)
			if passphrase != "" {
				logger.Info("passphrase = %s", strings.Repeat("x", len(passphrase))) // Do not display the password.
			}
			logger.Info("ssh port = %s", sshPort)
		} else {
			authProvider = helper.NewHTTPSAuthProvider(httpsUsername, httpsPassword)
			if len(httpsUsername) > 0 && len(httpsPassword) > 0 {
				logger.Info("https username = %s", httpsUsername)
				logger.Info("https password = %s", httpsPassword)
			}
		}

		updateService := service.NewSync(authProvider, homeDir, pwd, pwd)
		return updateService.Resolve(isForceUpdate)
	},
}

func initDepCmd() {
	upCmd.PersistentFlags().BoolP(forceUpdateFlag, "f", false, "update locked file and .proto vendors")
	upCmd.PersistentFlags().BoolP(httpsOnlyFlag, "o", false, "use https only for downloading dependencies")
	upCmd.PersistentFlags().StringP(httpsUsernameFlag, "u", "", "set username for https authentication")
	upCmd.PersistentFlags().StringP(httpsPasswordFlag, "s", "", "set password for https authentication")
	upCmd.PersistentFlags().StringP(sshIdentityFileFlag, "i", "id_rsa", "set name identity file for ssh-connection")
	upCmd.PersistentFlags().StringP(sshIdentityFilePassphraseFlag, "p", "", "set the passphrase for ssh-identity-file")
	upCmd.PersistentFlags().StringP(sshPortFlag, "P", "22", "set custom ssh-port")
}
