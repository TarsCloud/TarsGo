package main

import (
	"context"
)

type _SERVANT_Imp struct {
}

func (imp *_SERVANT_Imp) Init() (int, error) {
	//initialize servant here:
	//...
	return 0,nil
}

//////////////////////////////////////////////////////
func (imp *_SERVANT_Imp) Destroy() {
	//destroy servant here:
	//...
}

func (imp *_SERVANT_Imp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
func (imp *_SERVANT_Imp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
