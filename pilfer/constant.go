package pilfer

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"sort"
)

type takeConstant struct {
	Name    *ast.Ident
	Type    *takeType
	Const   *types.Const
	Value   constant.Value
	NewName string // Assigned only when inserted into a constantTable
}

type constantTable struct {
	types    typeTable
	consts   map[*types.Const]*takeConstant
	newNames map[string]*takeConstant
}

func newConstantTable(tys typeTable) constantTable {
	return constantTable{
		types:    tys,
		consts:   make(map[*types.Const]*takeConstant),
		newNames: make(map[string]*takeConstant),
	}
}

func (t constantTable) Has(cn *takeConstant) bool {
	_, has := t.consts[cn.Const]
	return has
}

func (t constantTable) NewNameTaken(newName string) bool {
	_, has := t.newNames[newName]
	if has {
		return true
	}
	return t.types.NewNameTaken(newName)
}

func (t constantTable) Add(cn *takeConstant) {
	newName := cn.Name.Name
	if t.NewNameTaken(newName) {
		num := 1
		for {
			newName = fmt.Sprintf("%s_%d", cn.Name.Name, num)
			if t.NewNameTaken(newName) {
				break
			}
			num++
		}
	}

	cn.NewName = newName
	t.consts[cn.Const] = cn
	t.newNames[newName] = cn
}

func (t constantTable) ConstantByObj(obj *types.Const) *takeConstant {
	return t.consts[obj]
}

func (t constantTable) ConstantByNewName(newName string) *takeConstant {
	return t.newNames[newName]
}

func (t constantTable) NewNames() []string {
	if len(t.newNames) == 0 {
		return nil
	}
	names := make([]string, 0, len(t.newNames))
	for newName := range t.newNames {
		names = append(names, newName)
	}
	sort.Strings(names)
	return names
}

func (t constantTable) NewNamesByTypeName() map[string][]string {
	ret := make(map[string][]string)
	if len(t.newNames) == 0 {
		return ret
	}

	for newName, cn := range t.newNames {
		ty := cn.Type
		ret[ty.NewName] = append(ret[ty.NewName], newName)
	}

	for n := range ret {
		sort.Strings(ret[n])
	}

	return ret
}
