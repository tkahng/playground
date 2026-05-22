//go:build !integration

package dataloader

// func TestDataloader_Load(t *testing.T) {
// 	ctx := context.Background()
// 	keys := []int{1, 2, 3}
// 	d := New(func(ctx context.Context, keys []int) ([]int, error) {
// 		return keys, nil
// 	})
// 	var newKeys []int
// 	for _, key := range keys {
// 		go func(key int) {
// 			res, err := d.Load(ctx, key)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			newKeys = append(newKeys, res)
// 		}(key)
// 	}
// 	err := d.Wait(ctx)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if len(newKeys) != len(keys) {
// 		t.Errorf("Expected %d keys, got %d", len(keys), len(newKeys))
// 	}
// }
