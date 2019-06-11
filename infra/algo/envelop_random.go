package algo

import (
	"math/rand"
	"time"
)


func generate(count int64, min int64, amount int64, seeds* []interface{}) {
	if count == 1 {
		*seeds = append(*seeds, amount)
		return
	}
	max := amount - count * min
	//计算最大可调度金额
	rand.Seed(time.Now().UnixNano())
	money := rand.Int63n(max) + min
	*seeds = append(*seeds, money)
	generate(count - 1, min, amount - money, seeds)
}


type EnvelopRandomStrategy interface {
	MinMoney () int64
	Generate (count int64, amount int64, seeds* []interface{})
}

type EnvelopSimpleRandom struct {}

func (this *EnvelopSimpleRandom) MinMoney() int64 {
	return 1
}

func (this *EnvelopSimpleRandom) Generate (count int64, amount int64, seeds* []interface{}) {
	generate(count, this.MinMoney(), amount, seeds)
}

type EnvelopAfterSuffleStrategy struct {

}

func (this *EnvelopAfterSuffleStrategy) MinMoney() int64 {
	return 1
}

func (this *EnvelopAfterSuffleStrategy) Generate (count int64, amount int64, seeds* []interface{}) {
	generate(count, this.MinMoney(), amount, seeds)

	rand.Shuffle(len(*seeds), func(i, j int) {
		(*seeds)[i], (*seeds)[j] = (*seeds)[j], (*seeds)[i]
	})
}

type EnvelopDoubleAvgStrategy struct {}

func (this *EnvelopDoubleAvgStrategy) MinMoney() int64 {
	return 1
}

func (this *EnvelopDoubleAvgStrategy) Generate (count int64, amount int64, seeds* []interface{}) {

	if count == 1 {
		*seeds = append(*seeds, amount)
		if len(*seeds) != 1 {
			rand.Shuffle(len(*seeds), func(i, j int) {
				(*seeds)[i], (*seeds)[j] = (*seeds)[j], (*seeds)[i]
			})
		}
		return
	}

	max := amount - count * this.MinMoney()
	avg := max / count
	avg2 := avg * 2
	rand.Seed(time.Now().UnixNano())
	money := rand.Int63n(avg2) + this.MinMoney()
	*seeds = append(*seeds, money)
	this.Generate(count - 1, amount - money, seeds)
}