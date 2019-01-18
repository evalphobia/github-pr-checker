package prchecker

import (
	"github.com/evalphobia/aws-sdk-go-wrapper/config"
	"github.com/evalphobia/aws-sdk-go-wrapper/kms"
)

// DecryptKMS decrypts base64 text encrypted by AWS KMS to plain text.
func DecryptKMS(base64Text string) (string, error) {
	cli, err := kms.New(config.Config{})
	if err != nil {
		return "", err
	}

	return cli.DecryptString(base64Text)
}
