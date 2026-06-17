package main

import (
	"reflect"
	"testing"
)

func TestUniqArrs(t *testing.T) {
	tests := []struct {
		name string
		arrs [][]string
		want []string
	}{
		{
			name: "single empty",
			arrs: [][]string{{}},
			want: []string{},
		},
		{
			name: "no duplicate, sorted",
			arrs: [][]string{{"B", "A"}, {"C"}},
			want: []string{"A", "B", "C"},
		},
		{
			name: "duplicates",
			arrs: [][]string{{"A", "B", "A"}, {"B", "C"}},
			want: []string{"A", "B", "C"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := uniqArrs(tt.arrs...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uniqArrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDepth(t *testing.T) {
	// Nil node
	if got := getDepth(nil); got != 0 {
		t.Errorf("getDepth(nil) = %d, want 0", got)
	}

	// Single leaf node (no type)
	nodeLeafNoType := &Node{}
	if got := getDepth(nodeLeafNoType); got != 0 {
		t.Errorf("getDepth(leaf with no type) = %d, want 0", got)
	}

	// Single leaf node with type
	nodeLeaf := &Node{Type: "chitieu"}
	if got := getDepth(nodeLeaf); got != 1 {
		t.Errorf("getDepth(leaf with type) = %d, want 1", got)
	}

	// Node with children
	nodeParent := &Node{
		Type: "phanto",
		Children: []*Node{
			{Type: "phanto_value"},
		},
	}
	if got := getDepth(nodeParent); got != 2 {
		t.Errorf("getDepth(parent) = %d, want 2", got)
	}
}

func TestCopyNode(t *testing.T) {
	node := &Node{
		Value: "test",
		Type:  "chitieu",
		IDbm:  1,
		Children: []*Node{
			{Value: "child", Type: "phanto"},
		},
	}

	clone := copyNode(node)

	if clone == node {
		t.Errorf("copyNode() returned same pointer")
	}
	if clone.Value != node.Value || clone.Type != node.Type {
		t.Errorf("copyNode() base properties mismatch: got %+v, want %+v", clone, node)
	}
	if clone.IDbm != node.IDbm {
		t.Errorf("copyNode() IDbm mismatch: got %d, want %d", clone.IDbm, node.IDbm)
	}
	if len(clone.Children) != len(node.Children) {
		t.Errorf("copyNode() children length mismatch")
	}
	if clone.Children[0] == node.Children[0] {
		t.Errorf("copyNode() copied child pointer instead of cloning")
	}
}

func TestFlatTableFlow(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "flat"
	bm.setupFull()

	// 1. Verify initial layout properties
	// depth of ColTree (headerRows) = 5
	// depth of RowTree (headerCols) = 0
	if got := getDepth(bm.ColTree); got != 5 {
		t.Errorf("Expected ColTree depth 5, got %d", got)
	}
	if got := getDepth(bm.RowTree); got != 0 {
		t.Errorf("Expected RowTree depth 0, got %d", got)
	}

	// 2. Verify initial data cell values in Content
	// Col 0: Bắc, Doanh thu, Bán lẻ -> 100
	// Col 1: Bắc, Chi phí, Vận hành -> 80
	// Col 2: Nam, Doanh thu, Bán lẻ -> 150
	// Col 3: Nam, Chi phí, Vận hành -> 120
	expectedInitial := map[CellIndex]string{
		{Ri: 5, Ci: 0}: "100",
		{Ri: 5, Ci: 1}: "80",
		{Ri: 5, Ci: 2}: "150",
		{Ri: 5, Ci: 3}: "120",
	}

	for idx, val := range expectedInitial {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("Cell at %+v not found in Content", idx)
		} else if cell.Value != val {
			t.Errorf("Cell at %+v value = %q, want %q", idx, cell.Value, val)
		}
	}

	// 3. Edit and Import
	editedMatrix := [][]string{
		{"Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền", "Vùng miền"},
		{"Miền Bắc", "Miền Trung", "Miền Nam", "Miền Bắc", "Miền Trung", "Miền Nam", "Miền Nam"},
		{"Doanh thu", "Doanh thu", "Doanh thu", "Chi phí", "Chi phí", "Chi phí", "Chi phí"},
		{"Hình thức", "Hình thức", "Hình thức", "Loại CP", "Loại CP", "Loại CP", "Loại CP"},
		{"Bán lẻ", "Bán lẻ", "Bán lẻ", "Vận hành", "Vận hành", "Vận hành", "Bán buôn"},
		{"100", "110", "150", "80", "90", "120", "200"},
	}

	bm.replaceContent(editedMatrix)

	// After replacement, ColTree depth remains 5, but there are 9 leaf columns:
	// Bắc (Doanh thu Bán lẻ, Chi phí Vận hành, Chi phí Bán buôn) -> cols 0, 1, 2
	// Nam (Doanh thu Bán lẻ, Chi phí Vận hành, Chi phí Bán buôn) -> cols 3, 4, 5
	// Trung (Doanh thu Bán lẻ, Chi phí Vận hành, Chi phí Bán buôn) -> cols 6, 7, 8
	expectedImported := map[CellIndex]string{
		{Ri: 5, Ci: 0}: "100", // Bắc, Doanh thu, Bán lẻ
		{Ri: 5, Ci: 1}: "80",  // Bắc, Chi phí, Vận hành
		{Ri: 5, Ci: 2}: "",    // Bắc, Chi phí, Bán buôn (empty)
		{Ri: 5, Ci: 3}: "150", // Nam, Doanh thu, Bán lẻ
		{Ri: 5, Ci: 4}: "120", // Nam, Chi phí, Vận hành
		{Ri: 5, Ci: 5}: "200", // Nam, Chi phí, Bán buôn
		{Ri: 5, Ci: 6}: "110", // Trung, Doanh thu, Bán lẻ
		{Ri: 5, Ci: 7}: "90",  // Trung, Chi phí, Vận hành
		{Ri: 5, Ci: 8}: "",    // Trung, Chi phí, Bán buôn (empty)
	}

	for idx, val := range expectedImported {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("Imported Cell at %+v not found in Content", idx)
		} else if cell.Value != val {
			t.Errorf("Imported Cell at %+v value = %q, want %q", idx, cell.Value, val)
		}
	}
}

func TestMatrixChitieuInRowsFlow(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()

	// 1. Verify layout properties
	// depth of ColTree (headerRows) = 2
	// depth of RowTree (headerCols) = 3
	if got := getDepth(bm.ColTree); got != 2 {
		t.Errorf("Expected ColTree depth 2, got %d", got)
	}
	if got := getDepth(bm.RowTree); got != 3 {
		t.Errorf("Expected RowTree depth 3, got %d", got)
	}

	// 2. Verify initial data cell values in Content
	// Rows (headerCols = 3): Doanh thu/Hình thức/Bán lẻ -> Ri = 2, Chi phí/Loại CP/Vận hành -> Ri = 3
	// Cols (headerRows = 2): Miền Bắc -> Ci = 3, Miền Nam -> Ci = 4
	expectedInitial := map[CellIndex]string{
		{Ri: 2, Ci: 3}: "100", // Doanh thu, Bán lẻ & Miền Bắc
		{Ri: 2, Ci: 4}: "150", // Doanh thu, Bán lẻ & Miền Nam
		{Ri: 3, Ci: 3}: "80",  // Chi phí, Vận hành & Miền Bắc
		{Ri: 3, Ci: 4}: "120", // Chi phí, Vận hành & Miền Nam
	}

	for idx, val := range expectedInitial {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("Cell at %+v not found in Content", idx)
		} else if cell.Value != val {
			t.Errorf("Cell at %+v value = %q, want %q", idx, cell.Value, val)
		}
	}

	// 3. Edit and Import
	editedMatrix := [][]string{
		{"-", "-", "-", "Vùng miền", "Vùng miền", "Vùng miền"},
		{"-", "-", "-", "Miền Bắc", "Miền Trung", "Miền Nam"},
		{"Doanh thu", "Hình thức", "Bán lẻ", "100", "110", "150"},
		{"Chi phí", "Loại CP", "Vận hành", "80", "90", "120"},
		{"Chi phí", "Loại CP", "Bán buôn", "-", "-", "200"},
	}

	bm.replaceContent(editedMatrix)

	// After replacement, ColTree depth = 2 (leaves: Bắc [col 3], Nam [col 4], Trung [col 5])
	// RowTree depth = 3 (leaves: Doanh thu Bán lẻ [row 2], Chi phí Vận hành [row 3], Chi phí Bán buôn [row 4])
	expectedImported := map[CellIndex]string{
		{Ri: 2, Ci: 3}: "100", // Doanh thu, Bán lẻ & Miền Bắc
		{Ri: 2, Ci: 4}: "150", // Doanh thu, Bán lẻ & Miền Nam
		{Ri: 2, Ci: 5}: "110", // Doanh thu, Bán lẻ & Miền Trung
		{Ri: 3, Ci: 3}: "80",  // Chi phí, Vận hành & Miền Bắc
		{Ri: 3, Ci: 4}: "120", // Chi phí, Vận hành & Miền Nam
		{Ri: 3, Ci: 5}: "90",  // Chi phí, Vận hành & Miền Trung
		{Ri: 4, Ci: 3}: "-",   // Chi phí, Bán buôn & Miền Bắc (empty)
		{Ri: 4, Ci: 4}: "200", // Chi phí, Bán buôn & Miền Nam
		{Ri: 4, Ci: 5}: "-",   // Chi phí, Bán buôn & Miền Trung (empty)
	}

	for idx, val := range expectedImported {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("Imported Cell at %+v not found in Content", idx)
		} else if cell.Value != val {
			t.Errorf("Imported Cell at %+v value = %q, want %q", idx, cell.Value, val)
		}
	}
}

func TestColCollapseGenContent(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_cols"
	bm.ColCollapse = true
	bm.setupFull()

	// headerRows=1, headerCols=getDepth(RowTree)=2
	// ColTree collapsed: all col-header nodes at Ri=0
	// Col headers (Ri=0): DT@Ci=2, Hình thức@Ci=3, Bán lẻ@Ci=4, CP@Ci=5, Loại CP@Ci=6, Vận hành@Ci=7
	// Row headers (non-collapsed): Vùng miền@(Ri=1,Ci=0), Bắc@(Ri=1,Ci=1), Nam@(Ri=2,Ci=1)
	// Data: (Ri=1,Ci=4)=100, (Ri=1,Ci=7)=80, (Ri=2,Ci=4)=150, (Ri=2,Ci=7)=120

	// All col-header cells must be in Ri=0
	for idx, cell := range bm.Content {
		if cell.Node != nil && cell.Node.Type != "" {
			// check if it's a col-tree node
			isColNode := false
			var checkNode func(n *Node) bool
			checkNode = func(n *Node) bool {
				if n == cell.Node {
					return true
				}
				for _, c := range n.Children {
					if checkNode(c) {
						return true
					}
				}
				return false
			}
			isColNode = checkNode(bm.ColTree)
			if isColNode && idx.Ri != 0 {
				t.Errorf("ColCollapse: col-header node at Ri=%d (want 0), cell=%+v", idx.Ri, idx)
			}
		}
	}

	expectedHeaders := map[CellIndex]string{
		{Ri: 0, Ci: 2}: "Doanh thu",
		{Ri: 0, Ci: 3}: "Hình thức",
		{Ri: 0, Ci: 4}: "Bán lẻ",
		{Ri: 0, Ci: 5}: "Chi phí",
		{Ri: 0, Ci: 6}: "Loại CP",
		{Ri: 0, Ci: 7}: "Vận hành",
	}
	for idx, want := range expectedHeaders {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("ColCollapse header cell at %+v not found", idx)
		} else if cell.Value != want {
			t.Errorf("ColCollapse header cell at %+v = %q, want %q", idx, cell.Value, want)
		}
	}

	expectedData := map[CellIndex]string{
		{Ri: 1, Ci: 4}: "100",
		{Ri: 1, Ci: 7}: "80",
		{Ri: 2, Ci: 4}: "150",
		{Ri: 2, Ci: 7}: "120",
	}
	for idx, want := range expectedData {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("ColCollapse data cell at %+v not found", idx)
		} else if cell.Value != want {
			t.Errorf("ColCollapse data cell at %+v = %q, want %q", idx, cell.Value, want)
		}
	}
}

func TestRowRollapseGenContent(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.RowRollapse = true
	bm.setupFull()

	// headerCols=1, headerRows=getDepth(ColTree)=2
	// RowTree collapsed: all row-header nodes at Ci=0
	// Row headers (Ci=0): DT@Ri=2, Hình thức@Ri=3, Bán lẻ@Ri=4, CP@Ri=5, Loại CP@Ri=6, Vận hành@Ri=7
	// Col headers (non-collapsed): Vùng miền@(Ri=0,Ci=1), Bắc@(Ri=1,Ci=1), Nam@(Ri=1,Ci=2)
	// Data: (Ri=4,Ci=1)=100, (Ri=4,Ci=2)=150, (Ri=7,Ci=1)=80, (Ri=7,Ci=2)=120

	// All row-header cells must be in Ci=0
	for idx, cell := range bm.Content {
		if cell.Node != nil && cell.Node.Type != "" {
			isRowNode := false
			var checkNode func(n *Node) bool
			checkNode = func(n *Node) bool {
				if n == cell.Node {
					return true
				}
				for _, c := range n.Children {
					if checkNode(c) {
						return true
					}
				}
				return false
			}
			isRowNode = checkNode(bm.RowTree)
			if isRowNode && idx.Ci != 0 {
				t.Errorf("RowRollapse: row-header node at Ci=%d (want 0), cell=%+v", idx.Ci, idx)
			}
		}
	}

	expectedHeaders := map[CellIndex]string{
		{Ri: 2, Ci: 0}: "Doanh thu",
		{Ri: 3, Ci: 0}: "Hình thức",
		{Ri: 4, Ci: 0}: "Bán lẻ",
		{Ri: 5, Ci: 0}: "Chi phí",
		{Ri: 6, Ci: 0}: "Loại CP",
		{Ri: 7, Ci: 0}: "Vận hành",
	}
	for idx, want := range expectedHeaders {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("RowRollapse header cell at %+v not found", idx)
		} else if cell.Value != want {
			t.Errorf("RowRollapse header cell at %+v = %q, want %q", idx, cell.Value, want)
		}
	}

	expectedData := map[CellIndex]string{
		{Ri: 4, Ci: 1}: "100",
		{Ri: 4, Ci: 2}: "150",
		{Ri: 7, Ci: 1}: "80",
		{Ri: 7, Ci: 2}: "120",
	}
	for idx, want := range expectedData {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("RowRollapse data cell at %+v not found", idx)
		} else if cell.Value != want {
			t.Errorf("RowRollapse data cell at %+v = %q, want %q", idx, cell.Value, want)
		}
	}
}

func TestImportFromMatrix(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()

	// headerRows = getDepth(ColTree) = 2, headerCols = getDepth(RowTree) = 3
	// Matrix has:
	// - non-ChiTieu cells at [0][0]="BÁO CÁO", [0][1]="Tháng 1"
	// - region order: Bắc (col 3), Trung (col 4), Nam (col 5)
	//   which differs from alphabetical order (Bắc < Nam < Trung)
	// - total column (col 6): sum of each data row  (360=100+110+150, 290=80+90+120)
	// - total row    (row 4): sum of each data col  (180=100+80, 200=110+90, 270=150+120)
	editedMatrix := [][]string{
		{"BÁO CÁO", "Tháng 1", "", "Vùng miền", "Vùng miền", "Vùng miền", "Tổng"},
		{"", "", "", "Miền Bắc", "Miền Trung", "Miền Nam", ""},
		{"Doanh thu", "Hình thức", "Bán lẻ", "100", "110", "150", "360"},
		{"Chi phí", "Loại CP", "Vận hành", "80", "90", "120", "290"},
		{"", "", "Tổng", "180", "200", "270", ""},
	}

	bm.importFromMatrix(editedMatrix)

	// 1. Non-ChiTieu cells (outside ChiTieu structure) must be preserved with no Node ref
	for _, tc := range []struct {
		idx  CellIndex
		want string
	}{
		{CellIndex{Ri: 0, Ci: 0}, "BÁO CÁO"},
		{CellIndex{Ri: 0, Ci: 1}, "Tháng 1"},
		// total column header
		{CellIndex{Ri: 0, Ci: 6}, "Tổng"},
		{CellIndex{Ri: 1, Ci: 6}, ""},
		// total column values
		{CellIndex{Ri: 2, Ci: 6}, "360"},
		{CellIndex{Ri: 3, Ci: 6}, "290"},
		// total row label
		{CellIndex{Ri: 4, Ci: 2}, "Tổng"},
		// total row values
		{CellIndex{Ri: 4, Ci: 3}, "180"},
		{CellIndex{Ri: 4, Ci: 4}, "200"},
		{CellIndex{Ri: 4, Ci: 5}, "270"},
	} {
		cell, ok := bm.Content[tc.idx]
		if !ok {
			t.Errorf("cell at %+v missing from Content", tc.idx)
		} else if cell.Value != tc.want {
			t.Errorf("cell at %+v = %q, want %q", tc.idx, cell.Value, tc.want)
		} else if cell.Node != nil {
			t.Errorf("cell at %+v should have no Node (non-ChiTieu), got %+v", tc.idx, cell.Node)
		}
	}

	// 2. Tree nodes must be at matrix positions, not alphabetical order
	// "Miền Trung" is at col 4 in matrix → must be at Ci=4, NOT Ci=5 (alphabetical)
	// "Miền Nam" is at col 5 in matrix → must be at Ci=5, NOT Ci=4
	for _, tc := range []struct {
		idx       CellIndex
		nodeValue string
	}{
		{CellIndex{Ri: 1, Ci: 3}, "Miền Bắc"},
		{CellIndex{Ri: 1, Ci: 4}, "Miền Trung"},
		{CellIndex{Ri: 1, Ci: 5}, "Miền Nam"},
	} {
		cell, ok := bm.Content[tc.idx]
		if !ok {
			t.Errorf("header cell at %+v missing from Content", tc.idx)
			continue
		}
		if cell.Node == nil {
			t.Errorf("header cell at %+v has no Node", tc.idx)
			continue
		}
		if cell.Node.Value != tc.nodeValue {
			t.Errorf("header cell at %+v Node.Value = %q, want %q", tc.idx, cell.Node.Value, tc.nodeValue)
		}
	}

	// 3. Data cells must be preserved at matrix positions
	expectedData := map[CellIndex]string{
		{Ri: 2, Ci: 3}: "100",
		{Ri: 2, Ci: 4}: "110",
		{Ri: 2, Ci: 5}: "150",
		{Ri: 3, Ci: 3}: "80",
		{Ri: 3, Ci: 4}: "90",
		{Ri: 3, Ci: 5}: "120",
	}
	for idx, want := range expectedData {
		cell, ok := bm.Content[idx]
		if !ok {
			t.Errorf("data cell at %+v not found in Content", idx)
		} else if cell.Value != want {
			t.Errorf("data cell at %+v = %q, want %q", idx, cell.Value, want)
		}
	}

	// 4. BangChiTieus must reflect data at matrix positions (Trung=110/90, not Nam's values)
	solieu := func(chiTieuName, vuongMienVal string) string {
		for _, bct := range bm.BangChiTieus {
			if bct.ChiTieuName != chiTieuName {
				continue
			}
			for _, ddl := range bct.DongDuLieus {
				for _, kv := range ddl.Dims {
					if kv.Key == "Vùng miền" && kv.Value == vuongMienVal {
						return ddl.Solieu
					}
				}
			}
		}
		return ""
	}

	for _, tc := range []struct {
		chiTieu, vung, want string
	}{
		{"Doanh thu", "Miền Bắc", "100"},
		{"Doanh thu", "Miền Trung", "110"},
		{"Doanh thu", "Miền Nam", "150"},
		{"Chi phí", "Miền Bắc", "80"},
		{"Chi phí", "Miền Trung", "90"},
		{"Chi phí", "Miền Nam", "120"},
	} {
		if got := solieu(tc.chiTieu, tc.vung); got != tc.want {
			t.Errorf("BangChiTieu[%s][%s] = %q, want %q", tc.chiTieu, tc.vung, got, tc.want)
		}
	}
}

func TestMatrixChitieuInColsFlow(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_cols"
	bm.setupFull()

	// 1. Verify layout properties
	// depth of ColTree (headerRows) = 3
	// depth of RowTree (headerCols) = 2
	if got := getDepth(bm.ColTree); got != 3 {
		t.Errorf("Expected ColTree depth 3, got %d", got)
	}
	if got := getDepth(bm.RowTree); got != 2 {
		t.Errorf("Expected RowTree depth 2, got %d", got)
	}

	// 2. Verify initial data cell values in Content
	// Rows (headerCols = 2): Miền Bắc -> Ri = 3, Miền Nam -> Ri = 4
	// Cols (headerRows = 3): Doanh thu/Hình thức/Bán lẻ -> Ci = 2, Chi phí/Loại CP/Vận hành -> Ci = 3
	expectedInitial := map[CellIndex]string{
		{Ri: 3, Ci: 2}: "100", // Miền Bắc & Doanh thu, Bán lẻ
		{Ri: 3, Ci: 3}: "80",  // Miền Bắc & Chi phí, Vận hành
		{Ri: 4, Ci: 2}: "150", // Miền Nam & Doanh thu, Bán lẻ
		{Ri: 4, Ci: 3}: "120", // Miền Nam & Chi phí, Vận hành
	}

	for idx, val := range expectedInitial {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("Cell at %+v not found in Content", idx)
		} else if cell.Value != val {
			t.Errorf("Cell at %+v value = %q, want %q", idx, cell.Value, val)
		}
	}

	// 3. Edit and Import
	editedMatrix := [][]string{
		{"-", "-", "Doanh thu", "Chi phí", "Chi phí"},
		{"-", "-", "Hình thức", "Loại CP", "Loại CP"},
		{"-", "-", "Bán lẻ", "Vận hành", "Bán buôn"},
		{"Vùng miền", "Miền Bắc", "100", "80", "-"},
		{"Vùng miền", "Miền Trung", "110", "90", "-"},
		{"Vùng miền", "Miền Nam", "150", "120", "200"},
	}

	bm.replaceContent(editedMatrix)

	// After replacement:
	// ColTree depth = 3 (leaves: Doanh thu Bán lẻ [col 2], Chi phí Vận hành [col 3], Chi phí Bán buôn [col 4])
	// RowTree depth = 2 (leaves: Bắc [row 3], Nam [row 4], Trung [row 5])
	expectedImported := map[CellIndex]string{
		{Ri: 3, Ci: 2}: "100", // Miền Bắc & Doanh thu, Bán lẻ
		{Ri: 3, Ci: 3}: "80",  // Miền Bắc & Chi phí, Vận hành
		{Ri: 3, Ci: 4}: "-",   // Miền Bắc & Chi phí, Bán buôn (empty)
		{Ri: 4, Ci: 2}: "150", // Miền Nam & Doanh thu, Bán lẻ
		{Ri: 4, Ci: 3}: "120", // Miền Nam & Chi phí, Vận hành
		{Ri: 4, Ci: 4}: "200", // Miền Nam & Chi phí, Bán buôn
		{Ri: 5, Ci: 2}: "110", // Miền Trung & Doanh thu, Bán lẻ
		{Ri: 5, Ci: 3}: "90",  // Miền Trung & Chi phí, Vận hành
		{Ri: 5, Ci: 4}: "-",   // Miền Trung & Chi phí, Bán buôn (empty)
	}

	for idx, val := range expectedImported {
		cell, exists := bm.Content[idx]
		if !exists {
			t.Errorf("Imported Cell at %+v not found in Content", idx)
		} else if cell.Value != val {
			t.Errorf("Imported Cell at %+v value = %q, want %q", idx, cell.Value, val)
		}
	}
}

// ─── findTableOrigin tests ───────────────────────────────────────────────────

// TestFindTableOriginShifted verifies origin detection when the table is
// shifted 2 rows down and 1 col right by free content.
func TestFindTableOriginShifted(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()
	// colHeaderDepth=2, rowHeaderDepth=3
	// colAnchors={"Vùng miền"}, rowAnchors={"Doanh thu","Chi phí"}

	matrix := [][]string{
		{"TIÊU ĐỀ BÁO CÁO", "", "", "", "", ""},
		{"", "", "", "", "", ""},
		{"", "", "", "", "Vùng miền", "Vùng miền"},
		{"", "", "", "", "Miền Bắc", "Miền Nam"},
		{"", "Doanh thu", "Hình thức", "Bán lẻ", "100", "150"},
		{"", "Chi phí", "Loại CP", "Vận hành", "80", "120"},
	}

	originR, originC, err := bm.findTableOrigin(matrix)
	if err != nil {
		t.Fatalf("findTableOrigin() error = %v", err)
	}
	if originR != 2 || originC != 1 {
		t.Errorf("findTableOrigin() = (%d, %d), want (2, 1)", originR, originC)
	}
}

// TestFindTableOriginDiagonal verifies detection when shifted diagonally (3 down, 2 right).
func TestFindTableOriginDiagonal(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()

	matrix := [][]string{
		{"Tài liệu nội bộ", "", "", "", "", "", "", ""},
		{"", "Đơn vị: triệu đồng", "", "", "", "", "", ""},
		{"", "", "", "", "", "", "", ""},
		{"", "", "", "", "", "Vùng miền", "Vùng miền", ""},
		{"", "", "", "", "", "Miền Bắc", "Miền Nam", ""},
		{"", "", "Doanh thu", "Hình thức", "Bán lẻ", "100", "150", ""},
		{"", "", "Chi phí", "Loại CP", "Vận hành", "80", "120", ""},
	}

	originR, originC, err := bm.findTableOrigin(matrix)
	if err != nil {
		t.Fatalf("findTableOrigin() error = %v", err)
	}
	if originR != 3 || originC != 2 {
		t.Errorf("findTableOrigin() = (%d, %d), want (3, 2)", originR, originC)
	}
}

// TestFindTableOriginNoAnchors verifies an error is returned when the matrix
// contains no known structural names.
func TestFindTableOriginNoAnchors(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()

	matrix := [][]string{
		{"100", "200", "300"},
		{"400", "500", "600"},
	}

	_, _, err := bm.findTableOrigin(matrix)
	if err == nil {
		t.Error("findTableOrigin() expected error for matrix with no anchors, got nil")
	}
}

// ─── findFreeRowsCols tests ──────────────────────────────────────────────────

// TestFindFreeColsTotalCol verifies that a "Tổng" column appended after the
// structured columns is detected as a free column (absolute index 5).
func TestFindFreeColsTotalCol(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()
	// colAnchors={"Vùng miền"}, colHeaderDepth=2, rowHeaderDepth=3

	matrix := [][]string{
		{"", "", "", "Vùng miền", "Vùng miền", "Tổng"},
		{"", "", "", "Miền Bắc", "Miền Nam", ""},
		{"Doanh thu", "Hình thức", "Bán lẻ", "100", "150", "250"},
		{"Chi phí", "Loại CP", "Vận hành", "80", "120", "200"},
	}

	freeRows, freeCols := bm.findFreeRowsCols(matrix, 0, 0)
	if len(freeRows) != 0 {
		t.Errorf("freeRows = %v, want []", freeRows)
	}
	if len(freeCols) != 1 || freeCols[0] != 5 {
		t.Errorf("freeCols = %v, want [5]", freeCols)
	}
}

// TestFindFreeRowsTotalRow verifies that a "Tổng cộng" row inserted between
// structural rows is detected as a free row (absolute index 3).
func TestFindFreeRowsTotalRow(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()
	// rowAnchors={"Doanh thu","Chi phí"}, colHeaderDepth=2

	matrix := [][]string{
		{"", "", "", "Vùng miền", "Vùng miền"},
		{"", "", "", "Miền Bắc", "Miền Nam"},
		{"Doanh thu", "Hình thức", "Bán lẻ", "100", "150"},
		{"Tổng cộng", "", "", "180", "270"},
		{"Chi phí", "Loại CP", "Vận hành", "80", "120"},
	}

	freeRows, freeCols := bm.findFreeRowsCols(matrix, 0, 0)
	if len(freeCols) != 0 {
		t.Errorf("freeCols = %v, want []", freeCols)
	}
	if len(freeRows) != 1 || freeRows[0] != 3 {
		t.Errorf("freeRows = %v, want [3]", freeRows)
	}
}

// TestFindFreeRowsCols_StrictValues verifies that with StrictValues=true, a
// PhanTo value absent from the predefined list is treated as a free column.
func TestFindFreeRowsCols_StrictValues(t *testing.T) {
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()
	// PhanToChungs["Vùng miền"].Values = ["Miền Bắc","Miền Nam"] after setup.
	for _, ptc := range bm.PhanToChungs {
		if ptc.Name == "Vùng miền" {
			ptc.StrictValues = true
		}
	}

	// Col 5: "Miền Trung" ∉ Values → free.
	matrix := [][]string{
		{"", "", "", "Vùng miền", "Vùng miền", "Vùng miền"},
		{"", "", "", "Miền Bắc", "Miền Nam", "Miền Trung"},
		{"Doanh thu", "Hình thức", "Bán lẻ", "100", "150", "110"},
		{"Chi phí", "Loại CP", "Vận hành", "80", "120", "90"},
	}

	freeRows, freeCols := bm.findFreeRowsCols(matrix, 0, 0)
	if len(freeRows) != 0 {
		t.Errorf("freeRows = %v, want []", freeRows)
	}
	if len(freeCols) != 1 || freeCols[0] != 5 {
		t.Errorf("freeCols = %v, want [5]", freeCols)
	}
}

// ─── extractSubMatrix tests ──────────────────────────────────────────────────

func TestExtractSubMatrix(t *testing.T) {
	matrix := [][]string{
		{"A", "B", "C", "D", "E"},
		{"F", "G", "H", "I", "J"},
		{"K", "L", "M", "N", "O"},
		{"P", "Q", "R", "S", "T"},
	}

	// From (1,1), skip abs-row 3 and abs-col 3.
	got, err := extractSubMatrix(matrix, 1, 1, []int{3}, []int{3})
	if err != nil {
		t.Fatalf("extractSubMatrix() error = %v", err)
	}
	want := [][]string{
		{"G", "H", "J"},
		{"L", "M", "O"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extractSubMatrix() = %v, want %v", got, want)
	}
}

func TestExtractSubMatrixNoFree(t *testing.T) {
	matrix := [][]string{{"1", "2"}, {"3", "4"}}
	got, err := extractSubMatrix(matrix, 0, 0, nil, nil)
	if err != nil {
		t.Fatalf("extractSubMatrix() error = %v", err)
	}
	if !reflect.DeepEqual(got, matrix) {
		t.Errorf("extractSubMatrix() = %v, want %v", got, matrix)
	}
}

// ─── importFromMatrixAuto integration test ───────────────────────────────────

// TestImportFromMatrixAutoFull tests the complete auto-import pipeline:
// the raw matrix has 2 extra header rows, 1 extra header col, an appended
// free "Tổng" column, and an inserted free "Tổng cộng" row. The resulting
// BangChiTieus must match a reference import from the clean sub-matrix.
func TestImportFromMatrixAutoFull(t *testing.T) {
	// Reference: import from the clean matrix directly.
	ref := createSampleBieuMau()
	ref.HeaderType = "matrix_chitieu_in_rows"
	ref.setupFull()
	ref.importFromMatrix([][]string{
		{"", "", "", "Vùng miền", "Vùng miền"},
		{"", "", "", "Miền Bắc", "Miền Nam"},
		{"Doanh thu", "Hình thức", "Bán lẻ", "100", "150"},
		{"Chi phí", "Loại CP", "Vận hành", "80", "120"},
	})

	// Auto-import from the raw shifted+polluted matrix.
	bm := createSampleBieuMau()
	bm.HeaderType = "matrix_chitieu_in_rows"
	bm.setupFull()

	// Table starts at abs (2,1); abs-col 6 is the "Tổng" free col;
	// abs-row 5 is the "Tổng cộng" free row.
	rawMatrix := [][]string{
		{"TIÊU ĐỀ", "", "", "", "", "", ""},
		{"", "", "", "", "", "", ""},
		{"", "", "", "", "Vùng miền", "Vùng miền", "Tổng"},
		{"", "", "", "", "Miền Bắc", "Miền Nam", ""},
		{"", "Doanh thu", "Hình thức", "Bán lẻ", "100", "150", "250"},
		{"", "Tổng cộng", "", "", "180", "270", "450"},
		{"", "Chi phí", "Loại CP", "Vận hành", "80", "120", "200"},
	}

	if err := bm.importFromMatrixAuto(rawMatrix); err != nil {
		t.Fatalf("importFromMatrixAuto() error = %v", err)
	}

	if len(bm.BangChiTieus) != len(ref.BangChiTieus) {
		t.Fatalf("len(BangChiTieus) = %d, want %d", len(bm.BangChiTieus), len(ref.BangChiTieus))
	}

	type dimKey struct{ vung, phanToVal string }
	extractSolieu := func(bcts []*BangChiTieu) map[string]map[dimKey]string {
		res := make(map[string]map[dimKey]string)
		for _, bct := range bcts {
			m := make(map[dimKey]string)
			for _, ddl := range bct.DongDuLieus {
				var vung, pt string
				for _, kv := range ddl.Dims {
					if kv.Key == "Vùng miền" {
						vung = kv.Value
					} else {
						pt = kv.Value
					}
				}
				m[dimKey{vung, pt}] = ddl.Solieu
			}
			res[bct.ChiTieuName] = m
		}
		return res
	}

	gotSL := extractSolieu(bm.BangChiTieus)
	wantSL := extractSolieu(ref.BangChiTieus)
	for ctName, wantM := range wantSL {
		gotM, ok := gotSL[ctName]
		if !ok {
			t.Errorf("BangChiTieu %q missing", ctName)
			continue
		}
		for dk, wantVal := range wantM {
			if gotM[dk] != wantVal {
				t.Errorf("BangChiTieu[%s][%+v] = %q, want %q", ctName, dk, gotM[dk], wantVal)
			}
		}
	}
}
