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
	httpsUsernameFlag             = "https-username"
	httpsPasswordFlag             = "https-password"
	sshIdentityFileFlag           = "ssh-identity-file"
	sshIdentityFilePassphraseFlag = "ssh-identity-file-passphrase"
	sshPortFlag                   = "ssh-port"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Populate .proto vendors existing protodep.toml and lock",
	RunE: func(cmd *cobra.Command, args []string) error {

		s, err := getData(cmd)
		if err != nil {
			return nil
		}

		authProviderSSH := helper.NewSSHAuthProvider(s.idFilePath, s.passphrase, s.sshPort)
		logger.Info("identity file = %s", s.idFilePath)
		if s.passphrase != "" {
			logger.Info("passphrase = %s", strings.Repeat("x", len(s.passphrase))) // Do not display the password.
		}
		logger.Info("ssh port = %s", s.sshPort)
		authProviderHTTPS := helper.NewHTTPSAuthProvider(s.httpsUsername, s.httpsPassword)
		if len(s.httpsUsername) > 0 && len(s.httpsPassword) > 0 {
			logger.Info("https username = %s", s.httpsUsername)
			logger.Info("https password = %s", strings.Repeat("x", len(s.httpsPassword)))
		}

		updateService := service.NewSync(authProviderSSH, authProviderHTTPS, s.homeDir, s.pwd, s.pwd)
		return updateService.Resolve(s.isForceUpdate)
	},
}

type stuff struct {
	isForceUpdate bool
	pwd           string
	homeDir       string
	identityFile  string
	passphrase    string
	idFilePath    string
	sshPort       string
	httpsUsername string
	httpsPassword string
}

func getData(cmd *cobra.Command) (*stuff, error) {

	stuff := &stuff{}

	var err error

	stuff.isForceUpdate, err = cmd.Flags().GetBool(forceUpdateFlag)
	if err != nil {
		return nil, err
	}
	logger.Info("force update = %t", stuff.isForceUpdate)

	stuff.pwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}
	logger.Info("current dir is = %s", stuff.pwd)

	stuff.homeDir, err = homedir.Dir()
	if err != nil {
		return nil, err
	}
	logger.Info("home dir is = %s", stuff.homeDir)

	stuff.identityFile, err = cmd.Flags().GetString(sshIdentityFileFlag)
	if err != nil {
		return nil, err
	}

	stuff.passphrase, err = cmd.Flags().GetString(sshIdentityFilePassphraseFlag)
	if err != nil {
		return nil, err
	}

	stuff.idFilePath = filepath.Join(stuff.homeDir, ".ssh", stuff.identityFile)
	if _, err := os.Stat(stuff.idFilePath); os.IsNotExist(err) {
		stuff.idFilePath = ""
	}

	stuff.sshPort, err = cmd.Flags().GetString(sshPortFlag)
	if err != nil {
		return nil, err
	}

	stuff.httpsUsername, err = cmd.Flags().GetString(httpsUsernameFlag)
	if err != nil {
		return nil, err
	}

	stuff.httpsPassword, err = cmd.Flags().GetString(httpsPasswordFlag)
	if err != nil {
		return nil, err
	}

	return stuff, nil
}

func initDepCmd() {
	upCmd.PersistentFlags().BoolP(forceUpdateFlag, "f", false, "update locked file and .proto vendors")
	upCmd.PersistentFlags().StringP(httpsUsernameFlag, "u", "", "set username for https authentication")
	upCmd.PersistentFlags().StringP(httpsPasswordFlag, "s", "", "set password for https authentication")
	upCmd.PersistentFlags().StringP(sshIdentityFileFlag, "i", "id_rsa", "set name identity file for ssh-connection")
	upCmd.PersistentFlags().StringP(sshIdentityFilePassphraseFlag, "p", "", "set the passphrase for ssh-identity-file")
	upCmd.PersistentFlags().StringP(sshPortFlag, "P", "22", "set custom ssh-port")
}
