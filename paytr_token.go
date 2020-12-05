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
	// MerchantId PayTR tarafından size verilen Mağaza numarası
	MerchantId     int   `schema:"merchant_id,required"`
	// UserIP İstek anında aldığınız müşteri ip numarası. En fazla 39 karakter (ipv4)
	UserIP         string   `schema:"user_ip,required"`
	// MerchantOid Satış işlemi için belirlediğiniz benzersizsipariş numarası.
	// En fazla 64 karakter, Alfa numerik
	MerchantOid    string   `schema:"merchant_oid,required"`
	// Mail Müşterinin sisteminizde kayıtlı olan eposta adresi.
	Mail           string   `schema:"email,required"`
	// PaymentAmount Siparişe ait toplam ödeme tutarının 100 ile çarpılmış hali.
	// Örn: 34.56 TL için 3456 gönderilmelidir
	PaymentAmount  int   `schema:"payment_amount,required"`
	// token İsteğin sizden geldiğine ve içeriğin değişmediğine emin olmamız
	// için oluşturacağınız değerdir.
	token          string   `schema:"paytr_token,required"`
	// basket Müşterinin sepet/sipariş içeriğinin encode olmuş hali
	basket         string   `schema:"user_basket,required"`
	// basketData Müşterinin sepet/sipariş içeriğinin sade hali
	basketData     []Basket `schema:"-"`
	// Debug Hata durumunda nedeni açıklaması için 1 yapın. paytr.Enable = 1
	Debug          mode     `schema:"debug_on"`
	// NoInstallment Taksit yapılmasını istemiyorsanız,
	// sadece tek çekim sunacaksanız 1 yapın. paytr.Enable = 1
	NoInstallment  mode     `schema:"no_installment"`
	// MaxInstallment Sayfada görüntülenecek taksit adedini sınırlamak istiyorsanız
	// uygun şekilde değiştirin. Sıfır (0) gönderilmesi durumunda yürürlükteki en fazla
	// izin verilen taksit geçerli olur.
	MaxInstallment int16    `schema:"max_installment"`
	// UserName Müşterinizin sitenizde kayıtlı veya form aracılığıyla aldığınız ad ve soyad bilgisi.
	UserName       string   `schema:"user_name,required"`
	// UserAddress Müşterinizin sitenizde kayıtlı veya form aracılığıyla aldığınız adres bilgisi.
	UserAddress    string   `schema:"user_address,required"`
	// UserAddress Müşterinizin sitenizde kayıtlı veya form aracılığıyla aldığınız telefon bilgisi.
	UserPhone      string   `schema:"user_phone,required"`
	// OkURL Başarılı ödeme sonrası müşterinizin yönlendirileceği sayfa.
	OkURL          string   `schema:"merchant_ok_url,required"`
	// FailURL Ödeme sürecinde beklenmedik bir hata oluşması durumunda müşterinizin yönlendirileceği sayfa.
	FailURL        string   `schema:"merchant_fail_url,required"`
	// TimeoutLimit İşlem zaman aşımı süresi - dakika cinsinden.
	TimeoutLimit   int16   `schema:"timeout_limit,required"`
	// Currency İşlemin yapılacağı para birimi. Örn. TL
	Currency       string   `schema:"currency,required"`
	// Test etmek istiyorsanız sanal iframe için bu modu 1 yapın. paytr.Enable = 1
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
	hashStr := strconv.Itoa(p.MerchantId) + p.UserIP + p.MerchantOid + p.Mail + strconv.Itoa(p.PaymentAmount) +
		p.basket + strconv.Itoa(int(p.NoInstallment)) + strconv.Itoa(int(p.MaxInstallment)) +
		p.Currency + strconv.Itoa(int(p.Test)) + merchantSalt

	tokenHmac := hmac.New(sha256.New, []byte(merchantKey))
	tokenHmac.Write([]byte(hashStr))

	p.token = base64.StdEncoding.EncodeToString(tokenHmac.Sum(nil))
	return p.token
}

func (p *PreparePayment) FetchToken() (TokenResponse, error) {
	var result TokenResponse
	var form url.Values
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
