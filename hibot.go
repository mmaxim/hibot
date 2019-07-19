package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
)

func main() {
	rc := mainInner()
	os.Exit(rc)
}

type Options struct {
	KeybaseLocation string
	Home            string
}

type BotServer struct {
	opts Options
	kbc  *kbchat.API
}

func NewBotServer(opts Options) *BotServer {
	return &BotServer{
		opts: opts,
	}
}

func (s *BotServer) debug(msg string, args ...interface{}) {
	fmt.Printf("BotServer: "+msg+"\n", args...)
}

func (s *BotServer) getCommand() string {
	return fmt.Sprintf("hi%s", s.kbc.GetUsername())
}

func (s *BotServer) getCommandBang() string {
	return "!" + s.getCommand()
}

func (s *BotServer) makeAdvertisement() kbchat.Advertisement {
	var listExtendedBody = fmt.Sprintf(`@%s will respond with a :wave:`, s.kbc.GetUsername())
	return kbchat.Advertisement{
		Alias: "Hi Bot",
		Advertisements: []kbchat.CommandsAdvertisement{
			kbchat.CommandsAdvertisement{
				Typ: "public",
				Commands: []kbchat.Command{
					kbchat.Command{
						Name:        s.getCommand(),
						Description: "Say hello!",
						ExtendedDescription: &kbchat.CommandExtendedDescription{
							Title: fmt.Sprintf(`*%s*
Spread cheer`, s.getCommandBang()),
							DesktopBody: listExtendedBody,
							MobileBody:  listExtendedBody,
						},
					},
				},
			},
		},
	}
}

func (s *BotServer) maybeReact(msg kbchat.Message) error {
	if strings.Trim(strings.Split(msg.Content.Text.Body, " ")[0], " ") == s.getCommandBang() {
		return s.kbc.ReactByConvID(msg.ConversationID, msg.MsgID, ":wave:")
	}
	return nil
}

func (s *BotServer) Start() (err error) {
	if s.kbc, err = kbchat.Start(kbchat.RunOptions{
		KeybaseLocation: s.opts.KeybaseLocation,
		HomeDir:         s.opts.Home,
	}); err != nil {
		return err
	}
	if err := s.kbc.AdvertiseCommands(s.makeAdvertisement()); err != nil {
		s.debug("advertise error: %s", err)
		return err
	}
	if err := s.kbc.SendMessageByTlfName(s.kbc.GetUsername(), "I'm running."); err != nil {
		s.debug("failed to announce self: %s", err)
		return err
	}
	sub, err := s.kbc.ListenForNewTextMessages()
	if err != nil {
		return err
	}
	s.debug("startup success, listening for messages...")
	for {
		msg, err := sub.Read()
		if err != nil {
			s.debug("Read() error: %s", err.Error())
			continue
		}
		if err := s.maybeReact(msg.Message); err != nil {
			s.debug("failed to react: %s", err)
		}
	}
}

func mainInner() int {
	var opts Options
	flag.StringVar(&opts.KeybaseLocation, "keybase", "keybase", "keybase command")
	flag.StringVar(&opts.Home, "home", "", "Home directory")
	flag.Parse()

	bs := NewBotServer(opts)
	if err := bs.Start(); err != nil {
		fmt.Printf("error running chat loop: %s\n", err.Error())
	}
	return 0
}
