package controllers

import (
	. "github.com/huacnlee/mediom/app/models"
	"github.com/revel/revel"
	"strconv"
)

type Nodes struct {
	App
}

func init() {
	revel.InterceptMethod((*Nodes).Before, revel.BEFORE)
}

func (c *Nodes) Before() revel.Result {
	c.requireAdmin()

	return nil
}

func (c *Nodes) loadNodeGroups() {
	groups := []NodeGroup{}
	DB.Order("sort desc").Find(&groups)
	c.RenderArgs["groups"] = groups
}

func (c *Nodes) Index() revel.Result {
	c.loadNodeGroups()
	nodes := FindAllNodes()
	c.RenderArgs["nodes"] = nodes
	return c.Render()
}

func (c *Nodes) Create() revel.Result {
	nodeGroupId, _ := strconv.Atoi(c.Params.Get("node_group_id"))
	n := Node{
		Name:        c.Params.Get("name"),
		NodeGroupId: nodeGroupId,
	}

	v := CreateNode(&n)
	if v.HasErrors() {
		c.loadNodeGroups()
		c.RenderArgs["node"] = n
		return c.renderValidation("nodes/index.html", v)
	}
	c.Flash.Success("节点创建成功")
	return c.Redirect("/nodes")
}

func (c Nodes) Edit() revel.Result {
	c.loadNodeGroups()

	node := Node{}
	err := DB.First(&node, c.Params.Get("id")).Error
	if err != nil {
		return c.RenderError(err)
	}

	c.RenderArgs["node"] = node
	return c.Render()
}

func (c Nodes) Update() revel.Result {
	node := Node{}
	err := DB.First(&node, c.Params.Get("id")).Error
	if err != nil {
		return c.RenderError(err)
	}
	node.Name = c.Params.Get("name")
	node.Summary = c.Params.Get("summary")
	node.NodeGroupId, _ = strconv.Atoi(c.Params.Get("node_group_id"))
	v := UpdateNode(&node)

	c.RenderArgs["node"] = node
	if v.HasErrors() {
		c.loadNodeGroups()
		return c.renderValidation("nodes/edit.html", v)
	}
	c.Flash.Success("节点更新成功")
	return c.Redirect("/nodes")
}

func (c Nodes) Delete() revel.Result {
	node := Node{}
	err := DB.First(&node, c.Params.Get("id")).Error
	if err != nil {
		return c.RenderError(err)
	}

	DB.Delete(&node)
	return c.Redirect("/nodes")
}
