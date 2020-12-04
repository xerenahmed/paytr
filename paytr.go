package paytr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/gorilla/schema"
	"strconv"
)

var schemaEncoder = schema.NewEncoder()

type HandlePayment struct {
	Hash                string `schema:"hash,required"`
	MerchantId          string `schema:"merchant_id"`
	MerchantOid         string `schema:"merchant_oid,required"`
	Status              string `schema:"status,required"`
	TotalAmount         int    `schema:"total_amount,required"`
	PaymentAmount       int    `schema:"payment_amount"`
	PaymentType         string `schema:"payment_type"`
	Currency            string `schema:"currency"`
	Test                bool   `schema:"test_mode"`
	FailedReasonCode    int    `schema:"failed_reason_code"`
	FailedReasonMessage string `schema:"failed_reason_msg"`
}

func (p HandlePayment) Valid(merchantKey, merchantSalt string) bool {
	salt := hmac.New(sha256.New, []byte(merchantKey))
	salt.Write([]byte(p.MerchantOid + merchantSalt + p.Status + strconv.Itoa(p.TotalAmount)))

	saltHash := base64.StdEncoding.EncodeToString(salt.Sum(nil))
	return p.Hash == saltHash
}
