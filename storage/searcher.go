package storage

import (
	"fmt"
	"github.com/go-ego/riot"
	"github.com/go-ego/riot/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"strconv"
)

type Searcher struct {
	engine *riot.Engine
}

func (Searcher) Init(name string) *Searcher {
	searcher := new(Searcher)
	searcher.engine = &riot.Engine{}
	searcher.engine.Init(types.EngineOpts{
		Model:     "zh",
		Using:     4,
		NotUseGse: true,
		//GseMode: true,
		//GseDict: viper.GetString("searcher.dict_path"),
		GseDict: viper.GetString(fmt.Sprintf("searcher.%s.dict_path", name)),
		IndexerOpts: &types.IndexerOpts{
			IndexType: types.DocIdsIndex,
		},
		UseStore: true,
		//StoreFolder: viper.GetString("searcher.store_path"),
		StoreFolder: viper.GetString(fmt.Sprintf("searcher.%s.store_path", name)),
		StoreEngine: "bg", // bg: badger, lbd: leveld
	})
	searcher.engine.Flush()
	logrus.Infoln("Searcher: " + name + " Init")
	return searcher
	//go func() {
	//	ticker := time.NewTicker(time.Duration(10) * time.Second)
	//	defer ticker.Stop()
	//	for {
	//		select {
	//		case <-ticker.C:
	//			this.engine.Flush()
	//			fmt.Println("Search Flush...")
	//		}
	//	}
	//}()
}

func (this *Searcher) RemoveIndex(docId int) {
	go func() {
		this.engine.RemoveDoc(strconv.Itoa(docId), true)
	}()
}

func (this *Searcher) Indexer(docId int, text string) {
	this.engine.Index(strconv.Itoa(docId), types.DocData{Content: text}, true)
	this.engine.Flush()
}

//func (this *Searcher) SearchApplyByFields(data map[string]string, fields []string, size int) (applyIds []int) {
//	applyIds = make([]int, 0)
//
//	fmt.Println("进入冲突查询函数")
//	_search := func(key, val string, wg *sync.WaitGroup) {
//		defer wg.Done()
//
//		result := this.engine.Search(types.SearchReq{
//			Text: fmt.Sprintf("%s %s", key, val),
//			RankOpts: &types.RankOpts{
//				OutputOffset: 0,
//				MaxOutputs:   size,
//			},
//			Logic: types.Logic{
//				Must: true,
//			},
//		})
//
//		ids := make([]int, 0)
//		if result.Docs.(types.ScoredDocs).Len() > 0 {
//			for _, d := range result.Docs.(types.ScoredDocs) {
//				_id, _ := strconv.Atoi(d.DocId)
//				ids = append(ids, _id)
//				applyIds = append(applyIds, _id)
//			}
//		}
//		fmt.Println(key, val, ids)
//	}
//
//	inArray := func(str string, strs []string) bool {
//		for _, s := range strs {
//			if s == str {
//				return true
//			}
//		}
//		return false
//	}
//
//	fieldNum := 0
//	var wg sync.WaitGroup
//	for key, val := range data {
//		if val == "" {
//			continue
//		}
//		if !inArray(key, fields) {
//			continue
//		}
//
//		fieldNum++
//		wg.Add(1)
//		_search(key, val, &wg)
//	}
//	wg.Wait()
//
//	if len(applyIds) == 0 || fieldNum == 0 {
//		fmt.Println("没有找到冲突数据")
//		return
//	}
//
//	scores := make(map[int]int)
//	for _, id := range applyIds {
//		times, found := scores[id]
//		if !found {
//			times = 0
//		}
//		times++
//		scores[id] = times
//	}
//
//	fmt.Println("scores:", scores)
//
//	applyIds = make([]int, 0)
//	for id, times := range scores {
//		rate := float64(times) / float64(fieldNum)
//		fmt.Println("Total:", float64(fieldNum), "Times:", float64(times), "Rate:", rate, "score:", viper.GetFloat64("searcher.apply_score"))
//		if rate >= viper.GetFloat64("searcher.apply_score") {
//			applyIds = append(applyIds, id)
//		}
//	}
//
//	fmt.Println("applyIds", applyIds)
//	return
//}

func (this *Searcher) Search(keywords string, offset, size int) ([]int, int) {
	result := this.engine.Search(types.SearchReq{
		Text: keywords,
		RankOpts: &types.RankOpts{
			OutputOffset: offset,
			MaxOutputs:   size,
		},
	})

	docIds := make([]int, 0)
	docs := result.Docs.(types.ScoredDocs)
	if docs.Len() > 0 {
		for _, d := range docs {
			_id, _ := strconv.Atoi(d.DocId)
			docIds = append(docIds, _id)
		}
	}

	return docIds, result.NumDocs
}
