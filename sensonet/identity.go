package sensonet

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/oauth"
	"github.com/evcc-io/evcc/util/request"
	"golang.org/x/oauth2"
)

const REALM_GERMANY = "vaillant-germany-b2c"

var Oauth2Config = Oauth2ConfigForRealm(REALM_GERMANY)

func Oauth2ConfigForRealm(realm string) *oauth2.Config {
	return &oauth2.Config{
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf(AUTH_URL, realm),
			TokenURL: fmt.Sprintf(TOKEN_URL, realm),
		},
		Scopes: []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess},
	}
}

type Identity struct {
	client *request.Helper
	realm  string
}

func NewIdentity(log *util.Logger, realm string) (*Identity, error) {
	client := request.NewHelper(log)
	client.Jar, _ = cookiejar.New(nil)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	v := &Identity{
		client: client,
		realm:  realm,
	}

	return v, nil
}

func (v *Identity) Login(user, password string) (*oauth2.Token, error) {
	cv := oauth2.GenerateVerifier()

	data := url.Values{
		"response_type":         {"code"},
		"client_id":             {CLIENT_ID},
		"code":                  {"code_challenge"},
		"redirect_uri":          {"enduservaillant.page.link://login"},
		"code_challenge_method": {"S256"},
		"code_challenge":        {oauth2.S256ChallengeFromVerifier(cv)},
	}

	uri := fmt.Sprintf(AUTH_URL, v.realm) + "?" + data.Encode()
	req, _ := http.NewRequest("GET", uri, nil)

	resp, err := v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	var code string
	if val, ok := resp.Header["Location"]; ok {
		parsedUrl, _ := url.Parse(val[0])
		code = parsedUrl.Query()["code"][0]
	}

	if code != "" {
		return nil, errors.New("missing code")
	}

	uri = computeLoginUrl(string(body), v.realm)
	if uri == "" {
		return nil, errors.New("missing login url")
	}

	params := url.Values{
		"username":     {user},
		"password":     {password},
		"credentialId": {""},
	}

	req, _ = http.NewRequest("POST", uri, strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = v.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not get code: %w", err)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, errors.New("could not find location header")
	}

	parsedUrl, _ := url.Parse(location)
	code = parsedUrl.Query()["code"][0]

	// get token
	var token TokenRequestStruct
	params = url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {CLIENT_ID},
		"code":          {code},
		"code_verifier": {cv},
		"redirect_uri":  {"enduservaillant.page.link://login"},
	}

	uri = fmt.Sprintf(TOKEN_URL, v.realm)
	req, _ = http.NewRequest("POST", uri, strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := v.client.DoJSON(req, &token); err != nil {
		return nil, fmt.Errorf("could not get token: %w", err)
	}

	token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)

	return &token.Token, nil
}

func (v *Identity) TokenSource(token *oauth2.Token) (oauth2.TokenSource, error) {
	ts := oauth2.ReuseTokenSource(token, oauth.RefreshTokenSource(token, v))
	_, err := ts.Token()
	return ts, err
}

func computeLoginUrl(body, realm string) string {
	url := fmt.Sprintf(LOGIN_URL, realm)
	index1 := strings.Index(body, "authenticate?")
	if index1 < 0 {
		return ""
	}
	index2 := strings.Index(body[index1:], "\"")
	if index2 < 0 {
		return ""
	}
	return html.UnescapeString(url + body[index1+12:index1+index2])
}

func (v *Identity) RefreshToken(token *oauth2.Token) (*oauth2.Token, error) {
	var res TokenRequestStruct
	params := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {CLIENT_ID},
		"refresh_token": {token.RefreshToken},
	}

	uri := fmt.Sprintf(TOKEN_URL, v.realm)
	req, _ := http.NewRequest("POST", uri, strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := v.client.DoJSON(req, &res); err != nil {
		return nil, err
	}

	res.Expiry = time.Now().Add(time.Duration(res.ExpiresIn) * time.Second)

	return &res.Token, nil
}
