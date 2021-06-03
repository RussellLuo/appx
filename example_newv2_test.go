package appx_test

import (
	"context"
	"fmt"
	"time"

	"github.com/RussellLuo/appx"
)

// Interface guards
var (
	_ appx.InitCleaner  = (*A)(nil)
	_ appx.StartStopper = (*A)(nil)

	_ appx.InitCleaner  = (*B)(nil)
	_ appx.StartStopper = (*B)(nil)
)

type A struct {
	Name  string
	Value string
}

func (a *A) Init(ctx appx.Context) error {
	a.Name = ctx.App.Name
	a.Value = "value_a"

	fmt.Printf("Initializing app %q, which requires %d app\n", a.Name, len(ctx.Required))
	return nil
}

func (a *A) Clean() error {
	fmt.Printf("Cleaning up app %q\n", a.Name)
	return nil
}

func (a *A) Start(ctx context.Context) error {
	fmt.Printf("Starting app %q\n", a.Name)
	return nil
}

func (a *A) Stop(ctx context.Context) error {
	fmt.Printf("Stopping app %q\n", a.Name)
	return nil
}

type B struct {
	Name  string
	Value string
}

func (b *B) Init(ctx appx.Context) error {
	b.Name = ctx.App.Name
	b.Value = "value_b"

	a := ctx.Required["a2"].Instance().(*A)
	fmt.Printf("Initializing app %q, which requires app %q, whose value is %q\n", b.Name, a.Name, a.Value)
	return nil
}

func (b *B) Clean() error {
	fmt.Printf("Cleaning up app %q\n", b.Name)
	return nil
}

func (b *B) Start(ctx context.Context) error {
	fmt.Printf("Starting app %q\n", b.Name)
	return nil
}

func (b *B) Stop(ctx context.Context) error {
	fmt.Printf("Stopping app %q\n", b.Name)
	return nil
}

func Example_newV2() {
	// Typically located in `func init()` of package a.
	appx.MustRegister(appx.NewV2("a2", new(A)))

	// Typically located in `func init()` of package b.
	appx.MustRegister(appx.NewV2("b2", new(B)).Require("a2"))

	ctx := context.Background()

	// Typically located in `func main()` of package main.
	if err := appx.Install(ctx, "b2"); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	defer appx.Uninstall()

	// In a typical scenario, we could just use appx.Run() here. Since we
	// don't want this example to run forever, we'll use the more explicit
	// Start and Stop.
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := appx.Start(startCtx); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println("Everything is running")

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	appx.Stop(stopCtx)

	// Output:
	// Initializing app "a2", which requires 0 app
	// Initializing app "b2", which requires app "a2", whose value is "value_a"
	// Starting app "a2"
	// Starting app "b2"
	// Everything is running
	// Stopping app "b2"
	// Stopping app "a2"
	// Cleaning up app "b2"
	// Cleaning up app "a2"
}
