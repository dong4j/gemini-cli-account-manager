package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gemini-cli-account-manager/internal/config"
)

const (
	GoogleAuthURL     = "https://accounts.google.com/o/oauth2/v2/auth"
	GoogleTokenURL    = "https://oauth2.googleapis.com/token"
	GoogleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

var Scopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/userinfo.email",
	"https://www.googleapis.com/auth/userinfo.profile",
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type UserInfo struct {
	Email string `json:"email"`
}

// OpenBrowser opens the specified URL in the default browser
func OpenBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // linux, freebsd, etc.
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

// Login performs OAuth login and adds the account to the pool
func Login(cfg *config.Config) error {
	// 1. Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d", port)

	// 2. Build Auth URL
	u, _ := url.Parse(GoogleAuthURL)
	q := u.Query()
	q.Set("client_id", cfg.OAuthClient.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(Scopes, " "))
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	u.RawQuery = q.Encode()

	fmt.Printf("Opening browser for OAuth login...\n")
	if err := OpenBrowser(u.String()); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Printf("Please open this URL manually: %s\n", u.String())
	}

	// 3. Start local server to receive code
	codeChan := make(chan string)
	errChan := make(chan error)
	server := &http.Server{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if code != "" {
			fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Login Success | GCAM</title>
    <script src="https://cdn.jsdelivr.net/npm/canvas-confetti@1.6.0/dist/confetti.browser.min.js"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            background-color: #0f172a;
            color: #f8fafc;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            overflow: hidden;
        }
        .card {
            background: rgba(30, 41, 59, 0.7);
            backdrop-filter: blur(10px);
            padding: 3rem;
            border-radius: 1.5rem;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
            text-align: center;
            border: 1px solid rgba(255, 255, 255, 0.1);
            animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1);
        }
        @keyframes slideUp {
            from { transform: translateY(30px); opacity: 0; }
            to { transform: translateY(0); opacity: 1; }
        }
        h1 { color: #38bdf8; margin-bottom: 0.5rem; font-size: 2.5rem; }
        p { color: #94a3b8; font-size: 1.1rem; }
        .icon {
            font-size: 5rem;
            margin-bottom: 1rem;
            display: inline-block;
            animation: bounce 2s infinite;
        }
        @keyframes bounce {
            0%%, 100%% { transform: translateY(0); }
            50%% { transform: translateY(-10px); }
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">🚀</div>
        <h1>Login Successful!</h1>
        <p>Your Google account has been captured into the pool.</p>
        <p style="margin-top: 2rem; font-size: 0.9rem; opacity: 0.6;">You can safely close this window now.</p>
    </div>
    <script>
        const end = Date.now() + (3 * 1000);
        const colors = ['#38bdf8', '#818cf8', '#c084fc', '#fb7185'];

        (function frame() {
            confetti({
                particleCount: 3,
                angle: 60,
                spread: 55,
                origin: { x: 0 },
                colors: colors
            });
            confetti({
                particleCount: 3,
                angle: 120,
                spread: 55,
                origin: { x: 1 },
                colors: colors
            });

            if (Date.now() < end) {
                requestAnimationFrame(frame);
            }
        }());
    </script>
</body>
</html>`)
			codeChan <- code
		} else {
			fmt.Fprintf(w, "<h1>Login Failed</h1><p>Please check the console for details.</p>")
			errChan <- fmt.Errorf("no code received")
		}
	})

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	var code string
	select {
	case code = <-codeChan:
		// Success
	case err := <-errChan:
		return err
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("login timeout")
	}

	// 4. Shutdown server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)

	// 5. Exchange code for tokens
	token, err := ExchangeCode(cfg, code, redirectURI)
	if err != nil {
		return err
	}

	// 6. Get user info
	userInfo, err := GetUserInfo(token.AccessToken)
	if err != nil {
		return err
	}

	// 7. Save to profile
	profileDir := filepath.Join(config.ProfilesDir, userInfo.Email)
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return err
	}

	credsData := map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.RefreshToken,
		"expiry_date":   time.Now().UnixMilli() + int64(token.ExpiresIn*1000),
		"client_id":     cfg.OAuthClient.ClientID,
		"client_secret": cfg.OAuthClient.ClientSecret,
	}

	data, _ := json.MarshalIndent(credsData, "", "  ")
	if err := os.WriteFile(filepath.Join(profileDir, "oauth_creds.json"), data, 0644); err != nil {
		return err
	}

	// Also create a placeholder google_account_id if needed, or just let it be
	_ = os.WriteFile(filepath.Join(profileDir, "google_account_id"), []byte(userInfo.Email), 0644)

	fmt.Printf("Successfully added account: %s\n", userInfo.Email)
	
	// Automatically switch to the new account
	_, err = FastSwitch(userInfo.Email)
	return err
}

func ExchangeCode(cfg *config.Config, code, redirectURI string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", cfg.OAuthClient.ClientID)
	data.Set("client_secret", cfg.OAuthClient.ClientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm(GoogleTokenURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func GetUserInfo(accessToken string) (*UserInfo, error) {
	req, _ := http.NewRequest("GET", GoogleUserInfoURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %d", resp.StatusCode)
	}

	var info UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}
