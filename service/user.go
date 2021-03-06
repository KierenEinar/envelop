package service

import (
	"encoding/json"
	"envelop/dao"
	"envelop/models"
	"github.com/astaxie/beego/logs"
)

type UserService interface {
	CreateUser (user *models.User) (int64, error)
	FindOne(id uint64) (*models.User, error)
}

type UserServiceImpl struct {
	UserDao *dao.UserDaoImpl `inject:""`
}

func (this *UserServiceImpl) CreateUser (user *models.User) (int64, error) {
	userjson, _ := json.Marshal(user)
	logs.Info("create user -> ", string(userjson))
	return this.UserDao.CreateUser(user)
}
func (this *UserServiceImpl) FindOne(id uint64) (*models.User, error) {
	logs.Info("find user, id -> ", id)
	return this.UserDao.FindUser(id)
}