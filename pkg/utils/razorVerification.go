package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
)

func RazorPaymentVerification(sign, orderId, paymentId string) error {
	signature := sign
	secret := "MU9PWzkhTBSCkPnxEUOAZdYW"
	data := orderId + "|" + paymentId

	h := hmac.New(sha256.New, []byte(secret))

	_, err := h.Write([]byte(data))

	if err != nil {
		panic(err)
	}

	sha := hex.EncodeToString(h.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(sha), []byte(signature)) != 1 {
		return errors.New("Payment failed")
	} else {
		return nil
	}
}
