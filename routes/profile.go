package routes

import (
	"database/sql"
	"net/http"
	"time"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

type ProfileStmts struct {
	getReplies *sql.Stmt
}

var profileStmts ProfileStmts

// TODO: Move these DbInits into some sort of abstraction
func init() {
	c.DbInits.Add(func(acc *qgen.Accumulator) error {
		profileStmts = ProfileStmts{
			getReplies: acc.SimpleLeftJoin("users_replies", "users", "users_replies.rid, users_replies.content, users_replies.createdBy, users_replies.createdAt, users_replies.lastEdit, users_replies.lastEditBy, users.avatar, users.name, users.group", "users_replies.createdBy = users.uid", "users_replies.uid = ?", "", ""),
		}
		return acc.FirstError()
	})
}

// TODO: Remove the View part of the name?
func ViewProfile(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	var err error
	var replyCreatedAt time.Time
	var replyContent, replyCreatedByName, replyAvatar string
	var rid, replyCreatedBy, replyLastEdit, replyLastEditBy, replyGroup int
	var replyList []*c.ReplyUser

	// TODO: Do a 301 if it's the wrong username? Do a canonical too?
	_, pid, err := ParseSEOURL(r.URL.Path[len("/user/"):])
	if err != nil {
		return c.SimpleError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, header)
	}

	// TODO: Preload this?
	header.AddSheet(header.Theme.Name + "/profile.css")
	if user.Loggedin {
		header.AddScriptAsync("profile_member.js")
	}

	var puser *c.User
	if pid == user.ID {
		user.IsMod = true
		puser = &user
	} else {
		// Fetch the user data
		// TODO: Add a shared function for checking for ErrNoRows and internal erroring if it's not that case?
		puser, err = c.Users.Get(pid)
		if err == sql.ErrNoRows {
			return c.NotFound(w, r, header)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
	}
	header.Title = phrases.GetTitlePhrasef("profile", puser.Name)
	header.Path = c.BuildProfileURL(c.NameToSlug(puser.Name), puser.ID)

	// Get the replies..
	rows, err := profileStmts.getReplies.Query(puser.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rid, &replyContent, &replyCreatedBy, &replyCreatedAt, &replyLastEdit, &replyLastEditBy, &replyAvatar, &replyCreatedByName, &replyGroup)
		if err != nil {
			return c.InternalError(err, w, r)
		}

		replyLiked := false
		replyLikeCount := 0
		ru := &c.ReplyUser{Reply: c.Reply{rid, puser.ID, replyContent, replyCreatedBy, replyGroup, replyCreatedAt, replyLastEdit, replyLastEditBy, 0, "", replyLiked, replyLikeCount, 0, ""}, ContentHtml: c.ParseMessage(replyContent, 0, "", user.ParseSettings), CreatedByName: replyCreatedByName, Avatar: replyAvatar, Level: 0}
		ru.Init()

		group, err := c.Groups.Get(ru.Group)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		if group.Tag != "" {
			ru.Tag = group.Tag
		} else if puser.ID == ru.CreatedBy {
			ru.Tag = phrases.GetTmplPhrase("profile_owner_tag")
		}

		// TODO: Add a hook here
		replyList = append(replyList, ru)
	}
	if err := rows.Err(); err != nil {
		return c.InternalError(err, w, r)
	}

	// Normalise the score so that the user sees their relative progress to the next level rather than showing them their total score
	prevScore := c.GetLevelScore(puser.Level)
	currentScore := puser.Score - prevScore
	nextScore := c.GetLevelScore(puser.Level+1) - prevScore
	var blocked, blockedInv bool
	if user.Loggedin {
		blocked, err = c.UserBlocks.IsBlockedBy(user.ID, puser.ID)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		blockedInv, err = c.UserBlocks.IsBlockedBy(puser.ID, user.ID)
		if err != nil {
			return c.InternalError(err, w, r)
		}
	}
	canMessage := (!blockedInv && user.Perms.UseConvos) || user.IsSuperMod
	canComment := !blockedInv && user.Perms.ViewTopic && user.Perms.CreateReply

	ppage := c.ProfilePage{header, replyList, *puser, currentScore, nextScore, blocked, canMessage, canComment}
	return renderTemplate("profile", w, r, header, ppage)
}
