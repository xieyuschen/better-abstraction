package dao

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type TestSetWithErrorResult struct {
	QueryId int
	ExpectedResult error
}

var ErrResultSets = []TestSetWithErrorResult{
	{
		QueryId: 6,
		ExpectedResult: context.DeadlineExceeded,
	},
	{
		QueryId:11,
		ExpectedResult: ErrNotFound,
	},
	{
		QueryId: 10,
		ExpectedResult: context.DeadlineExceeded,
	},
}
func Test_DaoWithError(t *testing.T) {
	dao:=NewUserDao()
	wg:=sync.WaitGroup{}
	wg.Add(len(ErrResultSets))
	for _,val:=range ErrResultSets{
		cpVal:=val
		go func() {
			defer wg.Done()
			ctx,cancel:=context.WithTimeout(context.Background(),1500*time.Millisecond)
			defer cancel()
			_,err:=dao.GetUserName(ctx,6)
			// put the cpVal.ExpectedResult first as the error might be wrapped.
			if errors.Is(cpVal.ExpectedResult,err){
				t.Errorf("failed when query id:%d, expect:%s, but receive:%s\n",cpVal.QueryId,cpVal.ExpectedResult,err)
			}
		}()
	}
	wg.Wait()
}
