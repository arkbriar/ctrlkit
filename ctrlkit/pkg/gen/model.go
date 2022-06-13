package gen

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type GvBind struct {
	Gv     string              `json:"gv"`
	Parsed schema.GroupVersion `json:"-"`
	Pkg    string              `json:"pkg"`
}

type GvReflections struct {
	GvPkgBinds map[string]GvBind `json:"binds"`
	GvkAliases map[string]string `json:"aliases"`
}

func (r *GvReflections) AddGvBind(gv string, pkg string, parsed schema.GroupVersion) bool {
	if _, ok := r.GvPkgBinds[gv]; ok {
		return false
	}
	r.GvPkgBinds[gv] = GvBind{
		Gv:     gv,
		Pkg:    pkg,
		Parsed: parsed,
	}
	return true
}

func (r *GvReflections) GetGvPkg(gv string) string {
	return r.GvPkgBinds[gv].Pkg
}

func (r *GvReflections) IsGvBound(gv string) bool {
	return r.GetGvPkg(gv) != ""
}

func (r *GvReflections) AddGvkAliases(gvk string, alias string) bool {
	if _, ok := r.GvkAliases[alias]; ok {
		return false
	}
	r.GvkAliases[alias] = gvk
	return true
}

func (r *GvReflections) GetGvkByAlias(alias string) string {
	return r.GvkAliases[alias]
}

func (r *GvReflections) DoesAliasExists(alias string) bool {
	return r.GetGvkByAlias(alias) != ""
}

type ControllerManagerDocument struct {
	FileName      string
	GvReflections `json:",inline"`
	Decls         map[string]ControllerManagerDeclaration `json:"decls"`
}

func (d *ControllerManagerDocument) DoesControllerManagerDeclarationExists(name string) bool {
	_, ok := d.Decls[name]
	return ok
}

type StateDeclaration struct {
	Comments  []string          `json:"comments"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	IsArray   bool              `json:"is_array"`
	Selectors map[string]string `json:"selectors"`
}

func (d *StateDeclaration) AddSelector(key, value string) bool {
	if _, ok := d.Selectors[key]; ok {
		return false
	}
	d.Selectors[key] = value
	return true
}

type ActionDeclaration struct {
	Comments []string `json:"comments"`
	Name     string   `json:"name"`
	Params   []string `json:"params"`
}

type ControllerManagerDeclaration struct {
	Comments   []string                     `json:"comments"`
	Name       string                       `json:"name"`
	TargetType string                       `json:"target_type"`
	States     map[string]StateDeclaration  `json:"states"`
	Actions    []ActionDeclaration          `json:"actions"`
	ActionMap  map[string]ActionDeclaration `json:"-"`
}

func (d *ControllerManagerDeclaration) AddStateDeclaration(s StateDeclaration) bool {
	if _, ok := d.States[s.Name]; ok {
		return false
	}
	d.States[s.Name] = s
	return true
}

func (d *ControllerManagerDeclaration) ContainsState(name string) bool {
	_, ok := d.States[name]
	return ok
}

func (d *ControllerManagerDeclaration) AddActionDeclaration(act ActionDeclaration) bool {
	if _, ok := d.ActionMap[act.Name]; ok {
		return false
	}
	d.ActionMap[act.Name] = act
	d.Actions = append(d.Actions, act)
	return true
}
