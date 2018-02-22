package easytraceCache

import (
	"fmt"
	"github.com/zamedic/go2hal/telegram"
	"gopkg.in/telegram-bot-api.v4"
)

type reloadEasyTraceCache struct {
	telegram telegram.Service
	service  Service
}

func ReloadCacheCallCommand(telegram telegram.Service, service Service) telegram.Command {
	return &reloadEasyTraceCache{telegram, service}
}

/* Set Heartbeat group */
func (s *reloadEasyTraceCache) CommandIdentifier() string {
	return "ReloadCache"
}

func (s *reloadEasyTraceCache) CommandDescription() string {
	return "Reload Easy Trace Cache?"
}

func (s *reloadEasyTraceCache) Execute(update tgbotapi.Update) {
	name := s.service.reloadCache()
	s.telegram.SendMessage(update.Message.Chat.ID, fmt.Sprintf(name), update.Message.MessageID)
}
