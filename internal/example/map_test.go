package example

import (
	"github.com/ameise84/rbmap"
	"log"
	"math/rand"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	//rand.Seed(1)
	count := int32(180)
	base := make([]int32, count)
	ids := make([]int32, count)
	for i := int32(0); i < count; i++ {
		ids[i] = i
		base[i] = rand.Int31()
	}

	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	mp := rbmap.New[int32, int32]()
	var list []int32
	for i := int32(0); i < count; i++ {
		n := rand.Int31n(10)
		if n < 8 || len(list) == 0 {
			mp.Store(ids[i], base[i])
			list = append(list, ids[i])
		} else {
			n2 := rand.Intn(len(list))
			mp.Delete(list[n2])
			list = append(list[:n2], list[n2+1:]...)
		}
	}
	//mp.PrintTree()
	//log.Println("===========================")
	//mp.PrintTreeIter()
	//log.Println("===========================")
	var iter rbmap.Iterator[int32, int32]
	for i := -2; i <= 64; i++ {
		iter = mp.Search(int32(i), rbmap.SearchModeET|rbmap.SearchModeLT)
		if iter == mp.EndIterator() {
			log.Println(strconv.Itoa(i) + "  -->  " + "end")
		} else {
			key, _ := iter.Key()
			log.Println(strconv.Itoa(i) + "  -->  " + strconv.Itoa(int(key)))
		}
	}

	log.Println("===========================")
	rand.Shuffle(mp.Size(), func(i, j int) {
		list[i], list[j] = list[j], list[i]
	})
	fix := rand.Intn(len(list))
	idx := list[fix]
	iter = mp.Search(idx)
	key, _ := iter.Key()
	value, _ := iter.Value()
	log.Println(strconv.Itoa(int(key)) + "  -->  " + strconv.Itoa(int(value)))
	for i := 0; i < len(list); i++ {
		if list[i] != idx {
			mp.Delete(list[i])
		}
	}
	key, _ = iter.Key()
	value, _ = iter.Value()
	log.Println(strconv.Itoa(int(key)) + "  -->  " + strconv.Itoa(int(value)))
}
