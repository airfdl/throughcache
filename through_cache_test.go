package throughcache

import (
	"context"
	"fmt"
	"testing"
)

func TestThroughCache(t *testing.T) {
	cacher := CacheTODO()
	baser := BaseTODO()
	o1 := WithSyncSetCache()
	o2 := WithSetCache()
	o3 := WithMustToBaser()
	o4 := WithSetEmptyValueCahe()
	x := NewThroughCache("testing", baser, cacher, o1, o2, o3, o4)
	_, _ = x.MGetValue(context.TODO(), nil)
	fmt.Println(x.isSetCache())
	fmt.Println(x.isSetEmptyValueCache())
	fmt.Println(x.syncSetCache())
	fmt.Println(x.mustToBaser())
	fmt.Printf("%b\n", x.Options.bit1)
}
