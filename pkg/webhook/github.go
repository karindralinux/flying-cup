package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

type GithubPRWebhook struct {
	Action      string      `json:"action"`
	Number      int         `json:"number"`
	PullRequest PullRequest `json:"pull_request"`
	Repository  Repository  `json:"repository"`
	Sender      Sender      `json:"sender"`
}

type Sender struct {
	Username string `json:"login"`
}

type Repository struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	CloneUrl string `json:"clone_url"`
}

type PullRequest struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Url   string `json:"url"`
	Head  struct {
		Ref string `json:"ref"`
		Sha string `json:"sha"`
	} `json:"head"`
}

func HandleGithubWebhook(
	webhookSecret string,
	onPROpened func(ctx context.Context, webhook *GithubPRWebhook) error,
	onPRClosed func(ctx context.Context, webhook *GithubPRWebhook) error,
) echo.HandlerFunc {
	return func(c echo.Context) error {

		body, err := io.ReadAll(c.Request().Body)

		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to read request body")
		}

		signature := c.Request().Header.Get("X-Hub-Signature-256")

		if err := validateSignature(body, signature, webhookSecret); err != nil {
			return c.String(http.StatusUnauthorized, "Invalid signature")
		}

		// Parse Webhook
		webhook, err := parseWebhook(body)

		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to parse webhook")
		}

		switch webhook.Action {
		case "opened":
			return handlePROpened(c, webhook, onPROpened)
		case "reopened":
			return handlePROpened(c, webhook, onPROpened)
		case "closed":
			return handlePRClosed(c, webhook, onPRClosed)
		default:
			return c.JSON(http.StatusOK, map[string]string{"status": "ignored"})
		}
	}
}

func handlePRClosed(c echo.Context, webhook *GithubPRWebhook, onPRClosed func(ctx context.Context, webhook *GithubPRWebhook) error) error {

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ Webhook handler panic for PR #%d: %v", webhook.Number, r)
			}
		}()

		log.Printf("📝 Processing PR closed event for PR #%d", webhook.Number)

		err := onPRClosed(context.Background(), webhook)

		if err != nil {
			log.Printf("❌ PR closed handler failed for PR #%d: %v", webhook.Number, err)
		} else {
			log.Printf("✅ PR closed handler completed for PR #%d", webhook.Number)
		}
	}()

	return c.JSON(http.StatusOK, map[string]string{"status": "handle pr closed triggered"})
}

func handlePROpened(c echo.Context, webhook *GithubPRWebhook, onPROpened func(ctx context.Context, webhook *GithubPRWebhook) error) error {

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ PR opened panic for PR #%d: %v", webhook.Number, r)
			}
		}()

		err := onPROpened(context.Background(), webhook)

		if err != nil {
			log.Printf("❌ PR opened failed for PR #%d: %v", webhook.Number, err)
		} else {
			log.Printf("✅ PR opened completed for PR #%d", webhook.Number)
		}
	}()

	return c.JSON(http.StatusOK, map[string]string{"status": "handle pr opened triggered"})
}

func parseWebhook(body []byte) (*GithubPRWebhook, error) {
	var webhook GithubPRWebhook

	formData, err := url.ParseQuery(string(body))
	if err != nil {
		log.Println("Error parsing form data:", err)
		return nil, echo.NewHTTPError(400, "Invalid form data")
	}

	payloadStr := formData.Get("payload")
	if payloadStr == "" {
		log.Println("Error: No payload field in form data")
		return nil, echo.NewHTTPError(400, "No payload field")
	}

	decodedPayload, err := url.QueryUnescape(payloadStr)
	if err != nil {
		log.Println("Error decoding payload:", err)
		return nil, echo.NewHTTPError(400, "Invalid URL encoding")
	}

	if err := json.Unmarshal([]byte(decodedPayload), &webhook); err != nil {
		log.Println("Error unmarshalling decoded payload:", err)
		return nil, echo.NewHTTPError(400, "Invalid JSON payload")
	}

	webhookJSON, err := json.MarshalIndent(webhook, "", "  ")

	if err != nil {
		log.Println("Failed to marshal webhook to JSON:", err)
		return nil, err
	}

	log.Println("Parsed Github Webhook:\n", string(webhookJSON))

	return &webhook, nil
}
func validateSignature(body []byte, signature string, webhookSecret string) error {

	fmt.Printf("webhookSecret: %s\n", webhookSecret)
	fmt.Printf("signature: %s\n", signature)

	if webhookSecret == "" {
		return nil
	}

	if !strings.HasPrefix(signature, "sha256=") {
		return fmt.Errorf("invalid signature")
	}

	expectedSignature := signature[7:]

	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	calculatedSignature := hex.EncodeToString(mac.Sum(nil))

	if calculatedSignature != expectedSignature {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
