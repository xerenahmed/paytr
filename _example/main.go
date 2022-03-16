package _example

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/gorilla/schema"
	"github.com/xerenahmed/paytr"
)

func main() {
	router := chi.NewRouter()

	router.Post("/paytr", HandlePayment)
	router.Post("/paytrerr", ErrorPayTR)

	fmt.Println(http.ListenAndServe(":8080", router))
}

func ExampleCreatePayment() {
	merchantId, err := strconv.Atoi(os.Getenv("MERCHANT_ID"))
	if err != nil {
		panic(errors.New("failed parse merchant id as number"))
	}

	var realIP, mail, name, address, phone string

	p := paytr.PreparePayment{
		MerchantId:   merchantId,
		UserIP:       realIP, // get real ip from request
		Mail:         mail,   // user mail
		Debug:        paytr.Disable,
		UserName:     name,                 // user name
		UserAddress:  address,              // user address
		UserPhone:    phone,                // user phone
		OkURL:        os.Getenv("OK_URL"),  // The address that paytr redirects after the transaction is successful
		FailURL:      os.Getenv("ERR_URL"), // The address that paytr redirects after the transaction is failed
		TimeoutLimit: 30,                   // 30 dakika zaman aşımı olsun
		Currency:     "TL",
		Test:         paytr.Disable,
	}

	p.AddBasket(paytr.Basket{
		Name:    "Ürün Adı",
		PerCost: 10, // 10 TL
		Amount:  2,  // 2 items
	})

	p.GenerateToken(os.Getenv("MERCHANT_KEY"), os.Getenv("MERCHANT_SALT"))
	tk, err := p.FetchToken()
	if err != nil {
		// failed fetch token from paytr
		panic(err)
	}

	if failed, reason := tk.Failed(); failed {
		// failed to generate token
		panic(fmt.Errorf("failed to generate token: %s", reason))
	}

	// token is tk.Token
	// you can send iframe html code with tk.IFrame()
}

// callback url için gereken işlemleri yapar.
func HandlePayment(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		// failed parse form data
		w.Write([]byte(err.Error()))
		return
	}

	schemaDecoder := schema.NewDecoder()
	var payment paytr.HandlePayment
	if err := schemaDecoder.Decode(&payment, r.PostForm); err != nil {
		// failed decode form data
		w.Write([]byte(err.Error()))
		return
	}

	valid := payment.Valid(os.Getenv("MERCHANT_KEY"), os.Getenv("MERCHANT_SALT"))
	if !valid {
		// failed validation
		return
	}

	if payment.FailedReasonMessage != "" || payment.Status != "success" {
		// already handled, skip
		w.Write([]byte("OK"))
		return
	}

	// payment is success, handle something (databese etc.)

	// finish with OK
	w.Write([]byte("OK"))
}

func ErrorPayTR(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "Error: %s", err.Error())
	}

	// handle error from paytr
	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(`<meta name="charset" content="utf-8" />`))
	w.Write([]byte("error <br> " + fmt.Sprintf("%+v", r.PostForm)))
}
