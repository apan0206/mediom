package controllers

import (
	"fmt"
	"strconv"
	"strings"
	//"fmt"
	. "github.com/huacnlee/mediom/app/models"
	"github.com/revel/revel"
)

type Topics struct {
	App
}

func (c Topics) Index(channel string) revel.Result {
	var page, nodeId int
	c.Params.Bind(&page, "page")
	c.Params.Bind(&nodeId, "node_id")
	node := Node{}
	if strings.EqualFold(channel, "node") {
		DB.Model(&Node{}).First(&node, nodeId)
		c.RenderArgs["node"] = node
	}
	topics, pageInfo := FindTopicPages(channel, nodeId, page, 20)
	pageInfo.Path = c.Request.URL.Path
	c.RenderArgs["channel"] = channel
	c.RenderArgs["topics"] = topics
	c.RenderArgs["page_info"] = pageInfo
	return c.Render()
}

func (c Topics) Feed() revel.Result {
	topics, _ := FindTopicPages("recent", 0, 1, 20)
	c.RenderArgs["topics"] = topics
	c.Response.ContentType = "text/xml"
	return c.Render()
}

func (c Topics) New() revel.Result {
	if r := c.requireUser(); r != nil {
		return r
	}
	t := &Topic{}
	c.RenderArgs["nodes"] = FindAllNodes()
	c.RenderArgs["topic"] = t
	return c.Render()
}

func (c Topics) Create() revel.Result {
	if r := c.requireUser(); r != nil {
		return r
	}
	var nodeId int32
	c.Params.Bind(&nodeId, "node_id")
	t := &Topic{
		Title:  c.Params.Get("title"),
		Body:   c.Params.Get("body"),
		NodeId: nodeId,
	}

	t.UserId = c.currentUser.Id
	v := CreateTopic(t)
	if v.HasErrors() {
		c.RenderArgs["topic"] = t
		c.RenderArgs["nodes"] = FindAllNodes()
		return c.renderValidation("topics/new.html", v)
	}
	return c.Redirect(fmt.Sprintf("/topics/%v", t.Id))
}

func (c Topics) Show() revel.Result {
	t := Topic{}
	DB.Preload("User").Preload("Node").First(&t, c.Params.Get("id"))
	replies := []Reply{}
	DB.Preload("User").Where("topic_id = ?", t.Id).Order("id asc").Find(&replies)
	c.RenderArgs["topic"] = t
	c.RenderArgs["replies"] = replies
	return c.Render()
}

func (c Topics) Edit() revel.Result {
	if r := c.requireUser(); r != nil {
		return r
	}
	t := &Topic{}
	DB.Where("id = ?", c.Params.Get("id")).First(t)
	if !c.isOwner(t) {
		c.Flash.Error("没有修改的权限")
		return c.Redirect("/")
	}
	c.RenderArgs["topic"] = t
	c.RenderArgs["nodes"] = FindAllNodes()
	return c.Render()
}

func (c Topics) Update() revel.Result {
	if r := c.requireUser(); r != nil {
		return r
	}
	t := Topic{}
	DB.First(&t, c.Params.Get("id"))
	if !c.isOwner(t) {
		c.Flash.Error("没有修改的权限")
		return c.Redirect("/")
	}
	nodeId, _ := strconv.Atoi(c.Params.Get("node_id"))
	t.NodeId = int32(nodeId)
	t.Title = c.Params.Get("title")
	t.Body = c.Params.Get("body")
	v := UpdateTopic(&t)
	if v.HasErrors() {
		c.RenderArgs["topic"] = t
		c.RenderArgs["nodes"] = FindAllNodes()
		return c.renderValidation("topics/edit.html", v)
	}
	return c.Redirect(fmt.Sprintf("/topics/%v", t.Id))
}

func (c Topics) Delete() revel.Result {
	if r := c.requireUser(); r != nil {
		return r
	}
	t := Topic{}
	DB.First(&t, c.Params.Get("id"))
	if !c.isOwner(t) {
		c.Flash.Error("没有修改的权限")
		return c.Redirect("/")
	}

	err := DB.Delete(&t).Error
	if err != nil {
		c.RenderError(err)
	}
	return c.Redirect("/topics")
}

func (c Topics) Watch() revel.Result {
	if r := c.requireUserForJSON(); r != nil {
		return r
	}
	t := Topic{}
	DB.First(&t, c.Params.Get("id"))
	c.currentUser.Watch(t)
	return c.successJSON(t.WatchesCount + 1)
}

func (c Topics) UnWatch() revel.Result {
	if r := c.requireUserForJSON(); r != nil {
		return r
	}
	t := Topic{}
	DB.First(&t, c.Params.Get("id"))
	c.currentUser.UnWatch(t)
	return c.successJSON(t.WatchesCount - 1)
}

func (c Topics) Star() revel.Result {
	if r := c.requireUserForJSON(); r != nil {
		return r
	}
	t := Topic{}
	DB.First(&t, c.Params.Get("id"))
	c.currentUser.Star(t)
	return c.successJSON(t.StarsCount + 1)
}

func (c Topics) UnStar() revel.Result {
	if r := c.requireUserForJSON(); r != nil {
		return r
	}
	t := Topic{}
	DB.First(&t, c.Params.Get("id"))
	c.currentUser.UnStar(t)
	return c.successJSON(t.StarsCount - 1)
}

func (c Topics) Rank() revel.Result {
	if r := c.requireAdmin(); r != nil {
		return r
	}

	rankVal := 0
	switch strings.ToLower(c.Params.Get("v")) {
	case "nopoint":
		rankVal = RankNoPoint
	case "awesome":
		rankVal = RankAwesome
	default:
		rankVal = RankNormal
	}

	t := Topic{}
	DB.First(&t, c.Params.Get("id"))
	err := t.UpdateRank(rankVal)
	if err != nil {
		return c.RenderError(err)
	}
	return c.Redirect(fmt.Sprintf("/topics/%v", t.Id))
}
