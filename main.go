package main

import (
	"slices"
)

type ChiTieu struct {
	Name    string
	PhanTos []*PhanTo

	// only in bieu mau
	IDbm int
}

type PhanTo struct {
	Name   string // uniq in ChiTieu
	Values []string

	// only in bieu mau
	ChiTieuIDbm int
	IDbm        int
	Children    []*PhanTo // only for phan to chung
}

type KV struct {
	Key   string
	Value string
}

type BangChiTieu struct {
	ChiTieuName string
	DongDuLieus []*DongDuLieu
}

type DongDuLieu struct {
	Dims   []*KV
	Solieu string // for easy
}

type CellIndex struct {
	Ci int
	Ri int
}

type Node struct {
	Children []*Node

	Value string
	Ci    int
	Ri    int

	IDbms []int  // chi tieu and phan to
	Type  string // chitieu, phanto, phanto_value
}

type BieuMau struct {
	ChiTieus     []*ChiTieu
	BangChiTieus []*BangChiTieu

	PhanToChungs []*PhanTo

	chiTieum map[int]*ChiTieu // derived
	phanTom  map[int]*PhanTo  // derived

	Cols    []int // phan to and chi tieu idbm
	ColTree *Node

	Rows    []int // phan to and chi tieu idbm
	RowTree *Node

	Content map[CellIndex]*Cell
}

type Cell struct {
	Value string
	Ci    int
	Ri    int

	Node *Node // optional
}

// require call if change ChiTieus, PhanTos, PhanToChungs, reinitID()
func (bm *BieuMau) derived() {
	bm.chiTieum = make(map[int]*ChiTieu)
	bm.phanTom = make(map[int]*PhanTo)

	for _, chiTieu := range bm.ChiTieus {
		bm.chiTieum[chiTieu.IDbm] = chiTieu
		for _, phanTo := range chiTieu.PhanTos {
			bm.phanTom[phanTo.IDbm] = phanTo
		}
	}

	for _, phanTo := range bm.PhanToChungs {
		bm.phanTom[phanTo.IDbm] = phanTo
	}

	// reinit ID
	var maxID int
	for _, chiTieu := range bm.ChiTieus {
		if chiTieu.IDbm > maxID {
			maxID = chiTieu.IDbm
		}
		for _, phanTo := range chiTieu.PhanTos {
			if phanTo.IDbm > maxID {
				maxID = phanTo.IDbm
			}
		}
	}

	for _, chiTieu := range bm.ChiTieus {
		if chiTieu.IDbm == 0 {
			maxID++
			chiTieu.IDbm = maxID
		}
		for _, phanTo := range chiTieu.PhanTos {
			if phanTo.IDbm == 0 {
				maxID++
				phanTo.IDbm = maxID
			}
			phanTo.ChiTieuIDbm = chiTieu.IDbm
		}
	}

	// TODO merge or join?, current use merge
	for _, phanTo := range bm.PhanToChungs {
		maxID++
		phanTo.IDbm = maxID

		phanTo.Values = []string{}
		for _, child := range phanTo.Children {
			phanTo.Values = append(phanTo.Values, child.Values...)
		}
		phanTo.Values = uniqArrs(phanTo.Values)
	}
}

func uniqArrs(arrs ...[]string) []string {
	idm := make(map[string]bool)
	for _, arr := range arrs {
		for _, idbm := range arr {
			idm[idbm] = true
		}
	}

	uniqarr := make([]string, 0)
	for k := range idm {
		uniqarr = append(uniqarr, k)
	}

	slices.Sort(uniqarr)

	return uniqarr
}

func (bm *BieuMau) genPhanToChung() {
	phanToChungm := make(map[string][]*PhanTo)
	for _, chitieu := range bm.ChiTieus {
		for _, phanto := range chitieu.PhanTos {
			phanToChungm[phanto.Name] = append(phanToChungm[phanto.Name], phanto)
		}
	}

	bm.PhanToChungs = make([]*PhanTo, 0)
	for k, v := range phanToChungm {
		if len(v) == len(bm.ChiTieus) {
			phanTo := &PhanTo{
				Name:     k,
				Children: v,
			}
			bm.PhanToChungs = append(bm.PhanToChungs, phanTo)
		}
	}
}

func (bm *BieuMau) setupBase() {
	bm.genPhanToChung()
	bm.derived()
}

func (bm *BieuMau) genHeaders(tableType string) {
	switch tableType {
	case "flat":
		var cols []int
		var rows []int
		idm := make(map[int]bool)
		for _, phanTo := range bm.PhanToChungs {
			cols = append(cols, phanTo.IDbm)
			idm[phanTo.IDbm] = true

			for _, child := range phanTo.Children {
				idm[child.IDbm] = true
			}
		}

		for _, chitieu := range bm.ChiTieus {
			idm[chitieu.IDbm] = true
			cols = append(cols, chitieu.IDbm)

			for _, phanTo := range chitieu.PhanTos {
				if !idm[phanTo.IDbm] {
					idm[phanTo.IDbm] = true
					cols = append(cols, phanTo.IDbm)
				}
			}
		}

		bm.Cols = cols
		bm.Rows = rows
	case "matrix_chitieu_in_rows":
		var cols []int
		var rows []int
		idm := make(map[int]bool)
		for _, phanTo := range bm.PhanToChungs {
			cols = append(cols, phanTo.IDbm)
			idm[phanTo.IDbm] = true

			for _, child := range phanTo.Children {
				idm[child.IDbm] = true
			}
		}

		for _, chitieu := range bm.ChiTieus {
			idm[chitieu.IDbm] = true
			rows = append(rows, chitieu.IDbm)

			for _, phanTo := range chitieu.PhanTos {
				if !idm[phanTo.IDbm] {
					idm[phanTo.IDbm] = true
					rows = append(rows, phanTo.IDbm)
				}
			}
		}

		bm.Cols = cols
		bm.Rows = rows
	case "matrix_chitieu_in_cols":
		var cols []int
		var rows []int
		idm := make(map[int]bool)
		for _, phanTo := range bm.PhanToChungs {
			rows = append(rows, phanTo.IDbm)
			idm[phanTo.IDbm] = true

			for _, child := range phanTo.Children {
				idm[child.IDbm] = true
			}
		}

		for _, chitieu := range bm.ChiTieus {
			idm[chitieu.IDbm] = true
			cols = append(cols, chitieu.IDbm)

			for _, phanTo := range chitieu.PhanTos {
				if !idm[phanTo.IDbm] {
					idm[phanTo.IDbm] = true
					cols = append(cols, phanTo.IDbm)
				}
			}
		}

		bm.Cols = cols
		bm.Rows = rows
	}
}

// TODO don't use recursive
func (bm *BieuMau) genTree(node *Node, idbms []int) {
	phanToChungIdbms := make([]int, 0)
	for _, idbm := range idbms {
		phanTo, has := bm.phanTom[idbm]
		if !has || len(phanTo.Children) == 0 {
			continue
		}
		phanToChungIdbms = append(phanToChungIdbms, idbm)
	}

	slices.Sort(phanToChungIdbms)
	bm.recursiveNode(node, phanToChungIdbms)

	if len(phanToChungIdbms) >= len(idbms) {
		return // not include chi tieu
	}

	chiTieuIdbms := make([]int, 0)
	for _, idbm := range idbms {
		if _, has := bm.phanTom[idbm]; !has {
			chiTieuIdbms = append(chiTieuIdbms, idbm)
		}
	}

	if len(chiTieuIdbms) == 0 {
		// TODO should handle
		return // ignore all remain idbms
	}
	slices.Sort(chiTieuIdbms)

	remainIdbmm := make(map[int][]int) // include phan to rieng (non-child), chi tieu
	for _, idbm := range chiTieuIdbms {
		remainIdbmm[idbm] = []int{}
	}
	for _, idbm := range idbms {
		phanTo, has := bm.phanTom[idbm]
		if !has || len(phanTo.Children) > 0 {
			continue
		}
		remainIdbmm[phanTo.ChiTieuIDbm] = append(remainIdbmm[phanTo.ChiTieuIDbm], idbm)
	}

	for idbm := range remainIdbmm {
		if len(remainIdbmm[idbm]) == 0 {
			continue
		}
		slices.Sort(remainIdbmm[idbm])
	}

	chitieuNodem := make(map[int]*Node) // branch node
	for _, idbm := range chiTieuIdbms {
		chiTieuNode := &Node{
			IDbms: []int{idbm},
			Type:  "chitieu",
		}
		bm.recursiveNode(chiTieuNode, remainIdbmm[idbm])
		chitieuNodem[idbm] = chiTieuNode
	}

	if len(phanToChungIdbms) == 0 {
		return
	}

	var leafNodes []*Node
	nodearr := []*Node{node}
	for len(nodearr) > 0 {
		nextNodearr := make([]*Node, 0)
		for _, n := range nodearr {
			if len(n.Children) == 0 {
				leafNodes = append(leafNodes, n)
			}

			nextNodearr = append(nextNodearr, n.Children...)
		}
		nodearr = nextNodearr
	}

	for _, lNode := range leafNodes {
		for _, idbm := range chiTieuIdbms {
			if n, has := chitieuNodem[idbm]; has {
				lNode.Children = append(lNode.Children, copyNode(n))
			}
		}
	}
}

// TODO don't use recursive
func copyNode(node *Node) *Node {
	clone := &Node{
		IDbms:    make([]int, len(node.IDbms)),
		Value:    node.Value,
		Type:     node.Type,
		Children: make([]*Node, len(node.Children)),
	}
	copy(clone.IDbms, node.IDbms)
	for i := range clone.Children {
		clone.Children[i] = copyNode(node.Children[i])
	}
	return clone
}

// only phan to, phan to value
func (bm *BieuMau) recursiveNode(node *Node, idbms []int) {
	if len(idbms) == 0 {
		return
	}
	for _, idbm := range idbms {
		phanTo, has := bm.phanTom[idbm]
		if !has {
			continue // ignore error
		}

		phanToNode := &Node{
			IDbms: []int{idbm},
			Type:  "phanto",
		}

		for _, value := range phanTo.Values {
			valueNode := &Node{
				Value: value,
				IDbms: []int{idbm},
				Type:  "phanto_value",
			}
			bm.recursiveNode(valueNode, idbms[1:])
			phanToNode.Children = append(phanToNode.Children, valueNode)
		}

		node.Children = append(node.Children, phanToNode)
	}
}

func (bm *BieuMau) genContent() {
	bm.Content = make(map[CellIndex]*Cell)

	stack := []*Node{bm.ColTree}

	for len(stack) > 0 {
		n := len(stack) - 1
		current := stack[n]
		stack = stack[:n]

		arr := make([]*Node, len(current.Children))
		copy(arr, current.Children)
		slices.Reverse(arr)
		stack = append(stack, arr...)

		cell := &Cell{}
		switch current.Type {
		case "chitieu":
			ct := bm.chiTieum[current.IDbms[0]]
			cell.Value = ct.Name
			cell.Node = current
		case "phanto":
			pt := bm.phanTom[current.IDbms[0]]
			cell.Value = pt.Name
			cell.Node = current
		case "phanto_value":
			cell.Value = current.Value
			cell.Node = current
		default:
			continue
		}

		// TODO add cell to bm.Content
	}

	stack = []*Node{bm.RowTree}

	for len(stack) > 0 {
		n := len(stack) - 1
		current := stack[n]
		stack = stack[:n]

		arr := make([]*Node, len(current.Children))
		copy(arr, current.Children)
		slices.Reverse(arr)
		stack = append(stack, arr...)

		cell := &Cell{}
		switch current.Type {
		case "chitieu":
			ct := bm.chiTieum[current.IDbms[0]]
			cell.Value = ct.Name
			cell.Node = current
		case "phanto":
			pt := bm.phanTom[current.IDbms[0]]
			cell.Value = pt.Name
			cell.Node = current
		case "phanto_value":
			cell.Value = current.Value
			cell.Node = current
		default:
			continue
		}

		// TODO add cell to bm.Content
	}
}

func (bm *BieuMau) replaceContent(matrix [][]string) {
	for i, row := range matrix {
		for j, value := range row {
			bm.Content[CellIndex{Ri: i, Ci: j}].Value = value
		}
	}

	// TODO update column tree and row tree
}

func (bm *BieuMau) setupFullFlattable() {
	bm.setupBase()
	bm.genHeaders("flat")
	bm.ColTree = &Node{}
	bm.genTree(bm.ColTree, bm.Cols)
	bm.RowTree = &Node{}
	bm.genTree(bm.RowTree, bm.Rows)
	bm.genContent()
}

func main() {

}
