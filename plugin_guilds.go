package main

import (
	//"fmt"

	"./extend/guilds/lib"
	"./query_gen/lib"
)

// TODO: Add a better way of splitting up giant plugins like this

// TODO: Add a plugin interface instead of having a bunch of argument to AddPlugin?
func init() {
	plugins["guilds"] = NewPlugin("guilds", "Guilds", "Azareal", "http://github.com/Azareal", "", "", "", initGuilds, nil, deactivateGuilds, installGuilds, nil)

	// TODO: Is it possible to avoid doing this when the plugin isn't activated?
	prebuildTmplList = append(prebuildTmplList, guilds.PrebuildTmplList)
}

func initGuilds() (err error) {
	plugins["guilds"].AddHook("intercept_build_widgets", guilds.Widgets)
	plugins["guilds"].AddHook("trow_assign", guilds.TrowAssign)
	plugins["guilds"].AddHook("topic_create_pre_loop", guilds.TopicCreatePreLoop)
	plugins["guilds"].AddHook("pre_render_view_forum", guilds.PreRenderViewForum)
	plugins["guilds"].AddHook("simple_forum_check_pre_perms", guilds.ForumCheck)
	plugins["guilds"].AddHook("forum_check_pre_perms", guilds.ForumCheck)
	// TODO: Auto-grant this perm to admins upon installation?
	registerPluginPerm("CreateGuild")
	router.HandleFunc("/guilds/", guilds.GuildList)
	router.HandleFunc("/guild/", guilds.ViewGuild)
	router.HandleFunc("/guild/create/", guilds.CreateGuild)
	router.HandleFunc("/guild/create/submit/", guilds.CreateGuildSubmit)
	router.HandleFunc("/guild/members/", guilds.MemberList)

	guilds.ListStmt, err = qgen.Builder.SimpleSelect("guilds", "guildID, name, desc, active, privacy, joinable, owner, memberCount, createdAt, lastUpdateTime", "", "", "")
	if err != nil {
		return err
	}
	guilds.GetGuildStmt, err = qgen.Builder.SimpleSelect("guilds", "name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime", "guildID = ?", "", "")
	if err != nil {
		return err
	}
	guilds.MemberListStmt, err = qgen.Builder.SimpleSelect("guilds_members", "guildID, uid, rank, posts, joinedAt", "", "", "")
	if err != nil {
		return err
	}
	guilds.MemberListJoinStmt, err = qgen.Builder.SimpleLeftJoin("guilds_members", "users", "users.uid, guilds_members.rank, guilds_members.posts, guilds_members.joinedAt, users.name, users.avatar", "guilds_members.uid = users.uid", "guilds_members.guildID = ?", "guilds_members.rank DESC, guilds_members.joinedat ASC", "")
	if err != nil {
		return err
	}
	guilds.GetMemberStmt, err = qgen.Builder.SimpleSelect("guilds_members", "rank, posts, joinedAt", "guildID = ? AND uid = ?", "", "")
	if err != nil {
		return err
	}
	guilds.CreateGuildStmt, err = qgen.Builder.SimpleInsert("guilds", "name, desc, active, privacy, joinable, owner, memberCount, mainForum, backdrop, createdAt, lastUpdateTime", "?,?,?,?,1,?,1,?,'',UTC_TIMESTAMP(),UTC_TIMESTAMP()")
	if err != nil {
		return err
	}
	guilds.AttachForumStmt, err = qgen.Builder.SimpleUpdate("forums", "parentID = ?, parentType = 'guild'", "fid = ?")
	if err != nil {
		return err
	}
	guilds.UnattachForumStmt, err = qgen.Builder.SimpleUpdate("forums", "parentID = 0, parentType = ''", "fid = ?")
	if err != nil {
		return err
	}
	guilds.AddMemberStmt, err = qgen.Builder.SimpleInsert("guilds_members", "guildID, uid, rank, posts, joinedAt", "?,?,?,0,UTC_TIMESTAMP()")
	if err != nil {
		return err
	}

	return nil
}

func deactivateGuilds() {
	plugins["guilds"].RemoveHook("intercept_build_widgets", guilds.Widgets)
	plugins["guilds"].RemoveHook("trow_assign", guilds.TrowAssign)
	plugins["guilds"].RemoveHook("topic_create_pre_loop", guilds.TopicCreatePreLoop)
	plugins["guilds"].RemoveHook("pre_render_view_forum", guilds.PreRenderViewForum)
	plugins["guilds"].RemoveHook("simple_forum_check_pre_perms", guilds.ForumCheck)
	plugins["guilds"].RemoveHook("forum_check_pre_perms", guilds.ForumCheck)
	deregisterPluginPerm("CreateGuild")
	_ = router.RemoveFunc("/guilds/")
	_ = router.RemoveFunc("/guild/")
	_ = router.RemoveFunc("/guild/create/")
	_ = router.RemoveFunc("/guild/create/submit/")
	_ = guilds.ListStmt.Close()
	_ = guilds.MemberListStmt.Close()
	_ = guilds.MemberListJoinStmt.Close()
	_ = guilds.GetMemberStmt.Close()
	_ = guilds.GetGuildStmt.Close()
	_ = guilds.CreateGuildStmt.Close()
	_ = guilds.AttachForumStmt.Close()
	_ = guilds.UnattachForumStmt.Close()
	_ = guilds.AddMemberStmt.Close()
}

// TODO: Stop accessing the query builder directly and add a feature in Gosora which is more easily reversed, if an error comes up during the installation process
func installGuilds() error {
	guildTableStmt, err := qgen.Builder.CreateTable("guilds", "utf8mb4", "utf8mb4_general_ci",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"guildID", "int", 0, false, true, ""},
			qgen.DB_Table_Column{"name", "varchar", 100, false, false, ""},
			qgen.DB_Table_Column{"desc", "varchar", 200, false, false, ""},
			qgen.DB_Table_Column{"active", "boolean", 1, false, false, ""},
			qgen.DB_Table_Column{"privacy", "smallint", 0, false, false, ""},
			qgen.DB_Table_Column{"joinable", "smallint", 0, false, false, "0"},
			qgen.DB_Table_Column{"owner", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"memberCount", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"mainForum", "int", 0, false, false, "0"}, // The board the user lands on when they click on a group, we'll make it possible for group admins to change what users land on
			//qgen.DB_Table_Column{"boards","varchar",255,false,false,""}, // Cap the max number of boards at 8 to avoid overflowing the confines of a 64-bit integer?
			qgen.DB_Table_Column{"backdrop", "varchar", 200, false, false, ""}, // File extension for the uploaded file, or an external link
			qgen.DB_Table_Column{"createdAt", "createdAt", 0, false, false, ""},
			qgen.DB_Table_Column{"lastUpdateTime", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{
			qgen.DB_Table_Key{"guildID", "primary"},
		},
	)
	if err != nil {
		return err
	}

	_, err = guildTableStmt.Exec()
	if err != nil {
		return err
	}

	guildMembersTableStmt, err := qgen.Builder.CreateTable("guilds_members", "", "",
		[]qgen.DB_Table_Column{
			qgen.DB_Table_Column{"guildID", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"uid", "int", 0, false, false, ""},
			qgen.DB_Table_Column{"rank", "int", 0, false, false, "0"},  /* 0: Member. 1: Mod. 2: Admin. */
			qgen.DB_Table_Column{"posts", "int", 0, false, false, "0"}, /* Per-Group post count. Should we do some sort of score system? */
			qgen.DB_Table_Column{"joinedAt", "datetime", 0, false, false, ""},
		},
		[]qgen.DB_Table_Key{},
	)
	if err != nil {
		return err
	}

	_, err = guildMembersTableStmt.Exec()
	return err
}

// TO-DO; Implement an uninstallation system into Gosora. And a better installation system.
func uninstallGuilds() error {
	return nil
}