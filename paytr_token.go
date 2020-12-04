package paytr

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type mode int16

var (
	Enable  = mode(1)
	Disable = mode(0)
)

type Basket struct {
	Name    string
	PerCost int
	Amount  int
}

type PreparePayment struct {
	MerchantId     string   `schema:"merchant_id,required"`
	MerchantOid    string   `schema:"merchant_oid,required"`
	UserIP         string   `schema:"user_ip,required"`
	Mail           string   `schema:"email,required"`
	PaymentAmount  string   `schema:"payment_amount,required"`
	token          string   `schema:"paytr_token,required"`
	basket         string   `schema:"user_basket,required"`
	basketData     []Basket `schema:"-"`
	Debug          mode     `schema:"debug_on"`
	NoInstallment  mode     `schema:"no_installment"`
	MaxInstallment int16    `schema:"max_installment"`
	UserName       string   `schema:"user_name,required"`
	UserAddress    string   `schema:"user_address,required"`
	UserPhone      string   `schema:"user_phone,required"`
	OkURL          string   `schema:"merchant_ok_url,required"`
	FailURL        string   `schema:"merchant_fail_url,required"`
	TimeoutLimit   string   `schema:"timeout_limit,required"`
	Currency       string   `schema:"currency,required"`
	Test           mode     `schema:"test_mode"`
}

func (p *PreparePayment) AddBasket(basket ...Basket) {
	p.basketData = append(p.basketData, basket...)

	var basketData []map[int]interface{}
	for _, basket := range p.basketData {
		basketData = append(basketData, map[int]interface{}{
			0: basket.Name,
			1: strconv.Itoa(basket.PerCost),
			2: basket.Amount,
		})
	}

	basketBytes, err := json.Marshal(basketData)
	if err != nil {
		panic(err)
	}
	p.basket = base64.StdEncoding.EncodeToString(basketBytes)
}

func (p *PreparePayment) GenerateToken(merchantKey, merchantSalt string) string {
	hashStr := p.MerchantId + p.UserIP + p.MerchantOid + p.Mail + p.PaymentAmount +
		p.basket + strconv.Itoa(int(p.NoInstallment)) + strconv.Itoa(int(p.MaxInstallment)) +
		p.Currency + strconv.Itoa(int(p.Test)) + merchantSalt

	tokenHmac := hmac.New(sha256.New, []byte(merchantKey))
	tokenHmac.Write([]byte(hashStr))

	p.token = base64.StdEncoding.EncodeToString(tokenHmac.Sum(nil))
	return p.token
}

func (p *PreparePayment) FetchToken() (TokenResponse, error) {
	var result TokenResponse
	form := url.Values{}
	if err := schemaEncoder.Encode(p, form); err != nil {
		return result, err
	}

	res, err := http.PostForm("https://www.paytr.com/odeme/api/get-token", form)
	if err != nil {
		return result, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	resText, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(resText, &result)
	return result, err
}
