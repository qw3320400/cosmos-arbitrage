package graph

// func TestFindCircle(t *testing.T) {
// 	pools := []*Pool{
// 		&Pool{
// 			ID:     1,
// 			Denoms: []string{"A", "B"},
// 		},
// 		&Pool{
// 			ID:     2,
// 			Denoms: []string{"B", "C"},
// 		},
// 		&Pool{
// 			ID:     3,
// 			Denoms: []string{"C", "A"},
// 		},
// 		&Pool{
// 			ID:     4,
// 			Denoms: []string{"C", "D"},
// 		},
// 		&Pool{
// 			ID:     5,
// 			Denoms: []string{"D", "A"},
// 		},
// 		&Pool{
// 			ID:     6,
// 			Denoms: []string{"B", "D"},
// 		},
// 	}
// 	circles := []*Circle{}
// 	circles = findPath(pools, "A", "A", 5, []*Pool{}, circles)
// 	for i, circle := range circles {
// 		t.Log("circle", i, circle.Path)
// 	}
// }

// func BenchmarkDij(b *testing.B) {
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		runDji(1, vertex)
// 	}
// }
