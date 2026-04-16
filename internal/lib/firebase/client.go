// Pacote firebase fornece integração com o cliente Firebase para autenticação e notificações.
package firebase

import (
	"context"
	"encoding/base64"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"

	"github.com/example/go-backend-boilerplate/internal/config"
)

type Client struct {
	authClient *auth.Client
	config     *config.FirebaseConfig
}

// NewClient inicializa Firebase Admin SDK
func NewClient(cfg *config.FirebaseConfig) (*Client, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	var opt option.ClientOption

	if cfg.ServiceAccountJSON != "" {
		// Produção: base64 encoded JSON
		decoded, err := base64.StdEncoding.DecodeString(cfg.ServiceAccountJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to decode service account JSON: %w", err)
		}
		opt = option.WithCredentialsJSON(decoded)
	} else if cfg.ServiceAccountPath != "" {
		// Local: arquivo JSON
		opt = option.WithCredentialsFile(cfg.ServiceAccountPath)
	} else {
		return nil, fmt.Errorf("firebase service account not configured")
	}

	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: cfg.ProjectID,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get auth client: %w", err)
	}

	return &Client{
		authClient: authClient,
		config:     cfg,
	}, nil
}

// VerifyIDToken valida Firebase ID token e retorna claims
func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (TokenAdapter, error) {
	token, err := c.authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return TokenAdapter{}, fmt.Errorf("failed to verify ID token: %w", err)
	}
	return TokenAdapter{token: token}, nil
}

// GetUser busca informações do usuário no Firebase
func (c *Client) GetUser(ctx context.Context, uid string) (UserAdapter, error) {
	user, err := c.authClient.GetUser(ctx, uid)
	if err != nil {
		return UserAdapter{}, fmt.Errorf("failed to get user: %w", err)
	}
	return UserAdapter{user: user}, nil
}

// TokenAdapter adapta auth.Token para a interface esperada
type TokenAdapter struct {
	token *auth.Token
}

func (t TokenAdapter) GetUID() string {
	return t.token.UID
}

func (t TokenAdapter) GetEmail() string {
	if email, ok := t.token.Claims["email"].(string); ok {
		return email
	}
	return ""
}

func (t TokenAdapter) GetName() string {
	if name, ok := t.token.Claims["name"].(string); ok {
		return name
	}
	return ""
}

func (t TokenAdapter) GetPicture() string {
	if picture, ok := t.token.Claims["picture"].(string); ok {
		return picture
	}
	return ""
}

func (t TokenAdapter) GetProviderID() string {
	if firebase, ok := t.token.Claims["firebase"].(map[string]interface{}); ok {
		if signInProvider, ok := firebase["sign_in_provider"].(string); ok {
			return signInProvider
		}
	}
	return ""
}

// UserAdapter adapta auth.UserRecord para a interface esperada
type UserAdapter struct {
	user *auth.UserRecord
}

func (u UserAdapter) GetUID() string {
	return u.user.UID
}

func (u UserAdapter) GetEmail() string {
	return u.user.Email
}

func (u UserAdapter) GetDisplayName() string {
	return u.user.DisplayName
}

func (u UserAdapter) GetPhotoURL() string {
	return u.user.PhotoURL
}
