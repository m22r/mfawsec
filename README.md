# mfawsec
Set AWS temporary credentials with MFA token

## Usage
```
$ ./mfawsec -h
Set AWS temporary credentials with MFA token

Usage:
  mfawsec [flags]
  mfawsec [command]

Available Commands:
  help        Help about any command
  version     Print version

Flags:
      --credential string      path of AWS credentials (default "$HOME/.aws/credentials")
  -h, --help                   help for mfawsec
      --log-format string      logging format (text or json) (default "text")
      --profile string         profile name
      --serial-number string   serial number of MFA Device (ex. arn:aws:iam::111111111111:mfa/foo)

Use "mfawsec [command] --help" for more information about a command.
```
