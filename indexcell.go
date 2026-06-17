package main

import (
	"errors"
	"slices"
)

type ChiTieu struct {
	Name    string
	PhanTos []*PhanTo

	// only in bieu mau
	IDbm int
}

type PhanTo struct {
	Name         string // uniq in ChiTieu
	Values       []string
	StrictValues bool // if true, unknown values during import are treated as free rows/cols

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

	IDbm int    // chi tieu and phan to idbm
	Type string // chitieu, phanto, phanto_value, or "" for root
}

type BieuMau struct {
	ChiTieus     []*ChiTieu
	BangChiTieus []*BangChiTieu

	PhanToChungs []*PhanTo

	chiTieum map[int]*ChiTieu // derived
	phanTom  map[int]*PhanTo  // derived

	HeaderType string // flat, matrix_chitieu_in_rows, matrix_chitieu_in_cols

	Cols        []int // phan to and chi tieu idbm
	ColTree     *Node
	ColLeafs    []int // derived from ColTree, "phan to and chi tieu idbm" are only in leaf nodes
	ColCollapse bool

	Rows        []int // phan to and chi tieu idbm
	RowTree     *Node
	RowLeafs    []int // derived from RowTree, "phan to and chi tieu idbm" are only in leaf nodes
	RowRollapse bool

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
	// 1. Find maxID
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
	for _, phanTo := range bm.PhanToChungs {
		if phanTo.IDbm > maxID {
			maxID = phanTo.IDbm
		}
	}

	// 2. Assign IDs if they are 0
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

	for _, phanTo := range bm.PhanToChungs {
		if phanTo.IDbm == 0 {
			maxID++
			phanTo.IDbm = maxID
		}
		phanTo.Values = []string{}
		for _, child := range phanTo.Children {
			phanTo.Values = append(phanTo.Values, child.Values...)
		}
		phanTo.Values = uniqArrs(phanTo.Values)
	}

	// 3. Populate maps
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

func (bm *BieuMau) genHeaders() {
	switch bm.HeaderType {
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
			IDbm:  idbm,
			Type:  "chitieu",
			Value: bm.chiTieum[idbm].Name,
		}
		bm.recursiveNode(chiTieuNode, remainIdbmm[idbm])
		chitieuNodem[idbm] = chiTieuNode
	}

	if len(phanToChungIdbms) == 0 {
		for _, idbm := range chiTieuIdbms {
			if n, has := chitieuNodem[idbm]; has {
				node.Children = append(node.Children, n)
			}
		}
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
		IDbm:     node.IDbm,
		Value:    node.Value,
		Type:     node.Type,
		Children: make([]*Node, len(node.Children)),
	}
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
	idbm := idbms[0]
	phanTo, has := bm.phanTom[idbm]
	if !has {
		bm.recursiveNode(node, idbms[1:])
		return
	}

	phanToNode := &Node{
		IDbm: idbm,
		Type: "phanto",
	}

	for _, value := range phanTo.Values {
		valueNode := &Node{
			Value: value,
			IDbm:  idbm,
			Type:  "phanto_value",
		}
		bm.recursiveNode(valueNode, idbms[1:])
		phanToNode.Children = append(phanToNode.Children, valueNode)
	}

	node.Children = append(node.Children, phanToNode)
}

type PathNode struct {
	Type  string
	IDbm  int
	Value string
}

func getDepth(node *Node) int {
	if node == nil {
		return 0
	}
	if len(node.Children) == 0 {
		if node.Type == "" {
			return 0
		}
		return 1
	}
	maxChildDepth := 0
	for _, child := range node.Children {
		d := getDepth(child)
		if d > maxChildDepth {
			maxChildDepth = d
		}
	}
	if node.Type == "" {
		return maxChildDepth
	}
	return 1 + maxChildDepth
}

func layoutColTree(node *Node, rowOffset int, colOffset *int) {
	if node == nil {
		return
	}
	if len(node.Children) == 0 {
		node.Ri = rowOffset
		node.Ci = *colOffset
		*colOffset++
		return
	}

	startCol := *colOffset
	for _, child := range node.Children {
		layoutColTree(child, rowOffset+1, colOffset)
	}
	node.Ri = rowOffset
	node.Ci = startCol
}

func layoutRowTree(node *Node, colOffset int, rowOffset *int) {
	if node == nil {
		return
	}
	if len(node.Children) == 0 {
		node.Ci = colOffset
		node.Ri = *rowOffset
		*rowOffset++
		return
	}

	startRow := *rowOffset
	for _, child := range node.Children {
		layoutRowTree(child, colOffset+1, rowOffset)
	}
	node.Ci = colOffset
	node.Ri = startRow
}

func getLeaves(node *Node) []*Node {
	if node == nil {
		return nil
	}
	if len(node.Children) == 0 {
		if node.Type == "" {
			return nil
		}
		return []*Node{node}
	}
	var leaves []*Node
	for _, child := range node.Children {
		leaves = append(leaves, getLeaves(child)...)
	}
	return leaves
}

func getLeafIdbms(node *Node) []int {
	leaves := getLeaves(node)
	var idbms []int
	seen := make(map[int]bool)
	for _, leaf := range leaves {
		if leaf.IDbm != 0 {
			id := leaf.IDbm
			if !seen[id] {
				seen[id] = true
				idbms = append(idbms, id)
			}
		}
	}
	return idbms
}

func getAllNodesPreOrder(node *Node) []*Node {
	if node == nil {
		return nil
	}
	var nodes []*Node
	if node.Type != "" {
		nodes = append(nodes, node)
	}
	for _, child := range node.Children {
		nodes = append(nodes, getAllNodesPreOrder(child)...)
	}
	return nodes
}

func layoutColTreeCollapsed(node *Node, rowOffset int, colOffset *int) {
	if node == nil {
		return
	}
	if node.Type != "" {
		node.Ri = rowOffset
		node.Ci = *colOffset
		*colOffset++
	}
	for _, child := range node.Children {
		layoutColTreeCollapsed(child, rowOffset, colOffset)
	}
}

func layoutRowTreeCollapsed(node *Node, colOffset int, rowOffset *int) {
	if node == nil {
		return
	}
	if node.Type != "" {
		node.Ci = colOffset
		node.Ri = *rowOffset
		*rowOffset++
	}
	for _, child := range node.Children {
		layoutRowTreeCollapsed(child, colOffset, rowOffset)
	}
}

func collectLeafPaths(node *Node, currentPath []PathNode, paths *map[*Node][]PathNode) {
	if node == nil {
		return
	}
	newPath := currentPath
	if node.Type != "" {
		newPath = append(currentPath, PathNode{
			Type:  node.Type,
			IDbm:  node.IDbm,
			Value: node.Value,
		})
	}

	if len(node.Children) == 0 {
		if node.Type != "" {
			(*paths)[node] = newPath
		}
		return
	}

	for _, child := range node.Children {
		collectLeafPaths(child, newPath, paths)
	}
}

func makeIntSet(ids []int) map[int]bool {
	s := make(map[int]bool, len(ids))
	for _, id := range ids {
		s[id] = true
	}
	return s
}

func findChiTieu(bm *BieuMau, val string) string {
	for _, ct := range bm.ChiTieus {
		if ct.Name == val {
			return ct.Name
		}
	}
	return ""
}

// isLeafPhanTo returns true if pt itself or any of its children (PhanToChung) is in the leaf set.
func isLeafPhanTo(pt *PhanTo, leafSet map[int]bool) bool {
	if len(pt.Children) == 0 {
		return leafSet[pt.IDbm]
	}
	for _, child := range pt.Children {
		if leafSet[child.IDbm] {
			return true
		}
	}
	return false
}

func buildCiPathMap(tree *Node) map[int][]PathNode {
	paths := make(map[*Node][]PathNode)
	collectLeafPaths(tree, nil, &paths)
	m := make(map[int][]PathNode)
	for _, leaf := range getLeaves(tree) {
		if path, ok := paths[leaf]; ok {
			m[leaf.Ci] = path
		}
	}
	return m
}

func buildRiPathMap(tree *Node) map[int][]PathNode {
	paths := make(map[*Node][]PathNode)
	collectLeafPaths(tree, nil, &paths)
	m := make(map[int][]PathNode)
	for _, leaf := range getLeaves(tree) {
		if path, ok := paths[leaf]; ok {
			m[leaf.Ri] = path
		}
	}
	return m
}

func extractFromPaths(bm *BieuMau, cPath, rPath []PathNode) (string, []*KV) {
	var chiTieuName string
	var dims []*KV
	for _, p := range cPath {
		switch p.Type {
		case "chitieu":
			chiTieuName = p.Value
		case "phanto_value":
			if pt, ok := bm.phanTom[p.IDbm]; ok {
				dims = append(dims, &KV{Key: pt.Name, Value: p.Value})
			}
		}
	}
	for _, p := range rPath {
		switch p.Type {
		case "chitieu":
			if chiTieuName == "" {
				chiTieuName = p.Value
			}
		case "phanto_value":
			if pt, ok := bm.phanTom[p.IDbm]; ok {
				dims = append(dims, &KV{Key: pt.Name, Value: p.Value})
			}
		}
	}
	return chiTieuName, dims
}

func matchDims(dims1, dims2 []*KV) bool {
	if len(dims1) != len(dims2) {
		return false
	}
	for _, kv1 := range dims1 {
		found := false
		for _, kv2 := range dims2 {
			if kv1.Key == kv2.Key && kv1.Value == kv2.Value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func findSoLieu(bm *BieuMau, chiTieuName string, dims []*KV) string {
	var targetBang *BangChiTieu
	for _, bct := range bm.BangChiTieus {
		if bct.ChiTieuName == chiTieuName {
			targetBang = bct
			break
		}
	}
	if targetBang == nil {
		return ""
	}

	for _, ddl := range targetBang.DongDuLieus {
		if matchDims(ddl.Dims, dims) {
			return ddl.Solieu
		}
	}
	return ""
}

func (bm *BieuMau) genContent() {
	bm.Content = make(map[CellIndex]*Cell)

	var headerRows, headerCols int
	if bm.ColCollapse {
		headerRows = 1
	} else {
		headerRows = getDepth(bm.ColTree)
	}
	if bm.RowRollapse {
		headerCols = 1
	} else {
		headerCols = getDepth(bm.RowTree)
	}

	colOffset := headerCols
	if bm.ColCollapse {
		layoutColTreeCollapsed(bm.ColTree, 0, &colOffset)
	} else {
		layoutColTree(bm.ColTree, -1, &colOffset)
	}

	rowOffset := headerRows
	if bm.RowRollapse {
		layoutRowTreeCollapsed(bm.RowTree, 0, &rowOffset)
	} else {
		layoutRowTree(bm.RowTree, -1, &rowOffset)
	}

	var traverse func(n *Node)
	traverse = func(n *Node) {
		if n == nil {
			return
		}
		if n.Type != "" {
			var val string
			switch n.Type {
			case "chitieu":
				if ct, ok := bm.chiTieum[n.IDbm]; ok {
					val = ct.Name
				}
			case "phanto":
				if pt, ok := bm.phanTom[n.IDbm]; ok {
					val = pt.Name
				}
			case "phanto_value":
				val = n.Value
			}
			cell := &Cell{
				Value: val,
				Ci:    n.Ci,
				Ri:    n.Ri,
				Node:  n,
			}
			bm.Content[CellIndex{Ri: n.Ri, Ci: n.Ci}] = cell
		}
		for _, child := range n.Children {
			traverse(child)
		}
	}
	traverse(bm.ColTree)
	traverse(bm.RowTree)

	colLeaves := getLeaves(bm.ColTree)
	rowLeaves := getLeaves(bm.RowTree)

	colPaths := make(map[*Node][]PathNode)
	collectLeafPaths(bm.ColTree, nil, &colPaths)

	rowPaths := make(map[*Node][]PathNode)
	collectLeafPaths(bm.RowTree, nil, &rowPaths)

	numCols := len(colLeaves)
	if numCols == 0 {
		numCols = 1
	}
	numRows := len(rowLeaves)
	if numRows == 0 {
		numRows = 1
	}

	for r := 0; r < numRows; r++ {
		var rLeaf *Node
		var rPath []PathNode
		if r < len(rowLeaves) {
			rLeaf = rowLeaves[r]
			rPath = rowPaths[rLeaf]
		}
		ri := headerRows + r
		if rLeaf != nil {
			ri = rLeaf.Ri
		}

		for c := 0; c < numCols; c++ {
			var cLeaf *Node
			var cPath []PathNode
			if c < len(colLeaves) {
				cLeaf = colLeaves[c]
				cPath = colPaths[cLeaf]
			}
			ci := headerCols + c
			if cLeaf != nil {
				ci = cLeaf.Ci
			}

			var chiTieuName string
			var dims []*KV

			for _, p := range rPath {
				switch p.Type {
				case "chitieu":
					chiTieuName = p.Value
				case "phanto_value":
					if pt, ok := bm.phanTom[p.IDbm]; ok {
						dims = append(dims, &KV{Key: pt.Name, Value: p.Value})
					}
				}
			}
			for _, p := range cPath {
				switch p.Type {
				case "chitieu":
					chiTieuName = p.Value
				case "phanto_value":
					if pt, ok := bm.phanTom[p.IDbm]; ok {
						dims = append(dims, &KV{Key: pt.Name, Value: p.Value})
					}
				}
			}

			solieu := findSoLieu(bm, chiTieuName, dims)
			cell := &Cell{
				Value: solieu,
				Ci:    ci,
				Ri:    ri,
			}
			bm.Content[CellIndex{Ri: ri, Ci: ci}] = cell
		}
	}
}

func (bm *BieuMau) replaceContent(matrix [][]string) {
	var headerRows, headerCols int
	if bm.ColCollapse {
		headerRows = 1
	} else {
		headerRows = getDepth(bm.ColTree)
	}
	if bm.RowRollapse {
		headerCols = 1
	} else {
		headerCols = getDepth(bm.RowTree)
	}

	collectedValues := make(map[*PhanTo][]string)
	seenValues := make(map[*PhanTo]map[string]bool)

	addPhanToValue := func(pt *PhanTo, val string) {
		if val == "" {
			return
		}
		if len(pt.Children) > 0 {
			for _, child := range pt.Children {
				if seenValues[child] == nil {
					seenValues[child] = make(map[string]bool)
				}
				if !seenValues[child][val] {
					seenValues[child][val] = true
					collectedValues[child] = append(collectedValues[child], val)
				}
			}
		} else {
			if seenValues[pt] == nil {
				seenValues[pt] = make(map[string]bool)
			}
			if !seenValues[pt][val] {
				seenValues[pt][val] = true
				collectedValues[pt] = append(collectedValues[pt], val)
			}
		}
	}

	findPhanTo := func(ptName string, chiTieuName string) *PhanTo {
		if chiTieuName != "" {
			var targetChiTieu *ChiTieu
			for _, ct := range bm.ChiTieus {
				if ct.Name == chiTieuName {
					targetChiTieu = ct
					break
				}
			}
			if targetChiTieu != nil {
				for _, pt := range targetChiTieu.PhanTos {
					if pt.Name == ptName {
						return pt
					}
				}
			}
		}
		for _, pt := range bm.PhanToChungs {
			if pt.Name == ptName {
				return pt
			}
		}
		for _, ct := range bm.ChiTieus {
			for _, pt := range ct.PhanTos {
				if pt.Name == ptName {
					return pt
				}
			}
		}
		return nil
	}

	// Parse col headers to collect new PhanTo values
	if bm.ColCollapse {
		colLeafSet := makeIntSet(bm.ColLeafs)
		if len(matrix) > 0 {
			var curPT *PhanTo
			var curCT string
			for c := headerCols; c < len(matrix[0]); c++ {
				val := matrix[0][c]
				if ct := findChiTieu(bm, val); ct != "" {
					curCT, curPT = ct, nil
					continue
				}
				if pt := findPhanTo(val, curCT); pt != nil {
					curPT = pt
				} else if curPT != nil && isLeafPhanTo(curPT, colLeafSet) {
					addPhanToValue(curPT, val)
				}
			}
		}
	} else {
		if len(matrix) > 0 {
			for c := headerCols; c < len(matrix[0]); c++ {
				var currentChiTieuName string
				for r := 0; r < headerRows; r++ {
					val := matrix[r][c]
					for _, ct := range bm.ChiTieus {
						if ct.Name == val {
							currentChiTieuName = val
							break
						}
					}
				}
				for r := 0; r < headerRows; {
					val := matrix[r][c]
					pt := findPhanTo(val, currentChiTieuName)
					if pt != nil && r+1 < headerRows {
						addPhanToValue(pt, matrix[r+1][c])
						r += 2
					} else {
						r++
					}
				}
			}
		}
	}

	// Parse row headers to collect new PhanTo values
	if bm.RowRollapse {
		rowLeafSet := makeIntSet(bm.RowLeafs)
		var curPT *PhanTo
		var curCT string
		for r := headerRows; r < len(matrix); r++ {
			if len(matrix[r]) == 0 {
				continue
			}
			val := matrix[r][0]
			if ct := findChiTieu(bm, val); ct != "" {
				curCT, curPT = ct, nil
				continue
			}
			if pt := findPhanTo(val, curCT); pt != nil {
				curPT = pt
			} else if curPT != nil && isLeafPhanTo(curPT, rowLeafSet) {
				addPhanToValue(curPT, val)
			}
		}
	} else {
		for r := headerRows; r < len(matrix); r++ {
			var currentChiTieuName string
			for c := 0; c < headerCols; c++ {
				val := matrix[r][c]
				for _, ct := range bm.ChiTieus {
					if ct.Name == val {
						currentChiTieuName = val
						break
					}
				}
			}
			for c := 0; c < headerCols; {
				val := matrix[r][c]
				pt := findPhanTo(val, currentChiTieuName)
				if pt != nil && c+1 < headerCols {
					addPhanToValue(pt, matrix[r][c+1])
					c += 2
				} else {
					c++
				}
			}
		}
	}

	for pt, vals := range collectedValues {
		pt.Values = vals
	}

	bm.derived()

	bm.ColTree = &Node{}
	bm.genTree(bm.ColTree, bm.Cols)
	bm.RowTree = &Node{}
	bm.genTree(bm.RowTree, bm.Rows)
	bm.ColLeafs = getLeafIdbms(bm.ColTree)
	bm.RowLeafs = getLeafIdbms(bm.RowTree)

	var newHR, newHC int
	if bm.ColCollapse {
		newHR = 1
	} else {
		newHR = getDepth(bm.ColTree)
	}
	if bm.RowRollapse {
		newHC = 1
	} else {
		newHC = getDepth(bm.RowTree)
	}

	colOffset := newHC
	if bm.ColCollapse {
		layoutColTreeCollapsed(bm.ColTree, 0, &colOffset)
	} else {
		layoutColTree(bm.ColTree, -1, &colOffset)
	}
	rowOffset := newHR
	if bm.RowRollapse {
		layoutRowTreeCollapsed(bm.RowTree, 0, &rowOffset)
	} else {
		layoutRowTree(bm.RowTree, -1, &rowOffset)
	}

	bm.BangChiTieus = []*BangChiTieu{}
	bangMap := make(map[string]*BangChiTieu)
	for _, ct := range bm.ChiTieus {
		bct := &BangChiTieu{
			ChiTieuName: ct.Name,
			DongDuLieus: []*DongDuLieu{},
		}
		bm.BangChiTieus = append(bm.BangChiTieus, bct)
		bangMap[ct.Name] = bct
	}

	if bm.ColCollapse || bm.RowRollapse {
		ciToPath := buildCiPathMap(bm.ColTree)
		riToPath := buildRiPathMap(bm.RowTree)
		colHasLeaves := len(getLeaves(bm.ColTree)) > 0
		rowHasLeaves := len(getLeaves(bm.RowTree)) > 0

		for r := headerRows; r < len(matrix); r++ {
			rPath, hasR := riToPath[r]
			if !hasR && rowHasLeaves {
				continue
			}
			for c := headerCols; c < len(matrix[r]); c++ {
				cPath, hasC := ciToPath[c]
				if !hasC && colHasLeaves {
					continue
				}
				chiTieuName, dims := extractFromPaths(bm, cPath, rPath)
				if chiTieuName != "" {
					bct := bangMap[chiTieuName]
					if bct != nil {
						bct.DongDuLieus = append(bct.DongDuLieus, &DongDuLieu{
							Dims:   dims,
							Solieu: matrix[r][c],
						})
					}
				}
			}
		}
	} else {
		for r := headerRows; r < len(matrix); r++ {
			for c := headerCols; c < len(matrix[r]); c++ {
				var chiTieuNameCol string
				var dimsCol []*KV
				for ri := 0; ri < headerRows; ri++ {
					val := matrix[ri][c]
					for _, ct := range bm.ChiTieus {
						if ct.Name == val {
							chiTieuNameCol = val
							break
						}
					}
				}
				for ri := 0; ri < headerRows; {
					val := matrix[ri][c]
					pt := findPhanTo(val, chiTieuNameCol)
					if pt != nil && ri+1 < headerRows {
						dimsCol = append(dimsCol, &KV{Key: pt.Name, Value: matrix[ri+1][c]})
						ri += 2
					} else {
						ri++
					}
				}

				var chiTieuNameRow string
				var dimsRow []*KV
				for ci := 0; ci < headerCols; ci++ {
					val := matrix[r][ci]
					for _, ct := range bm.ChiTieus {
						if ct.Name == val {
							chiTieuNameRow = val
							break
						}
					}
				}
				for ci := 0; ci < headerCols; {
					val := matrix[r][ci]
					pt := findPhanTo(val, chiTieuNameRow)
					if pt != nil && ci+1 < headerCols {
						dimsRow = append(dimsRow, &KV{Key: pt.Name, Value: matrix[r][ci+1]})
						ci += 2
					} else {
						ci++
					}
				}

				chiTieuName := chiTieuNameCol
				if chiTieuName == "" {
					chiTieuName = chiTieuNameRow
				}

				if chiTieuName != "" {
					dims := append(dimsCol, dimsRow...)
					val := matrix[r][c]
					bct := bangMap[chiTieuName]
					if bct != nil {
						bct.DongDuLieus = append(bct.DongDuLieus, &DongDuLieu{
							Dims:   dims,
							Solieu: val,
						})
					}
				}
			}
		}
	}

	bm.genContent()
}

func (bm *BieuMau) importFromMatrix(matrix [][]string) {
	var headerRows, headerCols int
	if bm.ColCollapse {
		headerRows = 1
	} else {
		headerRows = getDepth(bm.ColTree)
	}
	if bm.RowRollapse {
		headerCols = 1
	} else {
		headerCols = getDepth(bm.RowTree)
	}

	// Preserve all matrix cells in Content
	bm.Content = make(map[CellIndex]*Cell)
	for r, row := range matrix {
		for c, val := range row {
			bm.Content[CellIndex{Ri: r, Ci: c}] = &Cell{Value: val, Ri: r, Ci: c}
		}
	}

	collectedValues := make(map[*PhanTo][]string)
	seenValues := make(map[*PhanTo]map[string]bool)

	addPhanToValue := func(pt *PhanTo, val string) {
		if val == "" {
			return
		}
		if len(pt.Children) > 0 {
			for _, child := range pt.Children {
				if seenValues[child] == nil {
					seenValues[child] = make(map[string]bool)
				}
				if !seenValues[child][val] {
					seenValues[child][val] = true
					collectedValues[child] = append(collectedValues[child], val)
				}
			}
		} else {
			if seenValues[pt] == nil {
				seenValues[pt] = make(map[string]bool)
			}
			if !seenValues[pt][val] {
				seenValues[pt][val] = true
				collectedValues[pt] = append(collectedValues[pt], val)
			}
		}
	}

	findPhanTo := func(ptName string, chiTieuName string) *PhanTo {
		if chiTieuName != "" {
			var targetChiTieu *ChiTieu
			for _, ct := range bm.ChiTieus {
				if ct.Name == chiTieuName {
					targetChiTieu = ct
					break
				}
			}
			if targetChiTieu != nil {
				for _, pt := range targetChiTieu.PhanTos {
					if pt.Name == ptName {
						return pt
					}
				}
			}
		}
		for _, pt := range bm.PhanToChungs {
			if pt.Name == ptName {
				return pt
			}
		}
		for _, ct := range bm.ChiTieus {
			for _, pt := range ct.PhanTos {
				if pt.Name == ptName {
					return pt
				}
			}
		}
		return nil
	}

	// Parse col headers in matrix order
	if bm.ColCollapse {
		colLeafSet := makeIntSet(bm.ColLeafs)
		if len(matrix) > 0 {
			var curPT *PhanTo
			var curCT string
			for c := headerCols; c < len(matrix[0]); c++ {
				val := matrix[0][c]
				if ct := findChiTieu(bm, val); ct != "" {
					curCT, curPT = ct, nil
					continue
				}
				if pt := findPhanTo(val, curCT); pt != nil {
					curPT = pt
				} else if curPT != nil && isLeafPhanTo(curPT, colLeafSet) {
					addPhanToValue(curPT, val)
				}
			}
		}
	} else {
		if len(matrix) > 0 {
			for c := headerCols; c < len(matrix[0]); c++ {
				var currentChiTieuName string
				for r := 0; r < headerRows; r++ {
					val := matrix[r][c]
					for _, ct := range bm.ChiTieus {
						if ct.Name == val {
							currentChiTieuName = val
							break
						}
					}
				}
				for r := 0; r < headerRows; {
					val := matrix[r][c]
					pt := findPhanTo(val, currentChiTieuName)
					if pt != nil && r+1 < headerRows {
						addPhanToValue(pt, matrix[r+1][c])
						r += 2
					} else {
						r++
					}
				}
			}
		}
	}

	// Parse row headers in matrix order
	if bm.RowRollapse {
		rowLeafSet := makeIntSet(bm.RowLeafs)
		var curPT *PhanTo
		var curCT string
		for r := headerRows; r < len(matrix); r++ {
			if len(matrix[r]) == 0 {
				continue
			}
			val := matrix[r][0]
			if ct := findChiTieu(bm, val); ct != "" {
				curCT, curPT = ct, nil
				continue
			}
			if pt := findPhanTo(val, curCT); pt != nil {
				curPT = pt
			} else if curPT != nil && isLeafPhanTo(curPT, rowLeafSet) {
				addPhanToValue(curPT, val)
			}
		}
	} else {
		for r := headerRows; r < len(matrix); r++ {
			var currentChiTieuName string
			for c := 0; c < headerCols; c++ {
				val := matrix[r][c]
				for _, ct := range bm.ChiTieus {
					if ct.Name == val {
						currentChiTieuName = val
						break
					}
				}
			}
			for c := 0; c < headerCols; {
				val := matrix[r][c]
				pt := findPhanTo(val, currentChiTieuName)
				if pt != nil && c+1 < headerCols {
					addPhanToValue(pt, matrix[r][c+1])
					c += 2
				} else {
					c++
				}
			}
		}
	}

	for pt, vals := range collectedValues {
		pt.Values = vals
	}

	bm.derived()

	// Restore matrix order for PhanToChungs (derived() sorts alphabetically)
	for _, ptc := range bm.PhanToChungs {
		seen := make(map[string]bool)
		ordered := make([]string, 0)
		for _, child := range ptc.Children {
			for _, v := range child.Values {
				if !seen[v] {
					seen[v] = true
					ordered = append(ordered, v)
				}
			}
		}
		ptc.Values = ordered
	}

	bm.ColTree = &Node{}
	bm.genTree(bm.ColTree, bm.Cols)
	bm.RowTree = &Node{}
	bm.genTree(bm.RowTree, bm.Rows)
	bm.ColLeafs = getLeafIdbms(bm.ColTree)
	bm.RowLeafs = getLeafIdbms(bm.RowTree)

	// Layout trees; since Values follow matrix order, coordinates match matrix positions
	colOffset := headerCols
	if bm.ColCollapse {
		layoutColTreeCollapsed(bm.ColTree, 0, &colOffset)
	} else {
		layoutColTree(bm.ColTree, -1, &colOffset)
	}
	rowOffset := headerRows
	if bm.RowRollapse {
		layoutRowTreeCollapsed(bm.RowTree, 0, &rowOffset)
	} else {
		layoutRowTree(bm.RowTree, -1, &rowOffset)
	}

	// Attach Node refs to existing Content cells (header cells get linked to tree nodes)
	var attachNodes func(*Node)
	attachNodes = func(n *Node) {
		if n.Type != "" {
			idx := CellIndex{Ri: n.Ri, Ci: n.Ci}
			if cell, ok := bm.Content[idx]; ok {
				cell.Node = n
			}
		}
		for _, child := range n.Children {
			attachNodes(child)
		}
	}
	attachNodes(bm.ColTree)
	attachNodes(bm.RowTree)

	// Rebuild BangChiTieus by reading data values from Content
	bm.BangChiTieus = []*BangChiTieu{}
	bangMap := make(map[string]*BangChiTieu)
	for _, ct := range bm.ChiTieus {
		bct := &BangChiTieu{
			ChiTieuName: ct.Name,
			DongDuLieus: []*DongDuLieu{},
		}
		bm.BangChiTieus = append(bm.BangChiTieus, bct)
		bangMap[ct.Name] = bct
	}

	colLeaves := getLeaves(bm.ColTree)
	rowLeaves := getLeaves(bm.RowTree)

	colPaths := make(map[*Node][]PathNode)
	collectLeafPaths(bm.ColTree, nil, &colPaths)
	rowPaths := make(map[*Node][]PathNode)
	collectLeafPaths(bm.RowTree, nil, &rowPaths)

	numCols := len(colLeaves)
	if numCols == 0 {
		numCols = 1
	}
	numRows := len(rowLeaves)
	if numRows == 0 {
		numRows = 1
	}

	for r := 0; r < numRows; r++ {
		var rLeaf *Node
		var rPath []PathNode
		if r < len(rowLeaves) {
			rLeaf = rowLeaves[r]
			rPath = rowPaths[rLeaf]
		}
		ri := headerRows + r
		if rLeaf != nil {
			ri = rLeaf.Ri
		}

		for c := 0; c < numCols; c++ {
			var cLeaf *Node
			var cPath []PathNode
			if c < len(colLeaves) {
				cLeaf = colLeaves[c]
				cPath = colPaths[cLeaf]
			}
			ci := headerCols + c
			if cLeaf != nil {
				ci = cLeaf.Ci
			}

			chiTieuName, dims := extractFromPaths(bm, cPath, rPath)
			if chiTieuName == "" {
				continue
			}
			bct := bangMap[chiTieuName]
			if bct == nil {
				continue
			}
			val := ""
			if cell, ok := bm.Content[CellIndex{Ri: ri, Ci: ci}]; ok {
				val = cell.Value
			}
			bct.DongDuLieus = append(bct.DongDuLieus, &DongDuLieu{
				Dims:   dims,
				Solieu: val,
			})
		}
	}
}

func (bm *BieuMau) setupFull() {
	bm.genPhanToChung()
	bm.derived()
	bm.genHeaders()
	bm.ColTree = &Node{}
	bm.genTree(bm.ColTree, bm.Cols)
	bm.RowTree = &Node{}
	bm.genTree(bm.RowTree, bm.Rows)

	// for content
	bm.ColLeafs = getLeafIdbms(bm.ColTree)
	bm.RowLeafs = getLeafIdbms(bm.RowTree)
	bm.genContent()
}

func (bm *BieuMau) reset() {
	bm.BangChiTieus = nil
	bm.PhanToChungs = nil
	bm.chiTieum = nil
	bm.phanTom = nil
	bm.Cols = nil
	bm.ColTree = nil
	bm.Rows = nil
	bm.RowTree = nil
	bm.Content = nil
}

// colAnchorNames returns the display names of direct children of ColTree root.
func (bm *BieuMau) colAnchorNames() map[string]bool {
	anchors := make(map[string]bool)
	if bm.ColTree == nil {
		return anchors
	}
	for _, child := range bm.ColTree.Children {
		switch child.Type {
		case "phanto":
			if pt, ok := bm.phanTom[child.IDbm]; ok {
				anchors[pt.Name] = true
			}
		case "chitieu":
			if ct, ok := bm.chiTieum[child.IDbm]; ok {
				anchors[ct.Name] = true
			}
		}
	}
	return anchors
}

// rowAnchorNames returns the display names of direct children of RowTree root.
func (bm *BieuMau) rowAnchorNames() map[string]bool {
	anchors := make(map[string]bool)
	if bm.RowTree == nil {
		return anchors
	}
	for _, child := range bm.RowTree.Children {
		switch child.Type {
		case "phanto":
			if pt, ok := bm.phanTom[child.IDbm]; ok {
				anchors[pt.Name] = true
			}
		case "chitieu":
			if ct, ok := bm.chiTieum[child.IDbm]; ok {
				anchors[ct.Name] = true
			}
		}
	}
	return anchors
}

// headerDepths returns the number of header rows (for cols) and header cols (for rows).
func (bm *BieuMau) headerDepths() (colHeaderDepth, rowHeaderDepth int) {
	if bm.ColCollapse {
		colHeaderDepth = 1
	} else {
		colHeaderDepth = getDepth(bm.ColTree)
	}
	if bm.RowRollapse {
		rowHeaderDepth = 1
	} else {
		rowHeaderDepth = getDepth(bm.RowTree)
	}
	return
}

// findTableOrigin scans the matrix and returns the top-left corner (originR, originC)
// of the actual data table. It uses known ChiTieu/PhanTo names as structural anchors.
// Precondition: setupFull() must have been called so ColTree/RowTree and phanTom/chiTieum are initialised.
func (bm *BieuMau) findTableOrigin(matrix [][]string) (originR, originC int, err error) {
	if len(matrix) == 0 {
		return 0, 0, errors.New("matrix rỗng")
	}

	colHeaderDepth, rowHeaderDepth := bm.headerDepths()
	colAnchors := bm.colAnchorNames()
	rowAnchors := bm.rowAnchorNames()

	if len(colAnchors) == 0 && len(rowAnchors) == 0 {
		return 0, 0, errors.New("BieuMau không có cấu trúc ColTree/RowTree để làm neo phát hiện")
	}

	// candidateScore tracks score for each candidate origin.
	type key struct{ r, c int }
	candidates := make(map[key]int)

	for r, row := range matrix {
		for c, cell := range row {
			if cell == "" {
				continue
			}
			if colAnchors[cell] {
				// col anchors appear at row=originR, col >= originC+rowHeaderDepth
				oR := r
				oC := c - rowHeaderDepth
				if oR >= 0 && oC >= 0 {
					candidates[key{oR, oC}]++
				}
			}
			if rowAnchors[cell] {
				// row anchors appear at col=originC, row >= originR+colHeaderDepth
				oR := r - colHeaderDepth
				oC := c
				if oR >= 0 && oC >= 0 {
					candidates[key{oR, oC}]++
				}
			}
		}
	}

	if len(candidates) == 0 {
		return 0, 0, errors.New("không tìm thấy cấu trúc bảng trong matrix")
	}

	// Score each candidate: count how many anchors appear at expected positions.
	score := func(oR, oC int) int {
		s := 0
		// Check col anchors in row oR at cols >= oC+rowHeaderDepth
		if oR < len(matrix) {
			for c := oC + rowHeaderDepth; c < len(matrix[oR]); c++ {
				if colAnchors[matrix[oR][c]] {
					s++
				}
			}
		}
		// Check row anchors in col oC at rows >= oR+colHeaderDepth
		for r := oR + colHeaderDepth; r < len(matrix); r++ {
			if oC < len(matrix[r]) && rowAnchors[matrix[r][oC]] {
				s++
			}
		}
		return s
	}

	bestScore := -1
	bestR, bestC := 0, 0
	for k := range candidates {
		s := score(k.r, k.c)
		if s > bestScore || (s == bestScore && (k.r < bestR || (k.r == bestR && k.c < bestC))) {
			bestScore = s
			bestR, bestC = k.r, k.c
		}
	}

	if bestScore <= 0 {
		// Fall back to any candidate with non-negative coords.
		for k := range candidates {
			if k.r >= 0 && k.c >= 0 {
				return k.r, k.c, nil
			}
		}
		return 0, 0, errors.New("không tìm thấy cấu trúc bảng trong matrix")
	}

	return bestR, bestC, nil
}

// findFreeRowsCols returns the absolute indices (into the raw matrix) of rows and columns
// that are not part of the structural table (e.g. inserted totals, label columns).
// originR and originC are the top-left corner of the actual table as returned by findTableOrigin.
func (bm *BieuMau) findFreeRowsCols(matrix [][]string, originR, originC int) (freeRows, freeCols []int) {
	colHeaderDepth, rowHeaderDepth := bm.headerDepths()
	colAnchors := bm.colAnchorNames()
	rowAnchors := bm.rowAnchorNames()

	if len(matrix) == 0 {
		return
	}
	width := 0
	if originR < len(matrix) {
		width = len(matrix[originR])
	}

	// --- Free column detection ---
	// Scan cols from originC+rowHeaderDepth rightward.
	// At the first header row (originR), cells should be col anchors or empty (merged span).
	lastColAnchor := ""
	lastColAnchorPhanTo := (*PhanTo)(nil)

	for ac := originC + rowHeaderDepth; ac < width; ac++ {
		cell0 := ""
		if originR < len(matrix) && ac < len(matrix[originR]) {
			cell0 = matrix[originR][ac]
		}

		isFree := false
		if cell0 == "" {
			// Merged cell: inherit from lastColAnchor. Free only if no anchor is active.
			if lastColAnchor == "" {
				isFree = true
			}
		} else if colAnchors[cell0] {
			lastColAnchor = cell0
			// Resolve the PhanTo for StrictValues checking.
			lastColAnchorPhanTo = nil
			for _, pt := range bm.PhanToChungs {
				if pt.Name == cell0 {
					lastColAnchorPhanTo = pt
					break
				}
			}
		} else if bm.ColCollapse {
			// In collapse mode, value cells appear in the same header row as anchor names.
			// Only reject via StrictValues; otherwise treat as structural (potential new value).
			if lastColAnchorPhanTo != nil && lastColAnchorPhanTo.StrictValues {
				if !slices.Contains(lastColAnchorPhanTo.Values, cell0) {
					isFree = true
				}
			}
			// If no active PhanTo context, could be a ChiTieu-specific PhanTo name: treat as structural.
		} else {
			// Non-collapse mode: depth-0 should only have anchor names or empty.
			lastColAnchor = ""
			lastColAnchorPhanTo = nil
			isFree = true
		}

		// In non-collapse mode, also check StrictValues at depth-1 (value row).
		if !isFree && !bm.ColCollapse && lastColAnchorPhanTo != nil && lastColAnchorPhanTo.StrictValues && colHeaderDepth >= 2 {
			valR := originR + 1
			valCell := ""
			if valR < len(matrix) && ac < len(matrix[valR]) {
				valCell = matrix[valR][ac]
			}
			if valCell != "" && !slices.Contains(lastColAnchorPhanTo.Values, valCell) {
				isFree = true
			}
		}

		if isFree {
			freeCols = append(freeCols, ac)
		}
	}

	// --- Free row detection ---
	// Scan rows from originR+colHeaderDepth downward.
	// At the first header column (originC), cells should be row anchors or empty.
	lastRowAnchor := ""
	lastRowAnchorPhanTo := (*PhanTo)(nil)

	for ar := originR + colHeaderDepth; ar < len(matrix); ar++ {
		cell0 := ""
		if originC < len(matrix[ar]) {
			cell0 = matrix[ar][originC]
		}

		isFree := false
		if cell0 == "" {
			if lastRowAnchor == "" {
				isFree = true
			}
		} else if rowAnchors[cell0] {
			lastRowAnchor = cell0
			lastRowAnchorPhanTo = nil
			for _, pt := range bm.PhanToChungs {
				if pt.Name == cell0 {
					lastRowAnchorPhanTo = pt
					break
				}
			}
		} else if bm.RowRollapse {
			// In collapse mode, value cells appear in the same header column as anchor names.
			if lastRowAnchorPhanTo != nil && lastRowAnchorPhanTo.StrictValues {
				if !slices.Contains(lastRowAnchorPhanTo.Values, cell0) {
					isFree = true
				}
			}
		} else {
			lastRowAnchor = ""
			lastRowAnchorPhanTo = nil
			isFree = true
		}

		// In non-collapse mode, also check StrictValues at depth-1 (value col).
		if !isFree && !bm.RowRollapse && lastRowAnchorPhanTo != nil && lastRowAnchorPhanTo.StrictValues && rowHeaderDepth >= 2 {
			valC := originC + 1
			valCell := ""
			if valC < len(matrix[ar]) {
				valCell = matrix[ar][valC]
			}
			if valCell != "" && !slices.Contains(lastRowAnchorPhanTo.Values, valCell) {
				isFree = true
			}
		}

		if isFree {
			freeRows = append(freeRows, ar)
		}
	}

	return
}

// extractSubMatrix extracts the sub-matrix starting at (originR, originC),
// skipping the absolute row and column indices listed in freeRows and freeCols.
func extractSubMatrix(matrix [][]string, originR, originC int, freeRows, freeCols []int) ([][]string, error) {
	if originR < 0 || originC < 0 {
		return nil, errors.New("origin âm")
	}
	freeRowSet := makeIntSet(freeRows)
	freeColSet := makeIntSet(freeCols)

	result := [][]string{}
	for r := originR; r < len(matrix); r++ {
		if freeRowSet[r] {
			continue
		}
		row := []string{}
		for c := originC; c < len(matrix[r]); c++ {
			if freeColSet[c] {
				continue
			}
			row = append(row, matrix[r][c])
		}
		result = append(result, row)
	}
	if len(result) == 0 {
		return nil, errors.New("sub-matrix rỗng sau khi loại bỏ hàng/cột tự do")
	}
	return result, nil
}

// importFromMatrixAuto imports data from an arbitrary Excel-pasted matrix that may be shifted
// (extra rows/cols of free content above/left) and may contain inserted free rows/cols (totals, labels).
// Precondition: setupFull() must have been called first.
func (bm *BieuMau) importFromMatrixAuto(matrix [][]string) error {
	originR, originC, err := bm.findTableOrigin(matrix)
	if err != nil {
		return err
	}
	freeRows, freeCols := bm.findFreeRowsCols(matrix, originR, originC)
	sub, err := extractSubMatrix(matrix, originR, originC, freeRows, freeCols)
	if err != nil {
		return err
	}
	bm.importFromMatrix(sub)
	return nil
}
