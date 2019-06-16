package controllers

import (
	"encoding/json"
	"envelop/models"
	"envelop/service"
	"github.com/go-redis/redis"
	redisClient "envelop/redis"
)

type EnvelopController struct {
	BaseController
	EnvelopService *service.EnvelopServiceImpl `inject:""`
}


// @router / [post]
func (this * EnvelopController) Create () {
	var envelop models.Envelop
	json.Unmarshal(this.Ctx.Input.RequestBody, &envelop)
	data, err:= this.EnvelopService.CreateEnvelop(&envelop)
	this.apiResponse(0, err, data)
}

// @router /take [post]
func (this * EnvelopController) Take () {
	var takeEnvelopVo models.TakeEnvelopVo
	json.Unmarshal(this.Ctx.Input.RequestBody, &takeEnvelopVo)
	envelop, err:= this.EnvelopService.TakeEnvelopNew(&takeEnvelopVo)
	this.apiResponse(0, err, envelop)
}

// @router /query [post]
func (this * EnvelopController) Query () {
	var queryEnvelopVo models.QueryEnvelopVo
	json.Unmarshal(this.Ctx.Input.RequestBody, &queryEnvelopVo)
	envelop, err:= this.EnvelopService.QueryEnvelop(&queryEnvelopVo)
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

