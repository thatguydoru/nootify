package nootify

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type nootOption struct {
	roleID  string
	keyword string
	content string
}

// Nootify structure, contains most of the necessary info and options.
type Nootify struct {
	session      *discordgo.Session
	message      *discordgo.Message
	prefixSymbol rune

	// mapping between emojiID and a noot option
	options map[string]*nootOption
}

// Make a new nootify associated to a target message. Thus, making that message
// able to add or remove noot roles by monitoring who reacted with the proper emoji.
func InitNootify(
	session *discordgo.Session,
	message *discordgo.Message,
	prefixSymbol rune,
) Nootify {
    // TODO: See if there's a config file. If there is, then read the persisted state.

	return Nootify{
		session,
		message,
		prefixSymbol,
		make(map[string]*nootOption),
	}
}

// Add a new noot option. If the emoji associated with the noot option is already
// registered, reregistering it will return an error.
func (self *Nootify) RegisterNootOption(
	emojiID string,
	roleID string,
	keyword string,
	content string,
) error {
	_, ok := self.options[emojiID]
	if ok {
		return fmt.Errorf("noot option already registered.\n")
	}
	self.options[emojiID] = &nootOption{
		roleID,
		keyword,
		content,
	}
	return nil
}

// The method that must be called last to run the nootifier.
// This attaches all the necessary event handlers.
func (self *Nootify) GoNoot() {
    // TODO: persist Nootify state in the background

	for emojiID, noot := range self.options {
		self.session.AddHandler(func(
			session *discordgo.Session,
			event *discordgo.MessageReactionAdd,
		) {
			emojiMatch := event.Emoji.ID == emojiID || event.Emoji.Name == emojiID
			if emojiMatch && event.MessageID == self.message.ID {
				err := session.GuildMemberRoleAdd(
					event.GuildID,
					event.UserID,
					noot.roleID,
				)
				if err != nil {
					log.Println("failed to add role to user:", err)
					return
				}
				log.Println("member subscribed to nootify role:", noot.roleID)
			}
		})

		self.session.AddHandler(func(
			session *discordgo.Session,
			event *discordgo.MessageReactionRemove,
		) {
			emojiMatch := event.Emoji.ID == emojiID || event.Emoji.Name == emojiID
			if emojiMatch && event.MessageID == self.message.ID {
				err := session.GuildMemberRoleRemove(
					event.GuildID,
					event.UserID,
					noot.roleID,
				)
				if err != nil {
					log.Println("failed to remove role from user:", err)
					return
				}
				log.Println("member unsubscribed from nootify role:", noot.roleID)
			}
		})

		self.session.AddHandler(func(
			session *discordgo.Session,
			event *discordgo.MessageCreate,
		) {
			prefix := fmt.Sprintf("%v%v", self.prefixSymbol, noot.keyword)
			isOwner := event.Author.ID == session.State.Application.Owner.ID
			if isOwner && strings.HasPrefix(event.Content, prefix) {
				role, err := getRole(self.session, self.message.GuildID, noot.roleID)
				if err != nil {
					log.Println("failed to find role:", err)
					return
				}
				session.ChannelMessageSend(
					self.message.ChannelID,
					fmt.Sprintf("%v %v", role.Mention(), noot.content),
				)
				log.Println("noot option command called:", noot.keyword)
				session.ChannelMessageDelete(self.message.ChannelID, event.Message.ID)
			}
		})
	}
}

func getRole(session *discordgo.Session, guildID string, roleID string) (*discordgo.Role, error) {
	roles, err := session.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.ID == roleID {
			return role, nil
		}
	}
	return nil, fmt.Errorf("role not found")
}
