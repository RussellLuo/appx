package appx_test

import (
	"context"
	"fmt"
	"time"

	"github.com/RussellLuo/appx"
)

// Interface guards
var (
	_ appx.Initializer  = (*A)(nil)
	_ appx.Cleaner      = (*A)(nil)
	_ appx.StartStopper = (*A)(nil)

	_ appx.Initializer  = (*B)(nil)
	_ appx.Cleaner      = (*B)(nil)
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

	a := ctx.MustLoad("a").(*A)
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

func Example() {
	r := appx.NewRegistry()

	// Typically located in `func init()` of package a.
	r.MustRegister(appx.New("a", new(A)))

	// Typically located in `func init()` of package b.
	r.MustRegister(appx.New("b", new(B)).Require("a"))

	// Typically located in `func main()` of package main.
	if err := r.Install(context.Background(), "b"); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	defer r.Uninstall()

	// In a typical scenario, we could just use r.Run() here. Since we
	// don't want this example to run forever, we'll use the more explicit
	// Start and Stop.
	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := r.Start(startCtx); err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println("Everything is running")

	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	r.Stop(stopCtx)

	// Output:
	// Initializing app "a", which requires 0 app
	// Initializing app "b", which requires app "a", whose value is "value_a"
	// Starting app "a"
	// Starting app "b"
	// Everything is running
	// Stopping app "b"
	// Stopping app "a"
	// Cleaning up app "b"
	// Cleaning up app "a"
}
