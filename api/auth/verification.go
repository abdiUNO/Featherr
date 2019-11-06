package auth

import (
	"context"
	"fmt"
	"github.com/abdiUNO/featherr/config"
	"github.com/abdiUNO/featherr/utils"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"time"
)

func KeyFromUser(user *User) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "featherrapp.com",
		AccountName: user.Email,
		Secret:      []byte(user.Password),
	})

	return key, err
}

func CreateCode(user *User) (string, error) {
	key, err := KeyFromUser(user)
	if err != nil {
		return "", utils.NewError(utils.ECONFLICT, "could not create otp key", nil)
	}

	code, codeErr := totp.GenerateCodeCustom(key.Secret(), time.Now(), totp.ValidateOpts{
		Period:    660,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	if codeErr != nil {
		return "", utils.NewError(utils.ECONFLICT, "could not create otp code", nil)
	}

	return code, nil
}

func ValidateCode(passcode string, user *User) (bool, error) {
	key, err := KeyFromUser(user)
	if err != nil {
		return false, utils.NewError(utils.ECONFLICT, "could not create otp key", nil)
	}

	valid, validErr := totp.ValidateCustom(passcode, key.Secret(), time.Now(), totp.ValidateOpts{
		Period:    660,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})

	return valid, validErr
}

func EmailCode(ctx context.Context, passcode string, user *User) error {
	cfg := config.GetConfig()

	mg := mailgun.NewMailgun("mg.featherrapp.com", cfg.MailGunApiKey)

	sender := "Featherr App <noreply@mg.featherrapp.com>"
	subject := "Verification code!"
	recipient := user.Email
	template := "otp-email"
	// The message object allows you to add attachments and Bcc recipients
	message := mg.NewMessage(sender, subject, "", recipient)
	message.SetTemplate(template)

	message.AddVariable("passcode", passcode)
	message.AddVariable("username", user.FullName)

	msg, id, err := mg.Send(ctx, message)

	if err != nil {
		return err
	}

	fmt.Println(msg)
	fmt.Println(id)

	return nil
}
