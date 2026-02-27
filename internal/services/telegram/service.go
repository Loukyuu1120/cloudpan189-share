package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Service interface {
	SendMessage(msg string) error
	SendNotification(title, content string) error
	IsEnabled() bool
	GetProxyURL() string
}

type telegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type service struct {
	botToken string
	chatID   string
	proxyURL string
	enabled  bool
	logger   *zap.Logger
	client   *http.Client
	mu       sync.RWMutex
}

func NewService(botToken, chatID, proxyURL string, logger *zap.Logger) Service {
	s := &service{
		botToken: botToken,
		chatID:   chatID,
		proxyURL: proxyURL,
		logger:   logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	if botToken != "" && chatID != "" {
		s.enabled = true
		s.setupProxy()
	}

	return s
}

func (s *service) setupProxy() {
	if s.proxyURL == "" {
		return
	}

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*http.URL, error) {
			return http.ProxyFromEnvironment(req)
		},
	}

	if strings.HasPrefix(s.proxyURL, "http://") || strings.HasPrefix(s.proxyURL, "https://") {
		s.client.Transport = &http.Transport{
			Proxy: http.ProxyURL(func() *http.URL {
				u, _ := http.ParseURL(s.proxyURL)
				return u
			}()),
		}
		s.logger.Info("Telegram proxy configured", zap.String("proxy", s.proxyURL))
	}
}

func (s *service) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

func (s *service) GetProxyURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.proxyURL
}

func (s *service) SendMessage(msg string) error {
	s.mu.RLock()
	if !s.enabled || s.botToken == "" || s.chatID == "" {
		s.mu.RUnlock()
		return nil
	}
	botToken := s.botToken
	chatID := s.chatID
	proxyURL := s.proxyURL
	s.mu.RUnlock()

	message := telegramMessage{
		ChatID:    chatID,
		Text:      msg,
		ParseMode: "HTML",
	}

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	var client *http.Client
	if proxyURL != "" {
		proxy := func(*http.Request) (*http.URL, error) {
			return http.ParseURL(proxyURL)
		}
		transport := &http.Transport{Proxy: proxy}
		client = &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		}
	} else {
		client = s.client
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %d", resp.StatusCode)
	}

	return nil
}

func (s *service) SendNotification(title, content string) error {
	message := fmt.Sprintf("<b>%s</b>\n\n%s", title, content)
	return s.SendMessage(message)
}
