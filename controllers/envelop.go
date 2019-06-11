package controllers

import (
	"encoding/json"
	"envelop/models"
	"envelop/service"
	"github.com/go-redis/redis"
	redisClient "envelop/redis"
)

var (
	//envelopRandomStrategy = algo.EnvelopSimpleRandom{}
	//envelopRandomAfterSuffleStrategy = algo.EnvelopAfterSuffleStrategy{}
	//envelopRandomDoubleStrategy = algo.EnvelopDoubleAvgStrategy{}
	envelopService = new (service.EnvelopServiceImpl)
)

type EnvelopController struct {
	BaseController
}


// @router / [post]
func (this * EnvelopController) Create () {
	var envelop models.Envelop
	json.Unmarshal(this.Ctx.Input.RequestBody, &envelop)
	orderNo, err:= envelopService.CreateEnvelop(&envelop)
	this.apiResponse(0, err, orderNo)
}

// @router /take [post]
func (this * EnvelopController) Take () {
	var takeEnvelopVo models.TakeEnvelopVo
	json.Unmarshal(this.Ctx.Input.RequestBody, &takeEnvelopVo)
	envelop, err:= envelopService.TakeEnvelop(takeEnvelopVo)
	this.apiResponse(0, err, envelop)
}


// @router /test [get]
func (this *EnvelopController) TestRandomAlog () {
	res, err := redisClient.Client.ZAddNX("1557452205052554000", redis.Z{
		Score: 0,
		Member : 10,
	}).Result()

	this.apiResponse(0, err, res)
}
//
//// @router /test-aftersuffle [get]
//func (this *EnvelopController) TestRandomAlogAfterSuffle () {
//	seeds := make([]int64, 0)
//	envelopRandomAfterSuffleStrategy.Generate(10, 100, &seeds)
//	logs.Info("seeds", seeds)
//	this.apiResponse(0, nil, seeds)
//}
//
//
//// @router /double-aftersuffle [get]
//func (this *EnvelopController) TestRandomAlogDouble () {
//	seeds := make([]int64, 0)
//	envelopRandomDoubleStrategy.Generate(10, 500, &seeds)
//	logs.Info("seeds", seeds)
//	this.apiResponse(0, nil, seeds)
//}

