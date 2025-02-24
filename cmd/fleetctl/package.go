package main

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fleetdm/fleet/v4/orbit/pkg/packaging"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/skratchdot/open-golang/open"
	"github.com/urfave/cli/v2"
)

var (
	opt               packaging.Options
	disableOpenFolder bool
)

func packageCommand() *cli.Command {
	return &cli.Command{
		Name:        "package",
		Aliases:     nil,
		Usage:       "Create an Orbit installer package",
		Description: "An easy way to create fully boot-strapped installer packages for Windows, macOS, or Linux",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "type",
				Usage:    "Type of package to build",
				Required: true,
			},
			&cli.StringFlag{
				Name:        "enroll-secret",
				Usage:       "Enroll secret for authenticating to Fleet server",
				Destination: &opt.EnrollSecret,
			},
			&cli.StringFlag{
				Name:        "fleet-url",
				Usage:       "URL (host:port) of Fleet server",
				Destination: &opt.FleetURL,
			},
			&cli.StringFlag{
				Name:        "fleet-certificate",
				Usage:       "Path to server certificate chain",
				Destination: &opt.FleetCertificate,
			},
			&cli.StringFlag{
				Name:        "identifier",
				Usage:       "Identifier for package product",
				Value:       "com.fleetdm.orbit",
				Destination: &opt.Identifier,
			},
			&cli.StringFlag{
				Name:        "version",
				Usage:       "Version for package product",
				Destination: &opt.Version,
			},
			&cli.BoolFlag{
				Name:        "insecure",
				Usage:       "Disable TLS certificate verification",
				Destination: &opt.Insecure,
			},
			&cli.BoolFlag{
				Name:        "service",
				Usage:       "Install orbit/osquery with a persistence service (launchd, systemd, etc.)",
				Value:       true,
				Destination: &opt.StartService,
			},
			&cli.StringFlag{
				Name:        "sign-identity",
				Usage:       "Identity to use for macOS codesigning",
				Destination: &opt.SignIdentity,
			},
			&cli.BoolFlag{
				Name:        "notarize",
				Usage:       "Whether to notarize macOS packages",
				Destination: &opt.Notarize,
			},
			&cli.StringFlag{
				Name:        "osqueryd-channel",
				Usage:       "Update channel of osqueryd to use",
				Value:       "stable",
				Destination: &opt.OsquerydChannel,
			},
			&cli.StringFlag{
				Name:        "desktop-channel",
				Usage:       "Update channel of desktop to use",
				Value:       "stable",
				Destination: &opt.DesktopChannel,
			},
			&cli.StringFlag{
				Name:        "orbit-channel",
				Usage:       "Update channel of Orbit to use",
				Value:       "stable",
				Destination: &opt.OrbitChannel,
			},
			&cli.BoolFlag{
				Name:        "disable-updates",
				Usage:       "Disable auto updates on the generated package",
				Destination: &opt.DisableUpdates,
			},
			&cli.StringFlag{
				Name:        "update-url",
				Usage:       "URL for update server",
				Value:       "https://tuf.fleetctl.com",
				Destination: &opt.UpdateURL,
			},
			&cli.StringFlag{
				Name:        "update-roots",
				Usage:       "Root key JSON metadata for update server (from fleetctl updates roots)",
				Destination: &opt.UpdateRoots,
			},
			&cli.StringFlag{
				Name:        "osquery-flagfile",
				Usage:       "Flagfile to package and provide to osquery",
				Destination: &opt.OsqueryFlagfile,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Enable debug logging in orbit",
				Destination: &opt.Debug,
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Log detailed information when building the package",
			},
			&cli.BoolFlag{
				Name:        "fleet-desktop",
				Usage:       "Include the Fleet Desktop Application in the package",
				Destination: &opt.Desktop,
			},
			&cli.DurationFlag{
				Name:        "update-interval",
				Usage:       "Interval that Orbit will use to check for new updates (10s, 1h, etc.)",
				Value:       15 * time.Minute,
				Destination: &opt.OrbitUpdateInterval,
			},
			&cli.BoolFlag{
				Name:        "disable-open-folder",
				Usage:       "Disable opening the folder at the end",
				Destination: &disableOpenFolder,
			},
			&cli.BoolFlag{
				Name:        "native-tooling",
				Usage:       "Build the package using native tooling (only available in Linux)",
				EnvVars:     []string{"FLEETCTL_NATIVE_TOOLING"},
				Destination: &opt.NativeTooling,
			},
			&cli.StringFlag{
				Name:        "macos-devid-pem-content",
				Usage:       "Dev ID certificate keypair content in PEM format",
				EnvVars:     []string{"FLEETCTL_MACOS_DEVID_PEM_CONTENT"},
				Destination: &opt.MacOSDevIDCertificateContent,
			},
			&cli.StringFlag{
				Name:        "app-store-connect-api-key-id",
				Usage:       "App Store Connect API key used for notarization",
				EnvVars:     []string{"FLEETCTL_APP_STORE_CONNECT_API_KEY_ID"},
				Destination: &opt.AppStoreConnectAPIKeyID,
			},
			&cli.StringFlag{
				Name:        "app-store-connect-api-key-issuer",
				Usage:       "Issuer of the App Store Connect API key",
				EnvVars:     []string{"FLEETCTL_APP_STORE_CONNECT_API_KEY_ISSUER"},
				Destination: &opt.AppStoreConnectAPIKeyIssuer,
			},
			&cli.StringFlag{
				Name:        "app-store-connect-api-key-content",
				Usage:       "Contents of the .p8 App Store Connect API key",
				EnvVars:     []string{"FLEETCTL_APP_STORE_CONNECT_API_KEY_CONTENT"},
				Destination: &opt.AppStoreConnectAPIKeyContent,
			},
		},
		Action: func(c *cli.Context) error {
			if opt.FleetURL != "" || opt.EnrollSecret != "" {
				if opt.FleetURL == "" || opt.EnrollSecret == "" {
					return errors.New("--enroll-secret and --fleet-url must be provided together")
				}
			}

			if opt.Insecure && opt.FleetCertificate != "" {
				return errors.New("--insecure and --fleet-certificate may not be provided together")
			}

			if runtime.GOOS == "windows" && c.String("type") != "msi" {
				return errors.New("Windows can only build MSI packages.")
			}

			if opt.NativeTooling && runtime.GOOS != "linux" {
				return errors.New("native tooling is only available in Linux")
			}

			if opt.FleetCertificate != "" {
				err := checkPEMCertificate(opt.FleetCertificate)
				if err != nil {
					return fmt.Errorf("failed to read certificate %q: %w", opt.FleetCertificate, err)
				}
			}

			var buildFunc func(packaging.Options) (string, error)
			switch c.String("type") {
			case "pkg":
				buildFunc = packaging.BuildPkg
			case "deb":
				buildFunc = packaging.BuildDeb
			case "rpm":
				buildFunc = packaging.BuildRPM
			case "msi":
				buildFunc = packaging.BuildMSI
			default:
				return errors.New("type must be one of ('pkg', 'deb', 'rpm', 'msi')")
			}

			// disable detailed logging unless verbose is set
			if !c.Bool("verbose") {
				zlog.Logger = zerolog.Nop()
			}

			const maxAttempts = 9 // see #5732
			var (
				attempts int
				path     string
				err      error
			)
			for attempts < maxAttempts {
				attempts++

				if attempts > 1 {
					fmt.Printf("Generating your osquery installer [attempt %d/%d]...\n\n", attempts, maxAttempts)
				} else {
					fmt.Println("Generating your osquery installer...")
				}
				path, err = buildFunc(opt)
				if err == nil || !shouldRetry(c.String("type"), opt, err) {
					break
				}
			}
			if err != nil {
				return err
			}

			path, _ = filepath.Abs(path)
			fmt.Printf(`
Success! You generated an osquery installer at %s

To add this device to Fleet, double-click to open your installer.

To add other devices to Fleet, distribute this installer using Chef, Ansible, Jamf, or Puppet. Learn how: https://fleetdm.com/docs/using-fleet/adding-hosts
`, path)
			if !disableOpenFolder {
				open.Start(filepath.Dir(path)) //nolint:errcheck
			}
			return nil
		},
	}
}

func shouldRetry(pkgType string, opt packaging.Options, err error) bool {
	if pkgType != "msi" || runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return false
	}

	// building an MSI on macos M1, check if the error is one that should be retried
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "package root files: heat failed"):
		return true
	case strings.Contains(errStr, "build package: candle failed"):
		return true
	case strings.Contains(errStr, "build package: light failed"):
		return true
	default:
		return false
	}
}

func checkPEMCertificate(path string) error {
	cert, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if p, _ := pem.Decode(cert); p == nil {
		return errors.New("invalid PEM file")
	}
	return nil
}
