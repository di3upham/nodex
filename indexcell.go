package main

import "slices"

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
			IDbms: []int{idbm},
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
	idbm := idbms[0]
	phanTo, has := bm.phanTom[idbm]
	if !has {
		bm.recursiveNode(node, idbms[1:])
		return
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
		if len(leaf.IDbms) > 0 {
			id := leaf.IDbms[0]
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
		idbm := 0
		if len(node.IDbms) > 0 {
			idbm = node.IDbms[0]
		}
		newPath = append(currentPath, PathNode{
			Type:  node.Type,
			IDbm:  idbm,
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
				if ct, ok := bm.chiTieum[n.IDbms[0]]; ok {
					val = ct.Name
				}
			case "phanto":
				if pt, ok := bm.phanTom[n.IDbms[0]]; ok {
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

func (bm *BieuMau) setupFull() {
	bm.setupBase()
	bm.genHeaders()
	bm.ColTree = &Node{}
	bm.genTree(bm.ColTree, bm.Cols)
	bm.RowTree = &Node{}
	bm.genTree(bm.RowTree, bm.Rows)
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
