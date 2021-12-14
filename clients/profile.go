package clients

import (
	"net/http"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// ClientProfile
type ClientProfile struct {
	*profile.ClientProfile
	HttpTransport *http.Transport
}

// NewClientProfile
func NewClientProfile() *ClientProfile {
	clientProFile := &ClientProfile{
		ClientProfile: profile.NewClientProfile(),
	}
	clientProFile.ClientProfile.Language = "en-US"
	clientProFile.ClientProfile.HttpProfile.ReqTimeout = 90
	return clientProFile

}
