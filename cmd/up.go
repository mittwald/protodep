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

		bd, err := getBaseData(cmd)
		if err != nil {
			return nil
		}

		authProviderSSH := helper.NewSSHAuthProvider(bd.idFilePath, bd.passphrase, bd.sshPort)
		logger.Info("identity file = %s", bd.idFilePath)
		if bd.passphrase != "" {
			logger.Info("passphrase = %s", strings.Repeat("x", len(bd.passphrase))) // Do not display the password.
		}
		logger.Info("ssh port = %s", bd.sshPort)
		authProviderHTTPS := helper.NewHTTPSAuthProvider(bd.httpsUsername, bd.httpsPassword)
		if len(bd.httpsUsername) > 0 && len(bd.httpsPassword) > 0 {
			logger.Info("https username = %s", bd.httpsUsername)
			logger.Info("https password = %s", strings.Repeat("x", len(bd.httpsPassword)))
		}

		updateService := service.NewSync(authProviderSSH, authProviderHTTPS, bd.homeDir, bd.pwd, bd.pwd)
		return updateService.Resolve(bd.isForceUpdate)
	},
}

type baseData struct {
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

func getBaseData(cmd *cobra.Command) (*baseData, error) {

	baseData := &baseData{}

	var err error

	baseData.isForceUpdate, err = cmd.Flags().GetBool(forceUpdateFlag)
	if err != nil {
		return nil, err
	}
	logger.Info("force update = %t", baseData.isForceUpdate)

	baseData.pwd, err = os.Getwd()
	if err != nil {
		return nil, err
	}
	logger.Info("current dir is = %s", baseData.pwd)

	baseData.homeDir, err = homedir.Dir()
	if err != nil {
		return nil, err
	}
	logger.Info("home dir is = %s", baseData.homeDir)

	baseData.identityFile, err = cmd.Flags().GetString(sshIdentityFileFlag)
	if err != nil {
		return nil, err
	}

	baseData.passphrase, err = cmd.Flags().GetString(sshIdentityFilePassphraseFlag)
	if err != nil {
		return nil, err
	}

	baseData.idFilePath = filepath.Join(baseData.homeDir, ".ssh", baseData.identityFile)
	if _, err := os.Stat(baseData.idFilePath); os.IsNotExist(err) {
		baseData.idFilePath = ""
	}

	baseData.sshPort, err = cmd.Flags().GetString(sshPortFlag)
	if err != nil {
		return nil, err
	}

	baseData.httpsUsername, err = cmd.Flags().GetString(httpsUsernameFlag)
	if err != nil {
		return nil, err
	}

	baseData.httpsPassword, err = cmd.Flags().GetString(httpsPasswordFlag)
	if err != nil {
		return nil, err
	}

	return baseData, nil
}

func initDepCmd() {
	upCmd.PersistentFlags().BoolP(forceUpdateFlag, "f", false, "update locked file and .proto vendors")
	upCmd.PersistentFlags().StringP(httpsUsernameFlag, "u", "", "set username for https authentication")
	upCmd.PersistentFlags().StringP(httpsPasswordFlag, "s", "", "set password for https authentication")
	upCmd.PersistentFlags().StringP(sshIdentityFileFlag, "i", "id_rsa", "set name identity file for ssh-connection")
	upCmd.PersistentFlags().StringP(sshIdentityFilePassphraseFlag, "p", "", "set the passphrase for ssh-identity-file")
	upCmd.PersistentFlags().StringP(sshPortFlag, "P", "22", "set custom ssh-port")
}
