package guilds // import "github.com/Azareal/Gosora/extend/guilds/lib"

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/routes"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

var ListStmt *sql.Stmt
var MemberListStmt *sql.Stmt
var MemberListJoinStmt *sql.Stmt
var GetMemberStmt *sql.Stmt
var AttachForumStmt *sql.Stmt
var UnattachForumStmt *sql.Stmt
var AddMemberStmt *sql.Stmt

// Guild is a struct representing a guild
type Guild struct {
	ID      int
	Link    string
	Name    string
	Desc    string
	Active  bool
	Privacy int /* 0: Public, 1: Protected, 2: Private */

	// Who should be able to accept applications and create invites? Mods+ or just admins? Mods is a good start, we can ponder over whether we should make this more flexible in the future.
	Joinable int /* 0: Private, 1: Anyone can join, 2: Applications, 3: Invite-only */

	MemberCount    int
	Owner          int
	Backdrop       string
	CreatedAt      string
	LastUpdateTime string

	MainForumID int
	MainForum   *c.Forum
	Forums      []*c.Forum
	ExtData     c.ExtData
}

type Page struct {
	Title    string
	Header   *c.Header
	ItemList []*c.TopicsRow
	Forum    *c.Forum
	Guild    *Guild
	Page     int
	LastPage int
}

// ListPage is a page struct for constructing a list of every guild
type ListPage struct {
	Title     string
	Header    *c.Header
	GuildList []*Guild
}

type MemberListPage struct {
	Title    string
	Header   *c.Header
	ItemList []Member
	Guild    *Guild
	Page     int
	LastPage int
}

// Member is a struct representing a specific member of a guild, not to be confused with the global User struct.
type Member struct {
	Link       string
	Rank       int    /* 0: Member. 1: Mod. 2: Admin. */
	RankString string /* Member, Mod, Admin, Owner */
	PostCount  int
	JoinedAt   string
	Offline    bool // TODO: Need to track the online states of members when WebSockets are enabled

	User c.User
}

func PrebuildTmplList(user c.User, h *c.Header) c.CTmpl {
	guildList := []*Guild{
		&Guild{
			ID:             1,
			Name:           "lol",
			Link:           BuildGuildURL(c.NameToSlug("lol"), 1),
			Desc:           "A group for people who like to laugh",
			Active:         true,
			MemberCount:    1,
			Owner:          1,
			CreatedAt:      "date",
			LastUpdateTime: "date",
			MainForumID:    1,
			MainForum:      c.Forums.DirtyGet(1),
			Forums:         []*c.Forum{c.Forums.DirtyGet(1)},
		},
	}
	listPage := ListPage{"Guild List", user, h, guildList}
	return c.CTmpl{"guilds_guild_list", "guilds_guild_list.html", "templates/", "guilds.ListPage", listPage, []string{"./extend/guilds/lib"}}
}

// TODO: Do this properly via the widget system
// TODO: REWRITE THIS
func CommonAreaWidgets(header *c.Header) {
	// TODO: Hot Groups? Featured Groups? Official Groups?
	var b bytes.Buffer
	var menu = c.WidgetMenu{"Guilds", []c.WidgetMenuItem{
		c.WidgetMenuItem{"Create Guild", "/guild/create/", false},
	}}

	err := header.Theme.RunTmpl("widget_menu", pi, w)
	if err != nil {
		c.LogError(err)
		return
	}

	if header.Theme.HasDock("leftSidebar") {
		header.Widgets.LeftSidebar = template.HTML(string(b.Bytes()))
	} else if header.Theme.HasDock("rightSidebar") {
		header.Widgets.RightSidebar = template.HTML(string(b.Bytes()))
	}
}

// TODO: Do this properly via the widget system
// TODO: Make a better more customisable group widget system
func GuildWidgets(header *c.Header, guildItem *Guild) (success bool) {
	return false // Disabled until the next commit

	/*var b bytes.Buffer
	var menu WidgetMenu = WidgetMenu{"Guild Options", []WidgetMenuItem{
		WidgetMenuItem{"Join", "/guild/join/" + strconv.Itoa(guildItem.ID), false},
		WidgetMenuItem{"Members", "/guild/members/" + strconv.Itoa(guildItem.ID), false},
	}}

	err := templates.ExecuteTemplate(&b, "widget_menu.html", menu)
	if err != nil {
		c.LogError(err)
		return false
	}

	if themes[header.Theme.Name].Sidebars == "left" {
		header.Widgets.LeftSidebar = template.HTML(string(b.Bytes()))
	} else if themes[header.Theme.Name].Sidebars == "right" || themes[header.Theme.Name].Sidebars == "both" {
		header.Widgets.RightSidebar = template.HTML(string(b.Bytes()))
	} else {
		return false
	}
	return true*/
}

/*
	Custom Pages
*/

func RouteGuildList(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	header, ferr := c.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	CommonAreaWidgets(header)

	rows, err := ListStmt.Query()
	if err != nil && err != c.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	defer rows.Close()

	var guildList []*Guild
	for rows.Next() {
		guildItem := &Guild{ID: 0}
		err := rows.Scan(&guildItem.ID, &guildItem.Name, &guildItem.Desc, &guildItem.Active, &guildItem.Privacy, &guildItem.Joinable, &guildItem.Owner, &guildItem.MemberCount, &guildItem.CreatedAt, &guildItem.LastUpdateTime)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		guildItem.Link = BuildGuildURL(c.NameToSlug(guildItem.Name), guildItem.ID)
		guildList = append(guildList, guildItem)
	}
	err = rows.Err()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := ListPage{"Guild List", user, header, guildList}
	return routes.RenderTemplate("guilds_guild_list", w, r, header, pi)
}

func MiddleViewGuild(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, guildID, err := routes.ParseSEOURL(r.URL.Path[len("/guild/"):])
	if err != nil {
		return c.PreError("Not a valid guild ID", w, r)
	}

	guildItem, err := Gstore.Get(guildID)
	if err != nil {
		return c.LocalError("Bad guild", w, r, user)
	}
	// TODO: Build and pass header
	if !guildItem.Active {
		return c.NotFound(w, r, nil)
	}

	return nil

	// TODO: Re-implement this
	// Re-route the request to routeForums
	//var ctx = context.WithValue(r.Context(), "guilds_current_guild", guildItem)
	//return routeForum(w, r.WithContext(ctx), user, strconv.Itoa(guildItem.MainForumID))
}

func RouteCreateGuild(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	header, ferr := c.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = "Create Guild"
	// TODO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateGuild"] {
		return c.NoPermissions(w, r, user)
	}
	CommonAreaWidgets(header)

	return routes.RenderTemplate("guilds_create_guild", w, r, header, c.Page{header, tList, nil})
}

func RouteCreateGuildSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	// TODO: Add an approval queue mode for group creation
	if !user.Loggedin || !user.PluginPerms["CreateGuild"] {
		return c.NoPermissions(w, r, user)
	}

	var guildActive = true
	var guildName = c.SanitiseSingleLine(r.PostFormValue("group_name"))
	// TODO: Allow Markdown / BBCode / Limited HTML in the description?
	var guildDesc = c.SanitiseBody(r.PostFormValue("group_desc"))
	var gprivacy = r.PostFormValue("group_privacy")

	var guildPrivacy int
	switch gprivacy {
	case "0":
		guildPrivacy = 0 // Public
	case "1":
		guildPrivacy = 1 // Protected
	case "2":
		guildPrivacy = 2 // private
	default:
		guildPrivacy = 0
	}

	// Create the backing forum
	fid, err := c.Forums.Create(guildName, "", true, "")
	if err != nil {
		return c.InternalError(err, w, r)
	}

	gid, err := Gstore.Create(guildName, guildDesc, guildActive, guildPrivacy, user.ID, fid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// Add the main backing forum to the forum list
	err = AttachForum(gid, fid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	_, err = AddMemberStmt.Exec(gid, user.ID, 2)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, BuildGuildURL(c.NameToSlug(guildName), gid), http.StatusSeeOther)
	return nil
}

func RouteMemberList(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	header, ferr := c.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	_, guildID, err := routes.ParseSEOURL(r.URL.Path[len("/guild/members/"):])
	if err != nil {
		return c.PreError("Not a valid group ID", w, r)
	}

	guildItem, err := Gstore.Get(guildID)
	if err != nil {
		return c.LocalError("Bad group", w, r, user)
	}
	guildItem.Link = BuildGuildURL(c.NameToSlug(guildItem.Name), guildItem.ID)

	GuildWidgets(header, guildItem)

	rows, err := MemberListJoinStmt.Query(guildID)
	if err != nil && err != c.ErrNoRows {
		return c.InternalError(err, w, r)
	}

	var guildMembers []Member
	for rows.Next() {
		guildMember := Member{PostCount: 0}
		err := rows.Scan(&guildMember.User.ID, &guildMember.Rank, &guildMember.PostCount, &guildMember.JoinedAt, &guildMember.User.Name, &guildMember.User.RawAvatar)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		guildMember.Link = c.BuildProfileURL(c.NameToSlug(guildMember.User.Name), guildMember.User.ID)
		guildMember.User.Avatar, guildMember.User.MicroAvatar = c.BuildAvatar(guildMember.User.ID, guildMember.User.RawAvatar)
		guildMember.JoinedAt, _ = c.RelativeTimeFromString(guildMember.JoinedAt)
		if guildItem.Owner == guildMember.User.ID {
			guildMember.RankString = "Owner"
		} else {
			switch guildMember.Rank {
			case 0:
				guildMember.RankString = "Member"
			case 1:
				guildMember.RankString = "Mod"
			case 2:
				guildMember.RankString = "Admin"
			}
		}
		guildMembers = append(guildMembers, guildMember)
	}
	err = rows.Err()
	if err != nil {
		return c.InternalError(err, w, r)
	}
	rows.Close()

	pi := MemberListPage{"Guild Member List", user, header, guildMembers, guildItem, 0, 0}
	// A plugin with plugins. Pluginception!
	if c.RunPreRenderHook("pre_render_guilds_member_list", w, r, &user, &pi) {
		return nil
	}
	err = c.RunThemeTemplate(header.Theme.Name, "guilds_member_list", pi, w)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	return nil
}

func AttachForum(guildID int, fid int) error {
	_, err := AttachForumStmt.Exec(guildID, fid)
	return err
}

func UnattachForum(fid int) error {
	_, err := AttachForumStmt.Exec(fid)
	return err
}

func BuildGuildURL(slug string, id int) string {
	if slug == "" || !c.Config.BuildSlugs {
		return "/guild/" + strconv.Itoa(id)
	}
	return "/guild/" + slug + "." + strconv.Itoa(id)
}

/*
	Hooks
*/

// TODO: Prebuild this template
func PreRenderViewForum(w http.ResponseWriter, r *http.Request, user *c.User, data interface{}) (halt bool) {
	pi := data.(*c.ForumPage)
	if pi.Header.ExtData.Items != nil {
		if guildData, ok := pi.Header.ExtData.Items["guilds_current_group"]; ok {
			guildItem := guildData.(*Guild)

			guildpi := Page{pi.Title, pi.Header, pi.ItemList, pi.Forum, guildItem, pi.Page, pi.LastPage}
			err := routes.RenderTemplate("guilds_view_guild", w, r, header, guildpi)
			if err != nil {
				c.LogError(err)
				return false
			}
			return true
		}
	}
	return false
}

func TrowAssign(args ...interface{}) interface{} {
	var forum = args[1].(*c.Forum)
	if forum.ParentType == "guild" {
		var topicItem = args[0].(*c.TopicsRow)
		topicItem.ForumLink = "/guild/" + strings.TrimPrefix(topicItem.ForumLink, c.GetForumURLPrefix())
	}
	return nil
}

// TODO: It would be nice, if you could select one of the boards in the group from that drop-down rather than just the one you got linked from
func TopicCreatePreLoop(args ...interface{}) interface{} {
	var fid = args[2].(int)
	if c.Forums.DirtyGet(fid).ParentType == "guild" {
		var strictmode = args[5].(*bool)
		*strictmode = true
	}
	return nil
}

// TODO: Add privacy options
// TODO: Add support for multiple boards and add per-board simplified permissions
// TODO: Take js into account for routes which expect JSON responses
func ForumCheck(args ...interface{}) (skip bool, rerr c.RouteError) {
	var r = args[1].(*http.Request)
	var fid = args[3].(*int)
	var forum = c.Forums.DirtyGet(*fid)

	if forum.ParentType == "guild" {
		var err error
		w := args[0].(http.ResponseWriter)
		guildItem, ok := r.Context().Value("guilds_current_group").(*Guild)
		if !ok {
			guildItem, err = Gstore.Get(forum.ParentID)
			if err != nil {
				return true, c.InternalError(errors.New("Unable to find the parent group for a forum"), w, r)
			}
			if !guildItem.Active {
				return true, c.NotFound(w, r, nil) // TODO: Can we pull header out of args?
			}
			r = r.WithContext(context.WithValue(r.Context(), "guilds_current_group", guildItem))
		}

		user := args[2].(*c.User)
		var rank int
		var posts int
		var joinedAt string

		// TODO: Group privacy settings. For now, groups are all globally visible

		// Clear the default group permissions
		// TODO: Do this more efficiently, doing it quick and dirty for now to get this out quickly
		c.OverrideForumPerms(&user.Perms, false)
		user.Perms.ViewTopic = true

		err = GetMemberStmt.QueryRow(guildItem.ID, user.ID).Scan(&rank, &posts, &joinedAt)
		if err != nil && err != c.ErrNoRows {
			return true, c.InternalError(err, w, r)
		} else if err != nil {
			// TODO: Should we let admins / guests into public groups?
			return true, c.LocalError("You're not part of this group!", w, r, *user)
		}

		// TODO: Implement bans properly by adding the Local Ban API in the next commit
		// TODO: How does this even work? Refactor it along with the rest of this plugin!
		if rank < 0 {
			return true, c.LocalError("You've been banned from this group!", w, r, *user)
		}

		// Basic permissions for members, more complicated permissions coming in the next commit!
		if guildItem.Owner == user.ID {
			c.OverrideForumPerms(&user.Perms, true)
		} else if rank == 0 {
			user.Perms.LikeItem = true
			user.Perms.CreateTopic = true
			user.Perms.CreateReply = true
		} else {
			c.OverrideForumPerms(&user.Perms, true)
		}
		return true, nil
	}

	return false, nil
}

// TODO: Override redirects? I don't think this is needed quite yet

func Widgets(args ...interface{}) interface{} {
	zone := args[0].(string)
	header := args[2].(*c.Header)
	request := args[3].(*http.Request)
	if zone != "view_forum" {
		return false
	}

	forum := args[1].(*c.Forum)
	if forum.ParentType == "guild" {
		// This is why I hate using contexts, all the daisy chains and interface casts x.x
		guildItem, ok := request.Context().Value("guilds_current_group").(*Guild)
		if !ok {
			c.LogError(errors.New("Unable to find a parent group in the context data"))
			return false
		}

		if header.ExtData.Items == nil {
			header.ExtData.Items = make(map[string]interface{})
		}
		header.ExtData.Items["guilds_current_group"] = guildItem

		return GuildWidgets(header, guildItem)
	}
	return false
}
