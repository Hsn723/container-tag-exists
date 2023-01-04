package cmd

import (
	"net/http"
	"time"

	"github.com/Hsn723/container-tag-exists/pkg"
	"github.com/cybozu-go/log"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "container-tag-exists IMAGE TAG",
		Short: "check for the existence of a container tag",
		Long:  "check for the existence of a container tag against repositories using the Registry API v2",
		Args:  cobra.ExactArgs(2),
		RunE:  runRoot,
	}

	platforms []string
)

func init() {
	_ = rootCmd.LocalFlags().MarkHidden("logfile")
	_ = rootCmd.LocalFlags().MarkHidden("loglevel")
	_ = rootCmd.LocalFlags().MarkHidden("logformat")
	rootCmd.Flags().StringSliceVarP(&platforms, "platform", "p", nil, "specify platforms in the format os/arch to look for in container images. Default behavior is to look for any platform.")
}

func runRoot(cmd *cobra.Command, args []string) error {
	registryURL, err := pkg.ExtractRegistryURL(args[0])
	if err != nil {
		return err
	}
	registryName := pkg.NormalizeRegistryName(registryURL)
	imagePath, err := pkg.ExtractImagePath(args[0])
	if err != nil {
		return err
	}
	registryClient := &pkg.RegistryClient{
		RegistryName: registryName,
		RegistryURL:  registryURL,
		ImagePath:    imagePath,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
		Platforms: platforms,
	}
	hasTag, err := registryClient.IsTagExist(args[1])
	if err != nil {
		return err
	}
	if hasTag {
		cmd.Println("found")
	}
	return nil
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.ErrorExit(err)
	}
}
