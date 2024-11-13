package all_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.yuchanns.xyz/all"
)

var ctx context.Context

func TestMain(m *testing.M) {
	ctx = context.Background()

	m.Run()
}

func consumer(ctx context.Context, in int) (out string, err error) {
	ticker := time.NewTicker(time.Second * time.Duration(in))
	select {
	case <-ticker.C:
		return fmt.Sprintf("%d", in), nil
	case <-ctx.Done():
		return "", context.Cause(ctx)
	}
}

func TestAllowError(t *testing.T) {
	t.Parallel()

	x := all.NewVoid(ctx, func(ctx context.Context, in int) error {
		time.Sleep(time.Second * time.Duration(in))
		return fmt.Errorf("error %d", in)
	})
	all.Persist(x)
	for i := 0; i < 3; i++ {
		x.Assign(i)
	}
	_, err := all.Collect(x, false)
	require.NotNil(t, err)
}

func TestCollectVoid(t *testing.T) {
	t.Parallel()

	x := all.NewVoid(ctx, func(ctx context.Context, in int) error {
		return nil
	})

	var total = 20
	for i := 0; i < total; i++ {
		num := total - i

		x.Assign(num)
	}

	_, err := all.Collect(x, false)
	require.Nil(t, err)
}

func TestCollectError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancelCause(ctx)
	var expectedErr = errors.New("an error")
	go func() {
		time.Sleep(time.Second * 10)
		cancel(expectedErr)
	}()

	x := all.New(ctx, consumer)

	var total = 20
	for i := 0; i < total; i++ {
		num := total - i

		x.Assign(num)
	}

	_, err := all.Collect(x, false)
	require.Equal(t, expectedErr, err)
}

func TestNextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancelCause(ctx)
	var expectedErr = errors.New("an error")
	go func() {
		time.Sleep(time.Second * 10)
		cancel(expectedErr)
	}()

	x := all.New(ctx, consumer)

	var total = 20
	for i := 0; i < total; i++ {
		num := total - i

		x.Assign(num)
	}

	for x.Next() {
	}

	require.Equal(t, expectedErr, x.Error())
}

func TestNext(t *testing.T) {
	t.Parallel()

	x := all.New(ctx, consumer)

	var total = 20
	var expected = make([]string, 0, total)
	for i := 0; i < total; i++ {
		num := total - i

		x.Assign(num)

		expected = append(expected, fmt.Sprintf("%d", num))
	}

	var results []string

	for x.Next() {
		results = append(results, x.Each())
	}

	require.Nil(t, x.Error())

	slices.Sort(expected)
	slices.Sort(results)

	require.Equal(t, expected, results)
}

func TestCollect(t *testing.T) {
	t.Parallel()

	x := all.New(ctx, consumer)

	var total = 20
	var expected = make([]string, 0, total)
	for i := 1; i < total; i++ {
		num := total - i

		x.Assign(num)

		expected = append(expected, fmt.Sprintf("%d", num))
	}

	results, err := all.Collect(x, false)

	require.Nil(t, err)

	slices.Sort(expected)
	slices.Sort(results)

	require.Equal(t, expected, results)
}