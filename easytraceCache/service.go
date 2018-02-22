package easytraceCache

import (
	"os"
	"github.com/zamedic/go2hal/chef"
	"fmt"
)

type Service interface {
	reloadCache() (string)
}

type service struct {
	chefService chef.Service
}

func NewService(chefService chef.Service) Service {
	return &service{chefService}
}

func (s *service) reloadCache() (string) {
	nodes := s.chefService.FindNodesFromFriendlyNames("easytrace-app", "acceptance-chopchop-VirtualChannels-easytrace-app-master")
	fmt.Println(len(nodes))
	return "hello"
}

func getCacheURL() string {
	return os.Getenv("CACHE_URL")
}
