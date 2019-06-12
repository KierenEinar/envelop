package dao

import (
	"database/sql"
	"envelop/constant"
	"envelop/models"
	"envelop/redis"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/facebookgo/inject"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	dataSourceManager *DataSourceManager
)

type DataSourceManager struct {
	dataSource *DataSource
}

func initManager() {

	if dataSourceManager == nil {
		mutex := sync.Mutex{}
		mutex.Lock()
		defer mutex.Unlock()
		if dataSourceManager == nil {
			dataSourceManager = new (DataSourceManager)
		}
	}
}

func Db(this * DataSourceManager) *sql.DB {
	if this.dataSource == nil {
		mutex := sync.Mutex{}
		mutex.Lock()
		defer mutex.Unlock()
		if this.dataSource == nil {
			this.initDataSource()
		}
	}
	return this.dataSource.ConnectionPool.db
}


//开始一段数据库连接
func init () {

	initManager()

	err := dataSourceManager.initDataSource()

	if err != nil {
		log.Fatal("open mysql failed ...", err)
	}

	go observeSignal(func() {
		dataSourceManager.dataSource.Close()
		err := redis.Client.Close()
		logs.Info("redis conenct close ... ", err )
	})

}

func MustInit (g *inject.Graph) {
	g.Provide(
		&inject.Object{Value: &UserDaoImpl{}},
		&inject.Object{Value: &AccountDaoImpl{}},
		&inject.Object{Value: &AccountHistoryDaoImpl{}},
		&inject.Object{Value: &AccountBankTransferHistoryDaoImpl{}},
		&inject.Object{Value: &EnvelopDaoImpl{}},
		)
}

func (this * DataSourceManager) initDataSource () error {

	dataSource := new (DataSource)

	dataSource.Register("tcp(127.0.0.1:3306)/envelop_db",
		"root",
		"root",
		"mysql",
		16,
		200,
		time.Second * 60)

	err := dataSource.Open()
	this.dataSource = dataSource
	return err
}

func observeSignal (f func()) {
	c :=  make(chan os.Signal, 1)
	signal.Notify(c)
	s:=<-c
	fmt.Println("get signal :", s)
	if s == syscall.SIGINT || s == syscall.SIGKILL {
		f()
		os.Exit(0)
	}

}


type DataSource struct {
	Url string
	User string
	Pass string
	DriverName string
	MaxIdle int
	MaxConns int
	MaxOpenTime time.Duration
	ConnectionPool *ConnectionPool
}


func (this *DataSource) Register(Url string,
									User string,
									Pass string,
									DriverName string,
									MaxIdle int,
									MaxConns int,
									MaxOpenTime time.Duration) {

	this.User = User
	this.Url = Url
	this.Pass = Pass
	this.DriverName = DriverName
	this.MaxConns = MaxConns
	this.MaxIdle = MaxIdle
	this.MaxOpenTime = MaxOpenTime
}

type ConnectionPool struct {
	db *sql.DB
	orm *gorm.DB
}

func (this *DataSource ) Open () (error) {
	dns := fmt.Sprintf("%s:%s@%s", this.User, this.Pass, this.Url)
	log.Println("dns", dns)
	db, err := sql.Open(this.DriverName, dns)

	if err != nil {
		return err
	}

	db.SetConnMaxLifetime(this.MaxOpenTime)
	db.SetMaxIdleConns(this.MaxIdle)
	db.SetMaxOpenConns(this.MaxConns)


	orm, err := gorm.Open(this.DriverName, dns)
	if err!= nil {
		return err
	}

	pool := ConnectionPool{
		db,
		orm,
	}

	this.ConnectionPool = &pool
	return nil
}

func (this *DataSource ) Close() {
	if this.ConnectionPool != nil {
		if this.ConnectionPool.db != nil {
			this.ConnectionPool.db.Close()
			log.Print("close db connection ...")
		}
		if this.ConnectionPool.orm != nil {
			this.ConnectionPool.orm.Close()
			log.Print("close db connection ...")
		}
	}
}

type BaseDao struct {}

func (this *BaseDao) GetPool() *ConnectionPool {
	return dataSourceManager.dataSource.ConnectionPool
}

func (this *BaseDao) Tx(handler func (tx *sql.Tx) (error) )  (error) {
	db:=this.GetPool().db
	tx, _ := db.Begin()
	return handler(tx)
}

type SQLInsert struct {
	Tx *sql.Tx
	Prepare string
	Args[] interface{}
	Error constant.RuntimeError
}


func (this *BaseDao) Insert (sqlInsert *SQLInsert) (int64, error) {

	stmt, error:=sqlInsert.Tx.Prepare(sqlInsert.Prepare)
	if error != nil {
		return 0, error
	}
	res, err:= stmt.Exec(sqlInsert.Args[:]...)

	if err != nil {
		logs.Error(err)
		return 0, &sqlInsert.Error
	}
	rows,_:= res.RowsAffected()
	return rows, nil
}

type UserDao interface {
	CreateUser(user *models.User) (int, error)
	FindUser (id uint64) (*models.User, error)
	FindUsers () ([] *models.User, error)
}

type UserDaoImpl struct{
	BaseDao
}

func (this* UserDaoImpl) CreateUser(user *models.User) (int64, error) {
	pool := this.GetPool()
	user.CreateTime = time.Now().Unix()
	db:= pool.orm.Create(user)
	return db.RowsAffected, db.Error
}

func (this* UserDaoImpl) FindUser (id uint64) (*models.User, error) {
	pool := this.GetPool()
	user := new (models.User)
	err  := pool.orm.First(&user, id).Error
	if gorm.IsRecordNotFoundError(err) {
		user = nil
		err = &constant.RuntimeError{
			constant.UserNotFoundErrorCode,
			"user not found",
		}
	}
	return user, err
}


func (this* UserDaoImpl) FindUsers () ([] *models.User, error) {
	return nil,nil
}


type AccountDao interface {
	CreateAccount (account *models.Account) (int64, error)
	UpdateAccountBalance (tx *sql.Tx, userId uint64, amount uint64) (int64, error)
	FindIdByUserId(tx *sql.Tx, u uint64) (uint64, error)
}

type AccountDaoImpl struct {
	BaseDao
}

func (this *AccountDaoImpl) CreateAccount (account *models.Account) (int64, error) {
	pool:=this.GetPool()
	db:= pool.orm.Create(account)
	return db.RowsAffected, db.Error
}


func (this *AccountDaoImpl) UpdateAccountBalance (tx *sql.Tx, userId uint64, amount int64) (int64, error) {
	sql:= "update account set balance = balance + ?, update_time = now() where user_id = ?"
	stmt, error := tx.Prepare(sql)
	if error != nil {
		return 0, error
	}
	res, error := stmt.Exec(amount, userId)
	if error != nil {
		return 0, &constant.RuntimeError{
			constant.AccountBalanceErrorCode,
			"update failed",
		}
	}
	rows, _ :=res.RowsAffected()
	return rows, error
}

func (this *AccountDaoImpl) FindIdByUserId(tx *sql.Tx, u uint64) (uint64, error) {
	sql:= "select id from account where user_id = ?"
	stmt, error := tx.Prepare(sql)
	if error != nil {
		return 0, &constant.RuntimeError{
			constant.AccountBalanceErrorCode,
			"account not exist",
		}
	}
	var accountId uint64
	rs := stmt.QueryRow(u)

	rs.Scan(&accountId)
	return accountId, nil
}

type AccountHistoryDao interface {
	CreateAccountHistory (tx *sql.Tx, history* models.AccountHistory) (int64, error)
}

type AccountHistoryDaoImpl struct {
	BaseDao
}

func (this *AccountHistoryDaoImpl) CreateAccountHistory (tx *sql.Tx, history* models.AccountHistory) (int64, error) {
	sql:= "insert into account_history (user_id, account_id, trade_no, create_time, type, channel, currency, amount, pattern, description) values" +
		"(?,?,?,?,?,?,?,?,?,?);"

	args:=make([]interface{}, 0)

	args=append(args, history.UserId, history.AccountId, history.TradeNo, history.CreateTime, history.Type, history.Channel, history.Currency, history.Amount, history.Pattern, history.Description)
	sqlInsert := SQLInsert{
		tx,
		sql,
		args,
		constant.RuntimeError{
					constant.AccountBalanceErrorCode,
					"account_log insert failed ...",
		 },
	}


	return this.Insert(&sqlInsert)


}



type EnvelopDao interface {
	Create (tx *sql.Tx, envelop *models.Envelop) (error)
}

type EnvelopDaoImpl struct {
	BaseDao
}

func (this * EnvelopDaoImpl) Create (tx *sql.Tx, envelop *models.Envelop) (error) {
	//orm:= this.GetPool().orm
	//db:= orm.Create(envelop)
	//return db.RowsAffected, db.Error

	sql:= "insert into envelop (user_id, account_id, create_time, amount, type, quantity, version, pay_channel) values" +
		"(?,?,?,?,?,?,?,?);"

	args:=make([]interface{}, 0)

	args=append(args, envelop.UserId, envelop.AccountId, envelop.CreateTime, envelop.Amount, envelop.Type, envelop.Quantity, envelop.Version, envelop.PayChannel)
	sqlInsert := SQLInsert{
		tx,
		sql,
		args,
		constant.RuntimeError{
			constant.EnvelopCreateErrorCode,
			"envelop create failed ... ",
		},
	}


	rows, err:= this.Insert(&sqlInsert)

	if rows == 0 {
		return &constant.RuntimeError{
			constant.EnvelopCreateErrorCode,
			"envelop create failed ... ",
		}
	}

	return err
}


type AccountBankTransferHistoryDao interface {
	Create (tx *sql.Tx, history *models.AccountBankTransferHistory) (int64, error)
}

type AccountBankTransferHistoryDaoImpl struct {
	BaseDao
}


func (this * AccountBankTransferHistoryDaoImpl) Create (tx *sql.Tx, history *models.AccountBankTransferHistory) (int64, error) {
	//orm:= this.GetPool().orm
	//db:= orm.Create(envelop)
	//return db.RowsAffected, db.Error

	sql:= "insert into account_bank_transfer_history (trade_no, in_account_id, out_account_id, bank_no, bank_code, bank_name, create_time, amount) values" +
		"(?,?,?,?,?,?,?,?);"

	args:=make([]interface{}, 0)

	args=append(args, history.TradeNo, history.InAccountId, history.OutAccountId, history.BankNo, history.BankCode, history.BankName, history.CreateTime, history.Amount)
	sqlInsert := SQLInsert{
		tx,
		sql,
		args,
		constant.RuntimeError{
			constant.AccountBalanceErrorCode,
			"bank hisotry create failed",
		},
	}


	return this.Insert(&sqlInsert)
}



























