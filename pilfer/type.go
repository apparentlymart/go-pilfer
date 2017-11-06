package pilfer

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
)

type takeType struct {
	Name    *types.TypeName
	Ident   *ast.Ident
	Spec    *ast.TypeSpec
	Type    types.Type
	NewName string // Assigned only when inserted into a typeTable
}

func (ty *takeType) IsNamed() bool {
	_, isNamed := ty.Type.(*types.Named)
	return isNamed
}

func (ty *takeType) Underlying() types.Type {
	tn, isNamed := ty.Type.(*types.Named)
	if !isNamed {
		return nil
	}
	return tn.Underlying()
}

type typeTable struct {
	types    map[types.Type]*takeType
	newNames map[string]*takeType
}

func newTypeTable() typeTable {
	return typeTable{
		types:    make(map[types.Type]*takeType),
		newNames: make(map[string]*takeType),
	}
}

func (t typeTable) Has(ty *takeType) bool {
	_, has := t.types[ty.Type]
	return has
}

func (t typeTable) NewNameTaken(newName string) bool {
	_, has := t.newNames[newName]
	return has
}

func (t typeTable) Add(ty *takeType) {
	newName := ty.Name.Name()
	if _, conflict := t.newNames[newName]; conflict {
		num := 1
		for {
			newName = fmt.Sprintf("%s_%d", ty.Name.Name(), num)
			if _, conflict := t.newNames[newName]; !conflict {
				break
			}
			num++
		}
	}

	ty.NewName = newName
	t.types[ty.Type] = ty
	t.newNames[newName] = ty
}

func (t typeTable) TypeByName(name *types.TypeName) *takeType {
	return t.types[name.Type()]
}

func (t typeTable) TypeByNewName(newName string) *takeType {
	return t.newNames[newName]
}

func (t typeTable) NewNames() []string {
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
