package synx

import (
	"context"
	"errors"
	"testing"
)

func TestErrGroup(t *testing.T) {
	errDoom := errors.New("group_test: doomed")

	cases := []struct {
		errs []error
		want error
	}{
		{errs: []error{nil}, want: nil},
		{errs: []error{errDoom}, want: errDoom},
		{errs: []error{errDoom, nil}, want: errDoom},
	}

	for _, tc := range cases {
		g := NewErrGroup(context.Background())

		for _, err := range tc.errs {
			err := err
			g.Go(func(_ context.Context) error { return err })
		}

		if err := g.Wait(); err != tc.want {
			t.Errorf("after %T.Go(func(_ context.Context) error { return err }) for err in %v\n"+
				"g.Wait() = %v; want %v",
				g, tc.errs, err, tc.want)
		}
	}
}

func TestErrGroup_panic(t *testing.T) {
	g := NewErrGroup(context.Background())
	g.Go(func(_ context.Context) error {
		panic(`test panic`)
	})
	err := g.Wait()
	if err == nil {
		t.FailNow()
	}
}
