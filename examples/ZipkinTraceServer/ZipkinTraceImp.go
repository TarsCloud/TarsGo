package main

import "context"
import "time"

//ZipkinTraceImp struct
type ZipkinTraceImp struct {
}

//Add implement
func (imp *ZipkinTraceImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	*c = a * b
	time.Sleep(500 * time.Millisecond)
	//...
	return 0, nil
}

//Sub implement
func (imp *ZipkinTraceImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	*c = a / b
	//...
	return 0, nil
}
