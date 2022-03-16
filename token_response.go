package paytr

import "fmt"

// TokenResponse Token alma işleminin sonucu döndürür.
type TokenResponse struct {
	// İşlem sonucu.
	Status string
	// İşlem başarılı ise token bilgisi döndürür.
	Token string
	// Hata durumunda döndürülecek hata mesajı.
	Reason string
}

// Failed Hata varsa sorunu döndürür.
func (p TokenResponse) Failed() (bool, string) {
	return p.Status == "failed", p.Reason
}

// IFrame tokenin kullanıldığı bir iframe HTML kodu döndürür.
func (p TokenResponse) IFrame() string {
	return fmt.Sprintf(`
	<script src="https://www.paytr.com/js/iframeResizer.min.js"></script>
		<iframe src="https://www.paytr.com/odeme/guvenli/%s" id="paytriframe" frameborder="0" scrolling="no" style="width: 100%s"></iframe>
	<script>iFrameResize({}, "#paytriframe");</script>`, p.Token, "%;")
}
