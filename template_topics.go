// +build !no_templategen

// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "net/http"
import "./common"
import "strconv"

// nolint
func init() {
	common.Template_topics_handle = Template_topics
	common.Ctemplates = append(common.Ctemplates,"topics")
	common.TmplPtrMap["topics"] = &common.Template_topics_handle
	common.TmplPtrMap["o_topics"] = Template_topics
}

// nolint
func Template_topics(tmpl_topics_vars common.TopicsPage, w http.ResponseWriter) error {
w.Write(header_0)
w.Write([]byte(tmpl_topics_vars.Title))
w.Write(header_1)
w.Write([]byte(tmpl_topics_vars.Header.Site.Name))
w.Write(header_2)
w.Write([]byte(tmpl_topics_vars.Header.Theme.Name))
w.Write(header_3)
if len(tmpl_topics_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_topics_vars.Header.Stylesheets {
w.Write(header_4)
w.Write([]byte(item))
w.Write(header_5)
}
}
w.Write(header_6)
if len(tmpl_topics_vars.Header.Scripts) != 0 {
for _, item := range tmpl_topics_vars.Header.Scripts {
w.Write(header_7)
w.Write([]byte(item))
w.Write(header_8)
}
}
w.Write(header_9)
w.Write([]byte(tmpl_topics_vars.CurrentUser.Session))
w.Write(header_10)
w.Write([]byte(tmpl_topics_vars.Header.Site.URL))
w.Write(header_11)
if tmpl_topics_vars.Header.MetaDesc != "" {
w.Write(header_12)
w.Write([]byte(tmpl_topics_vars.Header.MetaDesc))
w.Write(header_13)
}
w.Write(header_14)
if !tmpl_topics_vars.CurrentUser.IsSuperMod {
w.Write(header_15)
}
w.Write(header_16)
w.Write(menu_0)
w.Write(menu_1)
w.Write([]byte(tmpl_topics_vars.Header.Site.ShortName))
w.Write(menu_2)
if tmpl_topics_vars.CurrentUser.Loggedin {
w.Write(menu_3)
w.Write([]byte(tmpl_topics_vars.CurrentUser.Link))
w.Write(menu_4)
w.Write([]byte(tmpl_topics_vars.CurrentUser.Session))
w.Write(menu_5)
} else {
w.Write(menu_6)
}
w.Write(menu_7)
w.Write(header_17)
if tmpl_topics_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_18)
}
w.Write(header_19)
if len(tmpl_topics_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_topics_vars.Header.NoticeList {
w.Write(header_20)
w.Write([]byte(item))
w.Write(header_21)
}
}
w.Write(topics_0)
if tmpl_topics_vars.CurrentUser.ID != 0 {
w.Write(topics_1)
}
w.Write(topics_2)
if tmpl_topics_vars.CurrentUser.ID != 0 {
if len(tmpl_topics_vars.ForumList) != 0 {
w.Write(topics_3)
w.Write(topics_4)
} else {
w.Write(topics_5)
}
w.Write(topics_6)
}
w.Write(topics_7)
if tmpl_topics_vars.CurrentUser.ID != 0 {
w.Write(topics_8)
if len(tmpl_topics_vars.ForumList) != 0 {
w.Write(topics_9)
if len(tmpl_topics_vars.ForumList) != 0 {
for _, item := range tmpl_topics_vars.ForumList {
w.Write(topics_10)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topics_11)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topics_12)
w.Write([]byte(item.Name))
w.Write(topics_13)
}
}
w.Write(topics_14)
w.Write([]byte(tmpl_topics_vars.CurrentUser.Avatar))
w.Write(topics_15)
if len(tmpl_topics_vars.ForumList) != 0 {
for _, item := range tmpl_topics_vars.ForumList {
w.Write(topics_16)
if item.ID == tmpl_topics_vars.DefaultForum {
w.Write(topics_17)
}
w.Write(topics_18)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topics_19)
w.Write([]byte(item.Name))
w.Write(topics_20)
}
}
w.Write(topics_21)
if tmpl_topics_vars.CurrentUser.Perms.UploadFiles {
w.Write(topics_22)
}
w.Write(topics_23)
}
}
w.Write(topics_24)
if len(tmpl_topics_vars.TopicList) != 0 {
for _, item := range tmpl_topics_vars.TopicList {
w.Write(topics_25)
w.Write([]byte(strconv.Itoa(item.ID)))
w.Write(topics_26)
if item.Sticky {
w.Write(topics_27)
} else {
if item.IsClosed {
w.Write(topics_28)
}
}
w.Write(topics_29)
w.Write([]byte(item.Creator.Link))
w.Write(topics_30)
w.Write([]byte(item.Creator.Avatar))
w.Write(topics_31)
w.Write([]byte(item.Creator.Name))
w.Write(topics_32)
w.Write([]byte(item.Creator.Name))
w.Write(topics_33)
w.Write([]byte(item.Link))
w.Write(topics_34)
w.Write([]byte(item.Title))
w.Write(topics_35)
if item.ForumName != "" {
w.Write(topics_36)
w.Write([]byte(item.ForumLink))
w.Write(topics_37)
w.Write([]byte(item.ForumName))
w.Write(topics_38)
}
w.Write(topics_39)
w.Write([]byte(item.Creator.Link))
w.Write(topics_40)
w.Write([]byte(item.Creator.Name))
w.Write(topics_41)
if item.IsClosed {
w.Write(topics_42)
}
if item.Sticky {
w.Write(topics_43)
}
w.Write(topics_44)
w.Write([]byte(strconv.Itoa(item.PostCount)))
w.Write(topics_45)
w.Write([]byte(strconv.Itoa(item.LikeCount)))
w.Write(topics_46)
if item.Sticky {
w.Write(topics_47)
} else {
if item.IsClosed {
w.Write(topics_48)
}
}
w.Write(topics_49)
w.Write([]byte(item.LastUser.Link))
w.Write(topics_50)
w.Write([]byte(item.LastUser.Avatar))
w.Write(topics_51)
w.Write([]byte(item.LastUser.Name))
w.Write(topics_52)
w.Write([]byte(item.LastUser.Name))
w.Write(topics_53)
w.Write([]byte(item.LastUser.Link))
w.Write(topics_54)
w.Write([]byte(item.LastUser.Name))
w.Write(topics_55)
w.Write([]byte(item.RelativeLastReplyAt))
w.Write(topics_56)
}
} else {
w.Write(topics_57)
if tmpl_topics_vars.CurrentUser.Perms.CreateTopic {
w.Write(topics_58)
}
w.Write(topics_59)
}
w.Write(topics_60)
w.Write(footer_0)
w.Write([]byte(common.BuildWidget("footer",tmpl_topics_vars.Header)))
w.Write(footer_1)
if len(tmpl_topics_vars.Header.Themes) != 0 {
for _, item := range tmpl_topics_vars.Header.Themes {
if !item.HideFromThemes {
w.Write(footer_2)
w.Write([]byte(item.Name))
w.Write(footer_3)
if tmpl_topics_vars.Header.Theme.Name == item.Name {
w.Write(footer_4)
}
w.Write(footer_5)
w.Write([]byte(item.FriendlyName))
w.Write(footer_6)
}
}
}
w.Write(footer_7)
w.Write([]byte(common.BuildWidget("rightSidebar",tmpl_topics_vars.Header)))
w.Write(footer_8)
	return nil
}
