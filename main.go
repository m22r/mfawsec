package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ini "gopkg.in/ini.v1"
)

// AppName is application name
const AppName = "mfawsec"

// Variables passed from go build -ldflag
var (
	commit  = "" //nolint
	version = "" //nolint
)

// Flags
type Flags struct {
	Profile      string
	Credential   string
	SerialNumber string
}

// AWSCredentials represents the set of attributes used to authenticate to AWS with a short lived session
type AWSCredentials struct {
	AWSAccessKey    string    `ini:"aws_access_key_id"`
	AWSSecretKey    string    `ini:"aws_secret_access_key"`
	AWSSessionToken string    `ini:"aws_session_token"`
	Expires         time.Time `ini:"aws_session_expiration"`
}

// Main functions
func main() {
	log.SetReportCaller(true)
	logger := createLogEntry()
	var f Flags
	var cmd = &cobra.Command{
		Use:   AppName,
		Short: "Set AWS temporary credentials with MFA token",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initFlags(&f); err != nil {
				logger.Fatal(err)
			}
			if err := runCmd(f); err != nil {
				logger.Fatal(err)
			}
			logger.Infof("%s completed", AppName)
		},
	}
	if err := parseFlags(cmd); err != nil {
		_ = cmd.Help()
		os.Exit(1)
	}
	cmd.AddCommand(versionCmd())
	_ = cmd.Execute()
}

// Create log entry
func createLogEntry() *log.Entry {
	return log.WithFields(log.Fields{
		"app":     AppName,
		"version": version,
		"commit":  commit,
	})
}

// Parse flags
func parseFlags(cmd *cobra.Command) error {
	viper.AutomaticEnv()
	cf := cmd.Flags()
	cf.String("profile", "", "profile name")
	cf.String("credential", os.Getenv("HOME")+"/.aws/credentials", "path of AWS credentials")
	cf.String("log-format", "text", "logging format (text or json)")
	cf.String("serial-number", "", "serial number of MFA Device (ex. arn:aws:iam::111111111111:mfa/foo)")
	if err := viper.BindPFlags(cf); err != nil {
		return err
	}
	for _, flag := range []string{"log-format", "serial-number"} {
		if err := viper.BindPFlag(strings.ReplaceAll(flag, "-", "_"), cf.Lookup(flag)); err != nil {
			return err
		}
	}
	return nil
}

// Initialize Flags
func initFlags(f *Flags) error {
	viper.AutomaticEnv()
	format := viper.GetString("log_format")
	switch format {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	default:
		return fmt.Errorf("log-format must be \"text\" or \"json\"")
	}
	f.Profile = viper.GetString("profile")
	f.Credential = viper.GetString("credential")
	f.SerialNumber = viper.GetString("serial_number")
	if f.Profile == "" || f.SerialNumber == "" {
		return fmt.Errorf("required flag(s) \"profile\" or \"serial-number\" is not set")
	}
	return nil
}

// Run command
func runCmd(f Flags) error {
	log := createLogEntry()
	cred, err := fetchSessionToken(f)
	if err != nil {
		return err
	}
	log.Info("succeeded fetchSessionToken()")

	if err := saveProfile(f, cred); err != nil {
		return err
	}
	log.Infof("succeeded save profile to %s", f.Credential)
	return nil
}

// Fetch session token
func fetchSessionToken(f Flags) (*sts.Credentials, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}))
	svc := sts.New(sess)
	res, err := svc.GetSessionToken(&sts.GetSessionTokenInput{
		SerialNumber: aws.String(f.SerialNumber),
		TokenCode:    aws.String(receiveUserInput("Enter MFA token: ")),
	})
	if err != nil {
		return nil, err
	}
	return res.Credentials, nil
}

// Receive user input from stdin
func receiveUserInput(description string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(description)
	text, _ := reader.ReadString('\n')
	return strings.TrimRight(text, "\n")
}

// Save profile to aws credential file
func saveProfile(f Flags, awsCred *sts.Credentials) error {
	cred := AWSCredentials{
		AWSAccessKey:    *awsCred.AccessKeyId,
		AWSSecretKey:    *awsCred.SecretAccessKey,
		AWSSessionToken: *awsCred.SessionToken,
		Expires:         *awsCred.Expiration,
	}

	config, err := ini.Load(f.Credential)
	if err != nil {
		return err
	}
	iniProfile, err := config.NewSection(f.Profile)
	if err != nil {
		return err
	}

	err = iniProfile.ReflectFrom(&cred)
	if err != nil {
		return err
	}

	return config.SaveTo(f.Credential)
}
