/*
*
*	Gosora Topic File
*	Copyright Azareal 2017 - 2018
*
 */
package common

//import "fmt"
import (
	"html"
	"html/template"
	"strconv"
	"time"
)

// This is also in reply.go
//var ErrAlreadyLiked = errors.New("This item was already liked by this user")

// ? - Add a TopicMeta struct for *Forums?

type Topic struct {
	ID                  int
	Link                string
	Title               string
	Content             string
	CreatedBy           int
	IsClosed            bool
	Sticky              bool
	CreatedAt           time.Time
	RelativeCreatedAt   string
	LastReplyAt         time.Time
	RelativeLastReplyAt string
	//LastReplyBy int
	ParentID  int
	Status    string // Deprecated. Marked for removal.
	IPAddress string
	PostCount int
	LikeCount int
	ClassName string // CSS Class Name
	Data      string // Used for report metadata
}

type TopicUser struct {
	ID                  int
	Link                string
	Title               string
	Content             string
	CreatedBy           int
	IsClosed            bool
	Sticky              bool
	CreatedAt           time.Time
	RelativeCreatedAt   string
	LastReplyAt         time.Time
	RelativeLastReplyAt string
	//LastReplyBy int
	ParentID  int
	Status    string // Deprecated. Marked for removal.
	IPAddress string
	PostCount int
	LikeCount int
	ClassName string
	Data      string // Used for report metadata

	UserLink      string
	CreatedByName string
	Group         int
	Avatar        string
	ContentLines  int
	ContentHTML   string
	Tag           string
	URL           string
	URLPrefix     string
	URLName       string
	Level         int
	Liked         bool
}

type TopicsRow struct {
	ID                  int
	Link                string
	Title               string
	Content             string
	CreatedBy           int
	IsClosed            bool
	Sticky              bool
	CreatedAt           string
	LastReplyAt         time.Time
	RelativeLastReplyAt string
	LastReplyBy         int
	ParentID            int
	Status              string // Deprecated. Marked for removal. -Is there anything we could use it for?
	IPAddress           string
	PostCount           int
	LikeCount           int
	ClassName           string
	Data                string // Used for report metadata

	Creator      *User
	CSS          template.CSS
	ContentLines int
	LastUser     *User

	ForumName string //TopicsRow
	ForumLink string
}

// Flush the topic out of the cache
// ? - We do a CacheRemove() here instead of mutating the pointer to avoid creating a race condition
func (topic *Topic) cacheRemove() {
	tcache, ok := topics.(TopicCache)
	if ok {
		tcache.CacheRemove(topic.ID)
	}
}

// TODO: Write a test for this
func (topic *Topic) AddReply(uid int) (err error) {
	_, err = stmts.addRepliesToTopic.Exec(1, uid, topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) Lock() (err error) {
	_, err = stmts.lockTopic.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) Unlock() (err error) {
	_, err = stmts.unlockTopic.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

// TODO: We might want more consistent terminology rather than using stick in some places and pin in others. If you don't understand the difference, there is none, they are one and the same.
func (topic *Topic) Stick() (err error) {
	_, err = stmts.stickTopic.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) Unstick() (err error) {
	_, err = stmts.unstickTopic.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

// TODO: Test this
// TODO: Use a transaction for this
func (topic *Topic) Like(score int, uid int) (err error) {
	var tid int // Unused
	err = stmts.hasLikedTopic.QueryRow(uid, topic.ID).Scan(&tid)
	if err != nil && err != ErrNoRows {
		return err
	} else if err != ErrNoRows {
		return ErrAlreadyLiked
	}

	_, err = stmts.createLike.Exec(score, tid, "topics", uid)
	if err != nil {
		return err
	}

	_, err = stmts.addLikesToTopic.Exec(1, tid)
	topic.cacheRemove()
	return err
}

// TODO: Implement this
func (topic *Topic) Unlike(uid int) error {
	return nil
}

// TODO: Use a transaction here
func (topic *Topic) Delete() error {
	topicCreator, err := users.Get(topic.CreatedBy)
	if err == nil {
		wcount := wordCount(topic.Content)
		err = topicCreator.decreasePostStats(wcount, true)
		if err != nil {
			return err
		}
	} else if err != ErrNoRows {
		return err
	}

	err = fstore.RemoveTopic(topic.ParentID)
	if err != nil && err != ErrNoRows {
		return err
	}

	_, err = stmts.deleteTopic.Exec(topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) Update(name string, content string) error {
	content = preparseMessage(content)
	parsed_content := parseMessage(html.EscapeString(content), topic.ParentID, "forums")

	_, err := stmts.editTopic.Exec(name, content, parsed_content, topic.ID)
	topic.cacheRemove()
	return err
}

func (topic *Topic) CreateActionReply(action string, ipaddress string, user User) (err error) {
	_, err = stmts.createActionReply.Exec(topic.ID, action, ipaddress, user.ID)
	if err != nil {
		return err
	}
	_, err = stmts.addRepliesToTopic.Exec(1, user.ID, topic.ID)
	topic.cacheRemove()
	// ? - Update the last topic cache for the parent forum?
	return err
}

// Copy gives you a non-pointer concurrency safe copy of the topic
func (topic *Topic) Copy() Topic {
	return *topic
}

// TODO: Refactor the caller to take a Topic and a User rather than a combined TopicUser
func getTopicUser(tid int) (TopicUser, error) {
	tcache, tok := topics.(TopicCache)
	ucache, uok := users.(UserCache)
	if tok && uok {
		topic, err := tcache.CacheGet(tid)
		if err == nil {
			user, err := users.Get(topic.CreatedBy)
			if err != nil {
				return TopicUser{ID: tid}, err
			}

			// We might be better off just passing separate topic and user structs to the caller?
			return copyTopicToTopicUser(topic, user), nil
		} else if ucache.Length() < ucache.GetCapacity() {
			topic, err = topics.Get(tid)
			if err != nil {
				return TopicUser{ID: tid}, err
			}
			user, err := users.Get(topic.CreatedBy)
			if err != nil {
				return TopicUser{ID: tid}, err
			}
			return copyTopicToTopicUser(topic, user), nil
		}
	}

	tu := TopicUser{ID: tid}
	err := stmts.getTopicUser.QueryRow(tid).Scan(&tu.Title, &tu.Content, &tu.CreatedBy, &tu.CreatedAt, &tu.IsClosed, &tu.Sticky, &tu.ParentID, &tu.IPAddress, &tu.PostCount, &tu.LikeCount, &tu.CreatedByName, &tu.Avatar, &tu.Group, &tu.URLPrefix, &tu.URLName, &tu.Level)
	tu.Link = buildTopicURL(nameToSlug(tu.Title), tu.ID)
	tu.UserLink = buildProfileURL(nameToSlug(tu.CreatedByName), tu.CreatedBy)
	tu.Tag = gstore.DirtyGet(tu.Group).Tag

	if tok {
		theTopic := Topic{ID: tu.ID, Link: tu.Link, Title: tu.Title, Content: tu.Content, CreatedBy: tu.CreatedBy, IsClosed: tu.IsClosed, Sticky: tu.Sticky, CreatedAt: tu.CreatedAt, LastReplyAt: tu.LastReplyAt, ParentID: tu.ParentID, IPAddress: tu.IPAddress, PostCount: tu.PostCount, LikeCount: tu.LikeCount}
		//log.Printf("the_topic: %+v\n", theTopic)
		_ = tcache.CacheAdd(&theTopic)
	}
	return tu, err
}

func copyTopicToTopicUser(topic *Topic, user *User) (tu TopicUser) {
	tu.UserLink = user.Link
	tu.CreatedByName = user.Name
	tu.Group = user.Group
	tu.Avatar = user.Avatar
	tu.URLPrefix = user.URLPrefix
	tu.URLName = user.URLName
	tu.Level = user.Level

	tu.ID = topic.ID
	tu.Link = topic.Link
	tu.Title = topic.Title
	tu.Content = topic.Content
	tu.CreatedBy = topic.CreatedBy
	tu.IsClosed = topic.IsClosed
	tu.Sticky = topic.Sticky
	tu.CreatedAt = topic.CreatedAt
	tu.LastReplyAt = topic.LastReplyAt
	tu.ParentID = topic.ParentID
	tu.IPAddress = topic.IPAddress
	tu.PostCount = topic.PostCount
	tu.LikeCount = topic.LikeCount
	tu.Data = topic.Data

	return tu
}

// For use in tests and for generating blank topics for forums which don't have a last poster
func getDummyTopic() *Topic {
	return &Topic{ID: 0, Title: ""}
}

func getTopicByReply(rid int) (*Topic, error) {
	topic := Topic{ID: 0}
	err := stmts.getTopicByReply.QueryRow(rid).Scan(&topic.ID, &topic.Title, &topic.Content, &topic.CreatedBy, &topic.CreatedAt, &topic.IsClosed, &topic.Sticky, &topic.ParentID, &topic.IPAddress, &topic.PostCount, &topic.LikeCount, &topic.Data)
	topic.Link = buildTopicURL(nameToSlug(topic.Title), topic.ID)
	return &topic, err
}

func buildTopicURL(slug string, tid int) string {
	if slug == "" {
		return "/topic/" + strconv.Itoa(tid)
	}
	return "/topic/" + slug + "." + strconv.Itoa(tid)
}

// I don't care if it isn't used,, it will likely be in the future. Nolint.
// nolint
func getTopicURLPrefix() string {
	return "/topic/"
}