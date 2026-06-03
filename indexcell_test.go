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
