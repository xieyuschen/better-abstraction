package dao

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type UserDao interface {
	GetUserName(context.Context,int) (string,error)
}

type Cache interface {
	GetUserName(context.Context,int) (string,bool,error)
}

type DB interface {
	GetUserName(context.Context,int) (string,bool,error)
}

// UserEntity implement the UserDao interface.
type UserEntity struct {
	// consider timeout, no hit in cache layer
	QueryCache Cache
	QueryDb DB
}

func (ue *UserEntity) GetUserName(ctx context.Context,id int) (string,error){
	name,hit,err:=ue.QueryCache.GetUserName(ctx,id)
	if err!=nil{
		return "", err
	}
	if hit{
		return name,nil
	}
	name,queried,err:= ue.QueryDb.GetUserName(ctx,id)
	if err!=nil{
		return name,err
	}
	if queried{
		return name,nil
	}
	return "",errors.New("no record is found")
}

type queryCache struct {}

func WrapId(source string,id int)string{
	return fmt.Sprintf("get a name from %s: %d",source,id)
}

// GetUserName mock the data cache such as redis,only id between [1,5] could hit.
// [6,7] will sleep 1 second and hit,[8,10] will sleep 1s and not hit
func (c *queryCache) GetUserName(ctx context.Context,id int) (string,bool,error){
	_,cancel:=context.WithCancel(ctx)
	hitCh:=make(chan bool)
	errCh:=make(chan error)
	nameCh:=make(chan string)
	go func(){
		defer cancel()
		if id>=1&&id<=5{
			nameCh<-WrapId("cache",id)
			hitCh<-true
			errCh<-nil
			return
		}

		if id>=6&&id<=7{
			time.Sleep(time.Second)
			nameCh<-WrapId("cache",id)
			hitCh<-true
			errCh<-nil
		}

		if id>=8&&id<=10{
			time.Sleep(time.Second)
			nameCh<-WrapId("cache",id)
			hitCh<-false
			errCh<-nil
		}
		nameCh<-""
		hitCh<-false
		errCh<-nil
	}()
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return "",false,fmt.Errorf("query cache timeout: %w",context.DeadlineExceeded)
		case context.Canceled:
			return <-nameCh,<-hitCh,<-errCh
		}
	}
	return "", false, nil
}

type queryDB struct {}

var ErrNotFound=errors.New("db record not is found")
// GetUserName mock the data db such as redis,only id between [1,5] exist.
// [6,8] will sleep 1 second and return value. [9,10] will sleep 1s and return a ErrNotFound error
// The other will not be queried.
func (c *queryDB) GetUserName(ctx context.Context,id int) (string,bool,error){
	_,cancel:=context.WithCancel(ctx)
	hitCh:=make(chan bool)
	errCh:=make(chan error)
	nameCh:=make(chan string)
	go func(){
		defer cancel()
		if id>0&&id<6{
			nameCh<-WrapId("db",id)
			hitCh<-true
			errCh<-nil
			return
		}
		if id>5&&id<9{
			time.Sleep(time.Second)
			nameCh<-WrapId("db",id)
			hitCh<-true
			errCh<-nil
		}

		if id>=9&&id<=10{
			time.Sleep(time.Second)
			nameCh<-WrapId("db",id)
			hitCh<-true
			errCh<-ErrNotFound
		}
		nameCh<-""
		hitCh<-false
		errCh<-ErrNotFound
	}()
	select {
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return "",false,fmt.Errorf("query cache timeout: %w",context.DeadlineExceeded)
		case context.Canceled:
			return <-nameCh,<-hitCh,<-errCh
		}
	}

	//unreachable, should be better if receive values from xxxCh and return here.
	return "", false, nil
}

var defaultQueryCache Cache=&queryCache{}
var defaultQueryDb DB=&queryDB{}

type Option func(entity *UserEntity)

func NewUserDao(opts... Option) UserDao{
	entity:= UserEntity{
		QueryCache: defaultQueryCache,
		QueryDb: defaultQueryDb,
	}
	for _,f:=range opts{
		f(&entity)
	}
	return &entity
}

