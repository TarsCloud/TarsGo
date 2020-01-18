package main

import "context"

//ZipkinClientImp struct
type ZipkinClientImp struct {
}

//Add implemnet
func (imp *ZipkinClientImp) Add(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	*c = a * b
	var d int32
	_, err := sapp.AddWithContext(ctx, *c, *c, &d)
	if err != nil {
		logger.Error("Error call add ", err)
	}

	return 0, nil
}

//Sub implement
func (imp *ZipkinClientImp) Sub(ctx context.Context, a int32, b int32, c *int32) (int32, error) {
	//Doing something in your function
	//...
	return 0, nil
}
