package service

import "github.com/facebookgo/inject"

func MustInit(g *inject.Graph) {

	g.Provide(
		&inject.Object{Value: &TransferStrategyPlat2Plat{}},
		&inject.Object{Value: &TransferStrategyPlat2UnionPay{}},
		&inject.Object{Value: &TransferStrategyUnionPay2UnionPay{}},
		&inject.Object{Value: &TransferStrategyUnionPay2Plat{}},
		&inject.Object{Value: &AccountServiceImpl{}},
		&inject.Object{Value: &EnvelopServiceImpl{}},
		&inject.Object{Value: &UserServiceImpl{}},
	)
}
