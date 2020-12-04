package paytr

import "fmt"

type TokenResponse struct {
	Status string
	Token  string
	Reason string
}

func (p TokenResponse) Failed() (bool, string) {
	return p.Status == "failed", p.Reason
}

func (p TokenResponse) IFrame() string {
	return fmt.Sprintf(`<script src="https://www.paytr.com/js/iframeResizer.min.js"></script>
			<iframe src="https://www.paytr.com/odeme/guvenli/%s" id="paytriframe" frameborder="0" scrolling="no" style="width: 100%s"></iframe>
			<script>iFrameResize({}, "#paytriframe");</script>	
		`, p.Token, "%;")
}
